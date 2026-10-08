//go:build windows

package cert

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	certKeyProvInfoPropID = 2 // CERT_KEY_PROV_INFO_PROP_ID

	// flags de PFXExportCertStoreEx
	reportNoPrivateKey              = 0x1
	reportNotAbleToExportPrivateKey = 0x2
	exportPrivateKeys               = 0x4
)

var (
	crypt32                  = windows.NewLazySystemDLL("crypt32.dll")
	procCertGetProperty      = crypt32.NewProc("CertGetCertificateContextProperty")
	procPFXExportCertStoreEx = crypt32.NewProc("PFXExportCertStoreEx")
)

// ErrChaveNaoExportavel indica que o Windows não permite exportar a chave privada.
var ErrChaveNaoExportavel = errors.New("a chave privada deste certificado não pode ser exportada: " +
	"certificados A3 (token/cartão) nunca permitem, e certificados A1 instalados sem a opção " +
	"\"marcar esta chave como exportável\" também não")

// ListStore lista os certificados pessoais ("Meu") do usuário atual,
// incluindo os A3 expostos pelo driver do token/cartão.
func ListStore() ([]Info, error) {
	name, _ := windows.UTF16PtrFromString("MY")
	store, err := windows.CertOpenSystemStore(0, name)
	if err != nil {
		return nil, fmt.Errorf("não foi possível abrir o repositório de certificados: %w", err)
	}
	defer windows.CertCloseStore(store, 0)

	now := time.Now()
	var out []Info
	var ctx *windows.CertContext
	for {
		ctx, err = windows.CertEnumCertificatesInStore(store, ctx)
		if err != nil || ctx == nil {
			break // fim da enumeração
		}
		raw := unsafe.Slice(ctx.EncodedCert, ctx.Length)
		c, perr := x509.ParseCertificate(append([]byte(nil), raw...))
		if perr != nil {
			continue
		}
		info := FromX509(c, now)
		info.PossuiChavePriv = hasPrivateKey(ctx)
		info.Origem = "Repositório do Windows"
		out = append(out, info)
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].ValidoAte > out[j].ValidoAte })
	return out, nil
}

func hasPrivateKey(ctx *windows.CertContext) bool {
	var size uint32
	r, _, _ := procCertGetProperty.Call(
		uintptr(unsafe.Pointer(ctx)),
		certKeyProvInfoPropID,
		0,
		uintptr(unsafe.Pointer(&size)),
	)
	return r != 0
}

// ExportPFX exporta um certificado instalado (identificado pelo thumbprint
// SHA-1) junto com a chave privada e a cadeia, protegido pela nova senha.
func ExportPFX(thumbprint, password string) ([]byte, error) {
	want := strings.ToUpper(strings.ReplaceAll(thumbprint, " ", ""))

	name, _ := windows.UTF16PtrFromString("MY")
	store, err := windows.CertOpenSystemStore(0, name)
	if err != nil {
		return nil, fmt.Errorf("não foi possível abrir o repositório de certificados: %w", err)
	}
	defer windows.CertCloseStore(store, 0)

	ctx, err := findByThumbprint(store, want)
	if err != nil {
		return nil, err
	}
	defer windows.CertFreeCertificateContext(ctx)

	if !hasPrivateKey(ctx) {
		return nil, errors.New("este certificado não possui chave privada instalada")
	}

	mem, err := windows.CertOpenStore(windows.CERT_STORE_PROV_MEMORY, 0, 0, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("falha ao preparar a exportação: %w", err)
	}
	defer windows.CertCloseStore(mem, 0)

	if err := windows.CertAddCertificateContextToStore(mem, ctx, windows.CERT_STORE_ADD_ALWAYS, nil); err != nil {
		return nil, fmt.Errorf("falha ao preparar a exportação: %w", err)
	}
	addChain(mem, ctx)

	pwd, err := windows.UTF16PtrFromString(password)
	if err != nil {
		return nil, errors.New("a senha contém caracteres inválidos")
	}
	flags := uintptr(exportPrivateKeys | reportNoPrivateKey | reportNotAbleToExportPrivateKey)

	var blob windows.CryptDataBlob
	if err := pfxExport(mem, &blob, pwd, flags); err != nil {
		return nil, err
	}
	if blob.Size == 0 {
		return nil, errors.New("o Windows não gerou o arquivo PFX")
	}
	buf := make([]byte, blob.Size)
	blob.Data = &buf[0]
	if err := pfxExport(mem, &blob, pwd, flags); err != nil {
		return nil, err
	}
	return buf[:blob.Size], nil
}

func pfxExport(store windows.Handle, blob *windows.CryptDataBlob, pwd *uint16, flags uintptr) error {
	r, _, callErr := procPFXExportCertStoreEx.Call(uintptr(store), uintptr(unsafe.Pointer(blob)),
		uintptr(unsafe.Pointer(pwd)), 0, flags)
	if r != 0 {
		return nil
	}
	var errno windows.Errno
	if errors.As(callErr, &errno) {
		switch uint32(errno) {
		case 0x8009000B, // NTE_BAD_KEY_STATE
			0x80090029, // NTE_NOT_SUPPORTED
			0x8009000D, // NTE_NO_KEY
			0x80092004: // CRYPT_E_NOT_FOUND
			return ErrChaveNaoExportavel
		case 0x800704C7: // ERROR_CANCELLED (PIN do token cancelado)
			return errors.New("exportação cancelada")
		}
	}
	return fmt.Errorf("o Windows recusou a exportação: %w", callErr)
}

func findByThumbprint(store windows.Handle, want string) (*windows.CertContext, error) {
	var ctx *windows.CertContext
	var err error
	for {
		ctx, err = windows.CertEnumCertificatesInStore(store, ctx)
		if err != nil || ctx == nil {
			return nil, errors.New("certificado não encontrado no repositório do Windows")
		}
		sum := sha1.Sum(unsafe.Slice(ctx.EncodedCert, ctx.Length))
		if strings.ToUpper(hex.EncodeToString(sum[:])) == want {
			// a enumeração libera o contexto anterior; duplicamos para mantê-lo vivo
			return windows.CertDuplicateCertificateContext(ctx), nil
		}
	}
}

// addChain inclui as ACs intermediárias e raiz no PFX, quando o Windows
// consegue montar a cadeia. Falhas aqui não impedem a exportação.
func addChain(mem windows.Handle, leaf *windows.CertContext) {
	para := windows.CertChainPara{}
	para.Size = uint32(unsafe.Sizeof(para))
	var chain *windows.CertChainContext
	if err := windows.CertGetCertificateChain(0, leaf, nil, leaf.Store, &para, 0, 0, &chain); err != nil || chain == nil {
		return
	}
	defer windows.CertFreeCertificateChain(chain)
	if chain.ChainCount == 0 {
		return
	}
	simple := unsafe.Slice(chain.Chains, chain.ChainCount)[0]
	elems := unsafe.Slice(simple.Elements, simple.NumElements)
	for _, e := range elems[1:] { // [0] é o próprio certificado
		_ = windows.CertAddCertificateContextToStore(mem, e.CertContext, windows.CERT_STORE_ADD_ALWAYS, nil)
	}
}

// DeleteFromStore remove um certificado do repositório pessoal do Windows.
// Assim como no certmgr, só o certificado é removido; para A3 o driver do
// token pode registrá-lo de novo quando o dispositivo for reconectado.
func DeleteFromStore(thumbprint string) error {
	want := strings.ToUpper(strings.ReplaceAll(thumbprint, " ", ""))

	name, _ := windows.UTF16PtrFromString("MY")
	store, err := windows.CertOpenSystemStore(0, name)
	if err != nil {
		return fmt.Errorf("não foi possível abrir o repositório de certificados: %w", err)
	}
	defer windows.CertCloseStore(store, 0)

	ctx, err := findByThumbprint(store, want)
	if err != nil {
		return err
	}
	// CertDeleteCertificateFromStore sempre libera o contexto, mesmo em erro.
	if err := windows.CertDeleteCertificateFromStore(ctx); err != nil {
		return fmt.Errorf("o Windows recusou a remoção: %w", err)
	}
	return nil
}

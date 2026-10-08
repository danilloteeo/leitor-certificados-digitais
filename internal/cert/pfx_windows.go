//go:build windows

package cert

import (
	"crypto/x509"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	certKeyContextPropID        = 5  // CERT_KEY_CONTEXT_PROP_ID
	certNCryptKeyHandlePropID   = 78 // CERT_NCRYPT_KEY_HANDLE_PROP_ID
	errorInvalidPasswordHResult = 0x80070056
)

// readPFXNative lê o PFX com o próprio Windows. Serve de alternativa para
// arquivos que o decodificador Go recusa, como os codificados em BER
// (comuns em PFX gerados por Java/BouncyCastle). A chave é importada só em
// memória e descartada ao fechar o store: nada é instalado.
func readPFXNative(data []byte, password string) (Info, error) {
	if len(data) == 0 {
		return Info{}, errors.New("arquivo vazio")
	}
	store, err := importPFX(data, password)
	if err != nil && password == "" && isInvalidPassword(err) {
		// "sem senha" pode ter sido gerado com senha vazia ou nula
		store, err = importPFXNilPassword(data)
	}
	if err != nil {
		if isInvalidPassword(err) {
			return Info{}, ErrSenhaIncorreta
		}
		return Info{}, err
	}
	defer windows.CertCloseStore(store, 0)

	type item struct {
		c      *x509.Certificate
		hasKey bool
	}
	var items []item
	var ctx *windows.CertContext
	for {
		ctx, err = windows.CertEnumCertificatesInStore(store, ctx)
		if err != nil || ctx == nil {
			break
		}
		raw := unsafe.Slice(ctx.EncodedCert, ctx.Length)
		c, perr := x509.ParseCertificate(append([]byte(nil), raw...))
		if perr != nil {
			continue
		}
		items = append(items, item{c, ctxHasKey(ctx)})
	}
	if len(items) == 0 {
		return Info{}, errors.New("nenhum certificado encontrado no arquivo")
	}

	leaf, hasKey := items[0].c, items[0].hasKey
	found := false
	for _, it := range items {
		if it.hasKey {
			leaf, hasKey, found = it.c, true, true
			break
		}
	}
	if !found {
		for _, it := range items {
			if !it.c.IsCA {
				leaf, hasKey = it.c, it.hasKey
				break
			}
		}
	}

	info := FromX509(leaf, time.Now())
	info.PossuiChavePriv = hasKey
	return info, nil
}

func importPFX(data []byte, password string) (windows.Handle, error) {
	pwd, err := windows.UTF16PtrFromString(password)
	if err != nil {
		return 0, errors.New("a senha contém caracteres inválidos")
	}
	return pfxImport(data, pwd)
}

func importPFXNilPassword(data []byte) (windows.Handle, error) {
	return pfxImport(data, nil)
}

func pfxImport(data []byte, pwd *uint16) (windows.Handle, error) {
	blob := windows.CryptDataBlob{Size: uint32(len(data)), Data: &data[0]}
	store, err := windows.PFXImportCertStore(&blob, pwd, windows.PKCS12_NO_PERSIST_KEY|windows.CRYPT_USER_KEYSET)
	if err != nil {
		if isInvalidPassword(err) {
			return 0, err
		}
		return 0, fmt.Errorf("o Windows também não conseguiu ler o arquivo: %w", err)
	}
	return store, nil
}

func isInvalidPassword(err error) bool {
	var errno windows.Errno
	if !errors.As(err, &errno) {
		return false
	}
	return errno == windows.ERROR_INVALID_PASSWORD || uint32(errno) == errorInvalidPasswordHResult
}

// ctxHasKey cobre chaves persistidas (CryptoAPI/CNG) e as importadas só em memória.
func ctxHasKey(ctx *windows.CertContext) bool {
	for _, prop := range []uintptr{certKeyProvInfoPropID, certKeyContextPropID, certNCryptKeyHandlePropID} {
		var size uint32
		r, _, _ := procCertGetProperty.Call(uintptr(unsafe.Pointer(ctx)), prop, 0, uintptr(unsafe.Pointer(&size)))
		if r != 0 {
			return true
		}
	}
	return false
}

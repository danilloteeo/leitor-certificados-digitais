//go:build windows

package cert

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestListStore(t *testing.T) {
	certs, err := ListStore()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d certificados no repositório", len(certs))
}

func TestExportPFXNaoEncontrado(t *testing.T) {
	if _, err := ExportPFX("0000000000000000000000000000000000000000", "x"); err == nil {
		t.Fatal("esperado erro para thumbprint inexistente")
	}
}

// Exporta (só em memória) o primeiro certificado com chave exportável e
// confere que o PFX gerado abre com a nova senha e traz a chave privada.
func TestExportPFXIdaEVolta(t *testing.T) {
	certs, err := ListStore()
	if err != nil {
		t.Fatal(err)
	}
	naoExportaveis := 0
	for _, c := range certs {
		if !c.PossuiChavePriv {
			continue
		}
		data, err := ExportPFX(c.ThumbprintSHA1, "Nova@Senha123")
		if errors.Is(err, ErrChaveNaoExportavel) {
			naoExportaveis++
			continue
		}
		if err != nil {
			t.Fatalf("exportação falhou: %v", err)
		}
		info, err := ReadPFX(data, "Nova@Senha123")
		if err != nil {
			t.Fatalf("PFX exportado não abre: %v", err)
		}
		if info.ThumbprintSHA1 != c.ThumbprintSHA1 || !info.PossuiChavePriv {
			t.Fatalf("PFX exportado não corresponde ao certificado original")
		}
		if _, err := ReadPFX(data, "outra"); !errors.Is(err, ErrSenhaIncorreta) {
			t.Fatalf("senha errada deveria falhar, veio %v", err)
		}
		t.Logf("exportação ok (%d bytes); %d não exportáveis antes dele", len(data), naoExportaveis)
		return
	}
	t.Skipf("nenhum certificado com chave exportável (%d não exportáveis)", naoExportaveis)
}

func TestDeleteFromStoreNaoEncontrado(t *testing.T) {
	if err := DeleteFromStore("0000000000000000000000000000000000000000"); err == nil {
		t.Fatal("esperado erro para thumbprint inexistente")
	}
}

// Instala um certificado descartável no repositório, remove e confere que sumiu.
func TestDeleteFromStoreRemove(t *testing.T) {
	c, _ := newCert(t, "TESTE DESCARTAVEL LEITOR CERTIFICADOS", nil, nil, time.Now().AddDate(0, 0, 1))
	thumb := FromX509(c, time.Now()).ThumbprintSHA1

	name, _ := windows.UTF16PtrFromString("MY")
	store, err := windows.CertOpenSystemStore(0, name)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING|windows.PKCS_7_ASN_ENCODING, &c.Raw[0], uint32(len(c.Raw)))
	if err != nil {
		t.Fatal(err)
	}
	err = windows.CertAddCertificateContextToStore(store, ctx, windows.CERT_STORE_ADD_REPLACE_EXISTING, nil)
	windows.CertFreeCertificateContext(ctx)
	windows.CertCloseStore(store, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := DeleteFromStore(thumb); err != nil {
		t.Fatal(err)
	}
	certs, err := ListStore()
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range certs {
		if x.ThumbprintSHA1 == thumb {
			t.Fatal("certificado continua no repositório após a remoção")
		}
	}
}

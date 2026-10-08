package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"math/big"
	"testing"
	"time"

	"software.sslmate.com/src/go-pkcs12"
)

type otherName struct {
	ID    asn1.ObjectIdentifier
	Value asn1.RawValue // [0] EXPLICIT, montado à mão
}

func sanOtherNames(t *testing.T, fields map[string]string, order []asn1.ObjectIdentifier) []byte {
	t.Helper()
	var names []asn1.RawValue
	for _, oid := range order {
		octets, err := asn1.Marshal([]byte(fields[oid.String()]))
		if err != nil {
			t.Fatal(err)
		}
		inner, err := asn1.Marshal(otherName{
			ID:    oid,
			Value: asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: octets},
		})
		if err != nil {
			t.Fatal(err)
		}
		// troca o SEQUENCE universal por [0] IMPLICIT
		var seq asn1.RawValue
		if _, err := asn1.Unmarshal(inner, &seq); err != nil {
			t.Fatal(err)
		}
		names = append(names, asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: seq.Bytes})
	}
	names = append(names, asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 1, Bytes: []byte("contato@empresa.com.br")})
	b, err := asn1.Marshal(names)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func newCert(t *testing.T, cn string, san []byte, policy []uint64, notAfter time.Time) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ca := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "AC TESTE"}}
	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(0x1A2B3C4D),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		NotAfter:     notAfter,
	}
	if policy != nil {
		oid, err := x509.OIDFromInts(policy)
		if err != nil {
			t.Fatal(err)
		}
		tpl.Policies = []x509.OID{oid}
	}
	if san != nil {
		tpl.ExtraExtensions = []pkix.Extension{{Id: oidSubjectAltName, Value: san}}
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, ca, &key.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	c, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return c, key
}

var now = time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

func TestECNPJ(t *testing.T) {
	fields := map[string]string{
		oidCNPJ.String():          "12345678000195",
		oidResponsavelPJ.String(): "FULANO DE TAL",
		oidDadosRespPJ.String():   "1503198012345678901" + "00000000000" + "000000000000000" + "000000",
	}
	san := sanOtherNames(t, fields, []asn1.ObjectIdentifier{oidDadosRespPJ, oidResponsavelPJ, oidCNPJ})
	c, _ := newCert(t, "EMPRESA EXEMPLO LTDA:12345678000195", san,
		[]uint64{2, 16, 76, 1, 2, 1, 51}, time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC))

	got := FromX509(c, now)

	checks := map[string][2]string{
		"Titular":        {got.Titular, "EMPRESA EXEMPLO LTDA"},
		"Tipo":           {got.Tipo, "e-CNPJ"},
		"Nivel":          {got.Nivel, "A1"},
		"CNPJ":           {got.CNPJ, "12.345.678/0001-95"},
		"Responsavel":    {got.Responsavel, "FULANO DE TAL"},
		"ResponsavelCPF": {got.ResponsavelCPF, "123.456.789-01"},
		"DataNascimento": {got.DataNascimento, "15/03/1980"},
		"NumeroSerie":    {got.NumeroSerie, "1A2B3C4D"},
		"Email":          {got.Email, "contato@empresa.com.br"},
		"Status":         {got.Status, StatusValido},
		"Emissor":        {got.Emissor, "AC TESTE"},
	}
	for name, v := range checks {
		if v[0] != v[1] {
			t.Errorf("%s = %q, esperado %q", name, v[0], v[1])
		}
	}
	if got.CPF != "" {
		t.Errorf("CPF deveria estar vazio em e-CNPJ, veio %q", got.CPF)
	}
}

func TestECPFPeloCommonName(t *testing.T) {
	c, _ := newCert(t, "MARIA DA SILVA:98765432100", nil,
		[]uint64{2, 16, 76, 1, 2, 3, 10}, time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC))
	got := FromX509(c, now)
	if got.CPF != "987.654.321-00" || got.Tipo != "e-CPF" || got.Titular != "MARIA DA SILVA" {
		t.Errorf("dados incorretos: %+v", got)
	}
	if got.Nivel != "A3" {
		t.Errorf("Nivel = %q, esperado A3", got.Nivel)
	}
	if got.Status != StatusVencendo || got.DiasRestantes != 11 {
		t.Errorf("status = %s / %d dias, esperado vencendo / 11", got.Status, got.DiasRestantes)
	}
}

func TestVencido(t *testing.T) {
	c, _ := newCert(t, "QUALQUER", nil, nil, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	got := FromX509(c, now)
	if got.Status != StatusVencido || got.DiasRestantes >= 0 || got.Tipo != "Outro" {
		t.Errorf("esperado vencido/Outro, veio %s %d %s", got.Status, got.DiasRestantes, got.Tipo)
	}
}

func TestReadPFX(t *testing.T) {
	fields := map[string]string{oidDadosPF.String(): "0101199011122233344" + "00000000000"}
	san := sanOtherNames(t, fields, []asn1.ObjectIdentifier{oidDadosPF})
	c, key := newCert(t, "JOAO PEREIRA:11122233344", san, nil, time.Now().AddDate(1, 0, 0))

	data, err := pkcs12.Modern.Encode(key, c, nil, "segredo")
	if err != nil {
		t.Fatal(err)
	}

	info, err := ReadPFX(data, "segredo")
	if err != nil {
		t.Fatal(err)
	}
	if info.CPF != "111.222.333-44" || info.DataNascimento != "01/01/1990" || !info.PossuiChavePriv {
		t.Errorf("dados incorretos: %+v", info)
	}

	if _, err := ReadPFX(data, "errada"); !errors.Is(err, ErrSenhaIncorreta) {
		t.Errorf("esperado ErrSenhaIncorreta, veio %v", err)
	}
}

func TestReadPFXEscolheTitularPelaChave(t *testing.T) {
	c, key := newCert(t, "TITULAR:11122233344", nil, nil, time.Now().AddDate(1, 0, 0))
	ac, _ := newCert(t, "AC INTERMEDIARIA", nil, nil, time.Now().AddDate(5, 0, 0))

	// codifica com a AC na posição de "leaf" para simular a ordem do Windows
	data, err := pkcs12.Modern.Encode(key, ac, []*x509.Certificate{c}, "s")
	if err == nil {
		info, err := ReadPFX(data, "s")
		if err != nil {
			t.Fatal(err)
		}
		if info.Titular != "TITULAR" {
			t.Errorf("titular = %q, esperado TITULAR", info.Titular)
		}
		return
	}
	t.Skipf("go-pkcs12 recusou montar o PFX fora de ordem: %v", err)
}

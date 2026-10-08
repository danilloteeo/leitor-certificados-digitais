// Package cert extrai as informações de certificados digitais X.509,
// com suporte aos campos específicos da ICP-Brasil (CPF, CNPJ, etc.).
package cert

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"math"
	"strings"
	"time"
)

// Status de validade do certificado.
const (
	StatusValido   = "valido"
	StatusVencendo = "vencendo" // vence em até DiasAlerta dias
	StatusVencido  = "vencido"
	StatusPendente = "pendente" // ainda não entrou em vigor
	DiasAlerta     = 30
)

// Info reúne os dados de um certificado prontos para exibição.
type Info struct {
	Titular          string `json:"titular"`
	CommonName       string `json:"commonName"`
	Tipo             string `json:"tipo"`           // e-CPF, e-CNPJ ou Outro
	Nivel            string `json:"nivel"`          // A1, A3... (vazio se desconhecido)
	CPF              string `json:"cpf"`            // formatado
	CNPJ             string `json:"cnpj"`           // formatado
	Responsavel      string `json:"responsavel"`    // responsável pelo e-CNPJ
	ResponsavelCPF   string `json:"responsavelCpf"` // formatado
	DataNascimento   string `json:"dataNascimento"` // dd/mm/aaaa
	Email            string `json:"email"`
	NumeroSerie      string `json:"numeroSerie"` // hexadecimal
	NumeroSerieDec   string `json:"numeroSerieDec"`
	Emissor          string `json:"emissor"`
	EmissorCompleto  string `json:"emissorCompleto"`
	SubjectCompleto  string `json:"subjectCompleto"`
	ValidoDe         string `json:"validoDe"`  // RFC 3339
	ValidoAte        string `json:"validoAte"` // RFC 3339
	DiasRestantes    int    `json:"diasRestantes"`
	Status           string `json:"status"`
	ThumbprintSHA1   string `json:"thumbprintSha1"`
	ThumbprintSHA256 string `json:"thumbprintSha256"`
	AlgoritmoAssinat string `json:"algoritmoAssinatura"`
	AlgoritmoChave   string `json:"algoritmoChave"`
	PossuiChavePriv  bool   `json:"possuiChavePrivada"`
	Origem           string `json:"origem"` // caminho do arquivo ou "Repositório do Windows"
}

// FromX509 converte um certificado em Info, usando now como referência
// para calcular status e dias restantes.
func FromX509(c *x509.Certificate, now time.Time) Info {
	info := Info{
		CommonName:       c.Subject.CommonName,
		NumeroSerie:      formatSerial(c),
		NumeroSerieDec:   c.SerialNumber.String(),
		Emissor:          c.Issuer.CommonName,
		EmissorCompleto:  c.Issuer.String(),
		SubjectCompleto:  c.Subject.String(),
		ValidoDe:         c.NotBefore.Format(time.RFC3339),
		ValidoAte:        c.NotAfter.Format(time.RFC3339),
		AlgoritmoAssinat: c.SignatureAlgorithm.String(),
		AlgoritmoChave:   c.PublicKeyAlgorithm.String(),
		Nivel:            nivelICP(c.PolicyIdentifiers),
	}
	if len(c.EmailAddresses) > 0 {
		info.Email = c.EmailAddresses[0]
	}

	sum1 := sha1.Sum(c.Raw)
	sum256 := sha256.Sum256(c.Raw)
	info.ThumbprintSHA1 = strings.ToUpper(hex.EncodeToString(sum1[:]))
	info.ThumbprintSHA256 = strings.ToUpper(hex.EncodeToString(sum256[:]))

	info.DiasRestantes = diasRestantes(c.NotAfter, now)
	info.Status = status(c, now)

	applyICPBrasil(&info, c)
	return info
}

func formatSerial(c *x509.Certificate) string {
	b := c.SerialNumber.Bytes()
	if len(b) == 0 {
		return "00"
	}
	return strings.ToUpper(hex.EncodeToString(b))
}

func diasRestantes(notAfter, now time.Time) int {
	d := notAfter.Sub(now).Hours() / 24
	if d < 0 {
		return int(math.Floor(d))
	}
	return int(d)
}

func status(c *x509.Certificate, now time.Time) string {
	switch {
	case now.Before(c.NotBefore):
		return StatusPendente
	case now.After(c.NotAfter):
		return StatusVencido
	case c.NotAfter.Sub(now) <= DiasAlerta*24*time.Hour:
		return StatusVencendo
	default:
		return StatusValido
	}
}

package cert

import (
	"crypto/x509"
	"encoding/asn1"
	"strings"
)

// OIDs dos campos "otherName" definidos pela ICP-Brasil (DOC-ICP-04).
var (
	oidSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
	oidDadosPF        = asn1.ObjectIdentifier{2, 16, 76, 1, 3, 1} // nascimento + CPF + NIS + RG + órgão
	oidResponsavelPJ  = asn1.ObjectIdentifier{2, 16, 76, 1, 3, 2} // nome do responsável
	oidCNPJ           = asn1.ObjectIdentifier{2, 16, 76, 1, 3, 3}
	oidDadosRespPJ    = asn1.ObjectIdentifier{2, 16, 76, 1, 3, 4} // nascimento + CPF do responsável ...
	oidPoliticaICP    = asn1.ObjectIdentifier{2, 16, 76, 1, 2}    // 2.16.76.1.2.<tipo>.<n>
)

// otherNames lê os campos otherName da extensão Subject Alternative Name,
// que o pacote x509 da biblioteca padrão não expõe.
func otherNames(c *x509.Certificate) map[string]string {
	out := map[string]string{}
	for _, ext := range c.Extensions {
		if !ext.Id.Equal(oidSubjectAltName) {
			continue
		}
		var seq asn1.RawValue
		if _, err := asn1.Unmarshal(ext.Value, &seq); err != nil {
			return out
		}
		rest := seq.Bytes
		for len(rest) > 0 {
			var gn asn1.RawValue
			var err error
			if rest, err = asn1.Unmarshal(rest, &gn); err != nil {
				break
			}
			// otherName ::= [0] IMPLICIT SEQUENCE { type-id OID, value [0] EXPLICIT ANY }
			if gn.Class != asn1.ClassContextSpecific || gn.Tag != 0 {
				continue
			}
			var oid asn1.ObjectIdentifier
			r, err := asn1.Unmarshal(gn.Bytes, &oid)
			if err != nil {
				continue
			}
			var wrapper asn1.RawValue
			if _, err := asn1.Unmarshal(r, &wrapper); err != nil {
				continue
			}
			var inner asn1.RawValue
			if _, err := asn1.Unmarshal(wrapper.Bytes, &inner); err != nil {
				continue
			}
			out[oid.String()] = string(inner.Bytes)
		}
	}
	return out
}

func applyICPBrasil(info *Info, c *x509.Certificate) {
	names := otherNames(c)

	if v, ok := names[oidCNPJ.String()]; ok {
		if cnpj := onlyDigits(v); len(cnpj) == 14 && !allZeros(cnpj) {
			info.CNPJ = FormatCNPJ(cnpj)
		}
	}
	if v, ok := names[oidDadosPF.String()]; ok {
		nasc, cpf := parseDadosPessoa(v)
		info.DataNascimento = nasc
		if cpf != "" {
			info.CPF = FormatCPF(cpf)
		}
	}
	if v, ok := names[oidDadosRespPJ.String()]; ok {
		nasc, cpf := parseDadosPessoa(v)
		if info.DataNascimento == "" {
			info.DataNascimento = nasc
		}
		if cpf != "" {
			info.ResponsavelCPF = FormatCPF(cpf)
		}
	}
	if v, ok := names[oidResponsavelPJ.String()]; ok {
		info.Responsavel = strings.TrimSpace(v)
	}

	// Muitos certificados trazem o documento no CN: "NOME:12345678901".
	nome, doc := splitCommonName(c.Subject.CommonName)
	info.Titular = nome
	switch {
	case len(doc) == 14 && info.CNPJ == "":
		info.CNPJ = FormatCNPJ(doc)
	case len(doc) == 11 && info.CPF == "":
		info.CPF = FormatCPF(doc)
	}

	switch {
	case info.CNPJ != "":
		info.Tipo = "e-CNPJ"
	case info.CPF != "":
		info.Tipo = "e-CPF"
	default:
		info.Tipo = "Outro"
	}
}

// parseDadosPessoa interpreta o layout ICP-Brasil: 8 dígitos de data de
// nascimento (ddmmaaaa) seguidos de 11 dígitos de CPF.
func parseDadosPessoa(v string) (nascimento, cpf string) {
	if len(v) < 19 {
		return "", ""
	}
	d, c := v[:8], v[8:19]
	if onlyDigits(d) == d && !allZeros(d) {
		nascimento = d[0:2] + "/" + d[2:4] + "/" + d[4:8]
	}
	if onlyDigits(c) == c && !allZeros(c) {
		cpf = c
	}
	return nascimento, cpf
}

func splitCommonName(cn string) (nome, doc string) {
	i := strings.LastIndex(cn, ":")
	if i < 0 {
		return strings.TrimSpace(cn), ""
	}
	d := strings.TrimSpace(cn[i+1:])
	if onlyDigits(d) != d || (len(d) != 11 && len(d) != 14) {
		return strings.TrimSpace(cn), ""
	}
	return strings.TrimSpace(cn[:i]), d
}

// nivelICP deduz o tipo (A1, A3...) pela política de certificação ICP-Brasil.
func nivelICP(policies []asn1.ObjectIdentifier) string {
	for _, p := range policies {
		if len(p) < 6 || !asn1.ObjectIdentifier(p[:5]).Equal(oidPoliticaICP) {
			continue
		}
		switch p[5] {
		case 1:
			return "A1"
		case 2:
			return "A2"
		case 3:
			return "A3"
		case 4:
			return "A4"
		case 101, 102, 103, 104:
			return "S" + string(rune('0'+p[5]-100))
		}
	}
	return ""
}

// FormatCPF formata 11 dígitos como 000.000.000-00.
func FormatCPF(d string) string {
	if len(d) != 11 {
		return d
	}
	return d[0:3] + "." + d[3:6] + "." + d[6:9] + "-" + d[9:11]
}

// FormatCNPJ formata 14 dígitos como 00.000.000/0000-00.
func FormatCNPJ(d string) string {
	if len(d) != 14 {
		return d
	}
	return d[0:2] + "." + d[2:5] + "." + d[5:8] + "/" + d[8:12] + "-" + d[12:14]
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func allZeros(s string) bool {
	return strings.Trim(s, "0") == ""
}

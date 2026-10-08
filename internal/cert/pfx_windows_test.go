//go:build windows

package cert

import (
	"errors"
	"strings"
	"testing"
	"time"

	"software.sslmate.com/src/go-pkcs12"
)

// toBER imita PFX gerados por Java/BouncyCastle: a estrutura externa (PFX,
// ContentInfo e [0]) usa comprimento indefinido (0x80 ... 00 00) e o resto
// continua DER. O conteúdo das OCTET STRINGs não muda, então o MAC segue válido.
func toBER(der []byte, depth, maxDepth int) []byte {
	var out []byte
	for len(der) > 0 {
		tag := der[0]
		hdr, n := 2, int(der[1])
		if der[1]&0x80 != 0 {
			k := int(der[1] & 0x7f)
			n = 0
			for _, b := range der[2 : 2+k] {
				n = n<<8 | int(b)
			}
			hdr = 2 + k
		}
		body := der[hdr : hdr+n]
		switch {
		case tag&0x20 == 0: // primitivo
			out = append(out, der[:hdr+n]...)
		case depth <= maxDepth:
			out = append(out, tag, 0x80)
			out = append(out, toBER(body, depth+1, maxDepth)...)
			out = append(out, 0, 0)
		default:
			out = append(out, der[:hdr+n]...)
		}
		der = der[hdr+n:]
	}
	return out
}

func TestReadPFXEmBER(t *testing.T) {
	c, key := newCert(t, "EMPRESA BER LTDA:12345678000195", nil, nil, time.Now().AddDate(1, 0, 0))

	for name, enc := range map[string]*pkcs12.Encoder{"3DES": pkcs12.LegacyDES, "AES": pkcs12.Modern} {
		t.Run(name, func(t *testing.T) {
			der, err := enc.Encode(key, c, nil, "segredo")
			if err != nil {
				t.Fatal(err)
			}
			ber := toBER(der, 0, 2)

			// garante que reproduzimos o erro relatado
			_, _, _, err = pkcs12.DecodeChain(ber, "segredo")
			if err == nil || !strings.Contains(err.Error(), "error reading P12 data: asn1: syntax error: indefinite length found") {
				t.Fatalf("esperado o erro de comprimento indefinido, veio %v", err)
			}

			info, err := ReadPFX(ber, "segredo")
			if err != nil {
				t.Fatalf("PFX em BER não foi lido: %v", err)
			}
			if info.CNPJ != "12.345.678/0001-95" || !info.PossuiChavePriv {
				t.Errorf("dados incorretos: cnpj=%q chave=%v", info.CNPJ, info.PossuiChavePriv)
			}

			if _, err := ReadPFX(ber, "errada"); !errors.Is(err, ErrSenhaIncorreta) {
				t.Errorf("esperado ErrSenhaIncorreta, veio %v", err)
			}
		})
	}
}

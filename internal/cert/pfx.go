package cert

import (
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"time"

	"software.sslmate.com/src/go-pkcs12"
)

// ErrSenhaIncorreta indica que a senha do arquivo PFX não confere.
var ErrSenhaIncorreta = errors.New("senha incorreta")

// ReadPFXFile lê um certificado A1 (.pfx/.p12) protegido por senha.
func ReadPFXFile(path, password string) (Info, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Info{}, fmt.Errorf("não foi possível abrir o arquivo: %w", err)
	}
	info, err := ReadPFX(data, password)
	if err != nil {
		return Info{}, err
	}
	info.Origem = path
	return info, nil
}

// ReadPFX decodifica o conteúdo de um PFX e devolve os dados do certificado do titular.
func ReadPFX(data []byte, password string) (Info, error) {
	key, first, others, err := pkcs12.DecodeChain(data, password)
	if err != nil {
		if errors.Is(err, pkcs12.ErrIncorrectPassword) {
			return Info{}, ErrSenhaIncorreta
		}
		// Formatos que o decodificador Go não aceita (ex.: BER) ficam com o Windows.
		info, nerr := readPFXNative(data, password)
		if nerr == nil || errors.Is(nerr, ErrSenhaIncorreta) {
			return info, nerr
		}
		return Info{}, fmt.Errorf("arquivo PFX inválido: %w", err)
	}
	leaf := titular(key, append([]*x509.Certificate{first}, others...))
	info := FromX509(leaf, time.Now())
	info.PossuiChavePriv = key != nil
	return info, nil
}

// titular escolhe o certificado do dono da chave: a ordem dentro do PFX não é
// garantida (o Windows, por exemplo, pode colocar as ACs antes).
func titular(key any, certs []*x509.Certificate) *x509.Certificate {
	if signer, ok := key.(crypto.Signer); ok {
		pub, ok := signer.Public().(interface{ Equal(crypto.PublicKey) bool })
		if ok {
			for _, c := range certs {
				if pub.Equal(c.PublicKey) {
					return c
				}
			}
		}
	}
	for _, c := range certs {
		if !c.IsCA {
			return c
		}
	}
	return certs[0]
}

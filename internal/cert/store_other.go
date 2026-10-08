//go:build !windows

package cert

import "errors"

// ListStore só é suportado no Windows.
func ListStore() ([]Info, error) {
	return nil, errors.New("repositório de certificados disponível apenas no Windows")
}

// ExportPFX só é suportado no Windows.
func ExportPFX(thumbprint, password string) ([]byte, error) {
	return nil, errors.New("exportação disponível apenas no Windows")
}

// DeleteFromStore só é suportado no Windows.
func DeleteFromStore(thumbprint string) error {
	return errors.New("remoção disponível apenas no Windows")
}

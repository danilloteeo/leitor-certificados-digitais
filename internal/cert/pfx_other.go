//go:build !windows

package cert

import "errors"

func readPFXNative(data []byte, password string) (Info, error) {
	return Info{}, errors.New("leitor alternativo disponível apenas no Windows")
}

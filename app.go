package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"certificado/internal/cert"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App expõe ao frontend a leitura de certificados digitais.
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectPFXFile abre o seletor nativo e devolve o caminho escolhido ("" se cancelado).
func (a *App) SelectPFXFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar certificado A1",
		Filters: []runtime.FileFilter{
			{DisplayName: "Certificado A1 (*.pfx, *.p12)", Pattern: "*.pfx;*.p12"},
		},
	})
}

// ReadPFX lê um arquivo .pfx/.p12 com a senha informada.
func (a *App) ReadPFX(path, password string) (cert.Info, error) {
	if strings.TrimSpace(path) == "" {
		return cert.Info{}, errors.New("selecione um arquivo .pfx ou .p12")
	}
	return cert.ReadPFXFile(path, password)
}

// ListStoreCertificates lista os certificados pessoais instalados no Windows.
func (a *App) ListStoreCertificates() ([]cert.Info, error) {
	return cert.ListStore()
}

// ExportCertificate exporta um certificado instalado para um novo .pfx com a
// senha informada. Devolve o caminho salvo, ou "" se o usuário cancelar.
func (a *App) ExportCertificate(thumbprint, password string) (string, error) {
	if len(password) < 4 {
		return "", errors.New("a nova senha deve ter pelo menos 4 caracteres")
	}

	data, err := cert.ExportPFX(thumbprint, password)
	if err != nil {
		return "", err
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Salvar certificado exportado",
		DefaultFilename: a.suggestedFileName(thumbprint),
		Filters: []runtime.FileFilter{
			{DisplayName: "Certificado A1 (*.pfx)", Pattern: "*.pfx"},
		},
	})
	if err != nil || path == "" {
		return "", err
	}
	if !strings.EqualFold(filepath.Ext(path), ".pfx") && !strings.EqualFold(filepath.Ext(path), ".p12") {
		path += ".pfx"
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("não foi possível salvar o arquivo: %w", err)
	}
	return path, nil
}

var invalidFileChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]+`)

func (a *App) suggestedFileName(thumbprint string) string {
	certs, _ := cert.ListStore()
	for _, c := range certs {
		if strings.EqualFold(c.ThumbprintSHA1, thumbprint) {
			name := c.Titular
			if doc := c.CNPJ + c.CPF; doc != "" {
				name += " - " + doc
			}
			name = strings.TrimSpace(invalidFileChars.ReplaceAllString(name, ""))
			if name != "" {
				return name + ".pfx"
			}
		}
	}
	return "certificado.pfx"
}

// DeleteCertificates remove do Windows os certificados escolhidos. Devolve
// quantos foram removidos; erros individuais são reunidos numa só mensagem.
func (a *App) DeleteCertificates(thumbprints []string) (int, error) {
	if len(thumbprints) == 0 {
		return 0, errors.New("nenhum certificado selecionado")
	}
	removed := 0
	var errs []error
	for _, t := range thumbprints {
		if err := cert.DeleteFromStore(t); err != nil {
			errs = append(errs, err)
			continue
		}
		removed++
	}
	if len(errs) > 0 {
		return removed, fmt.Errorf("%d de %d não puderam ser removidos: %w", len(errs), len(thumbprints), errors.Join(errs...))
	}
	return removed, nil
}

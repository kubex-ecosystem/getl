package utils

import (
	"os"
	"path/filepath"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
)

// GetWorkDir obtém o diretório de trabalho.
// Retorna o caminho do diretório de trabalho e um erro, se houver.
func GetWorkDir() (string, error) {
	var homeDir string
	var err error
	if homeDir, err = os.UserHomeDir(); homeDir == "" || err != nil {
		if homeDir, err = os.Getwd(); homeDir == "" || err != nil {
			homeDir = kbxGet.EnvOr("HOME", os.Getenv("GETL_CWD"))
		}
	}
	homeDir = filepath.Join(homeDir, ".kubex")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return "", err
	}
	return homeDir, nil
}

func GetKubexDir() (string, error) { return GetWorkDir() }

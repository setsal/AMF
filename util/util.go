package util

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetDefaultLogDir :
func GetDefaultLogDir() string {
	// modify to current dir for logging
	defaultLogDir := os.Getenv("PWD")
	if runtime.GOOS == "windows" {
		defaultLogDir = filepath.Join(os.Getenv("APPDATA"), "go-AMF")
	}
	return defaultLogDir
}

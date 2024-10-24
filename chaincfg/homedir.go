package chaincfg

import (
	stdlog "log"
	"os"
	"path/filepath"
)

const (
	HomeDirName = ".surgechain"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		stdlog.Printf("Failed to get home dir %v", err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, HomeDirName)
}

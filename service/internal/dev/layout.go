package dev

import (
	"os"
	"path/filepath"
)

const (
	defaultDataDirName      = ".opentdf/dev"
	defaultPlatformConfig   = "opentdf-dev.yaml"
	defaultIDPConfig        = "dev-idp.yaml"
	defaultRootKeyFile      = "root-key.txt"
	defaultKASRSAKey        = "kas-rsa-private.pem"
	defaultKASRSAPublicKey  = "kas-rsa-public.pem"
	defaultKASECKey         = "kas-ec-private.pem"
	defaultIDPPrivateKey    = "dev-idp-private.pem"
	defaultBinaryName       = "opentdf-dev"
	defaultPIDPlatformFile  = "platform.pid"
	defaultPIDIDPFile       = "dev-idp.pid"
	defaultDevKeysDirectory = "keys"
	defaultBinDirectory     = "bin"
	defaultPIDDirectory     = "pids"
)

type Layout struct {
	DataDir            string
	KeysDir            string
	BinDir             string
	PidsDir            string
	PlatformConfigPath string
	IDPConfigPath      string
	RootKeyPath        string
	KASRSAPrivatePath  string
	KASRSAPublicPath   string
	KASECPrivatePath   string
	IDPPrivateKeyPath  string
	PlatformBinaryPath string
	PlatformPIDPath    string
	IDPPIDPath         string
}

func DefaultDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, defaultDataDirName), nil
}

func EnsureLayout(dataDir string) (Layout, error) {
	absDir, err := filepath.Abs(dataDir)
	if err != nil {
		return Layout{}, err
	}

	keysDir := filepath.Join(absDir, defaultDevKeysDirectory)
	binDir := filepath.Join(absDir, defaultBinDirectory)
	pidsDir := filepath.Join(absDir, defaultPIDDirectory)

	for _, dir := range []string{absDir, keysDir, binDir, pidsDir} {
		if err := os.MkdirAll(dir, defaultDirPerm); err != nil {
			return Layout{}, err
		}
	}

	return Layout{
		DataDir:            absDir,
		KeysDir:            keysDir,
		BinDir:             binDir,
		PidsDir:            pidsDir,
		PlatformConfigPath: filepath.Join(absDir, defaultPlatformConfig),
		IDPConfigPath:      filepath.Join(absDir, defaultIDPConfig),
		RootKeyPath:        filepath.Join(absDir, defaultRootKeyFile),
		KASRSAPrivatePath:  filepath.Join(keysDir, defaultKASRSAKey),
		KASRSAPublicPath:   filepath.Join(keysDir, defaultKASRSAPublicKey),
		KASECPrivatePath:   filepath.Join(keysDir, defaultKASECKey),
		IDPPrivateKeyPath:  filepath.Join(keysDir, defaultIDPPrivateKey),
		PlatformBinaryPath: filepath.Join(binDir, defaultBinaryName),
		PlatformPIDPath:    filepath.Join(pidsDir, defaultPIDPlatformFile),
		IDPPIDPath:         filepath.Join(pidsDir, defaultPIDIDPFile),
	}, nil
}

package dev

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func EnsureRSAKeyPair(privatePath, publicPath string, regenerate bool) error {
	if !regenerate {
		if fileExists(privatePath) && (publicPath == "" || fileExists(publicPath)) {
			return nil
		}
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, defaultRSAKeySize)
	if err != nil {
		return fmt.Errorf("generate rsa key: %w", err)
	}

	privateBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("marshal rsa private key: %w", err)
	}

	if err := writePEM(privatePath, "PRIVATE KEY", privateBytes, defaultFilePerm); err != nil {
		return err
	}

	if publicPath == "" {
		return nil
	}

	publicBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshal rsa public key: %w", err)
	}

	return writePEM(publicPath, "PUBLIC KEY", publicBytes, defaultPublicFilePerm)
}

func EnsureECPrivateKey(privatePath string, regenerate bool) error {
	if !regenerate && fileExists(privatePath) {
		return nil
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate ec key: %w", err)
	}

	privateBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("marshal ec private key: %w", err)
	}

	return writePEM(privatePath, "PRIVATE KEY", privateBytes, defaultFilePerm)
}

func writePEM(path, blockType string, bytes []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), defaultDirPerm); err != nil {
		return err
	}
	block := &pem.Block{
		Type:  blockType,
		Bytes: bytes,
	}
	return os.WriteFile(path, pem.EncodeToMemory(block), perm)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func (m *Middleware) WithDecryptRSA() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Header.Get("Content-Encoding") != "encrypted" {
			ctx.Next()
			return
		}

		encryptedData, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		decryptedData, err := decryptOAEPChunks(encryptedData, m.privateKey)
		if err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		ctx.Request.Body = io.NopCloser(bytes.NewReader(decryptedData))
		ctx.Request.Header.Set("Content-Encoding", "gzip") // передаём дальше как gzip
		ctx.Next()
	}
}

func (m *Middleware) loadPrivateKey() error {
	keyData, err := os.ReadFile(m.cryptoKeyPath)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		return errors.New("failed to decode PEM block")
	}

	privKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	privKey, ok := privKeyInterface.(*rsa.PrivateKey)
	if !ok {
		return errors.New("not RSA private key")
	}
	m.privateKey = privKey
	return nil
}

func decryptOAEPChunks(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	hash := sha256.New()
	keySize := privateKey.PublicKey.Size()

	var decrypted []byte
	for start := 0; start < len(data); start += keySize {
		end := start + keySize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[start:end]
		decChunk, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("decrypt chunk error: %w", err)
		}

		decrypted = append(decrypted, decChunk...)
	}

	return decrypted, nil
}

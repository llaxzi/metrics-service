package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecryptOAEPChunks(t *testing.T) {
	keyPath := "test_private.pem"
	err := generateTestPrivateKeyPKCS8(keyPath)
	assert.NoError(t, err)
	defer os.Remove(keyPath)

	keyData, err := os.ReadFile(keyPath)
	assert.NoError(t, err)

	block, _ := pem.Decode(keyData)
	assert.NotNil(t, block)

	keyParsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	assert.NoError(t, err)

	privKey, ok := keyParsed.(*rsa.PrivateKey)
	assert.True(t, ok)

	// 3. Подготовка данных и шифрование по чанкам
	originalData := bytes.Repeat([]byte("X"), 512) // данные длиной > 1 чанка
	hash := sha256.New()
	maxChunkSize := privKey.Size() - 2*hash.Size() - 2

	var encrypted []byte
	for start := 0; start < len(originalData); start += maxChunkSize {
		end := start + maxChunkSize
		if end > len(originalData) {
			end = len(originalData)
		}
		chunk := originalData[start:end]

		enc, err := rsa.EncryptOAEP(hash, rand.Reader, &privKey.PublicKey, chunk, nil)
		assert.NoError(t, err)

		encrypted = append(encrypted, enc...)
	}

	decrypted, err := decryptOAEPChunks(encrypted, privKey)
	assert.NoError(t, err)
	assert.Equal(t, originalData, decrypted)
}

func generateTestPrivateKeyPKCS8(path string) error {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	return os.WriteFile(path, privPEM, 0644)
}

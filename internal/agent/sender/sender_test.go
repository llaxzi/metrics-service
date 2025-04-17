package sender

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSender_Send(t *testing.T) {
	// Фиктивный публичный ключ для теста
	pubKeyFile := "test_public.pem"
	err := generateTestPublicKey(pubKeyFile)
	assert.NoError(t, err)
	defer os.Remove(pubKeyFile)

	testTable := []struct {
		name       string
		metricsMap map[string]interface{}
		wantStatus int
		wantErr    bool
	}{
		{"OK", map[string]interface{}{
			"PollCount": int64(40),
			"Alloc":     float64(1234),
		}, http.StatusOK, false},

		{"Invalid metric type", map[string]interface{}{
			"PollCount": "invalid_type",
			"Alloc":     int64(1234),
		}, http.StatusBadRequest, true},

		{
			name:       "Empty metrics map",
			metricsMap: map[string]interface{}{},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/octet-stream", r.Header.Get("Content-Type"))
				assert.Equal(t, "encrypted", r.Header.Get("Content-Encoding"))

				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.NotEmpty(t, body)

				w.WriteHeader(test.wantStatus)
			}))
			defer server.Close()

			s, err := NewSender(server.URL, nil, pubKeyFile)
			assert.NoError(t, err)

			err = s.SendBatch(test.metricsMap)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func generateTestPublicKey(path string) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	pubASN1, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return err
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	return os.WriteFile(path, pubPEM, 0644)
}

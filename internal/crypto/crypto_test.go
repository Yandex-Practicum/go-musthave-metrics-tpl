package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeKeyPair создаёт временную пару RSA-ключей в PEM-файлах и возвращает
// пути к публичному и приватному.
func makeKeyPair(t *testing.T) (pubPath, privPath string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()

	privPath = filepath.Join(dir, "priv.pem")
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
		t.Fatal(err)
	}

	pubPath = filepath.Join(dir, "pub.pem")
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	if err := os.WriteFile(pubPath, pubPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	return pubPath, privPath
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	pubPath, privPath := makeKeyPair(t)
	pub, err := LoadPublicKey(pubPath)
	if err != nil {
		t.Fatal(err)
	}
	priv, err := LoadPrivateKey(privPath)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		data []byte
	}{
		{"small", []byte("hello")},
		{"empty", []byte{}},
		{"large", []byte(strings.Repeat("A", 10_000))},
		{"binary", []byte{0x00, 0xff, 0x10, 0x20, 0x30}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(pub, tt.data)
			if err != nil {
				t.Fatal(err)
			}
			decrypted, err := Decrypt(priv, encrypted)
			if err != nil {
				t.Fatal(err)
			}
			if string(decrypted) != string(tt.data) {
				t.Fatalf("decrypted=%q, want %q", decrypted, tt.data)
			}
		})
	}
}

func TestDecrypt_ShortInput(t *testing.T) {
	_, privPath := makeKeyPair(t)
	priv, _ := LoadPrivateKey(privPath)

	if _, err := Decrypt(priv, []byte{0x01}); err == nil {
		t.Fatal("ожидалась ошибка на слишком коротком входе")
	}
}

func TestLoadPublicKey_BadFile(t *testing.T) {
	if _, err := LoadPublicKey("/no/such/file"); err == nil {
		t.Fatal("ожидалась ошибка на несуществующем файле")
	}
	bad := filepath.Join(t.TempDir(), "bad.pem")
	os.WriteFile(bad, []byte("not a pem"), 0o600)
	if _, err := LoadPublicKey(bad); err == nil {
		t.Fatal("ожидалась ошибка на невалидном PEM")
	}
}

func TestLoadPrivateKey_BadFile(t *testing.T) {
	if _, err := LoadPrivateKey("/no/such/file"); err == nil {
		t.Fatal("ожидалась ошибка на несуществующем файле")
	}
	bad := filepath.Join(t.TempDir(), "bad.pem")
	os.WriteFile(bad, []byte("not a pem"), 0o600)
	if _, err := LoadPrivateKey(bad); err == nil {
		t.Fatal("ожидалась ошибка на невалидном PEM")
	}
}

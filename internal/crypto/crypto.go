// Package crypto реализует асимметричное шифрование тела HTTP-запросов
// между агентом и сервером по гибридной схеме RSA + AES-GCM.
//
// RSA-OAEP (SHA-256) шифрует одноразовый симметричный ключ AES-256,
// AES-GCM — само тело сообщения (позволяет любые размеры данных, RSA
// напрямую ограничен размером ключа). На выходе формируется двоичный
// пакет вида:
//
//	[2 bytes: длина ct_key][ct_key][nonce (12 bytes)][ciphertext + tag]
//
// Формат симметричен для Encrypt/Decrypt в этом же пакете и не совместим
// с внешними форматами (PKCS#7 CMS и т.п.).
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	// nonceSize — длина nonce для AES-GCM (стандартная — 12 байт).
	nonceSize = 12
	// aesKeySize — длина симметричного ключа AES-256.
	aesKeySize = 32
)

// LoadPublicKey читает RSA-публичный ключ из PEM-файла по указанному пути.
// Поддерживаются форматы PKIX ("PUBLIC KEY") и PKCS#1 ("RSA PUBLIC KEY").
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("чтение публичного ключа: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("публичный ключ: PEM-блок не найден")
	}
	switch block.Type {
	case "PUBLIC KEY":
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("разбор PKIX публичного ключа: %w", err)
		}
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("публичный ключ не RSA")
		}
		return rsaPub, nil
	case "RSA PUBLIC KEY":
		return x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		return nil, fmt.Errorf("неподдерживаемый тип PEM-блока: %s", block.Type)
	}
}

// LoadPrivateKey читает RSA-приватный ключ из PEM-файла по указанному пути.
// Поддерживаются форматы PKCS#8 ("PRIVATE KEY") и PKCS#1 ("RSA PRIVATE KEY").
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("чтение приватного ключа: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("приватный ключ: PEM-блок не найден")
	}
	switch block.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("разбор PKCS8 приватного ключа: %w", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("приватный ключ не RSA")
		}
		return rsaKey, nil
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("неподдерживаемый тип PEM-блока: %s", block.Type)
	}
}

// Encrypt шифрует data по гибридной схеме: генерирует случайный ключ
// AES-256, шифрует им данные (AES-GCM), а сам AES-ключ шифрует RSA-OAEP
// с помощью pub. Возвращает бинарный пакет, пригодный для передачи в
// теле HTTP-запроса.
func Encrypt(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	aesKey := make([]byte, aesKeySize)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, fmt.Errorf("генерация AES-ключа: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("создание AES-шифра: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("создание GCM: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("генерация nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	encKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("шифрование AES-ключа RSA: %w", err)
	}

	// Формат: [2 bytes: len(encKey)][encKey][nonce][ciphertext]
	buf := make([]byte, 0, 2+len(encKey)+len(nonce)+len(ciphertext))
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, uint16(len(encKey)))
	buf = append(buf, lenBuf...)
	buf = append(buf, encKey...)
	buf = append(buf, nonce...)
	buf = append(buf, ciphertext...)
	return buf, nil
}

// Decrypt расшифровывает пакет, созданный Encrypt: извлекает
// зашифрованный AES-ключ, расшифровывает его RSA-OAEP с помощью priv,
// затем расшифровывает тело AES-GCM.
func Decrypt(priv *rsa.PrivateKey, data []byte) ([]byte, error) {
	if len(data) < 2 {
		return nil, errors.New("шифротекст: слишком короткий заголовок")
	}
	keyLen := int(binary.BigEndian.Uint16(data[:2]))
	if len(data) < 2+keyLen+nonceSize {
		return nil, errors.New("шифротекст: некорректная длина")
	}
	encKey := data[2 : 2+keyLen]
	nonce := data[2+keyLen : 2+keyLen+nonceSize]
	ciphertext := data[2+keyLen+nonceSize:]

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, encKey, nil)
	if err != nil {
		return nil, fmt.Errorf("расшифровка AES-ключа: %w", err)
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("создание AES-шифра: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("создание GCM: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("расшифровка тела: %w", err)
	}
	return plaintext, nil
}

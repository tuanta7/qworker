package cipherx

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

type aead struct {
	aesgcm cipher.AEAD
}

func (a aead) Encrypt(plaintext []byte, opts ...[]byte) ([]byte, error) {
	var additionalData []byte
	if len(opts) > 0 {
		additionalData = opts[0]
	}

	nonce := make([]byte, a.aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ct := a.aesgcm.Seal(nil, nonce, plaintext, additionalData)
	return append(nonce, ct...), nil
}

func (a aead) Decrypt(ciphertext []byte, opts ...[]byte) ([]byte, error) {
	var additionalData []byte
	if len(opts) > 0 {
		additionalData = opts[0]
	}

	nonce := ciphertext[:a.aesgcm.NonceSize()]
	ct := ciphertext[a.aesgcm.NonceSize():]

	return a.aesgcm.Open(nil, nonce, ct, additionalData)
}

func (a aead) EncryptToStdBase64(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	ciphertext, err := a.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (a aead) DecryptFromStdBase64(base64Ciphertext string) (string, error) {
	if base64Ciphertext == "" {
		return "", nil
	}

	ciphertext := make([]byte, base64.StdEncoding.DecodedLen(len(base64Ciphertext)))
	n, err := base64.StdEncoding.Decode(ciphertext, []byte(base64Ciphertext))
	if err != nil {
		return "", err
	}

	plaintext, err := a.Decrypt(ciphertext[:n])
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

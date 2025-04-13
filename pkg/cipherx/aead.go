package cipherx

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

type aead struct {
	gcm cipher.AEAD
}

func (a aead) Encrypt(plaintext string, opts ...[]byte) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	ciphertext, err := a.encrypt([]byte(plaintext), opts...)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (a aead) encrypt(plaintext []byte, opts ...[]byte) ([]byte, error) {
	var additionalData []byte
	if len(opts) > 0 {
		additionalData = opts[0]
	}

	nonce := make([]byte, a.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := a.gcm.Seal(nil, nonce, plaintext, additionalData)
	return append(nonce, ciphertext...), nil
}

func (a aead) Decrypt(base64Ciphertext string, opts ...[]byte) (string, error) {
	if base64Ciphertext == "" {
		return "", nil
	}

	ciphertext := make([]byte, base64.StdEncoding.DecodedLen(len(base64Ciphertext)))
	n, err := base64.StdEncoding.Decode(ciphertext, []byte(base64Ciphertext))
	if err != nil {
		return "", err
	}

	plaintext, err := a.decrypt(ciphertext[:n], opts...)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (a aead) decrypt(c []byte, opts ...[]byte) ([]byte, error) {
	var additionalData []byte
	if len(opts) > 0 {
		additionalData = opts[0]
	}

	nonce := c[:a.gcm.NonceSize()]
	ciphertext := c[a.gcm.NonceSize():]
	return a.gcm.Open(nil, nonce, ciphertext, additionalData)
}

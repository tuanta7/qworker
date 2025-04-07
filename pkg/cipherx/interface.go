package cipherx

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

type Cipher interface {
	Encrypt(plaintext []byte, opts ...[]byte) ([]byte, error)
	Decrypt(ciphertext []byte, opts ...[]byte) ([]byte, error)
	EncryptToStdBase64(plaintext string) (string, error)
	DecryptFromStdBase64(base64Ciphertext string) (string, error)
}

type CipherType string

const (
	AEAD CipherType = "aead"
)

func New(secretKey []byte, t CipherType) (Cipher, error) {
	switch t {
	case AEAD:
		block, err := aes.NewCipher(secretKey)
		if err != nil {
			return nil, err
		}

		aesgcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		return &aead{aesgcm}, nil
	default:
		return nil, errors.New("cipher type not supported")
	}
}

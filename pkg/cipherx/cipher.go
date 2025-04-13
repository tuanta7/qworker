package cipherx

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

type CipherType string

const (
	AEAD CipherType = "AES_GCM_AEAD"
)

type Cipher interface {
	Encrypt(plaintext string, opts ...[]byte) (string, error)
	Decrypt(ciphertext string, opts ...[]byte) (string, error)
}

func New(t CipherType, secretKey []byte) (Cipher, error) {
	switch t {
	case AEAD:
		block, err := aes.NewCipher(secretKey)
		if err != nil {
			return nil, err
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		return &aead{gcm}, nil
	default:
		return nil, fmt.Errorf("cipher type not supported: %s", t)
	}
}

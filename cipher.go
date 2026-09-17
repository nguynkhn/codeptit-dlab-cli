package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

type Envelope struct {
	Ciphertext []byte `json:"data"`
	Nonce      []byte `json:"ptit"`
}

func generateNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return nonce, nil
}

func newAEAD(signKey []byte) (cipher.AEAD, error) {
	keyHash := sha256.Sum256(signKey)
	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

func EncryptEnvelope(writer io.Writer, signKey, plaintext []byte) error {
	aead, err := newAEAD(signKey)
	if err != nil {
		return err
	}

	nonce, err := generateNonce(aead.NonceSize())
	if err != nil {
		return err
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return json.NewEncoder(writer).Encode(Envelope{
		Ciphertext: ciphertext, Nonce: nonce,
	})
}

func DecryptEnvelope(reader io.Reader, signKey []byte) ([]byte, error) {
	aead, err := newAEAD(signKey)
	if err != nil {
		return nil, err
	}

	var envelope Envelope
	if err := json.NewDecoder(reader).Decode(&envelope); err != nil {
		return nil, err
	}

	return aead.Open(nil, envelope.Nonce, envelope.Ciphertext, nil)
}

type Signature struct {
	Nonce     string
	Sign      string
	Timestamp int64
}

func CreateSignature(signKey []byte, method, path string, body []byte, skewMs int64) (Signature, error) {
	var signature Signature

	nonce, err := generateNonce(16)
	if err != nil {
		return signature, err
	}

	signature.Timestamp = time.Now().UnixMilli() + skewMs
	signature.Nonce = hex.EncodeToString(nonce)

	canonical := fmt.Sprintf(
		"%s\n%s\n%d\n%x\n%x",
		strings.ToUpper(method),
		path,
		signature.Timestamp,
		nonce,
		sha256.Sum256(body),
	)
	mac := hmac.New(sha256.New, signKey)
	if _, err := mac.Write([]byte(canonical)); err != nil {
		return signature, err
	}

	sign := mac.Sum(nil)
	signature.Sign = hex.EncodeToString(sign)
	return signature, nil
}

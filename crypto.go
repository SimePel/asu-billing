package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
)

func createConfirmationLink(settings Settings) (string, error) {
	data, err := json.Marshal(&settings)
	if err != nil {
		return "", fmt.Errorf("Cannot marshal json: %v", err)
	}

	cipherData, err := encrypt(data)
	if err != nil {
		return "", fmt.Errorf("Cannot encrypt data: %v", err)
	}

	return "http://localhost:8080/confirm-settings?d=" + url.QueryEscape(cipherData), nil
}

func encrypt(text []byte) (string, error) {
	key := []byte(os.Getenv("AES_KEY"))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("Cannot make new cipher: %v", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("Cannot fill iv: %v", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(ciphertext string) ([]byte, error) {
	key := []byte(os.Getenv("AES_KEY"))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("Cannot make new cipher: %v", err)
	}

	text, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode cipher text to []byte: %v", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("Ciphertext too short")
	}

	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(text, text)

	return text, nil
}

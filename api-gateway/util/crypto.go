package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// PKCS7Padding 假设数据需要填充n个字节才对齐，那么填充n个字节，每个字节都是n
// 如果数据本身就已经对齐了，则填充一块长度为块大小的数据，每个字节都是块大小
func PKCS7Padding(b []byte) []byte {
	padding := aes.BlockSize - len(b)%aes.BlockSize
	result := make([]byte, len(b)+padding)
	copy(result, b)
	idx := len(b)
	for i := 0; i < padding; i++ {
		result[idx+i] = byte(padding)
	}
	return result
}

// AESEncrypt aes encrypt
func AESEncrypt(txt string, secret []byte) (string, error) {
	plaintext := PKCS7Padding([]byte(txt))

	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt aes decrypt
func AESDecrypt(s string, secret []byte) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("invalid cipher text: [%s]", s)
	}

	if len(ciphertext) <= aes.BlockSize || len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("invalid cipher text: [%s]", s)
	}

	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", err
	}
	iv := ciphertext[:aes.BlockSize]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)
	// remove padding
	padding := int(ciphertext[len(ciphertext)-1])
	if len(ciphertext)-padding <= aes.BlockSize {
		return "", fmt.Errorf("invalid cipher text: [%s]", s)
	}
	ciphertext = ciphertext[aes.BlockSize : len(ciphertext)-padding]
	return string(ciphertext), nil
}

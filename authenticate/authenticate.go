package authenticate

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

const (
	base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

var coder = base64.NewEncoding(base64Table)

func Base64Encode(plaintext []byte) []byte {
	return []byte(coder.EncodeToString(plaintext))
}

func Base64Decode(ciphertext []byte) ([]byte, error) {
	return coder.DecodeString(string(ciphertext))
}

func TokenEncode(plaintext []byte, token string) []byte {
	// TODO
	return plaintext
}

func TokenDecode(ciphertext []byte, token string) ([]byte, error) {
	// TODO
	return ciphertext, nil
}

func NewAesBlock(key []byte) cipher.Block {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}
	return block
}

func AesEncode(plaintext []byte, block cipher.Block) []byte {
	cfb := cipher.NewCFBEncrypter(block, commonIV)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	return []byte(ciphertext)
}

func AesDecode(cipherText []byte, plainLen int64, block cipher.Block) ([]byte, error) {
	cfbdec := cipher.NewCFBDecrypter(block, commonIV)
	plaintext := make([]byte, plainLen)
	cfbdec.XORKeyStream(plaintext, cipherText)
	return []byte(plaintext), nil
}

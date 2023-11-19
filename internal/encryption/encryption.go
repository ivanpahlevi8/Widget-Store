package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
)

// create encryption object
type Encryption struct {
	Key []byte
}

// crete function to enmcryption string
func (e *Encryption) EncryptionText(text string) (string, error) {
	// change text into bytes to be encryptedt
	textPlain := []byte(text)

	// create encryption block to be embed with plain text
	block, err := aes.NewCipher(e.Key)

	// check for an error
	if err != nil {
		log.Println("error when creating block ecnryption : ", err)
		return "", err
	}

	// make block to hold value of text and ecnryption
	holdData := make([]byte, aes.BlockSize+len(textPlain))

	// get block size of hold data to be assign with encryption
	iv := holdData[:aes.BlockSize]

	// assign encryption to block part
	_, err = io.ReadFull(rand.Reader, iv)

	// check for an error
	if err != nil {
		log.Println("error when write encryption to block of encryption : ", err)
		return "", err
	}

	// assign encryption
	stream := cipher.NewCFBEncrypter(block, iv)

	// assign text to result hold data with encryuption code
	stream.XORKeyStream(holdData[aes.BlockSize:], textPlain)

	// return hold data as base64 encryption
	return base64.URLEncoding.EncodeToString(holdData), nil
}

// create function to decode
func (e *Encryption) DecodeText(text string) (string, error) {
	// decode base 64
	decodeText, err := base64.URLEncoding.DecodeString(text)

	// check for an error
	if err != nil {
		log.Println("error when decoding text : ", err)
		return "", err
	}

	// create encryption block to be embed with plain text
	block, err := aes.NewCipher(e.Key)

	// check if decode text valid
	if len(decodeText) < aes.BlockSize {
		log.Println("not valid decode text : ", err)
		return "", err
	}

	// get text
	textEncry := decodeText[:aes.BlockSize]

	// get encryp code
	code := decodeText[aes.BlockSize:]

	// crete stream
	stream := cipher.NewCFBDecrypter(block, textEncry)

	stream.XORKeyStream(code, code)

	return fmt.Sprintf("%s", code), nil
}

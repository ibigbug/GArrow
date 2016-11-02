package arrow

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

type Cipher struct {
	iv    []byte
	block cipher.Block
	encer cipher.Stream
	decer cipher.Stream
}

func (c *Cipher) initEncer() {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		fmt.Println("Error initialize encer:", err)
	}
	c.iv = iv
	c.encer = cipher.NewCFBEncrypter(c.block, iv)
}

func (c *Cipher) Encrypt(plain []byte) (ciphered []byte) {
	ciphered = make([]byte, len(plain))
	c.encer.XORKeyStream(ciphered, plain)
	return
}

func (c *Cipher) initDecer(iv []byte) {
	c.decer = cipher.NewCFBDecrypter(c.block, iv)
}

func (c *Cipher) Decrypt(ciphered []byte) {
	c.decer.XORKeyStream(ciphered, ciphered)
}

func NewCipher(password string) (c *Cipher, err error) {
	key := sha256.Sum256([]byte(password))
	block, err := aes.NewCipher(key[:])
	c = &Cipher{
		block: block,
	}
	return
}

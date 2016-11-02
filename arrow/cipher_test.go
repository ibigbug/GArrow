package arrow

import "testing"

import "reflect"

func TestEncDec(t *testing.T) {
	plain := []byte("fuck1")
	password := "000"

	c, _ := NewCipher(password)
	iv := c.initEncer()
	ciphered := c.Encrypt(plain)
	c.initDecer(iv)
	c.Decrypt(ciphered)

	if !reflect.DeepEqual(plain, ciphered) {
		t.Error(1)
	}
}

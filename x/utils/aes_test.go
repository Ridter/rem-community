package utils

import (
	"testing"
)

func TestPKCS7Padding(t *testing.T) {
	print(string(PKCS7Padding([]byte{}, 16)))
}

package utils

import (
	"crypto/cipher"
	"io"
)

// XorStream 实现cipher.Stream接口
type XorStream struct {
	key     []byte
	iv      []byte
	counter int
}

// XORKeyStream 实现cipher.Stream接口
func (x *XorStream) XORKeyStream(dst, src []byte) {
	keyLen := len(x.key)
	ivLen := len(x.iv)

	for i := range src {
		index := x.counter + i
		keyByte := x.key[index%keyLen]
		ivByte := x.iv[index%ivLen]
		dst[i] = src[i] ^ keyByte ^ ivByte
	}

	x.counter += len(src)
}

// XorEncryptor 封装了加密和解密的Stream
type XorEncryptor struct {
	key    []byte
	iv     []byte
	stream cipher.Stream
}

// NewXorEncryptor 创建一个新的XorEncryptor
func NewXorEncryptor(key []byte, iv []byte) *XorEncryptor {
	stream := &XorStream{
		key:     key,
		iv:      iv,
		counter: 0,
	}
	return &XorEncryptor{
		key:    key,
		iv:     iv,
		stream: stream,
	}
}

// GetStream 获取当前的cipher.Stream实例
func (e *XorEncryptor) GetStream() cipher.Stream {
	return e.stream
}

// Encrypt 使用cipher.StreamWriter进行加密
func (e *XorEncryptor) Encrypt(dst io.Writer, src io.Reader) error {
	writer := &cipher.StreamWriter{S: e.stream, W: dst}
	_, err := io.Copy(writer, src)
	return err
}

// Decrypt 使用cipher.StreamReader进行解密
func (e *XorEncryptor) Decrypt(dst io.Writer, src io.Reader) error {
	reader := &cipher.StreamReader{S: e.stream, R: src}
	_, err := io.Copy(dst, reader)
	return err
}

// Reset 重置加密器状态
func (e *XorEncryptor) Reset() error {
	e.stream = &XorStream{
		key:     e.key,
		iv:      e.iv,
		counter: 0,
	}
	return nil
}

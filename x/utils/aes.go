package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
)

// AesStream 实现cipher.Stream接口
type AesStream struct {
	block   cipher.Block
	iv      []byte
	counter []byte
	stream  cipher.Stream
}

// NewAesStream 创建一个新的AES流加密器
func NewAesStream(key [32]byte, iv [16]byte) (cipher.Stream, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv[:])
	return stream, nil
}

// AesCtrEncryptor 封装了加密和解密的Stream
type AesCtrEncryptor struct {
	key    [32]byte
	iv     [16]byte
	stream cipher.Stream
}

// NewAesCtrEncryptor 创建一个新的AesCtrEncryptor
func NewAesCtrEncryptor(key [32]byte, iv [16]byte) (*AesCtrEncryptor, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv[:])
	return &AesCtrEncryptor{
		key:    key,
		iv:     iv,
		stream: stream,
	}, nil
}

// GetStream 获取当前的cipher.Stream实例
func (e *AesCtrEncryptor) GetStream() cipher.Stream {
	return e.stream
}

// Encrypt 使用cipher.StreamWriter进行加密
func (e *AesCtrEncryptor) Encrypt(dst io.Writer, src io.Reader) error {
	writer := &cipher.StreamWriter{S: e.stream, W: dst}
	_, err := io.Copy(writer, src)
	return err
}

// Decrypt 使用cipher.StreamReader进行解密
func (e *AesCtrEncryptor) Decrypt(dst io.Writer, src io.Reader) error {
	reader := &cipher.StreamReader{S: e.stream, R: src}
	_, err := io.Copy(dst, reader)
	return err
}

// Reset 重置加密器状态
func (e *AesCtrEncryptor) Reset() error {
	block, err := aes.NewCipher(e.key[:])
	if err != nil {
		return err
	}
	e.stream = cipher.NewCTR(block, e.iv[:])
	return nil
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// AES加密,CBC
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// AES解密
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS7UnPadding(origData)
	return origData, nil
}

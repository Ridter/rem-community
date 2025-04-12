package wrapper

import (
	"crypto/cipher"
	"io"

	"github.com/chainreactors/rem/protocol/core"
	"github.com/chainreactors/rem/x/utils"
)

func init() {
	core.WrapperRegister(core.XORWrapper, func(r io.Reader, w io.Writer, opt map[string]string) (core.Wrapper, error) {
		return NewXorWrapper(r, w, opt), nil
	})
	core.AvailableWrappers = append(core.AvailableWrappers, core.XORWrapper)
}

type XorWrapper struct {
	reader    io.Reader
	writer    io.Writer
	encStream cipher.Stream
	decStream cipher.Stream
	key       []byte
	iv        []byte
}

func NewXorWrapper(r io.Reader, w io.Writer, opt map[string]string) core.Wrapper {
	var key []byte
	if k, ok := opt["key"]; ok {
		key = []byte(k)
	} else {
		key = []byte{} // 使用空字节切片作为默认值
	}

	var iv []byte
	if i, ok := opt["iv"]; ok {
		iv = []byte(i)
	} else {
		iv = key // 如果没有提供iv，使用key作为iv
	}

	encryptor := utils.NewXorEncryptor(key, iv)
	decryptor := utils.NewXorEncryptor(key, iv)

	return &XorWrapper{
		reader:    &cipher.StreamReader{S: decryptor.GetStream(), R: r},
		writer:    &cipher.StreamWriter{S: encryptor.GetStream(), W: w},
		encStream: encryptor.GetStream(),
		decStream: decryptor.GetStream(),
		key:       key,
		iv:        iv,
	}
}

func (w *XorWrapper) Name() string {
	return core.XORWrapper
}

func (w *XorWrapper) Read(p []byte) (n int, err error) {
	return w.reader.Read(p)
}

func (w *XorWrapper) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *XorWrapper) Close() error {
	encryptor := utils.NewXorEncryptor(w.key, w.iv)
	decryptor := utils.NewXorEncryptor(w.key, w.iv)
	w.encStream = encryptor.GetStream()
	w.decStream = decryptor.GetStream()
	w.reader = &cipher.StreamReader{S: w.decStream, R: w.reader.(*cipher.StreamReader).R}
	w.writer = &cipher.StreamWriter{S: w.encStream, W: w.writer.(*cipher.StreamWriter).W}
	return nil
}

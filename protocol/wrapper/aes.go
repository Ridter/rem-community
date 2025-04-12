package wrapper

import (
	"crypto/cipher"
	"io"

	"github.com/chainreactors/rem/protocol/core"
	"github.com/chainreactors/rem/x/utils"
)

func init() {
	core.WrapperRegister(core.AESWrapper, func(r io.Reader, w io.Writer, opt map[string]string) (core.Wrapper, error) {
		return NewAesWrapper(r, w, opt)
	})
	core.AvailableWrappers = append(core.AvailableWrappers, core.AESWrapper)
}

type AESWrapper struct {
	reader    io.Reader
	writer    io.Writer
	encStream cipher.Stream
	decStream cipher.Stream
	key       [32]byte
	iv        [16]byte
}

func NewAesWrapper(r io.Reader, w io.Writer, opt map[string]string) (core.Wrapper, error) {
	var key [32]byte
	if k, ok := opt["key"]; ok {
		copy(key[:], k)
	}

	var iv [16]byte
	if i, ok := opt["iv"]; ok {
		copy(iv[:], i)
	} else {
		copy(iv[:], key[:16]) // 如果没有提供iv，使用key的前16字节
	}

	encStream, err := utils.NewAesStream(key, iv)
	if err != nil {
		return nil, err
	}

	decStream, err := utils.NewAesStream(key, iv)
	if err != nil {
		return nil, err
	}

	return &AESWrapper{
		reader:    &cipher.StreamReader{S: decStream, R: r},
		writer:    &cipher.StreamWriter{S: encStream, W: w},
		encStream: encStream,
		decStream: decStream,
		key:       key,
		iv:        iv,
	}, nil
}

func (w *AESWrapper) Name() string {
	return core.AESWrapper
}

func (w *AESWrapper) Read(p []byte) (n int, err error) {
	return w.reader.Read(p)
}

func (w *AESWrapper) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *AESWrapper) Close() error {
	encStream, err := utils.NewAesStream(w.key, w.iv)
	if err != nil {
		return err
	}
	decStream, err := utils.NewAesStream(w.key, w.iv)
	if err != nil {
		return err
	}
	w.encStream = encStream
	w.decStream = decStream
	w.reader = &cipher.StreamReader{S: w.decStream, R: w.reader.(*cipher.StreamReader).R}
	w.writer = &cipher.StreamWriter{S: w.encStream, W: w.writer.(*cipher.StreamWriter).W}
	return nil
}

package wrapper

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/chainreactors/rem/protocol/cio"
	"github.com/chainreactors/rem/protocol/core"
)

func init() {
	core.WrapperRegister(core.PaddingWrapper, func(r io.Reader, w io.Writer, opt map[string]string) (core.Wrapper, error) {
		return NewPaddingWrapper(r, w, opt), nil
	})
}

// PaddingOption 定义padding wrapper的配置选项
type PaddingOption struct {
	Prefix []byte
	Suffix []byte
}

type PaddingWrapper struct {
	reader      *cio.Reader
	writer      *cio.Writer
	parseLength func(reader *cio.Reader) (uint32, error)
	genLength   func(p []byte) []byte
	suffix      []byte
	prefix      []byte
}

func NewPaddingWrapper(r io.Reader, w io.Writer, opt map[string]string) core.Wrapper {
	var prefix, suffix string
	if p, ok := opt["prefix"]; ok {
		prefix = p
	}
	if s, ok := opt["suffix"]; ok {
		suffix = s
	}

	return &PaddingWrapper{
		reader:      cio.NewReader(r),
		writer:      cio.NewWriter(w),
		prefix:      []byte(prefix),
		suffix:      []byte(suffix),
		genLength:   defaultGenLength,
		parseLength: defaultParserLength,
	}
}

func (w *PaddingWrapper) Name() string {
	return "padding"
}

func (w *PaddingWrapper) Fill() error {
	err := w.reader.PeekAndRead(w.prefix)
	if err != nil {
		return err
	}
	n, err := w.parseLength(w.reader)
	if err != nil {
		return err
	}
	err = w.reader.FillN(int64(n))
	if err != nil {
		return err
	}

	return w.reader.PeekAndRead(w.suffix)
}

func (w *PaddingWrapper) Read(p []byte) (n int, err error) {
	if w.reader.Buffer.Size() == 0 {
		err = w.Fill()
		if err != nil {
			return 0, err
		}
	}
	return w.reader.Read(p)
}

func (w *PaddingWrapper) Write(p []byte) (n int, err error) {
	var buf bytes.Buffer
	buf.Write(w.prefix)
	buf.Write(w.genLength(p))
	buf.Write(p)
	buf.Write(w.suffix)

	n, err = w.writer.Write(buf.Bytes())
	if err != nil {
		return n, err
	}
	return len(p), nil
}

func (w *PaddingWrapper) Close() error {
	return nil
}

func defaultParserLength(reader *cio.Reader) (uint32, error) {
	p := make([]byte, 4)
	_, err := reader.Read(p)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(p), nil
}

func defaultGenLength(p []byte) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(p)))
	return buf
}

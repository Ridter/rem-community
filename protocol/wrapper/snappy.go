package wrapper

import (
	"github.com/chainreactors/rem/protocol/core"
	"github.com/golang/snappy"
	"io"
)

type SnappyWrapper struct {
	reader io.Reader
	writer io.Writer
}

func NewSnappyWrapper(r io.Reader, w io.Writer, opt map[string]string) core.Wrapper {
	return &SnappyWrapper{
		reader: snappy.NewReader(r), // 创建Snappy解压Reader
		writer: snappy.NewWriter(w), // 创建Snappy压缩Writer
	}
}

func (w *SnappyWrapper) Name() string {
	return "snappy"
}

func (w *SnappyWrapper) Read(p []byte) (n int, err error) {
	return w.reader.Read(p) // 从解压流读取数据
}

func (w *SnappyWrapper) Write(p []byte) (n int, err error) {
	return w.writer.Write(p) // 向压缩流写入数据
}

func (w *SnappyWrapper) Close() error {
	return nil
}

package wrapper

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestNewSnappyWrapper(t *testing.T) {
	originalData := []byte("Hello, this is a test for Snappy compression!")

	// 压缩
	var compressed bytes.Buffer
	writerWrapper := NewSnappyWrapper(nil, &compressed, nil)
	writerWrapper.Write(originalData)
	writerWrapper.Close()

	fmt.Printf("Original size: %d bytes, Compressed size: %d bytes\n", len(originalData), compressed.Len())

	// 解压
	readerWrapper := NewSnappyWrapper(&compressed, nil, nil)
	decompressed := new(bytes.Buffer)
	_, err := io.Copy(decompressed, readerWrapper)
	if err != nil {
		fmt.Println("Error decompressing:", err)
		return
	}

	fmt.Println("Decompressed data:", decompressed.String())
}

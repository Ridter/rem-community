package kcp

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

// Buffer 是一个线程安全的带大小限制的缓冲区
type Buffer struct {
	buf    *bytes.Buffer
	mu     sync.RWMutex
	cond   *sync.Cond
	maxLen int
}

// NewBuffer 创建一个新的Buffer
func NewBuffer(maxLen int) *Buffer {
	b := &Buffer{
		buf:    bytes.NewBuffer(nil),
		maxLen: maxLen,
	}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Write 写入数据,如果缓冲区满则阻塞等待
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	remaining := len(p)
	written := 0

	for remaining > 0 {
		// 计算当前可写入的空间
		availSpace := b.maxLen - b.buf.Len()
		if availSpace <= 0 {
			// 如果没有可用空间，等待
			b.cond.Wait() // Wait会自动释放锁并在返回时重新获取
			continue
		}

		// 确定本次写入的长度
		writeLen := remaining
		if writeLen > availSpace {
			writeLen = availSpace
		}

		n, _ = b.buf.Write(p[written : written+writeLen])

		written += n
		remaining -= n

		b.cond.Signal()
	}

	return written, nil
}

// Read 读取数据,如果没有数据则返回EOF
func (b *Buffer) Read(p []byte) (n int, err error) {
	if b.buf.Len() == 0 {
		return 0, io.EOF
	}

	b.mu.Lock()
	n, err = b.buf.Read(p)
	if n > 0 {
		b.cond.Signal()
	}
	b.mu.Unlock()
	return
}

// ReadAtLeast 读取数据，如果缓冲区为空则阻塞等待，直到有数据可读
func (b *Buffer) ReadAtLeast(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 如果没有数据，等待直到有数据
	for b.buf.Len() == 0 {
		b.cond.Wait()
	}

	n, err = b.buf.Read(p)
	if n > 0 {
		b.cond.Signal()
	}
	return
}

// Close 关闭缓冲区
func (b *Buffer) Close() error {
	b.mu.Lock()
	b.buf.Reset()
	// 关闭时需要唤醒所有等待者
	b.cond.Broadcast()
	b.mu.Unlock()
	return nil
}

// Size 返回当前缓冲区长度
func (b *Buffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock() // 这里可以用defer，因为没有Wait操作
	return b.buf.Len()
}

// Cap 返回缓冲区容量
func (b *Buffer) Cap() int {
	return b.maxLen
}

func NewChannel(size int, timeout time.Duration) *ChannelBuffer {
	if size == 0 {
		return &ChannelBuffer{
			ch:      make(chan []byte, 1),
			timeout: timeout,
		}
	} else {
		return &ChannelBuffer{
			ch:      make(chan []byte, size),
			timeout: timeout,
		}
	}
}

type ChannelBuffer struct {
	ch      chan []byte
	timeout time.Duration
}

func (ch *ChannelBuffer) Get() ([]byte, error) {
	select {
	case data := <-ch.ch:
		return data, nil
	case <-time.After(ch.timeout):
		return nil, fmt.Errorf("timeout")
	}
}

func (ch *ChannelBuffer) Put(data []byte) (int, error) {
	select {
	case ch.ch <- data:
		return len(data), nil
	case <-time.After(ch.timeout):
		return 0, fmt.Errorf("timeout")
	}
}

func (ch *ChannelBuffer) Close() error {
	close(ch.ch)
	return nil
}

func (ch *ChannelBuffer) Len() int {
	return len(ch.ch)
}

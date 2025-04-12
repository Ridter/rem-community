package kcp

import (
	"context"
	"io"
)

type KCPConn struct {
	*KCPSession
	ctx    context.Context
	cancel context.CancelFunc
}

func RadicalKCPConfig(conn *KCPSession) {
	conn.SetWriteDelay(false)
	conn.SetWindowSize(1024, 1024)
	conn.SetReadBuffer(1024 * 1024)
	conn.SetWriteBuffer(1024 * 1024)
	conn.SetNoDelay(1, 100, 2, 1)
	conn.SetMtu(1400)
	conn.SetACKNoDelay(true)
}

func HTTPKCPConfig(conn *KCPSession) {
	conn.SetWriteDelay(true)
	//conn.SetWindowSize(16, 16)
	conn.SetReadBuffer(1024 * 1024)
	conn.SetWriteBuffer(1024 * 1024)
	conn.SetNoDelay(0, 100, 10, 0)
	conn.SetMtu(mtuLimit - 16*1024)
	conn.SetACKNoDelay(true)
}

// NewKCPConn 创建新的KCP连接包装器
func NewKCPConn(conn *KCPSession, confn func(*KCPSession)) *KCPConn {
	ctx, cancel := context.WithCancel(context.Background())
	confn(conn)
	return &KCPConn{
		KCPSession: conn,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Read 重写Read方法，处理EOF问题，尽可能读满缓冲区
func (k *KCPConn) Read(p []byte) (n int, err error) {
	for n < len(p) {
		select {
		case <-k.ctx.Done():
			if n > 0 {
				return n, nil
			}
			return 0, io.ErrClosedPipe
		default:
			nr, err := k.KCPSession.Read(p[n:])
			if nr > 0 {
				n += nr
			}
			if err != nil {
				return n, err
			}
		}
	}
	return n, nil
}

// Write 重写Write方法，添加超时控制
func (k *KCPConn) Write(p []byte) (n int, err error) {
	select {
	case <-k.ctx.Done():
		return 0, io.ErrClosedPipe
	default:
		return k.KCPSession.Write(p)
	}
}

// Close 关闭连接
func (k *KCPConn) Close() error {
	k.cancel()
	return k.KCPSession.Close()
}

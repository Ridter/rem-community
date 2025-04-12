package cio

import (
	"sync/atomic"

	"golang.org/x/time/rate"
)

var GlobalLimiter *Limiter

func init() {
	GlobalLimiter = NewLimiter(rate.Inf, rate.Inf, 1024*1024)
}

// Limiter 全局限速器
type Limiter struct {
	readLimiter  *rate.Limiter
	writeLimiter *rate.Limiter
	readEnabled  atomic.Bool  // 读限速开关
	writeEnabled atomic.Bool  // 写限速开关
	readCount    atomic.Int64 // 读取计数器
	writeCount   atomic.Int64 // 写入计数器
}

func NewLimiter(readRate, writeRate rate.Limit, burstSize int) *Limiter {
	return &Limiter{
		readLimiter:  rate.NewLimiter(readRate, burstSize),
		writeLimiter: rate.NewLimiter(writeRate, burstSize),
	}
}

// GetCounts 获取读写计数
func (l *Limiter) GetCounts() (readCount, writeCount int64) {
	return l.readCount.Load(), l.writeCount.Load()
}

// SetReadRate 设置读取速率
func (l *Limiter) SetReadRate(readRate rate.Limit) {
	l.readLimiter.SetLimit(readRate)
}

// SetWriteRate 设置写入速率
func (l *Limiter) SetWriteRate(writeRate rate.Limit) {
	l.writeLimiter.SetLimit(writeRate)
}

// EnableReadLimit 启用/禁用读限速
func (l *Limiter) EnableReadLimit(enable bool) {
	l.readEnabled.Store(enable)
}

// EnableWriteLimit 启用/禁用写限速
func (l *Limiter) EnableWriteLimit(enable bool) {
	l.writeEnabled.Store(enable)
}

// GetLimits 获取当前的限速设置
func (l *Limiter) GetLimits() (readLimit, writeLimit rate.Limit) {
	return l.readLimiter.Limit(), l.writeLimiter.Limit()
}

// IsReadEnabled 检查读限速是否启用
func (l *Limiter) IsReadEnabled() bool {
	return l.readEnabled.Load()
}

// IsWriteEnabled 检查写限速是否启用
func (l *Limiter) IsWriteEnabled() bool {
	return l.writeEnabled.Load()
}

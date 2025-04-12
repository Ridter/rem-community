package cio

import (
	"fmt"
	"net"
	"sync"

	"github.com/chainreactors/go-metrics"
	"github.com/chainreactors/rem/x/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const (
	Sender   = 1
	Receiver = 2
	// 使用固定大小的数组来存储pending计数，避免使用map
	// 通常情况下同时存在的连接不会超过这个数
	maxConnections = 1024
)

type Message struct {
	ID      uint64
	Message proto.Message
}

// TrafficStats 流量统计
type TrafficStats struct {
	bytes        metrics.Counter
	packets      metrics.Counter
	rate         metrics.Meter
	pendingCount metrics.Counter // 待发送的数据包数量
	pendingSize  metrics.Counter // 待发送的数据大小
	pendingMap   sync.Map        // 存储每个连接的pending counter
}

func NewTrafficStats(name string) *TrafficStats {
	return &TrafficStats{
		bytes:        metrics.NewCounter(),
		packets:      metrics.NewCounter(),
		rate:         metrics.NewMeter(),
		pendingCount: metrics.NewCounter(),
		pendingSize:  metrics.NewCounter(),
	}
}

func (ts *TrafficStats) String(mod int) string {
	if mod == Sender {
		return fmt.Sprintf(
			"%d packets %d B, rate: %.2f B/s, pending: %d packets %d B",
			ts.packets.Count(),
			ts.bytes.Count(),
			ts.rate.Rate1(),
			ts.pendingCount.Count(),
			ts.pendingSize.Count(),
		)
	} else {
		return fmt.Sprintf(
			"%d packets %d B, rate: %.2f B/s",
			ts.packets.Count(),
			ts.bytes.Count(),
			ts.rate.Rate1(),
		)
	}
}

// AddPending 增加待发送计数
func (ts *TrafficStats) AddPending(id uint64, size int64) {
	counter := ts.getPendingCounter(id)
	counter.Inc(1)
	ts.pendingCount.Inc(1)
	ts.pendingSize.Inc(size)
}

// RemovePending 减少待发送计数
func (ts *TrafficStats) RemovePending(id uint64, size int64) {
	counter := ts.getPendingCounter(id)
	counter.Dec(1)
	ts.pendingCount.Dec(1)
	ts.pendingSize.Dec(size)
}

// GetPendingCount 获取指定ID的待发送包数量
func (ts *TrafficStats) GetPendingCount(id uint64) int64 {
	counter := ts.getPendingCounter(id)
	return counter.Count()
}

// ClearPending 清理所有待发送计数
func (ts *TrafficStats) ClearPending() {
	ts.pendingMap.Range(func(key, value interface{}) bool {
		counter := value.(metrics.Counter)
		counter.Clear()
		return true
	})
	ts.pendingCount.Clear()
	ts.pendingSize.Clear()
}

// getPendingCounter 获取或创建指定ID的counter
func (ts *TrafficStats) getPendingCounter(id uint64) metrics.Counter {
	value, ok := ts.pendingMap.Load(id)
	if !ok {
		counter := metrics.NewCounter()
		ts.pendingMap.Store(id, counter)
		return counter
	}
	return value.(metrics.Counter)
}

func NewChan(name string, capacity int) *Channel {
	var ch chan *Message
	if capacity == 0 {
		ch = make(chan *Message)
	} else {
		ch = make(chan *Message, capacity)
	}
	return &Channel{
		Name:   name,
		C:      ch,
		StopCh: make(chan struct{}),
		stats:  NewTrafficStats(name),
	}
}

type Channel struct {
	Mod    int
	Name   string
	C      chan *Message
	StopCh chan struct{}
	stats  *TrafficStats // 流量统计
}

// Send 发送消息
func (m *Channel) Send(id uint64, msg proto.Message) error {
	select {
	case <-m.StopCh:
		return errors.New("channel has closed")
	default:
	}

	msgWithID := &Message{
		ID:      id,
		Message: msg,
	}

	if m.Mod == Sender {
		size := int64(proto.Size(msg))
		m.stats.AddPending(id, size)
	}

	select {
	case <-m.StopCh:
		return errors.New("channel has closed")
	case m.C <- msgWithID:
	}

	return nil
}

func (m *Channel) Receiver(conn net.Conn) error {
	m.Mod = Receiver
	for {
		select {
		case <-m.StopCh:
			return nil
		default:
			msg, err := ReadMsg(conn)
			if err != nil {
				utils.Log.Debugf("[channel.receive] %s break, %s", m.Name, err.Error())
				m.Close()
				return err
			}

			size := int64(proto.Size(msg))
			m.stats.packets.Inc(1)
			m.stats.bytes.Inc(size)
			m.stats.rate.Mark(size)

			err = m.Send(0, msg)
			if err != nil {
				m.Close()
				return err
			}
		}
	}
}

func (m *Channel) Sender(conn net.Conn) error {
	m.Mod = Sender
	for {
		select {
		case <-m.StopCh:
			return nil
		case msg := <-m.C:
			if msg == nil || msg.Message == nil {
				return fmt.Errorf("nil message")
			}
			err := WriteMsg(conn, msg.Message)
			if err != nil {
				utils.Log.Errorf("[channel.send] %s break, %s", m.Name, err.Error())
				m.Close()
				return err
			}

			size := int64(proto.Size(msg.Message))
			m.stats.packets.Inc(1)
			m.stats.bytes.Inc(size)
			m.stats.rate.Mark(size)
			m.stats.RemovePending(msg.ID, size)
		}
	}
}

// GetPendingCount 获取指定ID的待发送包数量
func (m *Channel) GetPendingCount(id uint64) int64 {
	return m.stats.GetPendingCount(id)
}

func (m *Channel) GetStats() string {
	return m.stats.String(m.Mod)
}

func (m *Channel) Close() {
	defer func() {
		if recover() != nil {
		}
	}()

	utils.Log.Debugf("[channel.close] close channel %s", m.Name)
	close(m.StopCh)
	close(m.C)
	m.stats.ClearPending()
}

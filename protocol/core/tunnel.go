package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
)

type TunnelDialer interface {
	Dial(dst string) (net.Conn, error)
}

type WrappedDialer struct {
	TunnelDialer
	Dialer func(string) (net.Conn, error)
}

func (w *WrappedDialer) Dial(dst string) (net.Conn, error) {
	if w.Dialer != nil {
		return w.Dialer(dst)
	}
	return w.Dial(dst)
}

type TunnelListener interface {
	net.Listener
	Listen(dst string) (net.Listener, error)
}

type WrappedListener struct {
	TunnelListener
	Lns net.Listener
}

func (w *WrappedListener) Accept() (net.Conn, error) {
	if w.Lns != nil {
		return w.Lns.Accept()
	}
	return w.Accept()
}

// DialerCreatorFn 是创建 dialer 的工厂函数
type DialerCreatorFn func(ctx context.Context) (TunnelDialer, error)

// ListenerCreatorFn 是创建 listener 的工厂函数
type ListenerCreatorFn func(ctx context.Context) (TunnelListener, error)

var (
	dialerCreators   = make(map[string]DialerCreatorFn)
	listenerCreators = make(map[string]ListenerCreatorFn)
)

// DialerRegister 注册一个 dialer 类型
func DialerRegister(name string, fn DialerCreatorFn) {
	if _, exist := dialerCreators[name]; exist {
		panic(fmt.Sprintf("dialer [%s] is already registered", name))
	}
	dialerCreators[name] = fn
}

// ListenerRegister 注册一个 listener 类型
func ListenerRegister(name string, fn ListenerCreatorFn) {
	if _, exist := listenerCreators[name]; exist {
		panic(fmt.Sprintf("listener [%s] is already registered", name))
	}
	listenerCreators[name] = fn
}

// DialerCreate 创建一个指定类型的 dialer
func DialerCreate(name string, ctx context.Context) (TunnelDialer, error) {
	if fn, ok := dialerCreators[name]; ok {
		return fn(ctx)
	}
	return nil, fmt.Errorf("dialer [%s] is not registered", name)
}

// ListenerCreate 创建一个指定类型的 listener
func ListenerCreate(name string, ctx context.Context) (TunnelListener, error) {
	if fn, ok := listenerCreators[name]; ok {
		return fn(ctx)
	}
	return nil, fmt.Errorf("listener [%s] is not registered", name)
}

func GetMetas(ctx context.Context) Metas {
	if m, ok := ctx.Value("meta").(Metas); ok {
		return m
	}
	return nil
}

type Metas map[string]interface{}

func (m Metas) Value(key string) interface{} {
	return m[key]
}

func (m Metas) GetString(key string) string {
	v, ok := m[key]
	if ok {
		s, ok := v.(string)
		if ok {
			return s
		}
		return ""
	}
	return ""
}
func (m Metas) URL() *URL {
	return m["url"].(*URL)
}

func (m Metas) TLSConfig() *tls.Config {
	if v, ok := m["tls"]; ok {
		return v.(*tls.Config)
	}
	return nil
}

type BeforeHook struct {
	DialHook   func(ctx context.Context, addr string) context.Context
	AcceptHook func(ctx context.Context) context.Context
	ListenHook func(ctx context.Context, addr string) context.Context
}

type AfterHookFunc func(ctx context.Context, c net.Conn, addr string) (context.Context, net.Conn, error)

type AfterHook struct {
	Priority   uint
	DialerHook func(ctx context.Context, c net.Conn, addr string) (context.Context, net.Conn, error)
	AcceptHook func(ctx context.Context, c net.Conn) (context.Context, net.Conn, error)
	ListenHook func(ctx context.Context, listener net.Listener) (context.Context, net.Listener, error)
}

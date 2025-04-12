package core

import (
	"fmt"
	"github.com/chainreactors/rem/x/utils"
	"io"
	"net"
)

type Outbound interface {
	Name() string

	Handle(conn io.ReadWriteCloser, realConn net.Conn) (net.Conn, error)

	ToClash() *utils.Proxies
}

var outBoundCreators = make(map[string]OutboundCreatorFn)

// params has prefix "plugin_"
type OutboundCreatorFn func(options map[string]string, dial ContextDialer) (Outbound, error)

func OutboundRegister(name string, fn OutboundCreatorFn) {
	if _, exist := outBoundCreators[name]; exist {
		panic(fmt.Sprintf("plugin [%s] is already registered", name))
	}
	outBoundCreators[name] = fn
}

func OutboundCreate(name string, options map[string]string, dialer ContextDialer) (p Outbound, err error) {
	if fn, ok := outBoundCreators[name]; ok {
		p, err = fn(options, dialer)
	} else {
		err = fmt.Errorf("plugin [%s] is not registered", name)
	}
	return
}

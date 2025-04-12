package core

import (
	"fmt"
	"github.com/chainreactors/rem/x/utils"
	"io"
	"net"
)

type Inbound interface {
	Name() string
	Relay(conn net.Conn, bridge io.ReadWriteCloser) (net.Conn, error)
	ToClash() *utils.Proxies
}

var inboundCreators = make(map[string]InboundCreatorFn)

// params has prefix "relay_"
type InboundCreatorFn func(options map[string]string) (Inbound, error)

func InboundRegister(name string, fn InboundCreatorFn) {
	if _, exist := inboundCreators[name]; exist {
		panic(fmt.Sprintf("relay [%s] is already registered", name))
	}
	inboundCreators[name] = fn
}

func InboundCreate(name string, options map[string]string) (p Inbound, err error) {
	if fn, ok := inboundCreators[name]; ok {
		p, err = fn(options)
	} else {
		err = fmt.Errorf("relay [%s] is not registered", name)
	}
	return
}

func NewPluginOption(options map[string]string, mod, typ string) *PluginOption {
	options["type"] = typ
	return &PluginOption{
		options: options,
		Proxy:   utils.NewProxies(options),
	}
}

type PluginOption struct {
	Proxy   *utils.Proxies
	Mod     string
	options map[string]string
}

func (relay *PluginOption) String() string {
	return fmt.Sprintf("%s %s %d %s %s",
		relay.Proxy.Type,
		relay.Proxy.Server,
		relay.Proxy.Port,
		relay.Proxy.Username,
		relay.Proxy.Password)
}

func (relay *PluginOption) URL() string {
	if relay.Proxy.Username != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%d",
			relay.Proxy.Type,
			relay.Proxy.Username,
			relay.Proxy.Password,
			relay.Proxy.Server,
			relay.Proxy.Port,
		)
	} else {
		return fmt.Sprintf("%s://%s:%d",
			relay.Proxy.Type,
			relay.Proxy.Server,
			relay.Proxy.Port,
		)
	}
}

func (relay *PluginOption) ToClash() *utils.Proxies {
	return relay.Proxy
}

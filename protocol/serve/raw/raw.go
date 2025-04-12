package raw

import (
	"io"
	"net"

	"github.com/chainreactors/rem/protocol/core"
	"github.com/chainreactors/rem/x/utils"
)

func init() {
	core.InboundRegister(core.RawServe, NewRawInbound)
}

func NewRawInbound(params map[string]string) (core.Inbound, error) {
	utils.Log.Importantf("[agent.inbound] raw relay serving")
	return &RawPlugin{}, nil
}

type RawPlugin struct {
}

func (r *RawPlugin) Name() string {
	return core.RawServe
}

func (r *RawPlugin) ToClash() *utils.Proxies {
	return nil
}

func (r *RawPlugin) Relay(conn net.Conn, bridge io.ReadWriteCloser) (net.Conn, error) {
	return conn, nil
}

package message

import (
	"fmt"
	"github.com/chainreactors/rem/protocol/core"
)

func (c *Control) LocalURL() *core.URL {
	u, _ := core.NewURL(c.Local)
	return u
}

func (c *Control) RemoteURL() *core.URL {
	u, _ := core.NewURL(c.Remote)
	return u
}

//
//func (c *Control) Username() string {
//	if c.Plugin != nil {
//		return c.Plugin.Options["username"]
//	}
//	return ""
//}
//
//func (c *Control) Password() string {
//	if c.Plugin != nil {
//		return c.Plugin.Options["password"]
//	}
//	return ""
//}

func (l *Login) ConsoleURL() *core.URL {
	u, _ := core.NewURL(fmt.Sprintf("%s://%s:%d", l.ConsoleProto, l.ConsoleIP, l.ConsolePort))
	return u
}

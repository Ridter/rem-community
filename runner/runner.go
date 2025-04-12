package runner

import (
	"github.com/chainreactors/rem/protocol/core"
	"github.com/chainreactors/rem/x/utils"
	"net/url"
	"runtime/debug"
	"sync"
)

type RunnerConfig struct {
	*Options
	URLs        *core.URLs
	ConsoleURLs []*core.URL
	Proxies     []*url.URL
}

func (r *RunnerConfig) NewURLs(con *core.URL) *core.URLs {
	urls := r.URLs.Copy()
	urls.ConsoleURL = con
	return urls
}

func (r *RunnerConfig) Run() error {
	//recover
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()

	var wg sync.WaitGroup
	for _, cURL := range r.ConsoleURLs {
		wg.Add(1)
		console, err := NewConsole(r, r.NewURLs(cURL))
		if err != nil {
			utils.Log.Error("[console] " + err.Error())
			return err
		}
		if r.Mod == "bind" {
			go func() {
				err := console.Bind()
				if err != nil {
					utils.Log.Error(err.Error())
				}
				wg.Done()
			}()
		} else {
			go func() {
				err := console.Run()
				if err != nil {
					utils.Log.Error(err)
				}
				wg.Done()
			}()
		}

	}

	wg.Wait()
	return nil
}

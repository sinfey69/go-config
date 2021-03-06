package microcli

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/micro/cli"
	"github.com/micro/go-config/source"
	"strings"
	"time"
)

type clisrc struct {
	opts source.Options
	ctx  *cli.Context
}

func (c *clisrc) Read() (*source.ChangeSet, error) {
	var changes map[string]interface{}

	for _, name := range c.ctx.GlobalFlagNames() {
		tmp := toEntry(name, c.ctx.GlobalGeneric(name))
		mergo.Map(&changes, tmp) // need to sort error handling
	}

	for _, name := range c.ctx.FlagNames() {
		tmp := toEntry(name, c.ctx.Generic(name))
		mergo.Map(&changes, tmp) // need to sort error handling
	}

	b, err := json.Marshal(changes)
	if err != nil {
		return nil, err
	}

	h := md5.New()
	h.Write(b)
	checksum := fmt.Sprintf("%x", h.Sum(nil))

	return &source.ChangeSet{
		Data:      b,
		Checksum:  checksum,
		Timestamp: time.Now(),
		Source:    c.String(),
	}, nil
}

func toEntry(name string, v interface{}) map[string]interface{} {
	n := strings.ToLower(name)
	keys := strings.Split(n, "-")
	reverse(keys)
	tmp := make(map[string]interface{})
	for i, k := range keys {
		if i == 0 {
			tmp[k] = v
			continue
		}

		tmp = map[string]interface{}{k: tmp}
	}
	return tmp
}

func reverse(ss []string) {
	for i := len(ss)/2 - 1; i >= 0; i-- {
		opp := len(ss) - 1 - i
		ss[i], ss[opp] = ss[opp], ss[i]
	}
}
func (c *clisrc) Watch() (source.Watcher, error) {
	return source.NewNoopWatcher()
}

func (c *clisrc) String() string {
	return "microcli"
}

// NewSource returns a config source for integrating parsed flags from a micro/cli.Context.
// Hyphens are delimiters for nesting, and all keys are lowercased.
//
// Example:
//      cli.StringFlag{Name: "db-host"},
//
//
//      {
//          "database": {
//              "host": "localhost"
//          }
//      }
func NewSource(ctx *cli.Context, opts ...source.Option) source.Source {
	var options source.Options
	for _, o := range opts {
		o(&options)
	}

	return &clisrc{opts: options, ctx: ctx}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	// Packages
	consul "github.com/hashicorp/consul/api"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Url   string `yaml:"url"`
	Token string `yaml:"token"`
	Path  string `yaml:"path"`
}

type plugin struct {
	cfg    *consul.Config
	client *consul.Client
	path   string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	pathSeparator = string(os.PathSeparator)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)

	// Load configuration
	var cfg Config
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	}

	// Set config
	p.cfg = consul.DefaultConfig()
	p.cfg.Token = cfg.Token

	url, err := url.Parse(cfg.Url)
	if err != nil {
		provider.Print(ctx, "Error: ", err)
		return nil
	}
	p.cfg.Address = url.Host
	p.cfg.Scheme = url.Scheme

	if cred := url.User; cred != nil && cred.Username() != "" {
		password, _ := cred.Password()
		p.cfg.HttpAuth = &consul.HttpBasicAuth{
			Username: cred.Username(),
			Password: password,
		}
	}

	// Set path
	p.path = cfg.Path

	// Set client
	if client, err := consul.NewClient(p.cfg); err != nil {
		provider.Print(ctx, "Error: ", err)
		return nil
	} else {
		p.client = client
	}

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<consul"
	if p.cfg != nil {
		str += fmt.Sprintf(" scheme=%q", p.cfg.Scheme)
		str += fmt.Sprintf(" url=%q", p.cfg.Address)
	}
	if p.path != "" {
		str += fmt.Sprintf(" path=%q", p.path)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// USAGE

func Usage(w io.Writer) {
	fmt.Fprintln(w, "\n  Consul Key/Value store\n")
	fmt.Fprintln(w, "  Configuration:")
	fmt.Fprintln(w, "    url: <url>")
	fmt.Fprintln(w, "      URL of consul server. If user or password are provided")
	fmt.Fprintln(w, "      these are used as consul token details if no token parameter")
	fmt.Fprintln(w, "      is otherwise provided")
	fmt.Fprintln(w, "    token: <string>")
	fmt.Fprintln(w, "      Optional, Consul session token")
	fmt.Fprintln(w, "    path: <string>")
	fmt.Fprintln(w, "      Optional, Root path for key")
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "consul"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	<-ctx.Done()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - ENV

func (p *plugin) GetBytes(ctx context.Context, key string) ([]byte, error) {
	key = p.key(key)
	opts := consul.QueryOptions{}
	if pair, _, err := p.client.KV().Get(key, opts.WithContext(ctx)); err != nil {
		return nil, err
	} else if pair == nil {
		return nil, ErrNotFound.With(key)
	} else {
		return pair.Value, nil
	}
}

func (p *plugin) SetBytes(ctx context.Context, key string, value []byte) error {
	key = p.key(key)
	opts := consul.WriteOptions{}
	pair := consul.KVPair{
		Key:   key,
		Value: value,
	}
	if _, err := p.client.KV().Put(&pair, opts.WithContext(ctx)); err != nil {
		return err
	} else {
		return nil
	}
}

func (p *plugin) GetUint(ctx context.Context, key string) (uint64, error) {
	if data, err := p.GetBytes(ctx, key); err != nil {
		return 0, err
	} else if value, err := strconv.ParseUint(string(data), 0, 64); err != nil {
		return 0, err
	} else {
		return value, nil
	}
}

func (p *plugin) SetUint(ctx context.Context, key string, value uint64) error {
	return p.SetBytes(ctx, key, []byte(fmt.Sprint(value)))
}

func (p *plugin) GetString(ctx context.Context, key string) (string, error) {
	if data, err := p.GetBytes(ctx, key); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func (p *plugin) SetString(ctx context.Context, key, value string) error {
	return p.SetBytes(ctx, key, []byte(value))
}

func (p *plugin) GetDuration(ctx context.Context, key string) (time.Duration, error) {
	if data, err := p.GetBytes(ctx, key); err != nil {
		return 0, err
	} else if value, err := time.ParseDuration(string(data)); err != nil {
		return 0, err
	} else {
		return value, nil
	}
}

func (p *plugin) SetDuration(ctx context.Context, key string, value time.Duration) error {
	return p.SetBytes(ctx, key, []byte(value.String()))
}

func (p *plugin) GetJson(ctx context.Context, key string, value interface{}) error {
	if data, err := p.GetBytes(ctx, key); err != nil {
		return err
	} else if err := json.Unmarshal(data, value); err != nil {
		return err
	} else {
		return nil
	}
}

func (p *plugin) SetJson(ctx context.Context, key string, value interface{}) error {
	if data, err := json.Marshal(value); err != nil {
		return err
	} else {
		return p.SetBytes(ctx, key, data)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (p *plugin) key(k string) string {
	k = filepath.Join(p.path, strings.TrimPrefix(k, pathSeparator))
	return strings.TrimSuffix(k, pathSeparator)
}

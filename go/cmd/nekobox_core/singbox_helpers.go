package main

import (
	"context"
	"net"
	"net/http"
	"os"
	runtimeDebug "runtime/debug"
	"sync"
	"time"

	box "github.com/sagernet/sing-box"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/deprecated"
	"github.com/sagernet/sing-box/experimental/v2rayapi"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/service"
)

var singBoxStdLoggerOnce sync.Once

func newSingBoxContext() context.Context {
	// Keep it close to sing-box/cmd/sing-box initialization, but without CLI flags.
	singBoxStdLoggerOnce.Do(func() {
		// Disable ANSI colors globally (sing-box config can't set it).
		log.SetStdLogger(log.NewDefaultFactory(
			context.Background(),
			log.Formatter{BaseTime: time.Now(), DisableColors: true},
			os.Stderr,
			"",
			nil,
			false,
		).Logger())
	})
	ctx := context.Background()
	ctx = include.Context(service.ContextWith(ctx, deprecated.NewStderrManager(log.StdLogger())))
	return ctx
}

func createInstance(configContent string, statsOutbounds []string) (*box.Box, context.CancelFunc, *v2rayapi.StatsService, error) {
	ctx := newSingBoxContext()
	options, err := json.UnmarshalExtendedContext[option.Options](ctx, []byte(configContent))
	if err != nil {
		return nil, nil, nil, err
	}
	// GUI config can't set disable_color (json:"-"), so force it here.
	if options.Log == nil {
		options.Log = &option.LogOptions{}
	}
	options.Log.DisableColor = true

	ctx, cancel := context.WithCancel(ctx)
	instance, err := box.New(box.Options{
		Context: ctx,
		Options: options,
	})
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}

	var statsService *v2rayapi.StatsService
	if len(statsOutbounds) > 0 {
		statsService = v2rayapi.NewStatsService(option.V2RayStatsServiceOptions{
			Enabled:   true,
			Outbounds: statsOutbounds,
		})
		if statsService != nil {
			instance.Router().AppendTracker(statsService)
		}
	}

	connTracker := newConnTracker(instance.Outbound())
	instance.Router().AppendTracker(connTracker)
	instance_conn_tracker = connTracker

	err = instance.Start()
	if err != nil {
		cancel()
		instance.Close()
		return nil, nil, nil, err
	}

	runtimeDebug.FreeOSMemory()
	return instance, cancel, statsService, nil
}

func createProxyHTTPClient(instance *box.Box) *http.Client {
	if instance == nil {
		return &http.Client{}
	}
	// Prefer NekoBox/NekoRay convention: outbound tag "proxy" is the currently selected profile.
	// Test configs often omit route.final, which would make Default() fall back to direct.
	dialer := instance.Outbound().Default()
	if proxyOutbound, ok := instance.Outbound().Outbound("proxy"); ok {
		dialer = proxyOutbound
	}
	return &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2:   true,
			TLSHandshakeTimeout: C.TCPTimeout,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, M.ParseSocksaddr(addr))
			},
		},
	}
}

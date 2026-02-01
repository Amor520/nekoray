package main

import (
	"context"
	"net"
	"net/http"

	"github.com/matsuridayo/libneko/neko_common"
	"github.com/matsuridayo/libneko/neko_log"
	box "github.com/sagernet/sing-box"
	M "github.com/sagernet/sing/common/metadata"
)

var instance *box.Box
var instance_cancel context.CancelFunc

func setupCore() {
	//
	neko_log.SetupLog(50*1024, "./neko.log")
	//
	neko_common.GetCurrentInstance = func() interface{} {
		return instance
	}
	neko_common.DialContext = func(ctx context.Context, specifiedInstance interface{}, network, addr string) (net.Conn, error) {
		var b *box.Box
		if i, ok := specifiedInstance.(*box.Box); ok {
			b = i
		} else {
			b = instance
		}
		if b == nil {
			return neko_common.DialContextSystem(ctx, network, addr)
		}
		outbound := b.Outbound().Default()
		if proxyOutbound, ok := b.Outbound().Outbound("proxy"); ok {
			outbound = proxyOutbound
		}
		return outbound.DialContext(ctx, network, M.ParseSocksaddr(addr))
	}
	neko_common.DialUDP = func(ctx context.Context, specifiedInstance interface{}) (net.PacketConn, error) {
		var b *box.Box
		if i, ok := specifiedInstance.(*box.Box); ok {
			b = i
		} else {
			b = instance
		}
		if b == nil {
			return neko_common.DialUDPSystem(ctx)
		}
		outbound := b.Outbound().Default()
		if proxyOutbound, ok := b.Outbound().Outbound("proxy"); ok {
			outbound = proxyOutbound
		}
		return outbound.ListenPacket(ctx, M.Socksaddr{})
	}
	neko_common.CreateProxyHttpClient = func(specifiedInstance interface{}) *http.Client {
		if i, ok := specifiedInstance.(*box.Box); ok {
			return createProxyHTTPClient(i)
		}
		return createProxyHTTPClient(instance)
	}
}

package main

import (
    "context"
    "net"

    "github.com/sagernet/sing-box/adapter"
    "github.com/sagernet/sing-box/experimental/clashapi/trafficontrol"
    N "github.com/sagernet/sing/common/network"
)

type connTracker struct {
    manager  *trafficontrol.Manager
    outbound adapter.OutboundManager
}

var instance_conn_tracker *connTracker

func newConnTracker(outbound adapter.OutboundManager) *connTracker {
    return &connTracker{
        manager:  trafficontrol.NewManager(),
        outbound: outbound,
    }
}

func (t *connTracker) RoutedConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext, matchedRule adapter.Rule, matchOutbound adapter.Outbound) net.Conn {
    return trafficontrol.NewTCPTracker(conn, t.manager, metadata, t.outbound, matchedRule, matchOutbound)
}

func (t *connTracker) RoutedPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext, matchedRule adapter.Rule, matchOutbound adapter.Outbound) N.PacketConn {
    return trafficontrol.NewUDPTracker(conn, t.manager, metadata, t.outbound, matchedRule, matchOutbound)
}

func (t *connTracker) Manager() *trafficontrol.Manager {
    return t.manager
}

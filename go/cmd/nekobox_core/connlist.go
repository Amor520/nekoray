package main

import (
    "encoding/binary"
    "encoding/json"
    "sort"

    "github.com/gofrs/uuid/v5"
    C "github.com/sagernet/sing-box/constant"
    "github.com/sagernet/sing-box/experimental/clashapi/trafficontrol"
)

type connItem struct {
    ID    int    `json:"ID"`
    Start int64  `json:"Start"`
    End   int64  `json:"End"`
    Tag   string `json:"Tag"`
    Dest  string `json:"Dest"`
    RDest string `json:"RDest"`
}

func connIDFromUUID(id uuid.UUID) int {
    return int(binary.BigEndian.Uint32(id[0:4]) & 0x7fffffff)
}

func buildConnItem(meta trafficontrol.TrackerMetadata) connItem {
    start := meta.CreatedAt.Unix()
    end := int64(0)
    if !meta.ClosedAt.IsZero() {
        end = meta.ClosedAt.Unix()
    }
    dest := meta.Metadata.Destination.String()
    rdest := ""
    if meta.Metadata.Domain != "" && meta.Metadata.Domain != meta.Metadata.Destination.Fqdn {
        rdest = meta.Metadata.Domain
    }
    if rdest == "" && len(meta.Metadata.DestinationAddresses) > 0 {
        rdest = meta.Metadata.DestinationAddresses[0].String()
    }
    return connItem{
        ID:    connIDFromUUID(meta.ID),
        Start: start,
        End:   end,
        Tag:   meta.Outbound,
        Dest:  dest,
        RDest: rdest,
    }
}

func connectionsToJSON(manager *trafficontrol.Manager) string {
    if manager == nil {
        return "[]"
    }
    items := make([]connItem, 0)
    for _, meta := range manager.Connections() {
        if meta.OutboundType == C.TypeDNS {
            continue
        }
        items = append(items, buildConnItem(meta))
    }
    for _, meta := range manager.ClosedConnections() {
        if meta.OutboundType == C.TypeDNS {
            continue
        }
        items = append(items, buildConnItem(meta))
    }
    sort.Slice(items, func(i, j int) bool {
        if items[i].Start == items[j].Start {
            return items[i].ID > items[j].ID
        }
        return items[i].Start > items[j].Start
    })
    payload, err := json.Marshal(items)
    if err != nil {
        return "[]"
    }
    return string(payload)
}

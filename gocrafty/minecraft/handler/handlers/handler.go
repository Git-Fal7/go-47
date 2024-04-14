package handlers

import (
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol/packet"
	"github.com/Git-Fal7/go-47/gocrafty/player"
)

type Handler interface {
	PacketID() int32
	Handle(p *player.Player, packet packet.Packet)
}

package handshake

import (
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/types"
)

const (
	NextStateStatus = 1
	NextStateLogin  = 2
)

type Handshake struct {
	ProtocolVersion int32
	ServerAddress   string
	ServerPort      uint16
	NextState       int32
}

func (h *Handshake) ID() int32 {
	return IDHandshake
}

func (h *Handshake) State() int32 {
	return types.StateHandshaking
}

func (h *Handshake) Marshal(w *protocol.Writer) {
	w.VarInt(h.ProtocolVersion)
	w.String(h.ServerAddress)
	w.UShort(h.ServerPort)
	w.VarInt(h.NextState)
}

func (h *Handshake) Unmarshal(r *protocol.Reader) {
	r.VarInt(&h.ProtocolVersion)
	r.String(&h.ServerAddress)
	r.UShort(&h.ServerPort)
	r.VarInt(&h.NextState)
}

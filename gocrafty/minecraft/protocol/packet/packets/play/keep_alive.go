package play

import (
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/types"
)

type KeepAlive struct {
	KeepAliveId int32
}

func (k *KeepAlive) ID() int32 {
	return IDKeepAlive
}

func (k *KeepAlive) State() int32 {
	return types.StatePlay
}

func (k *KeepAlive) Marshal(w *protocol.Writer) {
	w.VarInt(k.KeepAliveId)
}

func (k *KeepAlive) Unmarshal(r *protocol.Reader) {
	r.VarInt(&k.KeepAliveId)
}

package login

import (
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/types"
)

type SetCompression struct {
	Threshold int32
}

func (d *SetCompression) ID() int32 {
	return IDSetCompression
}

func (s *SetCompression) State() int32 {
	return types.StateLogin
}

func (d *SetCompression) Marshal(w *protocol.Writer) {
	w.VarInt(d.Threshold)
}

func (d *SetCompression) Unmarshal(r *protocol.Reader) {}

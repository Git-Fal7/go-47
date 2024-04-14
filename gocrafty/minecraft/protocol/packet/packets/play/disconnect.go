package play

import (
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/types"
)

type Disconnect struct {
	Reason types.Chat
}

func (d *Disconnect) ID() int32 {
	return IDDisconnect
}

func (s *Disconnect) State() int32 {
	return types.StatePlay
}

func (d *Disconnect) Marshal(w *protocol.Writer) {
	w.Chat(d.Reason)
}

func (d *Disconnect) Unmarshal(r *protocol.Reader) {}

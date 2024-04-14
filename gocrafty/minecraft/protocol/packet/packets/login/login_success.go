package login

import (
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/types"
)

type LoginSuccess struct {
	UUID     string
	Username string
}

func (d *LoginSuccess) ID() int32 {
	return IDLoginSuccess
}

func (s *LoginSuccess) State() int32 {
	return types.StateLogin
}

func (d *LoginSuccess) Marshal(w *protocol.Writer) {
	w.String(d.UUID)
	w.String(d.Username)
}

func (d *LoginSuccess) Unmarshal(r *protocol.Reader) {}

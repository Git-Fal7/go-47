package login

import (
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/types"
)

type LoginStart struct {
	Username string
}

func (d *LoginStart) ID() int32 {
	return IDLoginStart
}

func (s *LoginStart) State() int32 {
	return types.StateLogin
}

func (d *LoginStart) Marshal(w *protocol.Writer) {}

func (d *LoginStart) Unmarshal(r *protocol.Reader) {
	r.String(&d.Username)
}

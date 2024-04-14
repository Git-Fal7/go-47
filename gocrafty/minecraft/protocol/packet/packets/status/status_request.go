package status

import (
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/types"
)

type StatusRequest struct{}

func (s *StatusRequest) ID() int32 {
	return IDStatusRequest
}

func (s *StatusRequest) State() int32 {
	return types.StateStatus
}

func (s *StatusRequest) Marshal(w *protocol.Writer) {}

func (s *StatusRequest) Unmarshal(r *protocol.Reader) {}

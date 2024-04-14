package status

import (
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol"
)

type PingResponse struct {
	PingTime int64
}

func (s *PingResponse) ID() int32 {
	return IDPing
}

func (s *PingResponse) State() int32 {
	return 0 // not used
}

func (s *PingResponse) Marshal(w *protocol.Writer) {
	w.Long(s.PingTime)
}

func (s *PingResponse) Unmarshal(r *protocol.Reader) {}

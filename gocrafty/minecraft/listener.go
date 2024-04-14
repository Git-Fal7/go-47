package minecraft

import (
	"errors"
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Git-Fal7/go-47/gocrafty/logger"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/handler"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol/packet"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol/packet/packets/handshake"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol/packet/packets/login"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol/packet/packets/status"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/socket"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/types"
	"github.com/Git-Fal7/go-47/gocrafty/player"
	"github.com/google/uuid"
)

type Listener struct {
	sync.RWMutex

	logger     logger.Logger
	name       string
	address    string
	maxPlayers int
	onlineMode bool

	PlayerCount atomic.Int32
	listener    net.Listener
	pool        packet.Pool

	pmu     sync.RWMutex
	players map[string]*player.Player

	incoming   chan *player.Player
	disconnect chan *player.Player

	close chan struct{}
}

func NewListener(incoming, disconnect chan *player.Player, logger logger.Logger, name, addr string, maxPlayers int, onlineMode bool) *Listener {
	return &Listener{
		logger:     logger,
		name:       name,
		address:    addr,
		maxPlayers: maxPlayers,
		onlineMode: onlineMode,

		pool: packet.NewPool(),

		players: make(map[string]*player.Player),

		incoming:   incoming,
		disconnect: disconnect,

		close: make(chan struct{}),
	}
}

func (l *Listener) Listen() (*Listener, error) {
	listener, err := net.Listen("tcp", l.address)

	if err != nil {
		return nil, errors.New("failed to listen on address " + l.address + ": " + err.Error())
	}

	l.listener = listener

	go l.listen()

	return l, nil
}

func (l *Listener) listen() {
	go func() {
		ticker := time.NewTicker(time.Second * 4)
		defer ticker.Stop()

		for {
			select {
			case <-l.close:
				return
			}
		}
	}()

	defer func() {
		// TODO: incoming channel
		close(l.close)
		_ = l.Close()
	}()

	for {
		netConn, err := l.listener.Accept()
		if err != nil {
			return
		}

		l.createConn(netConn)
	}
}

func (l *Listener) createConn(netConn net.Conn) {
	conn := socket.NewConn(netConn)

	go l.handleConn(conn)
}

func (l *Listener) handleConn(conn *socket.Conn) {
	defer func() {
		_ = conn.Close("")

		if r := recover(); r != nil {
			l.logger.Debugf("panic in connection handler: %v\n", string(debug.Stack()))
		}
	}()

	for {
		p, err := conn.ReadPacket()

		if err != nil {
			if err != nil {
				if p, ok := l.players[conn.UUID().String()]; ok {
					l.logger.Infof("Player %s disconnected", conn.Username())

					l.PlayerCount.Add(-1)
					l.RemovePlayer(p)
				}

				if !errors.Is(err, io.EOF) {
					l.logger.Errorf("Got an error receving packet: %v", err)
				}

				break
			}
		}

		if p == nil {
			//l.logger.Errorf("Got an unknown packet with ID: %v", p)
			continue
		}

		if conn.State == types.StateDisconnect {
			break
		}

		// packet handling
		if conn.State == types.StatePlay {
			pl, ok := l.players[conn.UUID().String()]
			if !ok {
				l.logger.Errorf("Got a packet from a player that is not in the player list")
				break
			}

			handler, ok := handler.GetHandler(p.ID())
			if !ok {
				l.logger.Errorf("Got a packet but no handler for it: %v", p.ID())
				continue
			}

			handler.Handle(pl, p)
		} else {
			l.handlePacket(conn, p)
		}

		select {
		case <-l.close:
			return

		default:
			// loop
		}
	}
}

func (l *Listener) handlePacket(c *socket.Conn, p packet.Packet) {
	switch c.State {

	// Handshaking
	case types.StateHandshaking:
		switch dp := p.(type) {
		case *handshake.Handshake:
			switch dp.NextState {
			case 1:
				c.State = types.StateStatus

			case 2:
				if dp.ProtocolVersion == ProtocolVersion {
					if l.PlayerCount.Load() >= int32(l.MaxPlayers()) {
						c.WritePacket(&login.Disconnect{
							Reason: types.Chat{
								Text:  "Server is full",
								Bold:  true,
								Color: "red",
							},
						})

						c.State = types.StateDisconnect

						return
					}

					l.PlayerCount.Add(1)

					c.State = types.StateLogin
				} else {
					c.WritePacket(&login.Disconnect{
						Reason: types.Chat{
							Text:  fmt.Sprintf("Outdated server! I'm still on %s", Version),
							Bold:  true,
							Color: "red",
						},
					})

					c.State = types.StateDisconnect
				}
			}
		}

	// Status
	case types.StateStatus:
		switch dp := p.(type) {
		case *status.StatusRequest:
			c.WritePacket(&status.StatusResponse{
				JSONResponse: &status.StatusResponseData{
					Version: &status.StatusResponseDataVersion{
						Name:     Version,
						Protocol: ProtocolVersion,
					},
					Players: &status.StatusResponseDataPlayers{
						Max:    int32(l.MaxPlayers()),
						Online: l.PlayerCount.Load(),
					},
					Description: &status.StatusResponseDataDescription{
						Text: l.Name(),
					},
				},
			})

		case *status.PingRequest:
			c.WritePacket(&status.PingResponse{
				PingTime: dp.PingTime,
			})
		}

	// Login
	case types.StateLogin:
		switch dp := p.(type) {
		case *login.LoginStart:
			l.Logger().Infof("Player %s connected", dp.Username)
			c.State = types.StatePlay

			// TODO: Verify player online mode

			// TODO: Support compression
			c.WritePacket(&login.SetCompression{
				Threshold: -1,
			})

			if l.OnlineMode() {
				// TODO: Send encryption request
			} else {
				uuid := uuid.New()
				uuid.UnmarshalText([]byte("00000000-0000-0000-0000-000000000000"))

				c.SetUsername(dp.Username)
				c.SetUUID(uuid)

				c.WritePacket(&login.LoginSuccess{
					UUID:     "00000000-0000-0000-0000-000000000000",
					Username: dp.Username,
				})

				l.AddPlayer(c)
			}
		}
	}
}

func (l *Listener) AddPlayer(c *socket.Conn) {
	l.pmu.Lock()
	defer l.pmu.Unlock()

	p := player.New(l.logger, c.Username(), c.UUID(), c)

	if _, ok := l.players[c.UUID().String()]; ok {
		p.Disconnect("You are already logged in!")

		return
	}

	l.players[c.UUID().String()] = p

	go func() {
		l.incoming <- p
	}()

	go p.Conn().KeepAliveTicker()

	p.JoinGame()
}

func (l *Listener) RemovePlayer(p *player.Player) {
	l.pmu.Lock()
	defer l.pmu.Unlock()

	delete(l.players, p.UUID().String())
}

func (l *Listener) Close() error {
	return l.listener.Close()
}

func (l *Listener) Pool() packet.Pool {
	l.RLock()
	defer l.RUnlock()

	return l.pool
}

func (l *Listener) Name() string {
	l.RLock()
	defer l.RUnlock()

	return l.name
}

func (l *Listener) Address() string {
	l.RLock()
	defer l.RUnlock()

	return l.address
}

func (l *Listener) MaxPlayers() int {
	l.RLock()
	defer l.RUnlock()

	return l.maxPlayers
}

func (l *Listener) OnlineMode() bool {
	l.RLock()
	defer l.RUnlock()

	return l.onlineMode
}

func (l *Listener) Logger() logger.Logger {
	l.RLock()
	defer l.RUnlock()

	return l.logger
}

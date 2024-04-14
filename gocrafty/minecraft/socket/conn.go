package socket

import (
	"bufio"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol/packet"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/protocol/packet/packets/play"
	"github.com/Git-Fal7/go-47/gocrafty/minecraft/types"
	"github.com/google/uuid"
)

type Conn struct {
	net.Conn

	sendMutex sync.Mutex
	conn      net.Conn

	username string
	uuid     uuid.UUID

	reader *bufio.Reader

	// The buffer used for recving data
	recvBuffer   [65565]byte
	decompBuffer [65565]byte // TODO: implement compression/decompression and give sense to this buffer cuz it's useless rn xd

	State int32
	close chan bool
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		Conn: conn,

		conn:   conn,
		reader: bufio.NewReader(conn),
		State:  types.StateHandshaking,

		close: make(chan bool),
	}
}

func (c *Conn) WritePacket(p packet.Packet) error {
	packetWriter := &protocol.Writer{}
	packetWriter.VarInt(p.ID())
	p.Marshal(packetWriter)

	payload := packetWriter.Bytes()

	sendPacket := &protocol.Writer{}
	sendPacket.VarInt(int32(len(payload)))
	sendPacket.WriteBytes(payload)

	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	_, err := c.Write(sendPacket.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) ReadPacket() (packet.Packet, error) {
	packetLength, err := c.varInt()
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(c.reader, c.recvBuffer[:packetLength])
	if err != nil {
		return nil, err
	}

	reader := protocol.Reader{Data: c.recvBuffer[:packetLength]}
	pool := packet.NewPool()

	var packetId int32
	reader.VarInt(&packetId)

	p, ok := pool.Get(c.State, packetId)
	if !ok {
		return nil, err
	}

	p.Unmarshal(&reader)

	return p, nil
}

func (c *Conn) KeepAliveTicker() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.WritePacket(&play.KeepAlive{
				KeepAliveId: int32(time.Now().Unix()),
			})

		case <-c.close:
			return
		}
	}
}

func (c *Conn) Username() string {
	return c.username
}

func (c *Conn) SetUsername(username string) {
	c.username = username
}

func (c *Conn) UUID() uuid.UUID {
	return c.uuid
}

func (c *Conn) SetUUID(uuid uuid.UUID) {
	c.uuid = uuid
}

func (c *Conn) Close(reason string) error {
	c.WritePacket(&play.Disconnect{
		Reason: types.Chat{
			Text:  reason,
			Color: "red",
		},
	})

	c.State = types.StateDisconnect

	close(c.close)

	return c.conn.Close()
}

func (c *Conn) varInt() (int, error) {
	numRead := 0
	result := 0
	for {
		read, err := c.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		value := read & 0b01111111
		result |= int(value) << (7 * numRead)

		numRead++
		if numRead > 5 {
			return 0, errors.New("varint is too big")
		}

		if (read & 0b10000000) == 0 {
			return result, nil
		}
	}
}

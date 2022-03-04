package network

import (
	"github.com/OCharnyshevich/proxycraft/proxy/helper"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	mcNet "github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/google/uuid"
)

// handlerHeap is PriorityQueue<PacketHandlerFunc>
type handlerHeap []PacketHandler

func (h handlerHeap) Len() int            { return len(h) }
func (h handlerHeap) Less(i, j int) bool  { return h[i].Priority < h[j].Priority }
func (h handlerHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *handlerHeap) Push(x interface{}) { *h = append(*h, x.(PacketHandler)) }
func (h *handlerHeap) Pop() interface{} {
	old := *h
	n := len(old)
	*h = old[0 : n-1]
	return old[n-1]
}

type Events struct {
	generic  *handlerHeap           // for every packet
	handlers map[int32]*handlerHeap // for specific packet id only
}

func NewEvents() Events {
	return Events{
		handlers: make(map[int32]*handlerHeap),
	}
}

func (e *Events) AddListener(listeners ...PacketHandler) {
	for _, l := range listeners {
		var s *handlerHeap
		var ok bool
		if s, ok = e.handlers[l.ID]; !ok {
			s = &handlerHeap{l}
			e.handlers[l.ID] = s
		} else {
			s.Push(l)
		}
	}
}

// AddGeneric adds listeners like AddListener, but the packet ID is ignored.
// Generic listener is always called before specific packet listener.
func (e *Events) AddGeneric(listeners ...PacketHandler) {
	for _, l := range listeners {
		if e.generic == nil {
			e.generic = &handlerHeap{l}
		} else {
			e.generic.Push(l)
		}
	}
}

type PacketHandlerFunc func(client *mcNet.Conn, server *mcNet.Conn, p pk.Packet) error
type PacketHandler struct {
	ID       int32
	Priority int
	F        func(client *mcNet.Conn, server *mcNet.Conn, p pk.Packet) error
}

type EventsListener struct {
	GameStart      func() error
	ChatMsg        func(c chat.Message, pos byte, uuid uuid.UUID) error
	KickDisconnect func(reason chat.Message) error
	HealthChange   func(health float32) error
	Death          func() error
}

func (e EventsListener) Attach(n helper.Network) {
	(n.Events().(*Events)).AddListener(
		PacketHandler{Priority: 64, ID: packetid.Login, F: e.onJoinGame},
		PacketHandler{Priority: 64, ID: packetid.ChatClientbound, F: e.onChatMsg},
		PacketHandler{Priority: 64, ID: packetid.KickDisconnect, F: e.onKickDisconnect},
		PacketHandler{Priority: 64, ID: packetid.UpdateHealth, F: e.onUpdateHealth},
	)
}

func (e *EventsListener) onJoinGame(_ *mcNet.Conn, _ *mcNet.Conn, _ pk.Packet) error {
	if e.GameStart != nil {
		return e.GameStart()
	}
	return nil
}

func (e *EventsListener) onKickDisconnect(_ *mcNet.Conn, _ *mcNet.Conn, p pk.Packet) error {
	if e.KickDisconnect != nil {
		var reason chat.Message
		if err := p.Scan(&reason); err != nil {
			return PacketHandlerError{ID: p.ID, Err: err}
		}
		return e.KickDisconnect(reason)
	}
	return nil
}

func (e *EventsListener) onChatMsg(_ *mcNet.Conn, _ *mcNet.Conn, p pk.Packet) error {
	if e.ChatMsg != nil {
		var msg chat.Message
		var pos pk.Byte
		var sender pk.UUID

		if err := p.Scan(&msg, &pos, &sender); err != nil {
			return PacketHandlerError{ID: p.ID, Err: err}
		}

		return e.ChatMsg(msg, byte(pos), uuid.UUID(sender))
	}
	return nil
}

func (e *EventsListener) onUpdateHealth(_ *mcNet.Conn, _ *mcNet.Conn, p pk.Packet) error {
	if e.ChatMsg != nil {
		var health pk.Float
		var food pk.VarInt
		var foodSaturation pk.Float

		if err := p.Scan(&health, &food, &foodSaturation); err != nil {
			return PacketHandlerError{ID: p.ID, Err: err}
		}
		if e.HealthChange != nil {
			if err := e.HealthChange(float32(health)); err != nil {
				return err
			}
		}
		if e.Death != nil && health <= 0 {
			if err := e.Death(); err != nil {
				return err
			}
		}
	}
	return nil
}

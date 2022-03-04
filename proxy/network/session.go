package network

import (
	"errors"
	"fmt"
	"github.com/OCharnyshevich/proxycraft/proxy/helper"
	"github.com/OCharnyshevich/proxycraft/proxy/log"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	mcNet "github.com/Tnze/go-mc/net"
	mcPkt "github.com/Tnze/go-mc/net/packet"
	"io"
	"strconv"
	"time"
)

type State int

const (
	Status State = 1
	Login  State = 2
	Play   State = 3
)

type session struct {
	logger    *log.Logging
	startTime time.Time
	client    *mcNet.Conn
	server    *mcNet.Conn
	state     State
	events    *Events
}

func NewSession(localConn *mcNet.Listener, remoteHost string, remotePort int, events *Events) (sess session, err error) {
	sess.logger = log.New("session", log.EveryLevel...)
	sess.state = Status
	sess.events = events

	client, err := localConn.Accept()
	if err != nil {
		return sess, err
	}
	//client.SetThreshold(256)

	server, err := mcNet.DialMC(remoteHost + ":" + strconv.Itoa(remotePort))
	if err != nil {
		_ = client.Close()
		return sess, err
	}

	//server.SetThreshold(256)

	sess.client = &client
	sess.startTime = time.Now().UTC()

	sess.logger.InfoF("Accepted connection from %s", client.Socket.RemoteAddr().String())
	sess.logger.InfoF("Connected to backend on %s", server.Socket.RemoteAddr().String())
	sess.server = server

	return sess, err
}

func (s *session) Kill() {
	_ = s.client.Close()
	_ = s.server.Close()
}

func (s *session) StreamBidirectional() {
	errs := make(chan error, 2)
	closer := make(chan interface{}, 2)
	go s.ClientToServer(errs, closer)
	go s.ServerToClient(errs, closer)

	<-errs
	closer <- struct{}{}
	closer <- struct{}{}
	s.Kill()
}

func (s *session) ClientToServer(errs chan error, closer chan interface{}) {
	for {
		select {
		case <-closer:
			return
		default:
			var packet mcPkt.Packet
			err := s.client.ReadPacket(&packet)
			if err != nil {
				if errors.Is(err, io.EOF) {
					errs <- err
					break
				}
				s.logger.WarnF("Unable to read packet from client: %v", err)
				continue
			}

			if err := s.server.WritePacket(packet); err != nil {
				if errors.Is(err, io.EOF) {
					errs <- err
					break
				}
				s.logger.WarnF("Unable to send packet to server: %v", err)
			}
		}
	}
}

func (s *session) ServerToClient(errs chan error, closer chan interface{}) {
	for {
		select {
		case <-closer:
			return
		default:
			var packet mcPkt.Packet
			err := s.server.ReadPacket(&packet)
			if err != nil {
				if errors.Is(err, io.EOF) {
					errs <- err
					break
				}
				s.logger.WarnF("Unable to read packet from server: %v", err)
				continue
			}

			if err := s.handleServerbound(packet); err != nil {
				s.logger.WarnF("PacketHandlerError: %v", err)
			}

			if packet.ID == packetid.UpdateTime {
				continue
			}

			if err := s.client.WritePacket(packet); err != nil {
				if errors.Is(err, io.EOF) {
					errs <- err
					break
				}
				s.logger.WarnF("Unable to send packet to client: %v", err)
			}
		}
	}
}

func (s *session) handleServerbound(packet mcPkt.Packet) (err error) {
	if s.events.generic != nil {
		for _, handler := range *s.events.generic {
			if err = handler.F(s.client, s.server, packet); err != nil {
				return PacketHandlerError{ID: packet.ID, Err: err}
			}
		}
	}
	if listeners := s.events.handlers[packet.ID]; listeners != nil {
		for _, handler := range *listeners {
			err = handler.F(s.client, s.server, packet)
			if err != nil {
				return PacketHandlerError{ID: packet.ID, Err: err}
			}
		}
	}

	return nil
}

func (s *session) SendMessage(message ...interface{}) {
	err := s.client.WritePacket(mcPkt.Marshal(
		packetid.ChatClientbound,
		chat.Text(helper.ConvertToString(message)), mcPkt.Byte(2),
		mcPkt.UUID{},
	))
	if err != nil {
		s.logger.Fail(err)
	}

	err = s.client.WritePacket(mcPkt.Marshal(
		packetid.UpdateTime,
		mcPkt.Long(275690), mcPkt.Long(1019),
	))
	if err != nil {
		s.logger.Fail(err)
	}
}

type PacketHandlerError struct {
	ID  int32
	Err error
}

func (d PacketHandlerError) Error() string {
	return fmt.Sprintf("handle packet 0x%X error: %v", d.ID, d.Err)
}

func (d PacketHandlerError) Unwrap() error {
	return d.Err
}

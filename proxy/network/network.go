package network

import (
	"fmt"
	"github.com/OCharnyshevich/proxycraft/proxy/helper"
	"github.com/OCharnyshevich/proxycraft/proxy/log"
	mcNet "github.com/Tnze/go-mc/net"
	"strconv"
)

type network struct {
	localHost string
	localPort int

	remoteHost string
	remotePort int

	logger *log.Logging
	//packets base.Packets

	//join chan base.PlayerAndConnection
	//quit chan base.PlayerAndConnection

	localConn *mcNet.Listener
	sessions  []helper.Sessionable
	events    Events

	report chan helper.Message
}

func New(report chan helper.Message, lHost string, lPort int, rHost string, rPort int, events Events) helper.Network {
	return &network{
		localHost:  lHost,
		localPort:  lPort,
		remoteHost: rHost,
		remotePort: rPort,

		report: report,
		logger: log.New("network", log.EveryLevel...),
		events: events,
	}
}

func (n *network) Load() {
	if err := n.startListening(); err != nil {
		n.report <- helper.Make(helper.FAIL, err)
		return
	}
}

func (n *network) Kill() {
	for _, sess := range n.sessions {
		sess.Kill()
	}
}

func (n *network) Events() interface{} {
	return &n.events
}

func (n *network) Sessions() []helper.Sessionable {
	return n.sessions
}

func (n *network) startListening() error {
	localConn, err := mcNet.ListenMC(n.localHost + ":" + strconv.Itoa(n.localPort))
	if err != nil {
		return fmt.Errorf("failed to bind [%v]", err)
	}

	n.localConn = localConn

	n.logger.InfoF("listening on %s:%d", n.localHost, n.localPort)

	go func() {
		for {
			session, err := NewSession(n.localConn, n.remoteHost, n.remotePort, &n.events)
			if err != nil {
				//n.report <- helper.Make(helper.FAIL, err)
				n.logger.Warn(err)
				continue
			}
			n.sessions = append(n.sessions, &session)
			go session.StreamBidirectional()
		}
	}()

	return nil
}

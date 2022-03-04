package proxy

import (
	"github.com/OCharnyshevich/proxycraft/proxy/console"
	"github.com/OCharnyshevich/proxycraft/proxy/helper"
	"github.com/OCharnyshevich/proxycraft/proxy/log"
	"github.com/OCharnyshevich/proxycraft/proxy/network"
	"github.com/OCharnyshevich/proxycraft/proxy/player/chat"
)

type proxy struct {
	message chan helper.Message
	console *console.Console
	logging *log.Logging
	network helper.Network
}

func New(config *Config) (*proxy, error) {
	message := make(chan helper.Message)
	c := console.New(message)
	l := log.New("proxy", log.EveryLevel...)
	e := network.NewEvents()

	n := network.New(
		message,
		config.Local.Host, config.Local.Port,
		config.Remote.Host, config.Remote.Port,
		e,
	)

	return &proxy{
		message: message,
		console: c,
		logging: l,
		network: n,
	}, nil
}

func (p *proxy) Load() {
	p.console.Load()
	p.network.Load()

	p.wait()
}

func (p *proxy) Kill() {
	p.console.Kill()
	p.network.Kill()

	// push the stop message to the server exit channel
	p.message <- helper.Make(helper.STOP, "normal stop")
	close(p.message)

	p.logging.Info(chat.DarkRed, "server stopped")
}

func (p *proxy) wait() {
	// select over server commands channel
	select {
	case command := <-p.message:
		switch command.Command {
		// stop selecting when stop is received
		case helper.STOP:
			return
		case helper.FAIL:
			p.logging.Fail("internal server error: ", command.Message)
			p.logging.Fail("stopping server")
			return
		}
	}

	p.wait()
}

func (p *proxy) Logging() *log.Logging {
	return p.logging
}

func (p *proxy) Network() helper.Network {
	return p.network
}

func (p *proxy) Broadcast(message string) {
	//p.console.SendMessage(message)

	for _, session := range p.Network().Sessions() {
		session.SendMessage(message)
	}
}

package console

import (
	"bufio"
	"github.com/OCharnyshevich/proxycraft/proxy/helper"
	"github.com/OCharnyshevich/proxycraft/proxy/log"
	"io"
	"os"
)

type Console struct {
	i io.Reader
	o io.Writer

	logger *log.Logging

	IChannel chan string
	OChannel chan string

	report chan helper.Message
}

func New(report chan helper.Message) *Console {
	console := &Console{
		IChannel: make(chan string),
		OChannel: make(chan string),

		report: report,
	}

	console.i = io.MultiReader(os.Stdin)
	console.o = io.MultiWriter(os.Stdout)

	console.logger = log.NewWith("console", console.o, log.EveryLevel...)

	return console
}

func (c *Console) Load() {
	// handle i channel
	go func() {
		scanner := bufio.NewScanner(c.i)

		for scanner.Scan() {
			err := helper.Attempt(func() {
				c.IChannel <- scanner.Text()
			})

			if err != nil {
				c.report <- helper.Make(helper.FAIL, err)
			}
		}
	}()

	// handle o channel
	go func() {
		for line := range c.OChannel {
			c.logger.Info(line)
		}
	}()
}

func (c *Console) Kill() {
	defer func() {
		_ = recover() // ignore panic with closing closed channel
	}()

	// save the log file as YYYY-MM-DD-{index}.log{.gz optionally compressed}

	close(c.IChannel)
	close(c.OChannel)
}

func (c *Console) Name() string {
	return "ConsoleSender"
}

func (c *Console) SendMessage(message ...interface{}) {
	defer func() {
		if err := recover(); err != nil {
			c.report <- helper.Make(helper.FAIL, err)
		}
	}()

	c.OChannel <- helper.ConvertToString(message...)
}

package log

import (
	"fmt"
	"github.com/OCharnyshevich/proxycraft/proxy/helper"
	"github.com/OCharnyshevich/proxycraft/proxy/player/chat"
	"github.com/fatih/color"
	"io"
	"os"
	"time"
)

type LogLevel int

const (
	Info LogLevel = iota
	Warn
	Fail
	Data
)

var BasicLevel = []LogLevel{Info, Warn, Fail}
var EveryLevel = []LogLevel{Info, Warn, Fail, Data}

type Logging struct {
	name   string
	writer io.Writer
	show   []LogLevel
}

func (log *Logging) Name() string {
	return log.name
}

func (log *Logging) Show() []LogLevel {
	return log.show
}

func (log *Logging) formatPrint(level, message string) {
	_, _ = fmt.Fprint(log.writer, fmt.Sprintf("[%s] [%s] [%s] %s\n", color.HiGreenString(currentTimeAsText()), level, color.WhiteString(log.Name()), chat.TranslateConsole(message)))
}

func (log *Logging) info(message string) {
	log.formatPrint(color.CyanString("INFO"), message)
}

func (log *Logging) warn(message string) {
	log.formatPrint(color.YellowString("WARN"), message)
}

func (log *Logging) fail(message string) {
	log.formatPrint(color.RedString("FAIL"), message)
}

func (log *Logging) data(message string) {
	log.formatPrint(color.MagentaString("DATA"), message)
}

func (log *Logging) Info(message ...interface{}) {
	if !checkIfLevelShows(log, Info) {
		return
	}

	log.info(helper.ConvertToString(message...))
}

func (log *Logging) Warn(message ...interface{}) {
	if !checkIfLevelShows(log, Warn) {
		return
	}

	log.warn(helper.ConvertToString(message...))
}

func (log *Logging) Fail(message ...interface{}) {
	if !checkIfLevelShows(log, Fail) {
		return
	}

	log.fail(helper.ConvertToString(message...))
}

func (log *Logging) Data(message ...interface{}) {
	if !checkIfLevelShows(log, Data) {
		return
	}

	log.data(helper.ConvertToString(message...))
}

func (log *Logging) InfoF(format string, a ...interface{}) {
	if !checkIfLevelShows(log, Info) {
		return
	}

	log.info(fmt.Sprintf(format, a...))
}

func (log *Logging) WarnF(format string, a ...interface{}) {
	if !checkIfLevelShows(log, Warn) {
		return
	}

	log.warn(fmt.Sprintf(format, a...))
}

func (log *Logging) FailF(format string, a ...interface{}) {
	if !checkIfLevelShows(log, Fail) {
		return
	}

	log.fail(fmt.Sprintf(format, a...))
}

func (log *Logging) DataF(format string, a ...interface{}) {
	if !checkIfLevelShows(log, Data) {
		return
	}

	log.data(fmt.Sprintf(format, a...))
}

func New(name string, show ...LogLevel) *Logging {
	return NewWith(name, os.Stdout, show...)
}

func NewWith(name string, writer io.Writer, show ...LogLevel) *Logging {
	return &Logging{name: name, writer: writer, show: show}
}

func currentTimeAsText() string {
	h, m, s := time.Now().Clock()
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func checkIfLevelShows(log *Logging, lvl LogLevel) bool {
	for _, a := range log.Show() {
		if a == lvl {
			return true
		}
	}
	return false
}

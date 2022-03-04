package helper

import (
	"fmt"
	"strings"
)

type Loads interface {
	Load()
}

type Kills interface {
	Kill()
}

type State interface {
	Loads
	Kills
}

type Network interface {
	State
	Events() interface{}
	Sessions() []Sessionable
}

type Sessionable interface {
	Kills
	SendMessage(message ...interface{})
}

func ConvertToString(data ...interface{}) string {
	strs := make([]string, len(data))

	for i, str := range data {
		strs[i] = fmt.Sprintf("%v", str)
	}

	return strings.Join(strs, "")
}

func Attempt(function func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("caught: %v", r)
		}
	}()

	function()

	return
}

package transport

import (
	"github.com/treepeck/justchess/pkg/proto"
)

type Topic struct {
	Register   chan string
	Unregister chan string
	In         chan proto.Message
	Out        chan proto.Message
}

func NewTopic() Topic {
	return Topic{
		Register:   chan string,
		Unregister: chan string,
		In:         chan proto.Message,
		Out:        chan proto.Message,
	}
}

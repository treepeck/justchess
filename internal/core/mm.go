package core

import (
	"justchess/internal/randgen"
	"log"
)

type matchmaking struct {
	join     chan joinMatchmakingReq
	leave    chan string
	response chan<- addRoomRes
	pool     map[string]roomParams
}

func newMatchmaking(response chan<- addRoomRes) matchmaking {
	return matchmaking{
		join:     make(chan joinMatchmakingReq),
		leave:    make(chan string),
		response: response,
		pool:     make(map[string]roomParams),
	}
}

func (mm matchmaking) handleEvents() {
	for {
		select {
		case req := <-mm.join:
			mm.search(req.playerId, req.params)

		case playerId := <-mm.leave:
			if _, exists := mm.pool[playerId]; exists {
				delete(mm.pool, playerId)
				log.Printf("player %s leaved matchmaking", playerId)
			}
		}
	}
}

func (mm matchmaking) search(playerId string, params roomParams) {
	var similar string

	for playerId, p := range mm.pool {
		if p == params {
			similar = playerId
			break
		}
	}

	switch similar {
	case "":
		mm.pool[playerId] = params

	default:
		delete(mm.pool, similar)

		mm.response <- addRoomRes{
			players:     [2]string{similar, playerId},
			timeControl: params.TimeControl,
			timeBonus:   params.TimeBonus,
			roomId:      randgen.GenId(randgen.IdLen),
		}
	}
}

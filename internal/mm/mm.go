/*
Package mm implements matchmaking system.
*/
package mm

import "log"

/*
WaitRoom represents a single pending game, which has not been started yet.
TODO: add rating constraints.
*/
type WaitRoom struct {
	CreatorId   string `json:"-"`
	TimeControl int    `json:"tc"`
	TimeBonus   int    `json:"tb"`
}

/*
Matchmaking's methods are not safe for concurrent use.
*/
type Matchmaking struct {
	pool map[string]WaitRoom
}

func NewMatchmaking() *Matchmaking {
	return &Matchmaking{
		pool: make(map[string]WaitRoom),
	}
}

/*
Add adds a [WaitRoom] to the [Matchmaking] pool.
*/
func (m *Matchmaking) Add(id string, r WaitRoom) {
	m.pool[id] = r

	log.Printf("wait room \"%s\" added", id)
}

/*
Remove removes a [WaitRoom] from the [Matchmaking] pool.
*/
func (m *Matchmaking) Remove(id string) {
	delete(m.pool, id)

	log.Printf("wait room \"%s\" removed", id)
}

/*
FindByCreatorId searches for a [WaitRoom] with the specified creator id.  It
returns the room id if found, or an empty string if no such room exists.
*/
func (m *Matchmaking) FindByCreatorId(cid string) string {
	for id, r := range m.pool {
		if r.CreatorId == cid {
			return id
		}
	}
	return ""
}

/*
Pair searches for a [WaitRoom] with the specified parameters.  If a
matching room is found, it returns the room id to create an instant game.  The
empty string is returned in case the room does not exist.
*/
func (m *Matchmaking) Pair(timeControl int, timeBonus int) string {
	for id, r := range m.pool {
		if r.TimeControl == timeControl && r.TimeBonus == timeBonus {
			return id
		}
	}
	return ""
}

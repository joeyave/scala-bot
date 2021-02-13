package entities

import (
	"scalaChordsBot/configs"
)

type User struct {
	ID     int64   `bson:"_id"`
	States []State `bson:"states"`
}

type State struct {
	Index   int     `bson:"index"`
	Name    string  `bson:"name"`
	Context Context `bson:"context"`
}

type Context struct {
	CurrentSong Song   `bson:"currentSong"`
	Songs       []Song `bson:"songs"`
}

func (u *User) CurrentState() *State {
	if u.States != nil && len(u.States) > 0 {
		return &u.States[len(u.States)-1]
	} else {
		return &State{
			Index: 0,
			Name:  configs.SongSearchState,
		}
	}
}

func (u *User) AppendState(name string, context Context) *State {
	state := State{
		Index:   0,
		Name:    name,
		Context: context,
	}
	u.States = append(u.States, state)
	return &state
}

func (s *State) NextIndex() {
	s.Index++
}

func (s *State) PrevIndex() {
	s.Index--
}

func (s *State) ChangeTo(name string) {
	s.Index = 0
	s.Name = name
	s.Context = Context{}
}

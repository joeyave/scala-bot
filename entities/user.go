package entities

type User struct {
	ID    int64  `bson:"_id"`
	State *State `bson:"states"`
}

type State struct {
	Index   int     `bson:"index"`
	Name    string  `bson:"name"`
	Context Context `bson:"context"`

	Prev *State `bson:"prev"`
	Next *State `bson:"next"`
}

type Context struct {
	CurrentSong      *Song    `bson:"currentSong"`
	Songs            []Song   `bson:"songs"`
	Setlist          []string `bson:"setlist"`
	FoundSongs       []Song   `bson:"foundSongs"`
	MessagesToDelete []int    `bson:"messagesToDelete"`
	CurrentVoice     *Voice   `bson:"currentVoice"`
	Key              string   `bson:"key"`
}

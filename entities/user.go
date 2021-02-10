package entities

type User struct {
	ID     *int64  `bson:"_id"`
	States []State `bson:"states"`
}

type State struct {
	Index   int     `bson:"index"`
	Name    string  `bson:"name"`
	Context Context `bson:"context"`
}

type Context struct {
	Songs []Song `bson:"songs"`
}

func NewUser(ID int64) *User {
	return &User{
		ID: &ID,
		States: []State{
			{
				Index:   0,
				Context: Context{},
			},
		},
	}
}

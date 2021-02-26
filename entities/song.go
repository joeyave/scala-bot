package entities

type Song struct {
	ID           string   `bson:"_id"`
	Name         string   `bson:"name"`
	TgFileID     string   `bson:"tgFileId"`
	ModifiedTime string   `bson:"modifiedTime"`
	WebViewLink  string   `bson:"webViewLink"`
	Voices       []*Voice `bson:"voices"`
	Parents      []string `bson:"parents"`
}

type Voice struct {
	TgFileID string `bson:"tgFileId"`
	Caption  string `bson:"caption"`
}

package entities

type Song struct {
	ID           *string `bson:"_id"`
	Name         string  `bson:"name"`
	TgFileID     string  `bson:"tgFileId"`
	ModifiedTime string  `bson:"modifiedTime"`
	WebViewLink  string  `bson:"webViewLink"`
	Voices       []voice `bson:"voices"`
}

type voice struct {
	TgFileID       string `bson:"tgFileId"`
	TgFileUniqueID string `bson:"tgFileUniqueId"`
	Caption        string `bson:"caption"`
}

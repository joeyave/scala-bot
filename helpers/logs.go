package helpers

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"strings"
)

type LogsWriter struct {
	Bot       *telebot.Bot
	ChannelID int64
}

func (w LogsWriter) Write(p []byte) (n int, err error) {

	str := fmt.Sprintf("<code>%s</code>", string(p))

	if strings.Contains(str, "ERR") {
		str = "❗️❗️❗️" + str
		_, err = w.Bot.Send(telebot.ChatID(w.ChannelID), str, telebot.ModeHTML, telebot.NoPreview)
	} else {
		_, err = w.Bot.Send(telebot.ChatID(w.ChannelID), str, telebot.ModeHTML, telebot.NoPreview, telebot.Silent)
	}

	return len(p), nil
}

type Update struct {
	*telebot.Update
	Message  *Message  `json:"message,omitempty"`
	Callback *Callback `json:"callback_query,omitempty"`
}

type Message struct {
	*telebot.Message

	Entities           *struct{}          `json:"entities,omitempty"` // https://github.com/golang/go/issues/35501
	ID                 int                `json:"message_id,omitempty"`
	Sender             *User              `json:"from,omitempty"`
	Unixtime           int64              `json:"date,omitempty"`
	Chat               *Chat              `json:"chat,omitempty"`
	SenderChat         *Chat              `json:"sender_chat,omitempty"`
	OriginalSender     *User              `json:"forward_from,omitempty"`
	OriginalChat       *Chat              `json:"forward_from_chat,omitempty"`
	OriginalMessageID  int                `json:"forward_from_message_id,omitempty"`
	OriginalSignature  string             `json:"forward_signature,omitempty"`
	OriginalSenderName string             `json:"forward_sender_name,omitempty"`
	OriginalUnixtime   int                `json:"forward_date,omitempty"`
	AutomaticForward   bool               `json:"is_automatic_forward,omitempty"`
	ReplyTo            *Message           `json:"reply_to_message,omitempty"`
	Via                *User              `json:"via_bot,omitempty"`
	LastEdit           int64              `json:"edit_date,omitempty"`
	AlbumID            string             `json:"media_group_id,omitempty"`
	Signature          string             `json:"author_signature,omitempty"`
	Text               string             `json:"text,omitempty"`
	Audio              *telebot.Audio     `json:"audio,omitempty"`
	Document           *telebot.Document  `json:"document,omitempty"`
	Photo              *telebot.Photo     `json:"photo,omitempty"`
	Sticker            *telebot.Sticker   `json:"sticker,omitempty"`
	Voice              *telebot.Voice     `json:"voice,omitempty"`
	VideoNote          *telebot.VideoNote `json:"video_note,omitempty"`
	Video              *telebot.Video     `json:"video,omitempty"`
	Animation          *telebot.Animation `json:"animation,omitempty"`
	Contact            *telebot.Contact   `json:"contact,omitempty"`
	Location           *telebot.Location  `json:"location,omitempty"`
	Venue              *telebot.Venue     `json:"venue,omitempty"`
	Poll               *telebot.Poll      `json:"poll,omitempty"`
	Game               *telebot.Game      `json:"game,omitempty"`
	Dice               *telebot.Dice      `json:"dice,omitempty"`
	UserJoined         *User              `json:"new_chat_member,omitempty"`
	UserLeft           *User              `json:"left_chat_member,omitempty"`
	NewGroupTitle      string             `json:"new_chat_title,omitempty"`
	NewGroupPhoto      *telebot.Photo     `json:"new_chat_photo,omitempty"`
	UsersJoined        []*User            `json:"new_chat_members,omitempty"`
	GroupPhotoDeleted  bool               `json:"delete_chat_photo,omitempty"`
	GroupCreated       bool               `json:"group_chat_created,omitempty"`
	SuperGroupCreated  bool               `json:"supergroup_chat_created,omitempty"`
	ChannelCreated     bool               `json:"channel_chat_created,omitempty"`
	MigrateTo          int64              `json:"migrate_to_chat_id,omitempty"`
	MigrateFrom        int64              `json:"migrate_from_chat_id,omitempty"`
	PinnedMessage      *Message           `json:"pinned_message,omitempty"`
	Invoice            *telebot.Invoice   `json:"invoice,omitempty"`
	Payment            *telebot.Payment   `json:"successful_payment,omitempty"`
}

func MapMessage(telebotMessage *telebot.Message) *Message {

	if telebotMessage == nil {
		return nil
	}

	return &Message{
		Message:            telebotMessage,
		ID:                 telebotMessage.ID,
		Sender:             MapUser(telebotMessage.Sender),
		Unixtime:           telebotMessage.Unixtime,
		Chat:               MapChat(telebotMessage.Chat),
		SenderChat:         MapChat(telebotMessage.SenderChat),
		OriginalSender:     MapUser(telebotMessage.OriginalSender),
		OriginalChat:       MapChat(telebotMessage.OriginalChat),
		OriginalMessageID:  telebotMessage.OriginalMessageID,
		OriginalSignature:  telebotMessage.OriginalSignature,
		OriginalSenderName: telebotMessage.OriginalSenderName,
		OriginalUnixtime:   telebotMessage.OriginalUnixtime,
		AutomaticForward:   telebotMessage.AutomaticForward,
		ReplyTo:            MapMessage(telebotMessage.ReplyTo),
		Via:                MapUser(telebotMessage.Via),
		LastEdit:           telebotMessage.LastEdit,
		AlbumID:            telebotMessage.AlbumID,
		Signature:          telebotMessage.Signature,
		Text:               telebotMessage.Text,
		Audio:              telebotMessage.Audio,
		Document:           telebotMessage.Document,
		Photo:              telebotMessage.Photo,
		Sticker:            telebotMessage.Sticker,
		Voice:              telebotMessage.Voice,
		VideoNote:          telebotMessage.VideoNote,
		Video:              telebotMessage.Video,
		Animation:          telebotMessage.Animation,
		Contact:            telebotMessage.Contact,
		Location:           telebotMessage.Location,
		Venue:              telebotMessage.Venue,
		Poll:               telebotMessage.Poll,
		Game:               telebotMessage.Game,
		Dice:               telebotMessage.Dice,
		UserJoined:         MapUser(telebotMessage.UserJoined),
		UserLeft:           MapUser(telebotMessage.UserLeft),
		NewGroupTitle:      telebotMessage.NewGroupTitle,
		NewGroupPhoto:      telebotMessage.NewGroupPhoto,
		UsersJoined:        MapUsers(telebotMessage.UsersJoined),
		GroupPhotoDeleted:  telebotMessage.GroupPhotoDeleted,
		GroupCreated:       telebotMessage.GroupCreated,
		SuperGroupCreated:  telebotMessage.SuperGroupCreated,
		ChannelCreated:     telebotMessage.ChannelCreated,
		MigrateTo:          telebotMessage.MigrateTo,
		MigrateFrom:        telebotMessage.MigrateFrom,
		PinnedMessage:      MapMessage(telebotMessage.PinnedMessage),
		Invoice:            telebotMessage.Invoice,
		Payment:            telebotMessage.Payment,
	}
}

type Callback struct {
	ID        string   `json:"id,omitempty"`
	Sender    *User    `json:"from,omitempty"`
	Message   *Message `json:"message,omitempty"`
	MessageID string   `json:"inline_message_id,omitempty"`
	Data      string   `json:"data,omitempty"`
	Unique    string   `json:"-"`
}

func MapCallback(telebotCallback *telebot.Callback) *Callback {

	if telebotCallback == nil {
		return nil
	}

	return &Callback{
		ID:        telebotCallback.ID,
		Sender:    MapUser(telebotCallback.Sender),
		Message:   MapMessage(telebotCallback.Message),
		MessageID: telebotCallback.MessageID,
		Data:      telebotCallback.Data,
		Unique:    telebotCallback.Unique,
	}
}

type User struct {
	ID int64 `json:"id,omitempty"`

	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
	IsBot        bool   `json:"is_bot,omitempty"`

	// Returns only in getMe
	CanJoinGroups   bool `json:"can_join_groups,omitempty"`
	CanReadMessages bool `json:"can_read_all_group_messages,omitempty"`
	SupportsInline  bool `json:"supports_inline_queries,omitempty"`
}

func MapUser(telebotUser *telebot.User) *User {

	if telebotUser == nil {
		return nil
	}

	return &User{
		ID:              telebotUser.ID,
		FirstName:       telebotUser.FirstName,
		LastName:        telebotUser.LastName,
		Username:        telebotUser.Username,
		LanguageCode:    telebotUser.LanguageCode,
		IsBot:           telebotUser.IsBot,
		CanJoinGroups:   telebotUser.CanJoinGroups,
		CanReadMessages: telebotUser.CanReadMessages,
		SupportsInline:  telebotUser.SupportsInline,
	}
}

func MapUsers(telebotUsers []telebot.User) []*User {

	var users []*User
	for _, telebotUser := range telebotUsers {
		users = append(users, MapUser(&telebotUser))
	}

	return users
}

type Chat struct {
	*telebot.Chat

	ID        int64  `json:"id,omitempty"`
	Type      string `json:"type,omitempty"`
	Title     string `json:"title,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	Still     bool   `json:"is_member,omitempty,omitempty"`
}

func MapChat(telebotChat *telebot.Chat) *Chat {

	if telebotChat == nil {
		return nil
	}

	return &Chat{
		Chat:      telebotChat,
		ID:        telebotChat.ID,
		Type:      string(telebotChat.Type),
		Title:     telebotChat.Title,
		FirstName: telebotChat.FirstName,
		LastName:  telebotChat.LastName,
		Username:  telebotChat.Username,
		Still:     telebotChat.Still,
	}
}

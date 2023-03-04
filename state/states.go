package state

// Plain keyboard states.
const (
	GetEvents = iota + 1
	FilterEvents

	Search
	SearchSetlist

	GetSongs
	FilterSongs

	SongVoices_CreateVoice

	BandCreate

	RoleCreate_ChoosePosition

	TransposeAudio
)

// Inline states.
const (
	EventCB = iota + 1

	EventSetlistDocs
	EventSetlistMetronome

	EventSetlist
	EventSetlistDeleteOrRecoverSong

	EventMembers
	EventMembersDeleteOrRecoverMember
	EventMembersAddMemberChooseRole
	EventMembersAddMemberChooseUser
	EventMembersAddMember
	EventMembersDeleteMember

	EventDeleteConfirm
	EventDelete

	SongCB
	SongLike

	SongVoices
	SongVoicesCreateVoiceAskForAudio
	SongVoice
	SongVoiceDeleteConfirm
	SongVoiceDelete

	SongDeleteConfirm
	SongDelete

	SettingsCB
	SettingsBands
	SettingsChooseBand

	SettingsBandMembers
	SettingsBandAddAdmin

	SongCopyToMyBand
	SongStyle
	SongAddLyricsPage

	BandCreate_AskForName

	RoleCreate_AskForName
	RoleCreate
)

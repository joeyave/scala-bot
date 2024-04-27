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

	SongArchive

	SettingsCB
	SettingsBands
	SettingsChooseBand

	SettingsBandMembers
	SettingsCleanupDatabase
	SettingsBandAddAdmin

	SongCopyToMyBand
	SongStyle
	SongAddLyricsPage

	BandCreate_AskForName

	RoleCreate_AskForName
	RoleCreate

	TransposeAudio_AskForSemitonesNumber
	TransposeAudio
)

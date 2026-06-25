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

	_ // reserved for removed chat-based band creation state

	RoleCreate_ChoosePosition
)

// Inline states.
const (
	EventCB = iota + 1

	EventSetlistDocs

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

	SongStats

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

	_ // reserved for removed chat-based band creation callback

	RoleCreate_AskForName
	RoleCreate

	TransposeAudio_AskForSemitonesNumber
	TransposeAudio

	JoinRequestApprove
	JoinRequestDecline
)

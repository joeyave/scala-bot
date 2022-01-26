package helpers

const SongsPageSize = 50
const EventsPageSize = 25

const (
	SearchSongState = iota
	SetlistState
	SongActionsState
	GetVoicesState
	AddSongTagState
	UploadVoiceState
	DeleteVoiceState
	MainMenuState
	TransposeSongState
	StyleSongState
	AddLyricsPageState
	ChangeSongBPMState
	ChooseBandState
	CreateBandState
	CopySongState
	CreateSongState
	DeleteSongState
	AddBandAdminState
	GetEventsState
	CreateEventState
	EventActionsState
	CreateRoleState
	AddEventSongState
	DeleteEventState
	ChangeSongOrderState
	AddEventMemberState
	ChangeEventDateState
	ChangeEventNotesState
	DeleteEventMemberState
	DeleteEventSongState
	GetSongsFromMongoState
	EditInlineKeyboardState
	SettingsState
)

// Buttons constants.
const (
	LoadMore                    string = "👨‍👩‍👧‍👦 Загрузить еще"
	Cancel                      string = "🚫 Отмена"
	Skip                        string = "⏩ Пропустить"
	Help                        string = "Как пользоваться?"
	CreateDoc                   string = "➕ Создать документ"
	Voices                      string = "Партии"
	Tags                        string = "Теги"
	CreateTag                   string = "➕ Создать тег"
	Audios                      string = "Аудио"
	Transpose                   string = "🎛 Транспонировать"
	Style                       string = "🎨 Стилизовать"
	ChangeSongBPM               string = "🥁 Изменить BPM"
	AddLyricsPage               string = "🔤 Слова"
	Menu                        string = "💻 Меню"
	Back                        string = "↩︎ Назад"
	Forward                     string = "▶️ Вперед"
	No                          string = "⛔️ Нет"
	Yes                         string = "✅ Да"
	AppendSection               string = "В конец документа"
	CreateBand                  string = "Создать свою группу"
	CreateEvent                 string = "➕ Добавить собрание"
	SearchEverywhere            string = "🔎 Искать во всех группах"
	CopyToMyBand                string = "🖨 Копировать песню в свою группу"
	Schedule                    string = "🗓️ Расписание"
	FindChords                  string = "🎶 Аккорды"
	ChangeBand                  string = "Изменить группу"
	AddAdmin                    string = "➕ Добавить администратора"
	Settings                    string = "⚙ Настройки"
	CreateRole                  string = "Создать роль"
	Members                     string = "📈 Статистика"
	Songs                       string = "🎵 Песни"
	AddMember                   string = "➕ Участник"
	DeleteMember                string = "➖ Участник"
	AddSong                     string = "➕ Песня"
	DeleteSong                  string = "➖ Песня"
	SongsOrder                  string = "🔄 Порядок песен"
	Date                        string = "🗓️ Изменить дату"
	Notes                       string = "✏️ Заметки"
	Edit                        string = "︎✍️ Редактировать"
	Archive                     string = "📥"
	ByWeekday                   string = "День недели"
	GetEventsWithMe             string = "🙋‍♂️"
	End                         string = "⛔️ Закончить"
	Delete                      string = "❌ Удалить"
	BandSettings                string = "Настройки группы"
	ProfileSettings             string = "Настройки профиля"
	SongsByNumberOfPerforming   string = "🔢"
	SongsByLastDateOfPerforming string = "📆"
	LikedSongs                  string = "❤️‍🔥"
	NextPage                    string = "→"
	PrevPage                    string = "←"
	Today                       string = "⏰"
	LinkToTheDoc                string = "📎 Ссылка на документ"
	Setlist                     string = "📝 Список"
	Like                        string = "❤️‍🔥"
	Placeholder                 string = "Фраза из песни"
)

// Roles.
const (
	Admin string = "Admin"
)

var FilesChannelID int64
var LogsChannelID int64

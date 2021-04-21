package helpers

const PageSize = 50

const (
	SearchSongState = iota
	SetlistState
	SongActionsState
	GetVoicesState
	UploadVoiceState
	MainMenuState
	TransposeSongState
	StyleSongState
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
	DeleteEventMemberState
	DeleteEventSongState
	GetSongsFromMongoHandler
)

// Buttons constants.
const (
	Cancel                      string = "🚫 Отмена"
	Skip                        string = "⏩ Пропустить"
	Help                        string = "Как пользоваться?"
	CreateDoc                   string = "Создать документ"
	Voices                      string = "Партии"
	Audios                      string = "Аудио"
	Transpose                   string = "🎛 Транспонировать"
	Style                       string = "Стилизовать"
	Menu                        string = "Меню"
	Back                        string = "◀️ Назад"
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
	Members                     string = "🧑‍🤝‍🧑 Участники"
	Songs                       string = "🎵 Песни"
	AddMember                   string = "➕ Участник"
	DeleteMember                string = "➖ Участник"
	AddSong                     string = "➕ Песня"
	DeleteSong                  string = "➖ Песня"
	ChangeSongsOrder            string = "🔄 Изменить порядок песен"
	GetAllEvents                string = "Все собрания"
	GetEventsWithMe             string = "🙋‍♂️ Собрания, где я участвую"
	End                         string = "🔴 Закончить"
	Delete                      string = "Удалить"
	BandSettings                string = "Настройки группы"
	ProfileSettings             string = "Настройки профиля"
	AllSongs                    string = "Все песни"
	SongsByNumberOfPerforming   string = "🧮 По количеству исполнений"
	SongsByLastDateOfPerforming string = "📆 По последнему исполнению"
	NextPage                    string = "▶️ Следующая страница"
	PrevPage                    string = "◀️ Предыдущая страница"
	Today                       string = "⏰ Сегодня"
	LinkToTheDoc                string = "Ссылка на документ"
	Setlist                     string = "📝 Список"
)

// Roles.
const (
	Admin string = "Admin"
)

var FilesChannelID int64
var LogsChannelID int64

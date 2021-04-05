package helpers

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
	ScheduleState
	CreateSongState
	DeleteSongState
	AddBandAdminState
	GetEventsState
	CreateEventState
	EventActionsState
	CreateRoleState
	AddEventMemberState
	AddEventSongState
	DeleteEventState
	ChangeSongOrderState
	AddEventMemberCallbackState
	DeleteEventMemberCallbackState
)

// Buttons constants.
const (
	Cancel           string = "Отмена"
	Skip             string = "Пропустить"
	Help             string = "Как пользоваться?"
	CreateDoc        string = "Создать документ"
	Voices           string = "Партии"
	Audios           string = "Аудио"
	Transpose        string = "Изменить тональность"
	Style            string = "Стилизовать"
	Menu             string = "Меню"
	Delete           string = "Удалить"
	Back             string = "Назад"
	No               string = "Нет"
	Yes              string = "Да"
	AppendSection    string = "В конец документа"
	CreateBand       string = "Создать свою группу"
	CreateEvent      string = "Добавить собрание"
	SearchEverywhere string = "Искать во всех группах"
	CopyToMyBand     string = "Копировать песню в свою группу"
	Schedule         string = "Расписание"
	ScheduleBeta     string = "Расписание (beta)"
	FindChords       string = "Найти аккорды"
	ChangeBand       string = "Изменить группу"
	AddAdmin         string = "Добавить администратора"
	Settings         string = "Настройки"
	CreateRole       string = "Создать роль"
	Members          string = "Участники"
	Songs            string = "Песни"
	Add              string = "Добавить"
	GetAllEvents     string = "Все собрания"
	End              string = "Закончить"
	ChangeOrder      string = "Изменить порядок"
)

const (
	Admin string = "Admin"
)

var FilesChannelID int64
var LogsChannelID int64

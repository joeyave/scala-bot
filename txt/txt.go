package txt

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var locales = map[string]map[string]string{
	"button.schedule": {
		"ru": "🗓️ Расписание",
	},
	"button.menu": {
		"ru": "💻 Меню",
	},
	"button.songs": {
		"ru": "🎵 Песни",
	},
	"button.stats": {
		"ru": "📈 Статистика",
	},
	"button.settings": {
		"ru": "⚙ Настройки",
	},
	"button.next": {
		"ru": "→",
	},
	"button.prev": {
		"ru": "←",
	},
	"button.eventsWithMe": {
		"ru": "🙋‍♂️",
	},
	"button.archive": {
		"ru": "📥",
	},
	"button.like": {
		"ru": "❤️‍🔥",
	},
	"button.unlike": {
		"ru": "♡",
	},
	"button.calendar": {
		"ru": "📆",
	},
	"button.numbers": {
		"ru": "🔢",
	},
	"button.tag": {
		"ru": "🔖",
	},
	"button.globalSearch": {
		"ru": "🔎 Искать во всех группах",
	},
	"button.cancel": {
		"ru": "🚫 Отмена",
	},
	"button.skip": {
		"ru": "⏩ Пропустить",
	},
	"button.createDoc": {
		"ru": "➕ Создать документ",
	},
	"button.createBand": {
		"ru": "➕ Создать группу",
	},
	"button.createEvent": {
		"ru": "➕ Добавить собрание",
	},
	"button.createRole": {
		"ru": "➕ Создать роль",
	},
	"button.chords": {
		"ru": "🎶 Аккорды",
	},
	"button.metronome": {
		"ru": "🥁 Метроном",
	},
	"button.edit": {
		"ru": "✍️ Редактировать",
	},
	"button.setlist": {
		"ru": "📝 Список",
	},
	"button.members": {
		"ru": "🙋‍♂️ Участники",
	},
	"button.notes": {
		"ru": "✏️ Заметки",
	},
	"button.editDate": {
		"ru": "🗓️ Изменить дату",
	},
	"button.delete": {
		"ru": "🗑 Удалить",
	},
	"button.back": {
		"ru": "↩︎ Назад",
	},
	"button.changeSongsOrder": {
		"ru": "🔄 Изменить порядок песен",
	},
	"button.eventEditEtc": {
		"ru": "Список, дата, заметки...",
	},
	"button.addSong": {
		"ru": "➕ Добавить песню",
	},
	"button.addMember": {
		"ru": "➕ Добавить участника",
	},
	"button.loadMore": {
		"ru": "👩‍👧‍👦 Загрузить еще",
	},
	"button.docLink": {
		"ru": "📎 Ссылка на Google Doc",
	},
	"button.voices": {
		"ru": "🎤 Партии",
	},
	"button.tags": {
		"ru": "🔖 Теги",
	},
	"button.more": {
		//"ru": "💬",
		"ru": "•••",
	},
	"button.transpose": {
		"ru": "🎛 Транспонировать",
	},
	"button.style": {
		"ru": "🎨 Стилизовать",
	},
	"button.changeBpm": {
		"ru": "🥁 Изменить BPM",
	},
	"button.lyrics": {
		"ru": "🔤 Слова",
	},
	"button.copyToMyBand": {
		"ru": "🖨 Копировать песню в свою группу",
	},
	"button.yes": {
		"ru": "✅ Да",
	},
	"button.createTag": {
		"ru": "➕ Создать тег",
	},
	"button.addVoice": {
		"ru": "➕ Добавить партию",
	},
	"button.changeBand": {
		"ru": "Изменить группу",
	},
	"button.addAdmin": {
		"ru": "Добавить админа",
	},

	"text.defaultPlaceholder": {
		"ru": "Фраза из песни или список",
	},
	"text.chooseEvent": {
		"ru": "Выбери собрание:",
	},
	"text.chooseTag": {
		"ru": "Выбери тег:",
	},
	"text.chooseSong": {
		"ru": "Выбери песню:",
	},
	"text.chooseSongOrTypeAnotherQuery": {
		"ru": "Выбери песню по запросу %s или введи другое название:",
	},
	"text.chooseRoleForNewMember": {
		"ru": "Выбери роль для нового участника:",
	},
	"text.chooseVoice": {
		"ru": "Выбери партию:",
	},
	"text.chooseNewMember": {
		"ru": "Выбери нового участника на роль %s:",
	},
	"text.chooseMemberToMakeAdmin": {
		"ru": "Выбери пользователя, которого ты хочешь сделать администратором:",
	},
	"text.chooseBand": {
		"ru": "Выбери группу или создай свою.",
	},
	"text.addedToBand": {
		"ru": "Ты добавлен в группу %s.",
	},
	"text.nothingFound": {
		"ru": "Ничего не найдено. Попробуй еще раз.",
	},
	"text.nothingFoundByQuery": {
		"ru": "По запросу %s ничего не найдено. Напиши новое название или пропусти эту песню.",
	},
	"text.menu": {
		"ru": "Меню:",
	},
	"text.sendAudioOrVoice": {
		"ru": "Отправь мне аудио или голосовое сообщение.",
	},
	"text.sendVoiceName": {
		"ru": "Отправь мне название этой партии.",
	},
	"text.sendTagName": {
		"ru": "Введи название тега:",
	},
	"text.sendBandName": {
		"ru": "Введи название группы:",
	},
	"text.sendRoleName": {
		"ru": "Отправь название новой роли. К сожалению, пока что отредактировать или удалить роль нельзя, поэтому напиши без ошибок.\n\n" +
			"Пример:\n🎤 Вокалисты\n 🎹 Клавишники \n📽 Медиа",
	},
	"text.voiceDeleteConfirm": {
		"ru": "Удалить эту партию?",
	},
	"text.eventDeleteConfirm": {
		"ru": "Удалить это собрание?",
	},
	"text.songDeleteConfirm": {
		"ru": "Удалить эту песню?",
	},
	"text.eventDeleted": {
		"ru": "Собрание удалено.",
	},
	"text.songDeleted": {
		"ru": "Песня удалена.",
	},
	"text.styled": {
		"ru": "Стилизация закончена. Аккорды выделены и покрашены.",
	},
	"text.addedLyricsPage": {
		"ru": "На вторую страницу добавлены слова без аккордов.",
	},

	"command.menu": {
		"ru": "menu",
	},
}

func init() {
	for key, langToMsgMap := range locales {
		for lang, msg := range langToMsgMap {
			message.SetString(language.Make(lang), key, msg)
		}
	}
}

func Get(key, lang string, a ...interface{}) string {
	//switch lang {
	//case "en":
	//	return enPrinter.Sprintf(key, a)
	//case "uk":
	//	return ukPrinter.Sprintf(key, a)
	//}
	return ruPrinter.Sprintf(key, a...)
}

//var ukPrinter = message.NewPrinter(language.Ukrainian)
//var enPrinter = message.NewPrinter(language.English)
var ruPrinter = message.NewPrinter(language.Russian)

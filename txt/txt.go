package txt

import (
	locales2 "github.com/go-playground/locales"
	"github.com/go-playground/locales/ru"
	"github.com/go-playground/locales/uk"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var locales = map[string]map[string]string{
	"text.defaultDocTitle": {
		"ru": "Название - Исполнитель",
		"uk": "Назва - Виконавець",
	},
	"text.roleAdded": {
		"ru": "Добавлена новая роль: %s.",
		"uk": "Додано нову роль: %s.",
	},
	"text.addedToBandAsAdmin": {
		"ru": "Ты добавлен в группу \"%s\" как администратор.",
		"uk": "Тебе додано до групи \"%s\" як адміністратора.",
	},
	"text.cleanQueryInfo": {
		"ru": "Из запроса удаляются все числа, дефисы и скобки вместе с тем, что в них.",
		"uk": "З запиту видаляються всі числа, дефіси та дужки разом із тим, що в них.",
	},
	"text.memberRemovedNotification": {
		"ru": "Привет. %s только что удалил тебя как %s из собрания ☹️\n\n%s️",
		"uk": "Привіт. %s щойно видалив тебе як %s із заходу ☹️\n\n%s",
	},
	"text.memberAddedNotification": {
		"ru": "Привет. %s только что добавил тебя как %s в собрание!\n\n%s",
		"uk": "Привіт. %s щойно додав тебе як %s у захід!\n\n%s",
	},
	"button.moreInfo": {
		"ru": "ℹ️ Подробнее",
		"uk": "ℹ️ Докладніше",
	},
	"text.upcomingEventNotification": {
		"ru": "Привет. Ты участвуешь в собрании через несколько дней!\n\n%s",
		"uk": "Привіт. Ти береш участь у заході через кілька днів!\n\n%s",
	},
	"button.schedule": {
		"ru": "🗓️ Расписание",
		"uk": "🗓️ Розклад",
	},
	"button.menu": {
		"ru": "💻 Меню",
		"uk": "💻 Меню",
	},
	"button.songs": {
		"ru": "🎵 Песни",
		"uk": "🎵 Пісні",
	},
	"button.stats": {
		"ru": "📈 Статистика",
		"uk": "📈 Статистика",
	},
	"button.settings": {
		"ru": "⚙ Настройки",
		"uk": "⚙ Налаштування",
	},
	"button.next": {
		"ru": "→",
		"uk": "→",
	},
	"button.prev": {
		"ru": "←",
		"uk": "←",
	},
	"button.eventsWithMe": {
		"ru": "🙋‍♂️",
		"uk": "🙋‍♂️",
	},
	"button.archive": {
		"ru": "📥",
		"uk": "📥",
	},
	"button.like": {
		"ru": "❤️‍🔥",
		"uk": "❤️‍🔥",
	},
	"button.unlike": {
		"ru": "♡",
		"uk": "♡",
	},
	"button.calendar": {
		"ru": "📆",
		"uk": "📆",
	},
	"button.numbers": {
		"ru": "🔢",
		"uk": "🔢",
	},
	"button.tag": {
		"ru": "🔖",
		"uk": "🔖",
	},
	"button.globalSearch": {
		"ru": "🔎 Искать во всех группах",
		"uk": "🔎 Шукати у всіх групах",
	},
	"button.cancel": {
		"ru": "🚫 Отмена",
		"uk": "🚫 Скасувати",
	},
	"button.skip": {
		"ru": "⏩ Пропустить",
		"uk": "⏩ Пропустити",
	},
	"button.createDoc": {
		"ru": "➕ Создать документ",
		"uk": "➕ Створити документ",
	},
	"button.createBand": {
		"ru": "➕ Создать группу",
		"uk": "➕ Створити группу",
	},
	"button.createEvent": {
		"ru": "➕ Добавить собрание",
		"uk": "➕ Додати захід",
	},
	"button.createRole": {
		"ru": "➕ Создать роль",
		"uk": "➕ Створити роль",
	},
	"button.chords": {
		"ru": "🎶 Аккорды",
		"uk": "🎶 Акорди",
	},
	"button.metronome": {
		"ru": "🥁 Метроном",
		"uk": "🥁 Метроном",
	},
	"button.edit": {
		"ru": "✍️ Редактировать",
		"uk": "✍️ Редагувати",
	},
	"button.setlist": {
		"ru": "📝 Список",
		"uk": "📝 Список",
	},
	"button.members": {
		"ru": "🙋‍♂️ Участники",
		"uk": "🙋‍♂️ Учасники",
	},
	"button.notes": {
		"ru": "✏️ Заметки",
		"uk": "✏️ Нотатки",
	},
	"button.editDate": {
		"ru": "🗓️ Изменить дату",
		"uk": "🗓️ Змінити дату",
	},
	"button.delete": {
		"ru": "🗑 Удалить",
		"uk": "🗑 Видалити",
	},
	"button.archiveText": {
		"ru": "📥 В архив",
		"uk": "📥 До архіву",
	},
	"button.unarchiveText": {
		"ru": "📤 К актуальным",
		"uk": "📤 До актуальних",
	},
	"button.back": {
		"ru": "↩︎ Назад",
		"uk": "↩︎ Назад",
	},
	"button.changeSongsOrder": {
		"ru": "🔄 Изменить порядок песен",
		"uk": "🔄 Змінити порядок пісень",
	},
	"button.eventEditEtc": {
		"ru": "Список, дата, заметки...",
		"uk": "Список, дата, нотатки...",
	},
	"button.addSong": {
		"ru": "➕ Добавить песню",
		"uk": "➕ Додати пісню",
	},
	"button.addMember": {
		"ru": "➕ Добавить участника",
		"uk": "➕ Додати учасника",
	},
	"button.loadMore": {
		"ru": "👩‍👧‍👦 Загрузить еще",
		"uk": "👩‍👧‍👦 Завантажити ще",
	},
	"button.docLink": {
		"ru": "📎 Ссылка на Google Doc",
		"uk": "📎 Посилання на Google Doc",
	},
	"button.voices": {
		"ru": "🎤 Партии (%d)",
		"uk": "🎤 Партії (%d)",
	},
	"button.tags": {
		"ru": "🔖 Теги",
		"uk": "🔖 Теги",
	},
	"button.more": {
		// "ru": "💬",
		"ru": "•••",
		"uk": "•••",
	},
	"button.transpose": {
		"ru": "🎛 Транспонировать",
		"uk": "🎛 Транспонувати",
	},
	"button.style": {
		"ru": "🎨 Форматировать",
		"uk": "🎨 Форматувати",
	},
	"button.changeBpm": {
		"ru": "🥁 Изменить BPM",
		"uk": "🥁 Змінити BPM",
	},
	"button.lyrics": {
		"ru": "🔤 Слова",
		"uk": "🔤 Слова",
	},
	"button.copyToMyBand": {
		"ru": "🖨 Копировать песню в свою группу",
		"uk": "🖨 Копіювати пісню в свою групу",
	},
	"button.yes": {
		"ru": "✅ Да",
		"uk": "✅ Так",
	},
	"button.createTag": {
		"ru": "➕ Создать тег",
		"uk": "➕ Створити тег",
	},
	"button.addVoice": {
		"ru": "➕ Добавить партию",
		"uk": "➕ Додати партію",
	},
	"button.changeBand": {
		"ru": "Изменить группу",
		"uk": "Змінити групу",
	},
	"button.addAdmin": {
		"ru": "Добавить админа",
		"uk": "Додати адміна",
	},
	"button.cleanupDatabase": {
		"ru": "Почистить базу (admin)",
		"uk": "Почистити базу (admin)",
	},
	"button.continue": {
		"ru": "Продолжить",
		"uk": "Продовжити",
	},
	"button.qualitatively": {
		"ru": "Качественно, но долго",
		"uk": "Якісно, але довго",
	},
	"button.skipClippingCheck": {
		"ru": "Без проверки на клиппинг",
		"uk": "Без перевірки на кліппінг",
	},
	"button.fine": {
		"ru": "Качественно",
		"uk": "Якісно",
	},
	"button.fast": {
		"ru": "Быстро",
		"uk": "Швидко",
	},

	"text.statsCaption": {
		"ru": "Всего исполнений: <b>%d %s</b>\nДата последнего исполнения: <b>%s</b> (%s)",
		"uk": "Всього виконань: <b>%d %s</b>\nДата останнього виконання: <b>%s</b> (%s)",
	},
	"text.noStatsCaption": {
		"ru": "За этот период не было исполнений",
		"uk": "За цей період не було виконань",
	},
	"text.title": {
		"ru": "Название",
		"uk": "Назва",
	},
	"text.defaultPlaceholder": {
		"ru": "Слова или список",
		"uk": "Слова або список",
	},
	"text.chooseEvent": {
		"ru": "Выбери собрание:",
		"uk": "Вибери захід:",
	},
	"text.chooseTag": {
		"ru": "Выбери тег:",
		"uk": "Вибери тег:",
	},
	"text.chooseStatsPeriod": {
		"ru": "Выбери период для статистики:",
		"uk": "Вибери період для статистики:",
	},
	"text.chooseStatsSorting": {
		"ru": "Выбери сортировку для статистики:",
		"uk": "Вибери сортування для статистики:",
	},
	"text.chooseSong": {
		"ru": "Выбери песню:",
		"uk": "Вибери пісню:",
	},
	"text.startDateToDo": {
		"ru": "Статистика считается за последние пол года.",
		"uk": "Статистика рахується за останні пів року.",
	},
	"text.chooseSongOrTypeAnotherQuery": {
		"ru": "Выбери песню по запросу %s или введи другое название:",
		"uk": "Вибери пісню за запитом %s або введи іншу назву:",
	},
	"text.chooseRoleForNewMember": {
		"ru": "Выбери роль для нового участника:",
		"uk": "Вибери роль для нового учасника:",
	},
	"text.chooseVoice": {
		"ru": "Выбери партию:",
		"uk": "Вибери партію:",
	},
	"text.chooseNewMember": {
		"ru": "Выбери нового участника на роль %s:",
		"uk": "Вибери нового учасника на роль %s:",
	},
	"text.chooseMemberToMakeAdmin": {
		"ru": "Выбери пользователя, которого ты хочешь сделать администратором:",
		"uk": "Вибери користувача, якого ти хочеш зробити адміністратором:",
	},
	"text.chooseBand": {
		"ru": "Выбери группу или создай свою.",
		"uk": "Вибери групу або створи свою.",
	},
	"text.addedToBand": {
		"ru": "Ты добавлен в группу %s.",
		"uk": "Ти додан у групу %s.",
	},
	"text.removedFromBand": {
		"ru": "Ты удален с группы %s.",
		"uk": "Тебе видалено з групи %s.",
	},
	"text.nothingFound": {
		"ru": "Ничего не найдено. Попробуй еще раз.",
		"uk": "Нічого не знайдено. Спробуй ще раз.",
	},
	"text.nothingFoundByQuery": {
		"ru": "По запросу %s ничего не найдено. Напиши новое название или пропусти эту песню.",
		"uk": "За запитом %s нічого не знайдено. Напиши нову назву або пропусти цю пісню.",
	},
	"text.menu": {
		"ru": "Меню:",
		"uk": "Меню:",
	},
	"text.sendAudioOrVoice": {
		"ru": "Отправь мне аудио или голосовое сообщение.",
		"uk": "Відправ мені аудіо або голосове повідомлення.",
	},
	"text.selectOrCreateBand": {
		"uk": "Вибрати або створити свою групу",
		"ru": "Выбрать или создать свою группу",
	},
	"text.sendSemitones": {
		"ru": "На сколько полутонов транспонировать этот аудио файл?",
		"uk": "На скільки півтонів транспонувати цей аудіо файл?",
	},
	"text.sendVoiceName": {
		"ru": "Отправь мне название этой партии.",
		"uk": "Відправ мені назву цієї партії.",
	},
	"text.sendTagName": {
		"ru": "Введи название тега:",
		"uk": "Введи назву тега:",
	},
	"text.sendBandName": {
		"ru": "Введи название группы:",
		"uk": "Введи назву групи:",
	},
	"text.sendRoleName": {
		"ru": "Отправь название новой роли. К сожалению, пока что отредактировать или удалить роль нельзя, поэтому напиши без ошибок.\n\n" +
			"Пример:\n🎤 Вокалисты\n 🎹 Клавишники \n📽 Медиа",
		"uk": "Відправ назву нової ролі. На жаль, поки що відредагувати або видалити роль неможливо, тому напиши без помилок.\n\n" +
			"Приклад:\n🎤 Вокалісти\n 🎹 Клавішники \n📽 Медіа",
	},
	"text.voiceDeleteConfirm": {
		"ru": "Удалить эту партию?",
		"uk": "Видалити цю партію?",
	},
	"text.eventDeleteConfirm": {
		"ru": "Удалить это собрание?",
		"uk": "Видалити цей захід?",
	},
	"text.songDeleteConfirm": {
		"ru": "Удалить эту песню?",
		"uk": "Видалити цю пісню?",
	},
	"text.eventDeleted": {
		"ru": "Собрание удалено.",
		"uk": "Захід видалено.",
	},
	"text.songDeleted": {
		"ru": "Песня удалена.",
		"uk": "Пісня видалена.",
	},
	"text.styled": {
		"ru": "Форматирование завершено.",
		"uk": "Форматування завершено.",
	},
	"text.addedLyricsPage": {
		"ru": "На вторую страницу добавлены слова (без аккордов).",
		"uk": "На другу сторінку додані слова (без акордів).",
	},
	"text.copiedToMyBand": {
		"ru": "Песня скопирована!",
		"uk": "Пісню скопійовано!",
	},
	"text.noStats": {
		"ru": "Статистика временно не доступна.",
		"uk": "Статистика тимчасово недоступна.",
	},
	"text.serverError": {
		"ru": "Произошла ошибка.",
		"uk": "Сталася помилка.",
	},
	"text.sendEmail": {
		"ru": "Теперь добавь имейл scala-drive@scala-chords-bot.iam.gserviceaccount.com в папку на Гугл Диске как редактора. После этого отправь мне ссылку на эту папку.",
		"uk": "Тепер додай імейл scala-drive@scala-chords-bot.iam.gserviceaccount.com в папку на Гугл Диску як редактора. Після цього відправ мені посилання на цю папку.",
	},
	"text.roleIndex": {
		"ru": "Роли выводятся в определенном порядке. После какой роли должна быть эта роль?",
		"uk": "Ролі виводяться у заданому порядку. Після якої ролі має бути ця роль?",
	},
	"text.noDocs": {
		"ru": "В папке на Google Диске нет документов.",
		"uk": "У папці Google Диск немає документів.",
	},
	"text.added": {
		"ru": "Добавлено.",
		"uk": "Додано.",
	},
	"text.noSongs": {
		"ru": "В этом собраннии еще нет списка песен.",
		"uk": "У цьому заході ще нема списку пісень.",
	},
	"text.addSongOrSetlist": {
		"ru": "Добавить песню или список",
		"uk": "Додати пісню або список",
	},
	"text.changeSongsOrderHint": {
		"ru": "Песни можно перетаскивать.",
		"uk": "Пісні можна перетаскувати.",
	},

	"command.menu": {
		"ru": "menu",
		"uk": "menu",
	},
	"text.section": {
		"ru": "Вместо %d страницы",
		"uk": "Замість %d сторінки",
	},
	"text.whereToStore": {
		"ru": "Куда сохранить новую тональность?",
		"uk": "Куди зберегти нову тональність?",
	},
	"text.docEnd": {
		"ru": "В конец документа",
		"uk": "В кінець документу",
	},
	"text.lyricsWithChords": {
		"ru": "Слова с аккордами",
		"uk": "Слова з акордами",
	},
	"text.newTag": {
		"ru": "Новый тег",
		"uk": "Новий тег",
	},
	"text.create": {
		"ru": "Создать",
		"uk": "Створити",
	},
	"text.save": {
		"ru": "Сохранить",
		"uk": "Зберегти",
	},
	"text.songAdded": {
		"ru": "Песня добавлена в список!",
		"uk": "Пісню додано в список!",
	},
	"text.songExists": {
		"ru": "Песня уже есть в списке.",
		"uk": "Пісня вже є у списку.",
	},
	"text.role": {
		"ru": "Роль",
		"uk": "Роль",
	},
	"text.weekday": {
		"ru": "День недели",
		"uk": "День тижня",
	},

	"text.mon": {
		"ru": "Понедельник",
		"uk": "Понеділок",
	},
	"text.tue": {
		"ru": "Вторник",
		"uk": "Вівторок",
	},
	"text.wed": {
		"ru": "Среда",
		"uk": "Середа",
	},
	"text.thu": {
		"ru": "Четверг",
		"uk": "Четвер",
	},
	"text.fri": {
		"ru": "Пятница",
		"uk": "П'ятниця",
	},
	"text.sat": {
		"ru": "Суббота",
		"uk": "Субота",
	},
	"text.sun": {
		"ru": "Воскресенье",
		"uk": "Неділя",
	},
	"text.name": {
		"ru": "Имя",
		"uk": "Ім'я",
	},
	"text.memCount": {
		"ru": "Количество участий",
		"uk": "Кількість участей",
	},
	"text.cleanupDatabase": {
		"ru": "Чистим базу",
		"uk": "Чистимо базу",
	},
	"text.period": {
		"ru": "⏳ : %s",
		"uk": "⏳ : %s",
	},
	"text.sorting": {
		"ru": "📈 : %s",
		"uk": "📈 : %s",
	},
	"text.period.lastHalfYear": {
		"ru": "за последние пол года",
		"uk": "за останні пів року",
	},
	"text.period.lastYear": {
		"ru": "за последний год",
		"uk": "за останній рік",
	},
	"text.period.allTime": {
		"ru": "за все время",
		"uk": "за весь час",
	},
	"text.period.lastThreeMonths": {
		"ru": "за последние три месяца",
		"uk": "за останні три місяці",
	},
	"text.sorting.ascending": {
		"ru": "по возрастанию",
		"uk": "за зростанням",
	},
	"text.sorting.descending": {
		"ru": "по убыванию",
		"uk": "за спаданням",
	},
}

func init() {
	for key, langToMsgMap := range locales {
		for lang, msg := range langToMsgMap {
			_ = message.SetString(language.Make(lang), key, msg)
		}
	}
}

func Get(key, lang string, a ...any) string {
	switch lang {
	case "ru":
		return ruPrinter.Sprintf(key, a...)
	}
	return ukPrinter.Sprintf(key, a...)
}

func GetTranslator(lang string) locales2.Translator {
	switch lang {
	case "ru":
		return ruTranslator
	}
	return ukTranslator
}

var (
	ukPrinter    = message.NewPrinter(language.Ukrainian)
	ukTranslator = uk.New()
)

var (
	ruPrinter    = message.NewPrinter(language.Russian)
	ruTranslator = ru.New()
)

// var enPrinter = message.NewPrinter(language.English)

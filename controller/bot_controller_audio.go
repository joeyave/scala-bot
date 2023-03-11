package controller

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/joeyave/scala-bot/controller/mysemaphore"
	"github.com/joeyave/scala-bot/entity"
	"github.com/joeyave/scala-bot/state"
	"github.com/joeyave/scala-bot/txt"
	"github.com/joeyave/scala-bot/util"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var ffmpegAudioExt = "mp3"

func (c *BotController) TransposeAudio_AskForSemitonesNumber(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	if user.State.Name == state.SongVoices_CreateVoice {
		return c.SongVoices_CreateVoice(user.State.Index)(bot, ctx)
	}

	semitones := 0
	//useR3 := false
	//skipClipping := false
	fine := false
	more := false
	if ctx.CallbackQuery != nil {
		payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
		split := strings.Split(payload, ":")
		_semitones, err := strconv.Atoi(split[0])
		if err != nil {
			return err
		}
		semitones = _semitones

		_fine, err := strconv.ParseBool(split[1])
		if err != nil {
			return err
		}
		fine = _fine

		if len(split) > 2 {
			_more, err := strconv.ParseBool(split[2])
			if err != nil {
				return err
			}
			more = _more
		}
	} else if ctx.EffectiveMessage.Audio != nil {
		audio := ctx.EffectiveMessage.Audio
		// todo: remove what's not needed.
		user.CallbackCache.IsVoice = false
		user.CallbackCache.AudioFileId = audio.FileId
		user.CallbackCache.AudioDuration = audio.Duration
		user.CallbackCache.AudioPerformer = audio.Performer
		user.CallbackCache.AudioTitle = audio.Title
		user.CallbackCache.AudioFileName = audio.FileName
		user.CallbackCache.AudioMimeType = audio.MimeType
		user.CallbackCache.AudioFileSize = audio.FileSize

		if audio.Thumb != nil {
			user.CallbackCache.AudioThumbFileId = audio.Thumb.FileId
			user.CallbackCache.AudioThumbFileUniqueId = audio.Thumb.FileUniqueId
			user.CallbackCache.AudioThumbWidth = audio.Thumb.Width
			user.CallbackCache.AudioThumbHeight = audio.Thumb.Height
			user.CallbackCache.AudioThumbFileSize = audio.Thumb.FileSize
		}
	} else if ctx.EffectiveMessage.Voice != nil {
		voice := ctx.EffectiveMessage.Voice
		// todo: remove what's not needed.
		user.CallbackCache.IsVoice = true
		user.CallbackCache.AudioFileId = voice.FileId
		user.CallbackCache.AudioDuration = voice.Duration
		user.CallbackCache.AudioMimeType = voice.MimeType
		user.CallbackCache.AudioFileSize = voice.FileSize
	}

	markup := &gotgbot.InlineKeyboardMarkup{}

	buttonText1 := txt.Get("button.fast", ctx.EffectiveUser.LanguageCode)
	buttonText2 := txt.Get("button.fine", ctx.EffectiveUser.LanguageCode)

	if !fine {
		buttonText1 = fmt.Sprintf("〔%s〕", buttonText1)
	} else {
		buttonText2 = fmt.Sprintf("〔%s〕", buttonText2)
	}

	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{
		{Text: buttonText1, CallbackData: util.CallbackData(state.TransposeAudio_AskForSemitonesNumber, fmt.Sprintf("%d:%t", semitones, false))},
		{Text: buttonText2, CallbackData: util.CallbackData(state.TransposeAudio_AskForSemitonesNumber, fmt.Sprintf("%d:%t", semitones, true))},
	})

	limit := 4
	if more {
		limit = 12
	}
	for i := 0; i > -limit; i-- {
		if i%4 == 0 || i == 0 {
			markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{})
		}
		buttonText := fmt.Sprintf("%d", i-1)
		if semitones == i-1 {
			buttonText = fmt.Sprintf("〔%s〕", buttonText)
		}
		markup.InlineKeyboard[len(markup.InlineKeyboard)-1] = append(markup.InlineKeyboard[len(markup.InlineKeyboard)-1], gotgbot.InlineKeyboardButton{Text: buttonText, CallbackData: util.CallbackData(state.TransposeAudio_AskForSemitonesNumber, fmt.Sprintf("%d:%t:%t", i-1, fine, more))})
	}
	for i := 0; i < limit; i++ {
		if i%4 == 0 || i == 0 {
			markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{})
		}

		buttonText := fmt.Sprintf("+%d", i+1)
		if semitones == i+1 {
			buttonText = fmt.Sprintf("〔%s〕", buttonText)
		}
		markup.InlineKeyboard[len(markup.InlineKeyboard)-1] = append(markup.InlineKeyboard[len(markup.InlineKeyboard)-1], gotgbot.InlineKeyboardButton{Text: buttonText, CallbackData: util.CallbackData(state.TransposeAudio_AskForSemitonesNumber, fmt.Sprintf("%d:%t:%t", i+1, fine, more))})
	}
	buttonText := "▿"
	if more {
		buttonText = "△"
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: buttonText, CallbackData: util.CallbackData(state.TransposeAudio_AskForSemitonesNumber, fmt.Sprintf("%d:%t:%t", semitones, fine, !more))}})

	s := strconv.Itoa(semitones)
	if !strings.HasPrefix(s, "-") {
		s = "+" + s
	}

	text := user.CallbackCache.AddToText(txt.Get("text.sendSemitones", ctx.EffectiveUser.LanguageCode))
	//text := user.CallbackCache.AddToText(txt.Get("text.sendSemitones", ctx.EffectiveUser.LanguageCode, user.CallbackCache.AudioFileName, s))

	if ctx.CallbackQuery != nil {
		if semitones != 0 {
			markup.InlineKeyboard = append(markup.InlineKeyboard,
				[]gotgbot.InlineKeyboardButton{
					{Text: txt.Get("button.continue", ctx.EffectiveUser.LanguageCode, s), CallbackData: util.CallbackData(state.TransposeAudio, fmt.Sprintf("%d:%t", semitones, fine))},
				})
		}

		//ctx.EffectiveMessage.EditText(bot, text, &gotgbot.EditMessageTextOpts{
		//	ReplyMarkup:           *markup,
		//	ParseMode:             "HTML",
		//	DisableWebPagePreview: true,
		//})
		ctx.EffectiveMessage.EditReplyMarkup(bot, &gotgbot.EditMessageReplyMarkupOpts{
			ReplyMarkup: *markup,
		})
		ctx.CallbackQuery.Answer(bot, nil)
	} else {
		_, err := ctx.EffectiveChat.SendMessage(bot, text, &gotgbot.SendMessageOpts{
			ReplyMarkup:           markup,
			ParseMode:             "HTML",
			DisableWebPagePreview: true,
			ReplyToMessageId:      ctx.EffectiveMessage.MessageId,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

var sem = mysemaphore.NewWeighted(100)

func (c *BotController) TransposeAudio(bot *gotgbot.Bot, ctx *ext.Context) error {

	ctx.CallbackQuery.Answer(bot, nil)

	user := ctx.Data["user"].(*entity.User)
	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")
	semitones := split[0]
	fine, err := strconv.ParseBool(split[1])
	if err != nil {
		return err
	}

	processingMsg, _, err := ctx.EffectiveMessage.EditText(bot, "Starting...", &gotgbot.EditMessageTextOpts{})
	if err != nil {
		return err
	}

	ctxWithCancel, stopSendingQueueMessages := context.WithCancel(context.Background())

	go func(id int64, ctx context.Context) {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		prevPos := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pos := sem.Position(id) + 1
				if pos != prevPos {
					processingMsg.EditText(bot, fmt.Sprintf("Position in queue: %d", pos), nil)
					//fmt.Println("start", id, "position in queue:", pos)
				}
				prevPos = pos
			}
		}
	}(ctx.EffectiveMessage.MessageId, ctxWithCancel)

	weight := int64(1)
	if !user.CallbackCache.IsVoice {
		mb := user.CallbackCache.AudioFileSize / 1000000
		coef := int64(2)
		if fine {
			coef = 5
		}
		weight = mb * coef
	}

	converted, newFileBytes, err := c.transposeAudio(bot, ctx, stopSendingQueueMessages, sem, weight, user.CallbackCache.AudioMimeType, user.CallbackCache.AudioFileId, semitones, fine, processingMsg)
	if err != nil {
		processingMsg.EditText(bot, fmt.Sprintf("Error: %v", err), nil)
		return err
	}

	ctx.EffectiveChat.SendAction(bot, "upload_document", nil)

	s := split[0]
	if !strings.HasPrefix(s, "-") {
		s = "+" + s
	}

	if user.CallbackCache.IsVoice {
		opts := &gotgbot.SendVoiceOpts{
			Duration: user.CallbackCache.AudioDuration,
		}
		_, err := bot.SendVoice(ctx.EffectiveChat.Id, newFileBytes, opts)
		if err != nil {
			return err
		}
	} else {
		opts := &gotgbot.SendAudioOpts{
			Duration:  user.CallbackCache.AudioDuration,
			Performer: user.CallbackCache.AudioPerformer,
			Thumb:     user.CallbackCache.AudioThumbFileId,
		}
		if user.CallbackCache.AudioTitle != "" {
			opts.Title = fmt.Sprintf("%s (%s)", user.CallbackCache.AudioTitle, s)
		}

		extension := filepath.Ext(user.CallbackCache.AudioFileName)
		fileName := strings.TrimSuffix(user.CallbackCache.AudioFileName, extension)
		if converted {
			extension = "." + ffmpegAudioExt
		}

		file := gotgbot.NamedFile{
			File:     bytes.NewReader(newFileBytes),
			FileName: fmt.Sprintf("%s (%s)%s", fileName, s, extension),
		}

		_, err = bot.SendAudio(ctx.EffectiveChat.Id, file, opts)
		if err != nil {
			return err
		}
	}

	processingMsg.Delete(bot, nil)

	return nil
}

func (c *BotController) transposeAudio(bot *gotgbot.Bot, ctx *ext.Context, stopQueueMessages context.CancelFunc, sem *mysemaphore.Weighted, weight int64, mimeType string, audioFileID string, semitones string, fine bool, processingMsg *gotgbot.Message) (bool, []byte, error) {
	err := sem.Acquire(context.TODO(), weight, ctx.EffectiveMessage.MessageId)
	if err != nil {
		sem.Release(weight)
		stopQueueMessages()
		return false, nil, err
	}
	defer sem.Release(weight)
	stopQueueMessages()

	processingMsg.EditText(bot, "Downloading...", nil)

	f, err := bot.GetFile(audioFileID, nil)
	if err != nil {
		return false, nil, err
	}

	originalFileBytes, err := util.File(bot, f)
	if err != nil {
		return false, nil, err
	}
	defer originalFileBytes.Close()

	inputTmpFile, err := os.CreateTemp("", "input_audio_*")
	if err != nil {
		return false, nil, err
	}
	defer os.Remove(inputTmpFile.Name())

	converted := false
	if mimeType == "audio/mp4" {
		processingMsg.EditText(bot, "Converting...", nil)
		converted = true
		if err := inputTmpFile.Close(); err != nil {
			return false, nil, err
		}

		err := ffmpeg.
			Input("pipe:").
			Output(inputTmpFile.Name(), ffmpeg.KwArgs{"f": ffmpegAudioExt, "c:v": "copy", "c:a": "libmp3lame", "q:a": "4"}).
			WithInput(originalFileBytes).
			OverWriteOutput().
			//ErrorToStdOut().
			Run()
		if err != nil {
			return false, nil, err
		}
	} else {
		if _, err := io.Copy(inputTmpFile, originalFileBytes); err != nil {
			return false, nil, err
		}
		if err := inputTmpFile.Close(); err != nil {
			return false, nil, err
		}
	}

	processingMsg.EditText(bot, "Processing...", nil)

	outTmpFile, err := os.CreateTemp("", "output_audio_*")
	if err != nil {
		return false, nil, err
	}
	defer os.Remove(outTmpFile.Name())

	if err := outTmpFile.Close(); err != nil {
		return false, nil, err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	args := []string{"-p", semitones}

	if fine {
		args = append(args, "-3")
	} else {
		args = append(args, "-2", "--ignore-clipping")
	}

	args = append(args, inputTmpFile.Name(), outTmpFile.Name())

	cmd := exec.CommandContext(ctxWithTimeout, "rubberband", args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return false, nil, err
	}
	if err := cmd.Start(); err != nil {
		return false, nil, err
	}

	go sendProgressToUser(stderr, bot, processingMsg)

	if err := cmd.Wait(); err != nil {
		return false, nil, err
	}

	newFileBytes, err := os.ReadFile(outTmpFile.Name())
	if err != nil {
		return false, nil, err
	}
	return converted, newFileBytes, nil
}

func sendProgressToUser(stderr io.Reader, bot *gotgbot.Bot, processingMsg *gotgbot.Message) {
	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanWords)

	processingStage := false
	currPercentage := 0

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for scanner.Scan() {
		//fmt.Println(scanner.Text())

		if scanner.Text() == "Processing..." {
			processingStage = true
		} else if strings.EqualFold(scanner.Text(), "NOTE:") { // Clipping detected.
			processingStage = false
		}

		percentage, err := strconv.Atoi(strings.TrimSuffix(scanner.Text(), "%"))
		if err != nil {
			continue
		}

		if processingStage && percentage > currPercentage {
			currPercentage = percentage

			select {
			case <-ticker.C:
				processingMsg.EditText(bot, "Processing... "+scanner.Text(), nil)
				//fmt.Println(wordsScanner.Text())
			default:
			}
		}
	}
}

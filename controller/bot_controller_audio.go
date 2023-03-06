package controller

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
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

	semitones := 0
	useR3 := false
	skipClipping := false
	more := false
	if ctx.CallbackQuery != nil {
		payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
		split := strings.Split(payload, ":")
		_semitones, err := strconv.Atoi(split[0])
		if err != nil {
			return err
		}
		semitones = _semitones

		_useR3, err := strconv.ParseBool(split[1])
		if err != nil {
			return err
		}
		useR3 = _useR3

		_skipClipping, err := strconv.ParseBool(split[2])
		if err != nil {
			return err
		}
		skipClipping = _skipClipping

		if len(split) > 3 {
			_more, err := strconv.ParseBool(split[3])
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

	buttonText := txt.Get("button.qualitatively", ctx.EffectiveUser.LanguageCode)
	if useR3 {
		buttonText += " ✅"
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: buttonText, CallbackData: util.CallbackData(state.TransposeAudio_AskForSemitonesNumber, fmt.Sprintf("%d:%t:%t", semitones, !useR3, skipClipping))}})

	buttonText = txt.Get("button.skipClippingCheck", ctx.EffectiveUser.LanguageCode)
	if skipClipping {
		buttonText += " ✅"
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: buttonText, CallbackData: util.CallbackData(state.TransposeAudio_AskForSemitonesNumber, fmt.Sprintf("%d:%t:%t", semitones, useR3, !skipClipping))}})

	limit := 4
	if more {
		limit = 12
	}
	//markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{})
	for i := 0; i > -limit; i-- {
		if i%4 == 0 || i == 0 {
			markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{})
		}
		buttonText = fmt.Sprintf("%d", i-1)
		if semitones == i-1 {
			buttonText = fmt.Sprintf("〔%s〕", buttonText)
		}
		markup.InlineKeyboard[len(markup.InlineKeyboard)-1] = append(markup.InlineKeyboard[len(markup.InlineKeyboard)-1], gotgbot.InlineKeyboardButton{Text: buttonText, CallbackData: util.CallbackData(state.TransposeAudio_AskForSemitonesNumber, fmt.Sprintf("%d:%t:%t:%t", i-1, useR3, skipClipping, more))})
	}
	//markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{})
	for i := 0; i < limit; i++ {
		if i%4 == 0 || i == 0 {
			markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{})
		}

		buttonText = fmt.Sprintf("+%d", i+1)
		if semitones == i+1 {
			buttonText = fmt.Sprintf("〔%s〕", buttonText)
		}
		markup.InlineKeyboard[len(markup.InlineKeyboard)-1] = append(markup.InlineKeyboard[len(markup.InlineKeyboard)-1], gotgbot.InlineKeyboardButton{Text: buttonText, CallbackData: util.CallbackData(state.TransposeAudio_AskForSemitonesNumber, fmt.Sprintf("%d:%t:%t:%t", i+1, useR3, skipClipping, more))})
	}
	buttonText = "▿"
	if more {
		buttonText = "△"
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard, []gotgbot.InlineKeyboardButton{{Text: buttonText, CallbackData: util.CallbackData(state.TransposeAudio_AskForSemitonesNumber, fmt.Sprintf("%d:%t:%t:%t", semitones, useR3, skipClipping, !more))}})

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
					{Text: txt.Get("button.continue", ctx.EffectiveUser.LanguageCode, s), CallbackData: util.CallbackData(state.TransposeAudio, fmt.Sprintf("%d:%t:%t", semitones, useR3, skipClipping))},
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

func (c *BotController) TransposeAudio(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)
	payload := util.ParseCallbackPayload(ctx.CallbackQuery.Data)
	split := strings.Split(payload, ":")

	processingMsg, _, err := ctx.EffectiveMessage.EditText(bot, "Processing...", &gotgbot.EditMessageTextOpts{})
	if err != nil {
		return err
	}

	f, err := bot.GetFile(user.CallbackCache.AudioFileId, nil)
	if err != nil {
		return err
	}

	originalFileBytes, err := util.File(bot, f)
	if err != nil {
		return err
	}
	defer originalFileBytes.Close()

	inputTmpFile, err := os.CreateTemp("", "input_audio_*")
	if err != nil {
		return err
	}
	defer os.Remove(inputTmpFile.Name())

	converted := false
	if user.CallbackCache.AudioMimeType == "audio/mp4" {
		converted = true
		if err := inputTmpFile.Close(); err != nil {
			return err
		}

		err = ffmpeg.
			Input("pipe:").
			Output(inputTmpFile.Name(), ffmpeg.KwArgs{"f": ffmpegAudioExt, "c:v": "copy", "c:a": "libmp3lame", "q:a": "4"}).
			WithInput(originalFileBytes).
			OverWriteOutput().
			ErrorToStdOut().
			Run()
		if err != nil {
			return err
		}
	} else {
		if _, err := io.Copy(inputTmpFile, originalFileBytes); err != nil {
			return err
		}
		if err := inputTmpFile.Close(); err != nil {
			return err
		}
	}

	outTmpFile, err := os.CreateTemp("", "output_audio_*")
	if err != nil {
		return err
	}
	defer os.Remove(outTmpFile.Name())

	if err := outTmpFile.Close(); err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	args := []string{"-p", split[0], "-F"}

	useR3, err := strconv.ParseBool(split[1])
	if err != nil {
		return err
	}
	if useR3 {
		args = append(args, "-3")
	}

	skipClipping, err := strconv.ParseBool(split[2])
	if err != nil {
		return err
	}
	if skipClipping {
		args = append(args, "--ignore-clipping")
	}

	args = append(args, inputTmpFile.Name(), outTmpFile.Name())

	cmd := exec.CommandContext(ctxWithTimeout, "rubberband-r3", args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	wordsScanner := bufio.NewScanner(stderr)
	wordsScanner.Split(bufio.ScanWords)

	go func() {
		processingStage := false
		currPercentage := 0

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for wordsScanner.Scan() {
			//fmt.Println(wordsScanner.Text())

			if wordsScanner.Text() == "Processing..." {
				processingStage = true
			} else if strings.EqualFold(wordsScanner.Text(), "NOTE:") { // Clipping detected.
				processingStage = false
			}

			percentage, err := strconv.Atoi(strings.TrimSuffix(wordsScanner.Text(), "%"))
			if err != nil {
				continue
			}

			if processingStage && percentage > currPercentage {
				currPercentage = percentage

				select {
				case <-ticker.C:
					processingMsg.EditText(bot, "Processing... "+wordsScanner.Text(), nil)
					fmt.Println(wordsScanner.Text())
				default:
				}
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return err
	}

	newFileBytes, err := os.ReadFile(outTmpFile.Name())
	if err != nil {
		return err
	}

	//newFileBytes, err := cmd.Output()
	//if err != nil {
	//	return err
	//}

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
			Title:     fmt.Sprintf("%s (%s)", user.CallbackCache.AudioTitle, s),
			Thumb:     user.CallbackCache.AudioThumbFileId,
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

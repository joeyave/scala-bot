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
	"io"
	"os"
	"os/exec"
	"time"
)

func (c *BotController) TransposeAudio_AskForSemitonesNumber(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	markup := &gotgbot.ReplyKeyboardMarkup{
		Keyboard: [][]gotgbot.KeyboardButton{
			{{Text: "1"}, {Text: "2"}, {Text: "3"}, {Text: "4"}, {Text: "5"}, {Text: "6"}, {Text: "7"}, {Text: "8"}},
			{{Text: "-1"}, {Text: "-2"}, {Text: "-3"}, {Text: "-4"}, {Text: "-5"}, {Text: "-6"}, {Text: "-7"}, {Text: "-8"}},
			{{Text: txt.Get("button.menu", ctx.EffectiveUser.LanguageCode)}},
		},
		ResizeKeyboard: true,
	}
	_, err := ctx.EffectiveChat.SendMessage(bot, txt.Get("text.sendSemitones", ctx.EffectiveUser.LanguageCode), &gotgbot.SendMessageOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		return err
	}

	user.State = entity.State{
		Name: state.TransposeAudio,
	}
	user.Cache = entity.Cache{
		Audio: ctx.EffectiveMessage.Audio,
	}

	//_, err = c.UserService.UpdateOne(*user)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (c *BotController) TransposeAudio(bot *gotgbot.Bot, ctx *ext.Context) error {

	user := ctx.Data["user"].(*entity.User)

	ctx.EffectiveChat.SendMessage(bot, "Processing...", nil)

	audio := user.Cache.Audio
	f, err := bot.GetFile(audio.FileId, nil)
	if err != nil {
		return err
	}

	reader, err := util.File(bot, f)
	if err != nil {
		return err
	}

	inputTmpFile, err := os.CreateTemp("", "input_audio_*")
	if err != nil {
		return err
	}
	defer os.Remove(inputTmpFile.Name())

	if _, err := io.Copy(inputTmpFile, reader); err != nil {
		return err
	}
	if err := inputTmpFile.Close(); err != nil {
		return err
	}

	outTmpFile, err := os.CreateTemp("", "output_audio_*")
	if err != nil {
		return err
	}
	defer os.Remove(outTmpFile.Name())

	if err := outTmpFile.Close(); err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctxWithTimeout, "rubberband-r3", "-p", ctx.EffectiveMessage.Text, inputTmpFile.Name(), outTmpFile.Name())
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	scanner = bufio.NewScanner(stderr)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
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

	opts := &gotgbot.SendAudioOpts{
		Duration:  audio.Duration,
		Performer: audio.Performer,
		Title:     audio.Title,
	}
	if audio.Thumb != nil {
		thumbFileID := gotgbot.InputFile(audio.Thumb.FileId)
		opts.Thumb = &thumbFileID
	}
	_, err = bot.SendAudio(ctx.EffectiveChat.Id, &gotgbot.NamedFile{
		File:     bytes.NewReader(newFileBytes),
		FileName: fmt.Sprintf("%s", audio.FileName),
	}, opts)
	if err != nil {
		return err
	}

	return nil
}

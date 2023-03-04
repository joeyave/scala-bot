package controller

import (
	"bufio"
	"bytes"
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

	audio := user.Cache.Audio
	f, err := bot.GetFile(audio.FileId, nil)
	if err != nil {
		return err
	}

	reader, err := util.File(bot, f)
	if err != nil {
		return err
	}

	// Write the input data to a temporary file
	tmpFile, err := os.CreateTemp("", "input")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, reader); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	ctx.EffectiveChat.SendMessage(bot, "Processing file. It can take some time.", nil)

	cmd := exec.Command("rubberband", "-p", ctx.EffectiveMessage.Text, tmpFile.Name(), "-")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return err
	}
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		return err
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
		return err
	}

	var stdBuffer bytes.Buffer
	cmd.Stderr = &stdBuffer

	//newFileBytes, err := cmd.Output()
	//// todo: remove.
	//ctx.EffectiveChat.SendMessage(bot, string(stdBuffer.Bytes()), nil)
	//if err != nil {
	//	return err
	//}
	//
	//ctx.EffectiveChat.SendAction(bot, "upload_document", nil)
	//
	//opts := &gotgbot.SendAudioOpts{
	//	Duration:  audio.Duration,
	//	Performer: audio.Performer,
	//	Title:     audio.Title,
	//}
	//if audio.Thumb != nil {
	//	thumbFileID := gotgbot.InputFile(audio.Thumb.FileId)
	//	opts.Thumb = &thumbFileID
	//}
	//_, err = bot.SendAudio(ctx.EffectiveChat.Id, &gotgbot.NamedFile{
	//	File:     bytes.NewReader(newFileBytes),
	//	FileName: fmt.Sprintf("%s", audio.FileName),
	//}, opts)
	//if err != nil {
	//	return err
	//}

	return nil
}

package main

import (
	"context"
	"fmt"
	api "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/patrick-me/tg-bot/client"
	pb "github.com/patrick-me/tg-bot/proto"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	bot, err := api.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic("TELEGRAM_APITOKEN isn't found as an env variable", err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := api.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		fmt.Println(update)
		processUpdateMessage(bot, update)
		processCallbackQuery(bot, update)
	}
}

func processCallbackQuery(bot *api.BotAPI, update api.Update) {
	if update.CallbackQuery != nil {
		// Respond to the callback query, telling Telegram to show the user
		// a message with the data received.

		callback := api.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
		if _, err := bot.Request(callback); err != nil {
			fmt.Println(err)
			return
		}
		response, err := processCallback(update)

		if err != nil {
			sendErrorMessage(bot, update, err)
			return
		}
		editMsg(bot, update, response)
	}
}

func processUpdateMessage(bot *api.BotAPI, update api.Update) {
	if update.Message != nil {
		log.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)
		serverMessage, err := process(update)

		if err != nil {
			sendErrorMessage(bot, update, err)
			return
		}
		sendMsg(bot, update, serverMessage)
	}
}

func createInlineKeyboard(keyboard *pb.Keyboard) api.InlineKeyboardMarkup {
	kb := api.NewInlineKeyboardMarkup()

	for _, row := range keyboard.GetRows() {
		rw := api.NewInlineKeyboardRow()
		for _, button := range row.GetButtons() {
			button := api.NewInlineKeyboardButtonData(button.GetName(), button.GetValue())
			rw = append(rw, button)
		}
		kb.InlineKeyboard = append(kb.InlineKeyboard, rw)
	}
	return api.NewInlineKeyboardMarkup(kb.InlineKeyboard...)
}

func createInlineKeyboardLink(keyboard *pb.Keyboard) *api.InlineKeyboardMarkup {
	kb := createInlineKeyboard(keyboard)
	return &api.InlineKeyboardMarkup{
		InlineKeyboard: kb.InlineKeyboard,
	}
}

func sendMsg(bot *api.BotAPI, update api.Update, serverMessage *pb.ProxyResponse) {
	message := serverMessage.GetMessage()
	length := len(message)
	var msgs []string

	msgs = splitLongLengthMessages(length, msgs, message)

	for id, m := range msgs {
		msg := api.NewMessage(update.Message.Chat.ID, m)
		msg.ReplyToMessageID = update.Message.MessageID
		applyToOnlyTheFirstMessage := id == 0

		if serverMessage.ApplyReplyMarkupKeyboard && applyToOnlyTheFirstMessage {
			msg.ReplyMarkup = createInlineKeyboard(serverMessage.GetReplyMarkupKeyboard())
		}

		if serverMessage.ApplyMarkdownV2 {
			msg.ParseMode = "MarkdownV2"
		}

		send, err := bot.Send(msg)
		log.Println(send)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func editMsg(bot *api.BotAPI, update api.Update, serverMessage *pb.ProxyResponse) {

	message := serverMessage.GetMessage()
	length := len(message)
	var msgs []string
	msgs = splitLongLengthMessages(length, msgs, message)

	for id, m := range msgs {
		msg := api.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			m)

		applyToOnlyTheFirstMessage := id == 0

		if serverMessage.ApplyReplyMarkupKeyboard && applyToOnlyTheFirstMessage {
			msg.ReplyMarkup = createInlineKeyboardLink(serverMessage.GetReplyMarkupKeyboard())
		} else if applyToOnlyTheFirstMessage {
			msg.ReplyMarkup = update.CallbackQuery.Message.ReplyMarkup
		}

		if serverMessage.ApplyMarkdownV2 {
			msg.ParseMode = "MarkdownV2"
		}

		send, err := bot.Send(msg)
		log.Println(send)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func splitLongLengthMessages(length int, msgs []string, message string) []string {
	const TgMsgMaxLen = 4096
	if length <= TgMsgMaxLen {
		msgs = append(msgs, message)
	} else {
		splits := strings.Split(message, "\n")
		newMsg := ""
		for _, part := range splits {
			if len(part)+1 > TgMsgMaxLen {
				continue
			}

			if len(newMsg)+len(part)+1 <= TgMsgMaxLen {
				newMsg += part + "\n"
			} else {
				msgs = append(msgs, newMsg)
				newMsg = part + "\n"
			}
		}

		if len(newMsg) > 0 {
			msgs = append(msgs, newMsg)
			newMsg = ""
		}
	}
	return msgs
}

func sendErrorMessage(bot *api.BotAPI, update api.Update, err error) {
	msg := api.NewMessage(update.Message.Chat.ID, err.Error())
	msg.ReplyToMessageID = update.Message.MessageID
	send, err := bot.Send(msg)
	log.Println(send)
	if err != nil {
		log.Println(err)
		return
	}
}

func process(update api.Update) (response *pb.ProxyResponse, err error) {
	client := client.CreateProxyClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err = client.Process(ctx, &pb.ProxyRequest{
		Message:   update.Message.Text,
		Username:  update.Message.From.UserName,
		MessageID: int64(update.Message.MessageID),
		ChatID:    update.Message.Chat.ID,
	})

	if err != nil {
		log.Printf("could not process: %v", err)
		err = fmt.Errorf("can't process msg:%s, err:%v", update.Message.Text, err)
		return
	}
	return
}

func processCallback(update api.Update) (response *pb.ProxyResponse, err error) {
	client := client.CreateProxyClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	message := update.CallbackQuery.Message
	var sourceUserMessage = message.ReplyToMessage.Text

	response, err = client.Process(ctx, &pb.ProxyRequest{
		Message:      sourceUserMessage,
		MessageID:    int64(message.ReplyToMessage.MessageID),
		Username:     message.From.UserName,
		ChatID:       message.ReplyToMessage.Chat.ID,
		CallbackData: update.CallbackQuery.Data,
	})

	if err != nil {
		log.Printf("could not process: %v", err)
		err = fmt.Errorf("can't process msg:%s, err:%v", update.Message.Text, err)
		return
	}
	return
}

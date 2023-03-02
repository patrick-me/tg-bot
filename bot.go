package main

import (
	"context"
	"fmt"
	api "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	pb "github.com/patrick-me/tg-bot/proto"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	bot, err := api.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := api.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)

			serverMessage, err := process(update.Message.Text)

			if err != nil {
				sendErrorMessage(bot, update, err)
				continue
			}

			sendMsg(bot, update, serverMessage)
		}
	}
}

func sendMsg(bot *api.BotAPI, update api.Update, serverMessage *pb.ProxyResponse) {

	message := serverMessage.GetMessage()
	length := len(message)
	var msgs []string

	msgs = splitLongLengthMessages(length, msgs, message)

	for _, m := range msgs {
		msg := api.NewMessage(update.Message.Chat.ID, m)
		msg.ReplyToMessageID = update.Message.MessageID

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

func process(msg string) (response *pb.ProxyResponse, err error) {
	client := CreateProxyClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	response, err = client.Process(ctx, &pb.ProxyRequest{Message: msg})
	if err != nil {
		log.Printf("could not process: %v", err)
		err = fmt.Errorf("Can't process msg:%s, err:%v", msg, err)
		return
	}
	return
}

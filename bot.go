package main

import (
	"context"
	"fmt"
	api "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	proxy "tg-bot/proto"
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
			}

			msg := api.NewMessage(update.Message.Chat.ID, serverMessage.GetMessage())
			msg.ReplyToMessageID = update.Message.MessageID

			if serverMessage.ApplyMarkdownV2 {
				msg.ParseMode = "MarkdownV2"
			}
			send, err := bot.Send(msg)
			log.Println(send)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
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

func process(msg string) (response *proxy.ProxyResponse, err error) {
	client := CreateProxyClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err = client.Process(ctx, &proxy.ProxyRequest{Message: msg})
	if err != nil {
		log.Printf("could not process: %v", err)
		err = fmt.Errorf("Can't process msg:%s, err:%v", msg, err)
		return
	}
	return
}

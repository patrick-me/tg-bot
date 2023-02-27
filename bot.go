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

			msg := api.NewMessage(update.Message.Chat.ID, process(update.Message.Text))
			msg.ReplyToMessageID = update.Message.MessageID
			//msg.ParseMode = "MarkdownV2"
			bot.Send(msg)
		}
	}
}

func process(msg string) string {
	client := CreateProxyClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := client.Process(ctx, &proxy.ProxyRequest{Message: msg})
	if err != nil {
		log.Printf("could not process: %v", err)
		return fmt.Sprintf("Can't process msg:%s, err:%v", msg, err)
	}
	return r.GetMessage()
}

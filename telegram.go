package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	telegramChatIDs = map[int64]bool{}
	bot             *tgbotapi.BotAPI
)

func init() {
	var err error
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if len(token) == 0 {
		log.Fatal("telegram token not set")
	}
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
}

func botRun() {
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		chatID := update.Message.Chat.ID

		log.Printf("new chat id %v", chatID)
		// store chat id
		telegramChatIDs[chatID] = true

		if err := sendMessage(chatID, "I will notify you of new announces"); err != nil {
			log.Println(err)
		}
	}
}

func sendToAllChats(text string) {
	for chatID := range telegramChatIDs {
		chatID := chatID
		go func() {
			if err := sendMessage(chatID, text); err != nil {
				log.Println(err)
			}
		}()
	}
}

func sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	return err
}

package main

import (
	"context"
	"log"
	"moviesbot/handle"
	"os"
	"os/signal"
	"syscall"
	"time"

	"moviesbot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	connStr := "user=godb password=0208 dbname=moviesbot sslmode=disable"
	db, err := storage.OpenDatabase(connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//botToken := "7022304427:AAHlYfBrdysAYZNzrFAyZ1laDzCD7I-N_Hk"

	botToken := "6902655696:AAHnk8V-ly7Q3W57QFNvOlatOTAZrgE4juA"
	botInstance, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	offset := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down bot...")
			return
		default:
			updates, err := botInstance.GetUpdates(tgbotapi.NewUpdate(offset))
			if err != nil {
				log.Printf("Error getting updates: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}
			for _, update := range updates {
				handle.HandleUpdate(update, db, botInstance)
				offset = update.UpdateID + 1
			}
		}
	}
}

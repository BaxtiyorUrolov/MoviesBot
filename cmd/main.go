package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"moviesbot/admin"
	"moviesbot/movie"
	"moviesbot/state"
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

	botToken := "6902655696:AAEtKAL78CG86DhjAYb-QVQrTVAGysTpLDA"
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
				handleUpdate(update, db, botInstance)
				offset = update.UpdateID + 1
			}
		}
	}
}

func handleUpdate(update tgbotapi.Update, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	if update.Message != nil {
		handleMessage(update.Message, db, botInstance)
	} else if update.CallbackQuery != nil {
		handleCallbackQuery(update.CallbackQuery, db, botInstance)
	} else {
		log.Printf("Unsupported update type: %T", update)
	}
}

func handleMessage(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	text := msg.Text

	log.Printf("Received message: %s", text)

	if userState, exists := state.UserStates[chatID]; exists {
		switch userState {
		case "waiting_for_broadcast_message":
			admin.HandleBroadcastMessage(msg, db, botInstance)
			delete(state.UserStates, chatID)
			return
		case "waiting_for_channel_link":
			admin.HandleChannelLink(msg, db, botInstance)
			delete(state.UserStates, chatID)
			return
		case "waiting_for_admin_id":
			admin.HandleAdminAdd(msg, db, botInstance)
			delete(state.UserStates, chatID)
			return
		case "waiting_for_admin_id_remove":
			admin.HandleAdminRemove(msg, db, botInstance)
			delete(state.UserStates, chatID)
			return
		case "waiting_for_movie_id":
			movie.HandleMovieID(msg, db, botInstance)
			state.UserStates[chatID] = "waiting_for_movie_link"
			return
		case "waiting_for_movie_link":
			movie.HandleMovieLink(msg, db, botInstance)
			state.UserStates[chatID] = "waiting_for_movie_title"
			return
		case "waiting_for_movie_title":
			movie.HandleMovieTitle(msg, db, botInstance)
			delete(state.UserStates, chatID)
			return
		case "waiting_for_search_movie_id":
			movie.HandleSearchMovieID(msg, db, botInstance)
			delete(state.UserStates, chatID)
			return
		}
	}

	if text == "/start" {
		handleStartCommand(msg, db, botInstance)
		storage.AddUserToDatabase(db, int(msg.Chat.ID))
	} else if text == "/admin" {
		admin.HandleAdminCommand(msg, db, botInstance)
	} else {
		handleDefaultMessage(msg, db, botInstance)
	}
}

func handleStartCommand(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	userID := msg.From.ID
	firstName := msg.From.FirstName

	log.Printf("Adding user to database: %d ", userID)
	err := storage.AddUserToDatabase(db, userID)
	if err != nil {
		log.Printf("Error adding user to database: %v", err)
		return
	}

	channels, err := storage.GetChannelsFromDatabase(db)
	if err != nil {
		log.Printf("Error getting channels from database: %v", err)
		return
	}

	if isUserSubscribedToChannels(chatID, channels, botInstance) {
		photoURL := "https://t.me/photolabsuz/15"
		welcomeMessage := fmt.Sprintf("👋 Assalomu alaykum [%s](tg://user?id=%d) botimizga xush kelibsiz.\n\n✍🏻 Kino kodini yuboring.", firstName, userID)

		// Send the photo
		photo := tgbotapi.NewPhotoShare(chatID, photoURL)
		photo.Caption = welcomeMessage
		photo.ParseMode = "Markdown" // Ensure the message is parsed as Markdown
		_, err = botInstance.Send(photo)
		if err != nil {
			log.Printf("Error sending photo: %v", err)
			return
		}

		// Send inline keyboard
		startTestButton := tgbotapi.NewInlineKeyboardButtonData("Kino izlash", "kino_izlash")
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(startTestButton),
		)
		msgReply := tgbotapi.NewMessage(chatID, "")
		msgReply.ReplyMarkup = inlineKeyboard
		botInstance.Send(msgReply)
	} else {
		msg := tgbotapi.NewMessage(chatID, "Iltimos, kanallarga azo bo'ling.")
		inlineKeyboard := createSubscriptionKeyboard(channels)
		msg.ReplyMarkup = inlineKeyboard
		botInstance.Send(msg)
	}
}

func handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := callbackQuery.Message.Chat.ID
	messageID := callbackQuery.Message.MessageID
	firstName := callbackQuery.From.FirstName
	userID := callbackQuery.From.ID

	channels, err := storage.GetChannelsFromDatabase(db)
	if err != nil {
		log.Printf("Error getting channels from database: %v", err)
		return
	}

	if callbackQuery.Data == "check_subscription" {
		if isUserSubscribedToChannels(chatID, channels, botInstance) {
			deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
			botInstance.Send(deleteMsg)

			photoURL := "https://t.me/photolabsuz/15"
			welcomeMessage := fmt.Sprintf("👋 Assalomu alaykum [%s](tg://user?id=%d) botimizga xush kelibsiz.\n\n✍🏻 Kino kodini yuboring.", firstName, userID)

			// Send the photo
			photo := tgbotapi.NewPhotoShare(chatID, photoURL)
			photo.Caption = welcomeMessage
			photo.ParseMode = "Markdown" // Ensure the message is parsed as Markdown
			_, err = botInstance.Send(photo)
			if err != nil {
				log.Printf("Error sending photo: %v", err)
				return
			}
		} else {
			msg := tgbotapi.NewMessage(chatID, "Iltimos, kanallarga azo bo'ling.")
			inlineKeyboard := createSubscriptionKeyboard(channels)
			msg.ReplyMarkup = inlineKeyboard
			botInstance.Send(msg)
		}
	} else if strings.HasPrefix(callbackQuery.Data, "delete_channel_") {
		channel := strings.TrimPrefix(callbackQuery.Data, "delete_channel_")
		admin.AskForChannelDeletionConfirmation(chatID, messageID, channel, botInstance)
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
		botInstance.Send(deleteMsg)
	} else if strings.HasPrefix(callbackQuery.Data, "confirm_delete_channel_") {
		channel := strings.TrimPrefix(callbackQuery.Data, "confirm_delete_channel_")
		admin.DeleteChannel(chatID, messageID, channel, db, botInstance)
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
		botInstance.Send(deleteMsg)
	} else if callbackQuery.Data == "cancel_delete_channel" {
		admin.CancelChannelDeletion(chatID, messageID, botInstance)
	} else if callbackQuery.Data == "kino_izlash" {
		state.UserStates[chatID] = "waiting_for_search_movie_id"
		msg := tgbotapi.NewMessage(chatID, "Iltimos, kino ID sini kiriting:")
		botInstance.Send(msg)
	}
}

func handleDefaultMessage(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	text := msg.Text

	switch text {
	case "Kanal qo'shish":
		state.UserStates[chatID] = "waiting_for_channel_link"
		msgResponse := tgbotapi.NewMessage(chatID, "Kanal linkini yuboring (masalan, https://t.me/your_channel):")
		botInstance.Send(msgResponse)
	case "Admin qo'shish":
		state.UserStates[chatID] = "waiting_for_admin_id"
		msgResponse := tgbotapi.NewMessage(chatID, "Iltimos, yangi admin ID sini yuboring:")
		botInstance.Send(msgResponse)
	case "Admin o'chirish":
		state.UserStates[chatID] = "waiting_for_admin_id_remove"
		msgResponse := tgbotapi.NewMessage(chatID, "Iltimos, admin ID sini o'chirish uchun yuboring:")
		botInstance.Send(msgResponse)
	case "Kanal o'chirish":
		admin.DisplayChannelsForDeletion(chatID, db, botInstance)
	case "Statistika":
		admin.HandleStatistics(msg, db, botInstance)
	case "Habar yuborish":
		state.UserStates[chatID] = "waiting_for_broadcast_message"
		msgResponse := tgbotapi.NewMessage(chatID, "Iltimos, yubormoqchi bo'lgan habaringizni kiriting (Bekor qilish uchun /cancel):")
		botInstance.Send(msgResponse)
	case "Kino yuklash":
		state.UserStates[chatID] = "waiting_for_movie_id"
		msgResponse := tgbotapi.NewMessage(chatID, "Iltimos, kino ID sini kiriting:")
		botInstance.Send(msgResponse)
	default:
		movie.HandleSearchMovieID(msg, db, botInstance)
	}
}

func isUserSubscribedToChannels(chatID int64, channels []string, botInstance *tgbotapi.BotAPI) bool {
	for _, channel := range channels {
		log.Printf("Checking subscription to channel: %s", channel)
		chat, err := botInstance.GetChat(tgbotapi.ChatConfig{SuperGroupUsername: "@" + channel})
		if err != nil {
			log.Printf("Error getting chat info for channel %s: %v", channel, err)
			return false
		}

		member, err := botInstance.GetChatMember(tgbotapi.ChatConfigWithUser{
			ChatID: chat.ID,
			UserID: int(chatID),
		})
		if err != nil {
			log.Printf("Error getting chat member info for channel %s: %v", channel, err)
			return false
		}
		if member.Status == "left" || member.Status == "kicked" {
			log.Printf("User %d is not subscribed to channel %s", chatID, channel)
			return false
		}
	}
	return true
}

func createSubscriptionKeyboard(channels []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, channel := range channels {
		channelName := strings.TrimPrefix(channel, "https://t.me/")
		button := tgbotapi.NewInlineKeyboardButtonURL("Kanalga azo bo'lish", "https://t.me/"+channelName)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}
	checkButton := tgbotapi.NewInlineKeyboardButtonData("Azo bo'ldim", "check_subscription")
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(checkButton))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

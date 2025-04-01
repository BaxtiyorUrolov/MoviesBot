package movie

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"moviesbot/state"
	"moviesbot/storage"
)

func HandleMovieID(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID

	if !storage.IsAdmin(int(chatID), db) {
		return
	}

	err := storage.AddMovieIDToDatabase(db, msg.Text)
	if err != nil {
		log.Printf("Error adding movie ID to database: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Kino ID sini qo'shishda xatolik yuz berdi.")
		botInstance.Send(msgResponse)
		return
	}

	state.MovieStates[chatID] = msg.Text // Store the movie ID in the state

	msgResponse := tgbotapi.NewMessage(chatID, "Kino linkini kiriting:")
	botInstance.Send(msgResponse)
}

func HandleMovieLink(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	link := msg.Text

	movieID := state.MovieStates[chatID] // Retrieve the movie ID from the state
	err := storage.AddMovieLinkToDatabase(db, movieID, link)
	if err != nil {
		log.Printf("Error adding movie link to database: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Kino linkini qo'shishda xatolik yuz berdi.")
		botInstance.Send(msgResponse)
		return
	}

	msgResponse := tgbotapi.NewMessage(chatID, "Kino nomini kiriting:")
	botInstance.Send(msgResponse)
}

func HandleMovieTitle(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	title := msg.Text

	movieID := state.MovieStates[chatID] // Retrieve the movie ID from the state
	err := storage.AddMovieTitleToDatabase(db, movieID, title)
	if err != nil {
		log.Printf("Error adding movie title to database: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Kino nomini qo'shishda xatolik yuz berdi.")
		botInstance.Send(msgResponse)
		return
	}

	msgResponse := tgbotapi.NewMessage(chatID, "Kino muvaffaqiyatli qo'shildi.")
	botInstance.Send(msgResponse)
}

func HandleSearchMovieID(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID

	movie, err := storage.GetMovieByID(db, msg.Text)
	if err != nil {
		log.Printf("Error retrieving movie: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Kino topilmadi.")
		botInstance.Send(msgResponse)
		return
	}

	video := tgbotapi.NewVideoShare(chatID, movie.Link)
	caption := fmt.Sprintf("Mana siz izlagan kino.\n\n\n Bot tayyorlatish uchun: @BaxtiyorUrolov")
	video.Caption = caption

	_, err = botInstance.Send(video)
	if err != nil {
		log.Printf("Error sending video: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Videoni yuborishda xatolik yuz berdi.")
		botInstance.Send(msgResponse)
		return
	}
}

func HandleDeleteMovie(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	movieID := msg.Text

	if !storage.IsAdmin(int(chatID), db) {
		msgResponse := tgbotapi.NewMessage(chatID, "Siz admin emassiz.")
		botInstance.Send(msgResponse)
		return
	}

	err := storage.DeleteMovie(db, movieID)
	if err != nil {
		msgResponse := tgbotapi.NewMessage(chatID, "Kino o'chirishda xatolik yuz berdi.")
		botInstance.Send(msgResponse)
		log.Printf("Error deleting movie: %v", err)
		return
	}

	msgResponse := tgbotapi.NewMessage(chatID, "Kino muvaffaqiyatli o'chirildi.")
	botInstance.Send(msgResponse)
	delete(state.MovieStates, chatID) // Clear the state for this chat
	return
}

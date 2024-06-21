package movie

import (
	"database/sql"
	"fmt"
	"log"
	"moviesbot/state"
	"moviesbot/storage"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func HandleMovieID(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID

	if !storage.IsAdmin(int(chatID), db) {
		return
	}

	movieID, err := strconv.ParseInt(msg.Text, 10, 64)
	if err != nil {
		log.Printf("Error parsing movie ID: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Noto'g'ri kino ID formati.")
		botInstance.Send(msgResponse)
		return
	}

	err = storage.AddMovieIDToDatabase(db, movieID)
	if err != nil {
		log.Printf("Error adding movie ID to database: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Kino ID sini qo'shishda xatolik yuz berdi.")
		botInstance.Send(msgResponse)
		return
	}

	state.MovieStates[chatID] = movieID // Store the movie ID in the state

	msgResponse := tgbotapi.NewMessage(chatID, "Kino linkini kiriting:")
	botInstance.Send(msgResponse)
}

func HandleMovieLink(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	link := msg.Text

	movieID := state.MovieStates[chatID] // Retrieve the movie ID from the state
	fmt.Println("link uchun id ", movieID)
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
	fmt.Println("kino id: ", movieID)
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
	movieID, err := strconv.ParseInt(msg.Text, 10, 64)
	if err != nil {
		log.Printf("Error parsing movie ID: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Noto'g'ri kino ID formati.")
		botInstance.Send(msgResponse)
		return
	}

	movie, err := storage.GetMovieByID(db, movieID)
	if err != nil {
		log.Printf("Error retrieving movie: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Kino topilmadi.")
		botInstance.Send(msgResponse)
		return
	}

	video := tgbotapi.NewVideoShare(chatID, movie.Link)
	caption := fmt.Sprintf("Kino nomi: %s\n\nBot manzili: @MovieTVuz_Bot \n\nBizning loyihalar: @MRC_GROUPUZ", movie.Title)
	video.Caption = caption

	_, err = botInstance.Send(video)
	if err != nil {
		log.Printf("Error sending video: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Videoni yuborishda xatolik yuz berdi.")
		botInstance.Send(msgResponse)
		return
	}
}

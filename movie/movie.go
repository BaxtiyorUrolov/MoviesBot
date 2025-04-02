package movie

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"moviesbot/state"
	"moviesbot/storage"
	"strings"
)

func HandleMovieID(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID

	if !storage.IsAdmin(int(chatID), db) {
		return
	}

	// Instagram URL dan ID ni ajratib olish
	inputText := msg.Text
	var movieID string

	// URL da "/reel/" qismini qidirish
	if strings.Contains(inputText, "/reel/") {
		parts := strings.Split(inputText, "/reel/")
		if len(parts) > 1 {
			// "?" belgisi bo'lsa, undan oldingi qismni olish
			idPart := parts[1]
			if strings.Contains(idPart, "?") {
				movieID = strings.Split(idPart, "?")[0]
			} else {
				movieID = idPart
			}
		}
	} else {
		movieID = inputText // Agar URL bo'lmasa, xabar matnini o'zini ishlatish
	}

	if movieID == "" {
		msgResponse := tgbotapi.NewMessage(chatID, "To'g'ri Instagram reel URL kiriting.")
		botInstance.Send(msgResponse)
		return
	}

	err := storage.AddMovieIDToDatabase(db, movieID)
	if err != nil {
		log.Printf("Error adding movie ID to database: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Kino ID sini qo'shishda xatolik yuz berdi.")
		botInstance.Send(msgResponse)
		return
	}

	state.MovieStates[chatID] = movieID // Faqat ID ni saqlash

	msgResponse := tgbotapi.NewMessage(chatID, "Kino linkini kiriting:")
	botInstance.Send(msgResponse)
}

func HandleMovieLink(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	link := msg.Text

	movieID := state.MovieStates[chatID] // Retrieve the movie ID from the state
	err := storage.AddMovieLinkToDatabase(db, movieID, link)
	if err != nil {
		msgResponse := tgbotapi.NewMessage(chatID, "Kino linkini qo'shishda xatolik yuz berdi: "+err.Error())
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
		msgResponse := tgbotapi.NewMessage(chatID, "Kino nomini qo'shishda xatolik yuz berdi: "+err.Error())
		botInstance.Send(msgResponse)
		return
	}

	msgResponse := tgbotapi.NewMessage(chatID, "Kino muvaffaqiyatli qo'shildi.")
	botInstance.Send(msgResponse)
}

func HandleSearchMovieID(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID

	// Instagram URL dan ID ni ajratib olish
	inputText := msg.Text
	var movieID string

	// URL da "/reel/" qismini qidirish
	if strings.Contains(inputText, "/reel/") {
		parts := strings.Split(inputText, "/reel/")
		if len(parts) > 1 {
			// "?" belgisi bo'lsa, undan oldingi qismni olish
			idPart := parts[1]
			if strings.Contains(idPart, "?") {
				movieID = strings.Split(idPart, "?")[0]
			} else {
				movieID = idPart
			}
		}
	} else {
		movieID = inputText // Agar URL bo'lmasa, xabar matnini o'zini ishlatish
	}

	if movieID == "" {
		msgResponse := tgbotapi.NewMessage(chatID, "To'g'ri Instagram reel URL yoki ID kiriting.")
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
	caption := fmt.Sprintf("Mana siz izlagan kino.\n\n\n Bot tayyorlatish uchun: @BaxtiyorUrolov")
	video.Caption = caption

	_, err = botInstance.Send(video)
	if err != nil {
		log.Printf("Error sending video: %v", err)
		msgResponse := tgbotapi.NewMessage(chatID, "Videoni yuborishda xatolik yuz berdi: "+err.Error())
		botInstance.Send(msgResponse)
		return
	}
}

func HandleDeleteMovie(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	inputText := msg.Text

	if !storage.IsAdmin(int(chatID), db) {
		msgResponse := tgbotapi.NewMessage(chatID, "Siz admin emassiz.")
		botInstance.Send(msgResponse)
		return
	}

	// Instagram URL dan ID ni ajratib olish
	var movieID string
	if strings.Contains(inputText, "/reel/") {
		parts := strings.Split(inputText, "/reel/")
		if len(parts) > 1 {
			// "?" belgisi bo'lsa, undan oldingi qismni olish
			idPart := parts[1]
			if strings.Contains(idPart, "?") {
				movieID = strings.Split(idPart, "?")[0]
			} else {
				movieID = idPart
			}
		}
	} else {
		movieID = inputText // Agar URL bo'lmasa, xabar matnini o'zini ishlatish
	}

	if movieID == "" {
		msgResponse := tgbotapi.NewMessage(chatID, "To'g'ri Instagram reel URL yoki ID kiriting.")
		botInstance.Send(msgResponse)
		return
	}

	err := storage.DeleteMovie(db, movieID)
	if err != nil {
		msgResponse := tgbotapi.NewMessage(chatID, "Kino o'chirishda xatolik yuz berdi: "+err.Error())
		botInstance.Send(msgResponse)
		log.Printf("Error deleting movie: %v", err)
		return
	}

	msgResponse := tgbotapi.NewMessage(chatID, "Kino muvaffaqiyatli o'chirildi.")
	botInstance.Send(msgResponse)
	delete(state.MovieStates, chatID) // Clear the state for this chat
	return
}

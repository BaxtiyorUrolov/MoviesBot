package storage

import (
	"database/sql"
	"errors"
	"log"
	"moviesbot/models"
)

func AddMovieIDToDatabase(db *sql.DB, movieID string) error {
	query := `INSERT INTO movies (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`
	_, err := db.Exec(query, movieID)
	if err != nil {
		log.Println("Kino ID sini qo'shishda xtolik:", err)
		return err
	}
	return nil
}

func AddMovieLinkToDatabase(db *sql.DB, movieID string, link string) error {
	query := `UPDATE movies SET link = $2 WHERE id = $1`
	_, err := db.Exec(query, movieID, link)
	if err != nil {
		log.Println("Kino linkini qo'shishda xatolik:", err)
		return err
	}
	return nil
}

func AddMovieTitleToDatabase(db *sql.DB, movieID string, title string) error {

	query := `UPDATE movies SET title = $2 WHERE id = $1`
	result, err := db.Exec(query, movieID, title)
	if err != nil {
		log.Println("Kino nomini qo'shishda xatolik:", err)
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("no movie found to update")
	}
	return nil
}

func GetMovieByID(db *sql.DB, movieID string) (*models.Movie, error) {
	query := `SELECT id, link, title FROM movies WHERE id = $1`
	row := db.QueryRow(query, movieID)

	var movie models.Movie
	err := row.Scan(&movie.ID, &movie.Link, &movie.Title)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("movie not found")
		}
		log.Println("Kino ID sini olishda xatolik:", err)
		return nil, err
	}

	return &movie, nil
}

func DeleteMovie(db *sql.DB, movieID string) error {
	query := `DELETE FROM movies WHERE id = $1`
	_, err := db.Exec(query, movieID)
	if err != nil {
		log.Println("Kino o'chirishda xatolik:", err)
		return err
	}
	return nil
}

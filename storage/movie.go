package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"moviesbot/models"
)

func AddMovieIDToDatabase(db *sql.DB, movieID int64) error {
	query := `INSERT INTO movies (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`
	_, err := db.Exec(query, movieID)
	return err
}

func AddMovieLinkToDatabase(db *sql.DB, movieID int64, link string) error {
	fmt.Println("link uchun id ", movieID)
	query := `UPDATE movies SET link = $2 WHERE id = $1`
	_, err := db.Exec(query, movieID, link)
	return err
}

func AddMovieTitleToDatabase(db *sql.DB, movieID int64, title string) error {
	fmt.Println("kino id: ", movieID)

	query := `UPDATE movies SET title = $2 WHERE id = $1`
	result, err := db.Exec(query, movieID, title)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("no movie found to update")
	}
	return nil
}

func GetMovieByID(db *sql.DB, movieID int64) (*models.Movie, error) {
	query := `SELECT id, link, title FROM movies WHERE id = $1`
	row := db.QueryRow(query, movieID)

	var movie models.Movie
	err := row.Scan(&movie.ID, &movie.Link, &movie.Title)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("movie not found")
		}
		return nil, err
	}

	return &movie, nil
}

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

func AddMovieGenreToDatabase(db *sql.DB, movieID int64, genre string) error {
	query := `UPDATE movies SET genre = $2 WHERE id = $1`
	result, err := db.Exec(query, movieID, genre)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("no movie found to update")
	}
	return nil
}

func AddMovieReleaseYearToDatabase(db *sql.DB, movieID int64, releaseYear int) error {
	query := `UPDATE movies SET release_year = $2 WHERE id = $1`
	result, err := db.Exec(query, movieID, releaseYear)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("no movie found to update")
	}
	return nil
}


func GetMovieByID(db *sql.DB, movieID int64) (models.Movie, error) {
	var movie models.Movie
	query := `SELECT id, link, title, genre, release_year FROM movies WHERE id = $1`
	row := db.QueryRow(query, movieID)
	err := row.Scan(&movie.ID, &movie.Link, &movie.Title, &movie.Genre, &movie.ReleaseYear)
	if err != nil {
		if err == sql.ErrNoRows {
			return movie, errors.New("movie not found")
		}
		return movie, err
	}
	return movie, nil
}
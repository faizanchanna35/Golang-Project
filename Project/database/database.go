

package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

func Connect() (db *sql.DB) {

	cfg := PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "faizan",
		Database: "Authentication",
		SSLMode:  "disable",
	}

	db, err := sql.Open("pgx", cfg.String())

	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`
	
	
	CREATE TABLE IF NOT EXISTS posts(
			id SERIAL PRIMARY KEY,
			title TEXT,
			description TEXT,
			user_id integer,
			image TEXT
		);
	
		CREATE TABLE IF NOT EXISTS contact(
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		description TEXT,
		user_id integer
	);

	
	`)


	// _, err = db.Exec(`

	// CREATE TABLE IF NOT EXISTS posts(
	// 	id SERIAL PRIMARY KEY,
	// 	title TEXT,
	// 	description TEXT UNIQUE NOT NULL,
	// 	user_id integer
	// )
	
    // `)

	if err != nil {
		panic(err)
	} 
	return
}

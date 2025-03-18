// config/config.go
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL      string
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	APIURL     string
	ServerPort int
}

func LoadConfig() (*Config, error) {
	godotenv.Load()

	apiURL := os.Getenv("API_URL")
	serverPortStr := os.Getenv("SERVER_PORT")
	serverPort, err := strconv.Atoi(serverPortStr)
	if err != nil {
		serverPort = 8080
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbHost := os.Getenv("DB_HOST")
		dbPortStr := os.Getenv("DB_PORT")
		var dbPort int
		dbPort, err = strconv.Atoi(dbPortStr)
		if err != nil {
			dbPort = 5432
		}
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName) // Используем dbPort, объявленный выше
	}

	parsedDBURL, err := url.Parse(dbURL)
	if err != nil {
		return nil, err
	}

	dbHost := parsedDBURL.Hostname()
	dbPortParsed, _ := strconv.Atoi(parsedDBURL.Port())
	dbUser := parsedDBURL.User.Username()
	dbPassword, _ := parsedDBURL.User.Password()
	dbName := strings.TrimPrefix(parsedDBURL.Path, "/")

	return &Config{
		DBURL:      dbURL,
		DBHost:     dbHost,
		DBPort:     dbPortParsed,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBName:     dbName,
		APIURL:     apiURL,
		ServerPort: serverPort,
	}, nil
}

package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

const urlsTable = "urls"

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type PostgresDB struct {
	db *sql.DB
}

// Функция NewPostgresDB создает новый объект PostgresDB
func NewPostgresDB(cfg Config) (*PostgresDB, error) {
	// Строка подключения к БД
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Database, cfg.Password))
	if err != nil {
		return nil, err
	}
	// Проверка подключения к БД
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	// Создание таблицы urls
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (short_url varchar(255) PRIMARY KEY, long_url varchar(255) not null unique);", urlsTable)
	// Выполнение запроса
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	return &PostgresDB{db: db}, nil
}

// Функция GetUrl возвращает длинный URL по короткому
func (postgres *PostgresDB) GetUrl(shortUrl string) (string, error) {
	if shortUrl == "" {
		return "", errors.New("неправильные исходные данные для чтения из БД")
	}
	// Запрос на получение длинного URL по короткому
	query := fmt.Sprintf("SELECT long_url FROM %s WHERE short_url = $1", urlsTable)
	row := postgres.db.QueryRow(query, shortUrl)

	var longUrl string
	err := row.Scan(&longUrl)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return longUrl, nil
}

// Функция GetShortUrl возвращает короткий URL по длинному
func (postgres *PostgresDB) GetShortUrl(longUrl string) (string, error) {
	if longUrl == "" {
		return "", errors.New("неправильные исходные данные для чтения из БД")
	}
	query := fmt.Sprintf("SELECT short_url FROM %s WHERE long_url = $1", urlsTable)
	row := postgres.db.QueryRow(query, longUrl)

	var shortUrl string
	err := row.Scan(&shortUrl)

	if err != nil {
		return "", err
	}
	return shortUrl, nil
}

// Функция PostUrl добавляет новую запись в БД
func (postgres *PostgresDB) PostUrl(shortUrl string, longUrl string) error {
	if shortUrl == "" || longUrl == "" {
		return errors.New("неправильные исходные данные для записи в БД")
	}
	// Запрос на добавление новой записи в БД
	query := fmt.Sprintf("SELECT short_url FROM %s WHERE short_url = $1 OR long_url = $2", urlsTable)
	row := postgres.db.QueryRow(query, shortUrl, longUrl)

	var url string
	err := row.Scan(&url)

	if err == sql.ErrNoRows {
		query = fmt.Sprintf("INSERT INTO %s (short_url, long_url) values ($1, $2);", urlsTable)
		_, err = postgres.db.Query(query, shortUrl, longUrl)
		return err
	} else if err != nil {
		return err
	}
	// Если запись с таким коротким URL уже существует
	short, _ := postgres.GetShortUrl(longUrl)
	return errors.New("http://localhost:8080/" + short)
}

// Функция Close закрывает соединение с БД
func (postgres *PostgresDB) Close() error {
	return postgres.db.Close()
}

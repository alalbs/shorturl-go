package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"prod/pkg/postgres"

	"github.com/gin-gonic/gin"
)

// Объявляем переменные для хранения данных
var urls = make(map[string]string)
var db *postgres.PostgresDB

// Переменная для определения режима работы
var boolPtr = flag.Bool("d", false, "a bool")

func main() {
	flag.Parse()
	defer db.Close()

	//Определяем режим работы
	if *boolPtr {
		//Подключаемся к БД
		var err error
		db, err = postgres.NewPostgresDB(GetConf())
		fmt.Println("postgres mode")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(db)
		}
	} else {
		fmt.Println("memory mode")
	}
	//Запускаем сервер
	router := gin.Default()
	router.GET("/:hash", getUrls)
	router.POST("/", postUrl)

	router.Run(":8080")
}

// Обратка GET запроса
func getUrls(c *gin.Context) {
	//Получаем хеш из запроса
	hash := c.Param("hash")
	var url string
	var ok bool
	var err error
	//Проверяем режим работы
	if *boolPtr {
		//Получаем данные из БД по хешу
		url, err = db.GetUrl(hash)
	} else {
		//Получаем данные из памяти по хешу
		url, ok = urls[hash]
	}
	//Проверяем наличие данных
	if ok || err == nil {
		c.Redirect(http.StatusMovedPermanently, url)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "not found"})
	}

}

// Обратка Post запроса
func postUrl(c *gin.Context) {
	var newUrl string
	//Получаем данные из запроса
	if err := c.BindJSON(&newUrl); err != nil {
		return
	}
	//Получаем хеш
	hash := GetHash(newUrl)
	//Проверяем режим работы
	if *boolPtr {
		//Сохраняем данные в БД
		db.PostUrl(hash, newUrl)
	} else {
		//Сохраняем данные в памяти
		urls[hash] = newUrl
	}
	//Отправляем сокращенный URL
	c.IndentedJSON(http.StatusCreated, "http://localhost:8080/"+hash)
}

// Функция для получения хеша
func GetHash(url string) string {
	h := sha256.New()
	//Получаем хеш при помощи алгоритма sha256
	h.Write([]byte(url))
	//Кодируем хеш в base64
	sha := base64.URLEncoding.EncodeToString(h.Sum(nil)[:15])
	return sha
}

// Функция для настройки подключения к БД
func GetConf() postgres.Config {
	file, exist := os.Open("../config/config.json")
	if exist != nil {
		fmt.Println(exist)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	conf := postgres.Config{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("error:", err)
	}
	return conf

}

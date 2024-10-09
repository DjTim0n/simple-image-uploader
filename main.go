package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/mattn/go-sqlite3"
)

// Структура для хранения данных изображения
type ImageData struct {
	Hash     string
	FileName string
	URL      string
}

func hashFileFromBytes(fileData io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, fileData); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func uploadImage(c *fiber.Ctx, db *sql.DB) error {
	file, err := c.FormFile("image")
	if err != nil {
		log.Println("Error in uploading Image : ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	// Открываем файл без сохранения его в систему, чтобы вычислить хеш
	fileReader, err := file.Open()
	if err != nil {
		log.Println("Error in opening Image :", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}
	defer fileReader.Close()

	// Хешируем файл
	fileHash, err := hashFileFromBytes(fileReader)
	if err != nil {
		log.Println("Error in hashing Image :", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	// Проверяем, есть ли запись с таким хешем в базе данных
	var existingImage ImageData
	err = db.QueryRow("SELECT hash, file_name, url FROM images WHERE hash = ?", fileHash).Scan(&existingImage.Hash, &existingImage.FileName, &existingImage.URL)
	if err == nil {
		// Если запись существует в базе, возвращаем URL
		return c.JSON(fiber.Map{"status": 200, "message": "Image already exists", "data": fiber.Map{"imageUrl": existingImage.URL}})
	}

	// Если файл не найден в базе данных, сохраняем его
	savePath := c.Params("path", "")
	if savePath == "" {
		savePath = "anon"
	}

	if err := os.MkdirAll("images/"+savePath, os.ModePerm); err != nil {
		log.Println("Error creating directory: ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Error creating directory", "data": nil})
	}

	// Генерация читабельного имени файла
	dirtFilename := strings.Split(file.Filename, ".")[0]
	fileName := strings.ReplaceAll(dirtFilename, " ", "_")
	fileExt := strings.Split(file.Filename, ".")[1]
	readableFileName := fmt.Sprintf("%s.%s", fileName, fileExt)

	finalFilePath := filepath.Join("images/"+savePath, readableFileName)

	// Сохраняем файл
	err = c.SaveFile(file, finalFilePath)
	if err != nil {
		log.Println("Error in saving Image :", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	imageUrl := fmt.Sprintf("http://localhost:4000/images/%s/%s", savePath, readableFileName)

	// Сохраняем запись о новом изображении в базу данных
	_, err = db.Exec("INSERT INTO images (hash, file_name, url) VALUES (?, ?, ?)", fileHash, readableFileName, imageUrl)
	if err != nil {
		log.Println("Error in inserting image data to DB:", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	data := map[string]interface{}{
		"imageName": readableFileName,
		"imageUrl":  imageUrl,
		"header":    file.Header,
		"size":      file.Size,
	}

	return c.JSON(fiber.Map{"status": 201, "message": "Image uploaded successfully", "data": data})
}

func main() {
	// Подключение к базе данных SQLite
	db, err := sql.Open("sqlite3", "./images.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем таблицу, если ее еще нет
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash TEXT NOT NULL UNIQUE,
		file_name TEXT NOT NULL,
		url TEXT NOT NULL
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024,
	})
	app.Static("/images", "./images")

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
	}))

	// Маршрут с параметром пути
	app.Post("/upload/:path", func(c *fiber.Ctx) error {
		return uploadImage(c, db)
	})

	app.Post("/upload", func(c *fiber.Ctx) error {
		return uploadImage(c, db)
	})

	if err := app.Listen(":4000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

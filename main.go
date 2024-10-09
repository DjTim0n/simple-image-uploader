package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func hashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func uploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("image")

	if err != nil {
		log.Println("Error in uploading Image : ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	savePath := c.Params("path", "")
	if savePath == "" {
		savePath = "anon"
	}

	if err := os.MkdirAll("images/"+savePath, os.ModePerm); err != nil {
		log.Println("Error creating directory: ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Error creating directory", "data": nil})
	}

	tempFilePath := filepath.Join("images/"+savePath, file.Filename)
	err = c.SaveFile(file, tempFilePath)

	if err != nil {
		log.Println("Error in saving temporary Image :", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	fileHash, err := hashFile(tempFilePath)
	if err != nil {
		log.Println("Error in hashing Image :", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	hashFilePath := filepath.Join("images/"+savePath, fileHash)
	if _, err := os.Stat(hashFilePath); err == nil {
		imageUrl := fmt.Sprintf("http://localhost:4000/images/%s/%s", savePath, fileHash)
		return c.JSON(fiber.Map{"status": 200, "message": "Image already exists", "data": fiber.Map{"imageUrl": imageUrl}})
	}

	dirtFilename := strings.Split(file.Filename, ".")[0]
	fileName := strings.ReplaceAll(dirtFilename, " ", "_")
	fileExt := strings.Split(file.Filename, ".")[1]

	readableFileName := fmt.Sprintf("%s.%s", fileName, fileExt)

	finalFilePath := filepath.Join("images/"+savePath, readableFileName)
	err = os.Rename(tempFilePath, finalFilePath)
	if err != nil {
		log.Println("Error in renaming Image :", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	imageUrl := fmt.Sprintf("http://localhost:4000/images/%s/%s", savePath, readableFileName)

	data := map[string]interface{}{
		"imageName": readableFileName,
		"imageUrl":  imageUrl,
		"header":    file.Header,
		"size":      file.Size,
	}

	return c.JSON(fiber.Map{"status": 201, "message": "Image uploaded successfully", "data": data})
}

func main() {

	app := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024,
	})
	app.Static("/images", "./images")

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
	}))

	app.Post("/upload/:path", uploadImage)

	app.Post("/upload", uploadImage)

	if err := app.Listen(":4000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

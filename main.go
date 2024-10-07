package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
)

func uploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("image")

	if err != nil {
		log.Println("Error in uploading Image : ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	uniqueId := uuid.New()

	filename := strings.Replace(uniqueId.String(), "-", "", -1)

	fileExt := strings.Split(file.Filename, ".")[1]

	image := fmt.Sprintf("%s.%s", filename, fileExt)

	err = c.SaveFile(file, fmt.Sprintf("./images/%s", image))

	if err != nil {
		log.Println("Error in saving Image :", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	imageUrl := fmt.Sprintf("https://imageservice.tim-space.kz/images/%s", image)

	data := map[string]interface{}{
		"imageName": image,
		"imageUrl":  imageUrl,
		"header":    file.Header,
		"size":      file.Size,
	}

	return c.JSON(fiber.Map{"status": 201, "message": "Image uploaded successfully", "data": data})
}

func main() {

	app := fiber.New()

	app.Static("/images", "./images")

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "*",
	}))

	app.Post("/upload", uploadImage)

	if err := app.Listen(":4000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

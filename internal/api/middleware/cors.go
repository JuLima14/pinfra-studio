package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173,http://localhost:3000",
		AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	})
}

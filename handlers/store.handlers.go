package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func HandleGetWebSocket(ctx *fiber.Ctx) error {
	ctx.Write([]byte("ok"))
	return nil
}

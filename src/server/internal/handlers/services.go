package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/ivankprod/ivankprod.ru/src/server/internal/domain"
)

func HandlerServicesIndex(c *fiber.Ctx) error {
	uAuth, ok := c.Locals("user_auth").(*domain.User)
	if !ok {
		uAuth = nil
	}

	data := make(fiber.Map)

	if uAuth != nil {
		data["user"] = *uAuth
	}

	err := c.Render("services", fiber.Map{
		"urlBase":      c.BaseURL(),
		"urlCanonical": c.BaseURL() + c.Path(),
		"pageTitle":    "Сервисы - " + os.Getenv("INFO_TITLE_BASE"),
		"pageDesc":     os.Getenv("INFO_DESC_BASE"),
		"pageScope":    "services",
		"ogTags": fiber.Map{
			"title": "Услуги - " + os.Getenv("INFO_TITLE_BASE"),
			"type":  "website",
		},
		"activeServices": true,
		"data":           data,
	})

	return err
}

package routes

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"ivankprod.ru/src/backend/modules/models"
	"ivankprod.ru/src/backend/modules/utils"
)

func RouteHomeIndex(c *fiber.Ctx) error {
	uAuth, ok := c.Locals("user_auth").(*models.User)
	if !ok {
		uAuth = nil
	}

	data := make(fiber.Map)

	if uAuth != nil {
		data["user"] = *uAuth
	}

	err := c.Render("index", fiber.Map{
		"urlBase":      c.BaseURL(),
		"urlCanonical": c.BaseURL() + c.Path(),
		"pageTitle":    "Главная - " + os.Getenv("INFO_TITLE_BASE"),
		"pageDesc":     os.Getenv("INFO_DESC_BASE"),
		"pageScope":    "home",
		"ogTags": fiber.Map{
			"title": os.Getenv("INFO_TITLE_BASE"),
		},
		"activeHome": true,
		"data":       data,
	})
	if err == nil {
		go utils.Logger(c.Request().URI().String(), c.IP(), 200)

		return nil
	}

	return fiber.NewError(fiber.StatusNotFound, "Страница не найдена!")
}

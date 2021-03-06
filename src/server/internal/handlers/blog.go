package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/ivankprod/ivankprod.ru/src/server/internal/domain"
)

func HandlerBlogIndex(c *fiber.Ctx) error {
	uAuth, ok := c.Locals("user_auth").(*domain.User)
	if !ok {
		uAuth = nil
	}

	data := make(fiber.Map)

	if uAuth != nil {
		data["user"] = *uAuth
	}

	err := c.Render("blog", fiber.Map{
		"urlBase":      c.BaseURL(),
		"urlCanonical": c.BaseURL() + c.Path(),
		"pageTitle":    "Блог - " + os.Getenv("INFO_TITLE_BASE"),
		"pageDesc":     os.Getenv("INFO_DESC_BASE"),
		"pageScope":    "blog",
		"ogTags": fiber.Map{
			"title": "Блог - " + os.Getenv("INFO_TITLE_BASE"),
			"type":  "website",
		},
		"activeBlog": true,
		"data":       data,
	})

	return err
}

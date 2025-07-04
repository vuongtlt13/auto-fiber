package autofiber_test

import (
	"github.com/gofiber/fiber/v2"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

// newTestApp returns an AutoFiber app with a standardized error handler for testing
func newTestApp() *autofiber.AutoFiber {
	return autofiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			switch e := err.(type) {
			case *autofiber.ValidationRequestError:
				code := 400
				if e.Message == "Validation failed" {
					code = 422
				}
				return c.Status(code).JSON(e)
			case *autofiber.ValidationResponseError:
				return c.Status(500).JSON(e)
			default:
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
		},
	})
}

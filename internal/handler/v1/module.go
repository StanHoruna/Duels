package v1

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("v1",
		fx.Provide(
			NewAuthHandler,
			NewUserHandler,
		),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler) {
			authHandler.RegisterRoutes(app)
		}),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler, userHandler *UserHandler) {
			userHandler.RegisterRoutes(app, authHandler)
		}),
	)
}

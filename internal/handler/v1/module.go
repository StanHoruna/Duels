package v1

import (
	"duels-api/internal/handler/v1/swagger"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("v1",
		fx.Provide(
			NewAuthHandler,
			NewUserHandler,
			NewDuelHandler,
			swagger.NewSwaggerHandler,
		),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler) {
			authHandler.RegisterRoutes(app)
		}),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler, userHandler *UserHandler) {
			userHandler.RegisterRoutes(app, authHandler)
		}),
		fx.Invoke(func(app *fiber.App, authHandler *AuthHandler, duelHandler *DuelHandler) {
			duelHandler.RegisterRoutes(app, authHandler)
		}),
		fx.Invoke(func(app *fiber.App, swaggerHandler *swagger.Handler) {
			swaggerHandler.RegisterRoutes(app)
		}),
	)
}

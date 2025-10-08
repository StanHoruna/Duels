package v1

import (
	"duels-api/internal/service"
	"duels-api/pkg/apperrors"
	auth "duels-api/pkg/jwt"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	UserService *service.UserService
}

func NewUserHandler(
	userService *service.UserService,
) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

func (h *UserHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	userGroup := app.Group("/user")

	userGroup.Use(auth.AuthMiddleware)
	{
		userGroup.Get("/", h.GetUser)
		userGroup.Put("/profile-picture", h.SetProfilePicture)
	}
}

func (h *UserHandler) GetUser(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	user, err := h.UserService.GetByID(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"user": user})
}

func (h *UserHandler) SetProfilePicture(c fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	imageURL, err := h.UserService.UpdateProfilePicture(c.Context(), claims.UserID, file)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"image_url": imageURL,
	})
}

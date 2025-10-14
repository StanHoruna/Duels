package v1

import (
	"duels-api/internal/service"
	"duels-api/pkg/apperrors"
	auth "duels-api/pkg/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"time"
)

var (
	userImageUploadCache = make(map[uuid.UUID]int64)
)

type AuthHandler struct {
	UserService *service.UserService
	JWTService  *service.JWTService
}

func NewAuthHandler(us *service.UserService,
	jwtService *service.JWTService) *AuthHandler {
	return &AuthHandler{
		UserService: us,
		JWTService:  jwtService,
	}
}

func (h *AuthHandler) RegisterRoutes(app *fiber.App) {
	authGroup := app.Group("/auth")
	{
		authGroup.Post("/sign-in-wallet", h.SignInWithWallet)

		authGroup.Post("/refresh", h.RefreshTokens)
	}
}

func (h *AuthHandler) AuthMiddleware(c fiber.Ctx) error {
	token := c.Get("Authorization")

	claims, err := h.JWTService.ParseToken(token)
	if err != nil {
		return apperrors.Unauthorized("failed to parse token", err)
	}

	if claims == nil {
		return apperrors.Internal("failed to generate claims")
	}

	c.Locals("claims", *claims)

	return c.Next()
}

func (h *AuthHandler) UploadUserImageMiddleware(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	now := time.Now().Unix()
	lastUploadTime, ok := userImageUploadCache[claims.UserID]
	if ok {
		// add 60 seconds to last image upload
		if lastUploadTime+60 > now {
			return apperrors.TooManyRequests("only 3 images per minute allowed")
		}
	}

	userImageUploadCache[claims.UserID] = now

	return c.Next()
}

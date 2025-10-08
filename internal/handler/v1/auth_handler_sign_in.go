package v1

import (
	"duels-api/internal/model"
	"duels-api/pkg/apperrors"
	auth "duels-api/pkg/jwt"
	"github.com/gofiber/fiber/v3"
)

func (h *AuthHandler) SignInWithWallet(c fiber.Ctx) error {
	var req model.AuthWithWallet
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}

	user, err := h.UserService.SignInWithWallet(c.Context(), req)
	if err != nil {
		return err
	}

	claims, ok := auth.NewTokenClaimsByUser(user)
	if !ok {
		return apperrors.Internal("failed to generate token claims for user")
	}

	tokenPair, err := h.JWTService.GenerateTokenPair(c.Context(), claims)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"user":     user,
		"jwt_info": tokenPair,
	})
}

func (h *AuthHandler) RefreshTokens(c fiber.Ctx) error {
	token := c.Get("Authorization")

	tokenPair, err := h.JWTService.RefreshSession(c.Context(), token)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"jwt_info": tokenPair,
	})
}

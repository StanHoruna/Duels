package v1

import (
	"duels-api/internal/model"
	"duels-api/pkg/apperrors"
	auth "duels-api/pkg/jwt"
	"github.com/gofiber/fiber/v3"
)

// SignInWithWallet godoc
//
//	@Summary		Sign in with crypto wallet
//	@Description	Authenticates a user using a connected crypto wallet (e.g., Solana Phantom). On success returns the user profile and a new JWT token pair.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.AuthWithWallet							true	"Wallet sign-in payload"
//	@Success		200		{object}	object{user=model.User,jwt_info=auth.TokenPair}	"Authenticated successfully"
//	@Failure		400		{object}	apperrors.ErrorPublic							"Invalid request body"
//	@Failure		401		{object}	apperrors.ErrorPublic							"Unauthorized - invalid wallet data/signature"
//	@Failure		500		{object}	apperrors.ErrorPublic							"Internal server error"
//	@Router			/auth/sign-in-wallet [post]
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

// RefreshTokens godoc
//
//	@Summary		Refresh JWT tokens
//	@Description	Refreshes the user's session using the refresh token supplied in the Authorization header. Returns a new access/refresh token pair.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string							true	"Bearer refresh token"	default(Bearer <refresh_token>)
//	@Success		200				{object}	object{jwt_info=auth.TokenPair}	"Tokens refreshed"
//	@Failure		401				{object}	apperrors.ErrorPublic			"Unauthorized - invalid or expired refresh token"
//	@Failure		500				{object}	apperrors.ErrorPublic			"Internal server error"
//	@Router			/auth/refresh [post]
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

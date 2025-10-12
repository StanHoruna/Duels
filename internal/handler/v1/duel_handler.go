package v1

import (
	"duels-api/internal/model"
	"duels-api/internal/service"
	"duels-api/pkg/apperrors"
	auth "duels-api/pkg/jwt"
	"github.com/gofiber/fiber/v3"
)

type DuelHandler struct {
	DuelService *service.DuelService
}

func NewDuelHandler(
	duelService *service.DuelService,
) (*DuelHandler, error) {
	authHandler := &DuelHandler{
		DuelService: duelService,
	}

	return authHandler, nil
}

func (h *DuelHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	cryptoDuelGroup := app.Group("/crypto-duel", auth.AuthMiddleware)
	{

		solanaDuelGroup := cryptoDuelGroup.Group("/solana")
		{
			solanaDuelGroup.Post("/", h.CreateExternalWalletCryptoDuel)
			solanaDuelGroup.Post("/sign-tx", h.SignCreateCryptoDuelTransaction)
			solanaDuelGroup.Post("/join", h.JoinExternalWalletCryptoDuel)
			solanaDuelGroup.Post("/join/sign-tx", h.SignJoinCryptoDuelTransaction)
			solanaDuelGroup.Put("/resolve", h.ResolveCryptoDuelByOwner)
		}

	}
}

func (h *DuelHandler) CreateExternalWalletCryptoDuel(c fiber.Ctx) error {
	var req model.CreateDuelReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	resp, err := h.DuelService.CreateCryptoDuel(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (h *DuelHandler) SignCreateCryptoDuelTransaction(c fiber.Ctx) error {
	var req model.CreateDuelReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	tx, err := h.DuelService.SignCreateCryptoDuelTransaction(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"tx": tx})
}

func (h *DuelHandler) JoinExternalWalletCryptoDuel(c fiber.Ctx) error {
	var req model.JoinDuelReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	resp, err := h.DuelService.JoinExternalWalletCryptoDuel(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (h *DuelHandler) SignJoinCryptoDuelTransaction(c fiber.Ctx) error {
	var req model.JoinDuelReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	tx, err := h.DuelService.SignJoinCryptoDuelTransaction(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"tx": tx})
}

func (h *DuelHandler) ResolveCryptoDuelByOwner(c fiber.Ctx) error {
	var req model.DuelResolveReq
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	txHashes, err := h.DuelService.ResolveCryptoDuelByOwner(c.Context(), claims.UserID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"tx_hashes": txHashes})
}

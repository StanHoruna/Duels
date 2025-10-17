package v1

import (
	"duels-api/internal/model"
	"duels-api/internal/service"
	"duels-api/pkg/apperrors"
	auth "duels-api/pkg/jwt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
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
	cryptoDuel := app.Group("/crypto-duel", auth.AuthMiddleware)
	{
		solana := cryptoDuel.Group("/solana")
		{
			solana.Post("/", h.CreateExternalWalletCryptoDuel)         // create after client sent tx
			solana.Post("/sign-tx", h.SignCreateCryptoDuelTransaction) // server builds raw init tx
			solana.Post("/join", h.JoinExternalWalletCryptoDuel)       // join after client sent tx
			solana.Post("/join/sign-tx", h.SignJoinCryptoDuelTransaction)
			solana.Put("/resolve", h.ResolveCryptoDuelByOwner) // payouts from admin wallet
		}
	}

	public := app.Group("/duel/public")
	{
		public.Get("/all", h.GetAllDuelsPublic)
		public.Get("/count", h.CountAllDuelsPublic)
		public.Get("/:id", h.GetDuelByIDPublic)
	}

	duel := app.Group("/duel", auth.AuthMiddleware)
	{
		duel.Get("/all", h.GetAllDuelsAuthorized)
		duel.Get("/my", h.GetMyDuels)
		duel.Get("/my/participant", h.GetMyDuelsAsParticipant)
		duel.Get("/:id", h.GetDuelByIDAuthorized)
	}
}

// CreateExternalWalletCryptoDuel godoc
//
//	@Summary		Create crypto duel (external Solana wallet)
//	@Description	Creates a crypto duel after the client has submitted an init transaction from an external wallet. The backend validates on-chain data and persists the duel.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreateDuelReq			true	"Create duel request"
//	@Success		200		{object}	model.CreateCryptoDuelResp	"Duel created"
//	@Failure		400		{object}	apperrors.ErrorPublic		"Invalid request"
//	@Failure		401		{object}	apperrors.ErrorPublic		"Unauthorized"
//	@Failure		500		{object}	apperrors.ErrorPublic		"Internal error"
//	@Router			/crypto-duel/solana [post]
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

// SignCreateCryptoDuelTransaction godoc
//
//	@Summary		Build unsigned init transaction for creating a crypto duel
//	@Description	Builds a base64-encoded Solana init transaction. The client signs and sends it using an external wallet (e.g., Phantom/Solflare).
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreateDuelReq		true	"Duel parameters"
//	@Success		200		{object}	object{tx=string}	"Unsigned base64 transaction"
//	@Failure		400		{object}	apperrors.ErrorPublic	"Invalid request"
//	@Failure		401		{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		500		{object}	apperrors.ErrorPublic	"Internal error"
//	@Router			/crypto-duel/solana/sign-tx [post]
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

// JoinExternalWalletCryptoDuel godoc
//
//	@Summary		Join crypto duel (external Solana wallet)
//	@Description	Records user participation in a crypto duel after the client submits a join transaction from an external wallet.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.JoinDuelReq			true	"Duel ID, answer (0/1), optional tx hash"
//	@Success		200		{object}	model.JoinCryptoDuelResp	"Joined successfully"
//	@Failure		400		{object}	apperrors.ErrorPublic		"Invalid request"
//	@Failure		401		{object}	apperrors.ErrorPublic		"Unauthorized"
//	@Failure		404		{object}	apperrors.ErrorPublic		"Duel not found"
//	@Failure		500		{object}	apperrors.ErrorPublic		"Internal error"
//	@Router			/crypto-duel/solana/join [post]
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

// SignJoinCryptoDuelTransaction godoc
//
//	@Summary		Build unsigned join transaction for a crypto duel
//	@Description	Builds a base64-encoded Solana join transaction for the client to sign in an external wallet.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.JoinDuelReq		true	"Duel ID and answer (0/1)"
//	@Success		200		{object}	object{tx=string}	"Unsigned base64 transaction"
//	@Failure		400		{object}	apperrors.ErrorPublic	"Invalid request"
//	@Failure		401		{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		500		{object}	apperrors.ErrorPublic	"Internal error"
//	@Router			/crypto-duel/solana/join/sign-tx [post]
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

// ResolveCryptoDuelByOwner godoc
//
//	@Summary		Resolve crypto duel by owner (payouts)
//	@Description	Allows the duel owner to resolve the duel and trigger payouts. Returns transaction hashes of on-chain transfers.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.DuelResolveReq		true	"Duel ID and final answer (0/1)"
//	@Success		200		{object}	object{tx_hashes=[]string}	"Distribution transaction hashes"
//	@Failure		400		{object}	apperrors.ErrorPublic		"Invalid request"
//	@Failure		401		{object}	apperrors.ErrorPublic		"Unauthorized"
//	@Failure		500		{object}	apperrors.ErrorPublic		"Internal error"
//	@Router			/crypto-duel/solana/resolve [put]
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

// GetAllDuelsPublic godoc
//
//	@Summary		List duels (public)
//	@Description	Public list of duels with filtering/sorting/pagination (no user-specific flags like 'joined').
//	@Tags			duel-public
//	@Accept			json
//	@Produce		json
//	@Param			opts.pagination.page_size	query		uint64					false	"Page size"					default(10)
//	@Param			opts.pagination.page_num	query		uint64					false	"Page number (starts at 1)"	default(1)
//	@Param			opts.order.order_by			query		string					false	"Order by field"				default(created_at)
//	@Param			opts.order.order_type		query		string					false	"Order type"					Enums(desc,asc)	default(desc)
//	@Param			opts.filters[0].column		query		string					false	"Filter column"
//	@Param			opts.filters[0].operator	query		string					false	"Filter operator"
//	@Param			opts.filters[0].value		query		string					false	"Filter value"
//	@Param			opts.filters[0].where_or	query		bool					false	"Use OR between filters"
//	@Success		200							{array}		model.DuelShow			"Duels"
//	@Failure		400							{object}	apperrors.ErrorPublic	"Bad request"
//	@Failure		500							{object}	apperrors.ErrorPublic	"Internal error"
//	@Router			/duel/public/all [get]
func (h *DuelHandler) GetAllDuelsPublic(c fiber.Ctx) error {
	var req model.OptsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}
	duels, err := h.DuelService.GetAllDuelsUnauthorized(c.Context(), &req.Opts)
	if err != nil {
		return err
	}
	return c.JSON(duels)
}

// CountAllDuelsPublic godoc
//
//	@Summary		Count duels (public)
//	@Description	Public count of duels with optional filtering.
//	@Tags			duel-public
//	@Accept			json
//	@Produce		json
//	@Param			opts.filters[0].column		query		string					false	"Filter column"
//	@Param			opts.filters[0].operator	query		string					false	"Filter operator"
//	@Param			opts.filters[0].value		query		string					false	"Filter value"
//	@Param			opts.filters[0].where_or	query		bool					false	"Use OR between filters"
//	@Success		200							{integer}	int						"Total count"
//	@Failure		400							{object}	apperrors.ErrorPublic	"Bad request"
//	@Failure		500							{object}	apperrors.ErrorPublic	"Internal error"
//	@Router			/duel/public/count [get]
func (h *DuelHandler) CountAllDuelsPublic(c fiber.Ctx) error {
	var req model.OptsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}
	count, err := h.DuelService.CountAllDuels(c.Context(), &req.Opts)
	if err != nil {
		return err
	}
	return c.JSON(count)
}

// GetDuelByIDPublic godoc
//
//	@Summary		Get duel by ID (public)
//	@Description	Returns a duel by ID without authentication.
//	@Tags			duel-public
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Duel ID (UUID)"
//	@Success		200	{object}	model.Duel				"Duel"
//	@Failure		400	{object}	apperrors.ErrorPublic	"Invalid duel ID"
//	@Failure		404	{object}	apperrors.ErrorPublic	"Duel not found"
//	@Failure		500	{object}	apperrors.ErrorPublic	"Internal error"
//	@Router			/duel/public/{id} [get]
func (h *DuelHandler) GetDuelByIDPublic(c fiber.Ctx) error {
	duelID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid duel ID", err)
	}
	duel, err := h.DuelService.GetDuelByIDUnauthorized(c.Context(), duelID)
	if err != nil {
		return err
	}
	return c.JSON(duel)
}

// GetAllDuelsAuthorized godoc
//
//	@Summary		List duels (authorized, with participation flags)
//	@Description	Returns duels with user-specific flags (e.g., joined, your_answer). Requires authentication.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			opts.pagination.page_size	query		uint64					false	"Page size"					default(10)
//	@Param			opts.pagination.page_num	query		uint64					false	"Page number (starts at 1)"	default(1)
//	@Param			opts.order.order_by			query		string					false	"Order by field"
//	@Param			opts.order.order_type		query		string					false	"Order type"	Enums(desc,asc)
//	@Param			opts.filters[0].column		query		string					false	"Filter column"
//	@Param			opts.filters[0].operator	query		string					false	"Filter operator"
//	@Param			opts.filters[0].value		query		string					false	"Filter value"
//	@Param			opts.filters[0].where_or	query		bool					false	"Use OR between filters"
//	@Success		200							{array}		model.DuelShow			"Duels with participation flags"
//	@Failure		400							{object}	apperrors.ErrorPublic	"Bad request"
//	@Failure		401							{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		500							{object}	apperrors.ErrorPublic	"Internal error"
//	@Router			/duel/all [get]
func (h *DuelHandler) GetAllDuelsAuthorized(c fiber.Ctx) error {
	var req model.OptsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}
	duels, err := h.DuelService.GetAllDuels(c.Context(), claims.UserID, &req.Opts) // возвращает joined/your_answer
	if err != nil {
		return err
	}
	return c.JSON(duels)
}

// GetMyDuels godoc
//
//	@Summary		List my duels
//	@Description	Returns duels created by the authenticated user. Supports filtering/sorting/pagination.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			opts.pagination.page_size	query		uint64					false	"Page size"					default(10)
//	@Param			opts.pagination.page_num	query		uint64					false	"Page number (starts at 1)"	default(1)
//	@Param			opts.order.order_by			query		string					false	"Order by field"
//	@Param			opts.order.order_type		query		string					false	"Order type"	Enums(desc,asc)
//	@Param			opts.filters[0].column		query		string					false	"Filter column"
//	@Param			opts.filters[0].operator	query		string					false	"Filter operator"
//	@Param			opts.filters[0].value		query		string					false	"Filter value"
//	@Param			opts.filters[0].where_or	query		bool					false	"Use OR between filters"
//	@Success		200							{array}		model.DuelShow			"My duels"
//	@Failure		400							{object}	apperrors.ErrorPublic	"Bad request"
//	@Failure		401							{object}	apperrors.ErrorPublic	"Unauthorized"
//	@Failure		500							{object}	apperrors.ErrorPublic	"Internal error"
//	@Router			/duel/my [get]
func (h *DuelHandler) GetMyDuels(c fiber.Ctx) error {
	var req model.OptsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}
	duels, err := h.DuelService.GetMyDuels(c.Context(), claims.UserID, &req.Opts)
	if err != nil {
		return err
	}
	return c.JSON(duels)
}

func (h *DuelHandler) GetMyDuelsAsParticipant(c fiber.Ctx) error {
	var req model.OptsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}
	duels, err := h.DuelService.GetMyDuelsAsParticipant(c.Context(), claims.UserID, &req.Opts)
	if err != nil {
		return err
	}
	return c.JSON(duels)
}

// GetDuelByIDAuthorized godoc
//
//	@Summary		Get duel by ID (authorized)
//	@Description	Returns duel details and players for the authenticated user.
//	@Tags			duel
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string													true	"Duel ID (UUID)"
//	@Success		200	{object}	object{duel=model.DuelShow,players=[]model.PlayerShow}	"Duel and players"
//	@Failure		400	{object}	apperrors.ErrorPublic									"Invalid duel ID"
//	@Failure		401	{object}	apperrors.ErrorPublic									"Unauthorized"
//	@Failure		404	{object}	apperrors.ErrorPublic									"Duel not found"
//	@Failure		500	{object}	apperrors.ErrorPublic									"Internal error"
//	@Router			/duel/{id} [get]
func (h *DuelHandler) GetDuelByIDAuthorized(c fiber.Ctx) error {
	duelID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid duel ID", err)
	}
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}
	duel, players, err := h.DuelService.GetDuelByID(c.Context(), duelID, claims.UserID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"duel": duel, "players": players})
}

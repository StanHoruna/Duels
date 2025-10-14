package v1

import (
	"duels-api/internal/model"
	"duels-api/internal/service"
	"duels-api/pkg/apperrors"
	auth "duels-api/pkg/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	NotificationService *service.NotificationService
}

func NewNotificationHandler(
	notificationService *service.NotificationService,
) *NotificationHandler {
	return &NotificationHandler{
		NotificationService: notificationService,
	}
}

func (h *NotificationHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	notificationGroup := app.Group("/notification", auth.AuthMiddleware)
	{
		notificationGroup.Get("/", h.GetAllNotifications)
		notificationGroup.Put("/:id", h.MarkAsRead)
		notificationGroup.Put("/", h.MarkAllAsRead)
		notificationGroup.Get("/unread", h.GetUnreadCount)
	}
}

// GetAllNotifications godoc
//
//	@Summary		Get all notifications
//	@Description	Retrieve all notifications for the authenticated user
//	@Tags			notification
//	@Produce		json
//	@Security		BearerAuth
//	@Param			opts.pagination.page_size	query		uint64					false	"Number of items per page"		default(10)
//	@Param			opts.pagination.page_num	query		uint64					false	"Page number (starting from 1)"	default(1)
//	@Param			opts.order.order_by			query		string					false	"Field to order by"				default(created_at)
//	@Param			opts.order.order_type		query		string					false	"Order type"					Enums(desc,asc)	default("")	"Order type (asc or desc)"	Enums(asc,desc)	default(desc)
//	@Param			opts.filters[0].column		query		string					false	"First filter column name"
//	@Param			opts.filters[0].operator	query		string					false	"First filter operator"
//	@Param			opts.filters[0].value		query		string					false	"First filter value"
//	@Param			opts.filters[0].where_or	query		bool					false	"First filter OR condition"
//	@Param			Authorization				header		string					true	"Authorization Bearer token"
//	@Success		200							{array}		model.Notification		"List of notifications"
//	@Failure		401							{object}	apperrors.ErrorPublic	"Unauthorized - Invalid or missing claims"
//	@Failure		500							{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/notification [get]
func (h *NotificationHandler) GetAllNotifications(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	var req model.OptsReq
	if err := c.Bind().Query(&req); err != nil {
		return apperrors.BadRequest("invalid request params")
	}

	notifications, err := h.NotificationService.GetAllNotifications(c.Context(), claims.UserID, &req.Opts)
	if err != nil {
		return err
	}

	return c.JSON(notifications)
}

// MarkAsRead godoc
//
//	@Summary		Mark a notification as read
//	@Description	Mark a single notification as read by ID
//	@Tags			notification
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Param			id				path		string					true	"Notification ID (UUID)"
//	@Success		200				{object}	model.Notification		"Updated notification"
//	@Failure		400				{object}	apperrors.ErrorPublic	"Invalid request data"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized - Invalid or missing claims"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/notification/{id} [put]
func (h *NotificationHandler) MarkAsRead(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	notificationIDStr := c.Params("id")
	if len(notificationIDStr) == 0 {
		return apperrors.BadRequest("invalid request data")
	}

	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	notification, err := h.NotificationService.MarkAsRead(c.Context(), claims.UserID, notificationID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"notification": notification})
}

// MarkAllAsRead godoc
//
//	@Summary		Mark all notifications as read
//	@Description	Marks all notifications as read for the authenticated user
//	@Tags			notification
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Authorization Bearer token"
//	@Success		200				{array}		model.Notification		"Updated notifications"
//	@Failure		401				{object}	apperrors.ErrorPublic	"Unauthorized - Invalid or missing claims"
//	@Failure		500				{object}	apperrors.ErrorPublic	"Internal server error"
//	@Router			/notification [put]
func (h *NotificationHandler) MarkAllAsRead(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	notifications, err := h.NotificationService.MarkAllAsRead(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"notifications": notifications})
}

// GetUnreadCount godoc
//
//	@Summary		Get unread notifications count
//	@Description	Retrieve the number of unread notifications for the authenticated user
//	@Tags			notification
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string						true	"Authorization Bearer token"
//	@Success		200				{object}	object{notifications=int}	"Unread notifications count"
//	@Failure		401				{object}	apperrors.ErrorPublic		"Unauthorized - Invalid or missing claims"
//	@Failure		500				{object}	apperrors.ErrorPublic		"Internal server error"
//	@Router			/notification/unread [get]
func (h *NotificationHandler) GetUnreadCount(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	notifications, err := h.NotificationService.GetUnreadCount(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"notifications": notifications})
}

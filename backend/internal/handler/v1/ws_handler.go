package v1

import (
	"duels-api/config"
	"duels-api/internal/client/ws"
	"duels-api/internal/storage/cache"
	"duels-api/pkg/apperrors"
	"strings"
	"time"

	auth "duels-api/pkg/jwt"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type WSHandler struct {
	EventPubSub  *cache.EventPubSub
	AllowOrigins string
}

func NewWSHandler(
	eventPubSub *cache.EventPubSub,
	c *config.Config,
) *WSHandler {
	return &WSHandler{
		EventPubSub:  eventPubSub,
		AllowOrigins: c.HTTP.AllowOrigins,
	}
}

func (h *WSHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	wsGroup := app.Group("/ws", auth.AuthMiddlewareQuery)
	{
		upgrader := websocket.FastHTTPUpgrader{
			HandshakeTimeout: 10 * time.Second,
			CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
				origin := string(ctx.Request.Header.Peek("Origin"))
				return strings.Contains(h.AllowOrigins, origin) || h.AllowOrigins == "*"
			},
		}
		wsGroup.Get("/", h.NewWS(upgrader))
	}
}

func (h *WSHandler) NewWS(upgrader websocket.FastHTTPUpgrader) fiber.Handler {
	return func(c fiber.Ctx) error {
		claims, ok := c.Locals("claims").(auth.TokenClaims)
		if !ok {
			return apperrors.Unauthorized("claims not found")
		}

		// Ensure the hub has the pub/sub set
		ws.Stream().SetEventPubSub(h.EventPubSub)

		err := upgrader.Upgrade(
			c.RequestCtx(),
			func(conn *websocket.Conn) {
				defer func() {
					err := conn.Close()
					zap.L().Error("error closing websocket", zap.Error(err))
				}()

				h.Conn(conn, claims.UserID)
			})
		if err != nil {
			return fiber.ErrUpgradeRequired
		}

		return nil
	}
}

func (h *WSHandler) Conn(conn *websocket.Conn, userID uuid.UUID) {
	ws.Stream().Conn(conn, userID)
}

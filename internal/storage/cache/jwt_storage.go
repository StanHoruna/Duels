package cache

import (
	"context"
	"duels-api/config"
	"duels-api/pkg/apperrors"
	auth "duels-api/pkg/jwt"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"slices"
	"strconv"
	"time"
)

type JWTStorage struct {
	client          *redis.Client
	conf            *config.Config
	refreshTokenTTL time.Duration
}

func NewJWTCacheStorage(
	client *redis.Client,
	c *config.Config,
) *JWTStorage {

	return &JWTStorage{
		client:          client,
		conf:            c,
		refreshTokenTTL: c.Auth.RefreshTokenTTL,
	}
}

const userSessionsLimit = 5

func (s *JWTStorage) Save(
	ctx context.Context,
	token *auth.TokenExpiration,
	claims auth.TokenClaims,
) error {
	key := getUserSessionsHashKey(claims.UserID)

	sessionCount, err := s.client.HLen(ctx, key).Result()
	if err != nil {
		return apperrors.Internal("failed to check user session count", err)
	}

	if sessionCount >= int64(userSessionsLimit) {
		keys, err := s.client.HKeys(ctx, key).Result()
		if err != nil {
			return apperrors.Internal("failed to get user session", err)
		}

		if len(keys) >= userSessionsLimit {
			// session id is uuid V7 (time based) stored as integers, so they are sortable
			slices.Sort(keys)

			fieldsToDelete := keys[:len(keys)-userSessionsLimit+1]
			err = s.client.HDel(ctx, key, fieldsToDelete...).Err()
			if err != nil {
				return apperrors.Internal("failed to delete old user session", err)
			}
		}
	}

	field := getSessionHashField(claims.SessionID)

	pipe := s.client.Pipeline()
	pipe.HSet(ctx, key, field, token.Token)
	pipe.HExpireAt(ctx, key, token.ExpiresAt, field)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return apperrors.Internal("failed to save new user session", err)
	}

	return nil
}

func getUserSessionsHashKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s:session", userID.String())
}

func getSessionHashField(sessionID uuid.UUID) string {
	if t := sessionID.Time(); t != 0 {
		return strconv.FormatInt(int64(t), 10)
	}

	return sessionID.String()
}

func (s *JWTStorage) GetUserSession(ctx context.Context, userID, sessionID uuid.UUID) (string, error) {
	var (
		key   = getUserSessionsHashKey(userID)
		field = getSessionHashField(sessionID)
	)

	token, err := s.client.HGet(ctx, key, field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", apperrors.NotFound("token not found in storage")
		}

		return "", apperrors.Unauthorized("failed to find token by user_id and session_id", err)
	}

	return token, nil
}

func (s *JWTStorage) DeleteUserSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	key := getUserSessionsHashKey(userID)

	field := getSessionHashField(sessionID)

	err := s.client.HDel(ctx, key, field).Err()
	if err != nil {
		return apperrors.Internal("failed to delete user session from a storage", err)
	}

	return nil
}

package logger

import (
	"context"
	"duels-api/pkg/apperrors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strconv"
)

type ctxKey struct{}

var key = ctxKey{}

type LogContext struct {
	UserID uuid.UUID
}

func EmbedLogData(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, &LogContext{})
}

func WithLogUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, key, &LogContext{UserID: userID})
}

func GetLogContext(ctx context.Context) *LogContext {
	if logCtx, ok := ctx.Value(key).(*LogContext); ok {
		return logCtx
	}

	return nil
}

func LogCtxFields(data *LogContext) []zap.Field {
	if data == nil {
		return nil
	}

	logFields := make([]zap.Field, 0, 1)
	if data.UserID != uuid.Nil {
		logFields = append(logFields, zap.String("user_id", data.UserID.String()))
	}

	return logFields
}

func LogAppErrFields(err *apperrors.AppError) []zap.Field {
	if err == nil {
		return nil
	}

	logFields := make([]zap.Field, 0, 2)
	if err.BaseError != nil {
		logFields = append(logFields, zap.String("err", err.BaseError.Error()))
	}

	if err.FileErrOccurred != "" && err.LineErrOccurred != 0 {
		path := err.FileErrOccurred + ":" + strconv.FormatInt(int64(err.LineErrOccurred), 10)
		logFields = append(logFields, zap.String("path", path))
	}

	return logFields
}

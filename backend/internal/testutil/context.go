package testutil

import (
	"context"

	"github.com/hadfielj/taran/backend/internal/auth"
)

// ContextWithUserID returns a context with the given userID set,
// matching what auth.SessionAuth.Middleware produces.
func ContextWithUserID(userID string) context.Context {
	return context.WithValue(context.Background(), auth.UserIDKey, userID)
}

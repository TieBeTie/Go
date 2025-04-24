//go:build !solution

package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type User struct {
	Name  string
	Email string
}

type userContextKey struct{}

func ContextUser(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey{}).(*User)
	return user, ok
}

var ErrInvalidToken = errors.New("invalid token")

type TokenChecker interface {
	CheckToken(ctx context.Context, token string) (*User, error)
}

func CheckAuth(checker TokenChecker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			user, err := checker.CheckToken(r.Context(), token)

			if err != nil {
				if errors.Is(err, ErrInvalidToken) {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey{}, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

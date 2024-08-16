package appbase

import (
	"context"
	"net/http"

	"gorm.io/gorm"
)

func GormMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "gormDB", db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

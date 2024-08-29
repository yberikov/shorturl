package middleware

import (
	clientConn "apiGW/internal/http-server/client"
	"context"
	au "github.com/yberikov/us-protos/gen/auth-microservice"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"

func JwtMiddleware(client *clientConn.ClientConn) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			grpcReq := &au.ValidateTokenRequest{Token: token}

			grpcResp, err := client.AuthClient.ValidateToken(context.TODO(), grpcReq)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, grpcResp.UserId)
			// Token is valid, proceed with the request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

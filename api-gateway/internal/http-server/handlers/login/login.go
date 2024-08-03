package login

import (
	clientConn "apiGW/internal/http-server/client"
	"context"
	"encoding/json"
	au "gitea.com/yberikov/us-protos/gen/auth-service"
	"google.golang.org/grpc/status"
	"net/http"
)

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func New(client *clientConn.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		grpcReq := &au.LoginRequest{Email: req.Email, Password: req.Password}

		grpcResp, err := client.AuthClient.Login(context.Background(), grpcReq)
		if err != nil {
			grpcError, _ := status.FromError(err)
			http.Error(w, grpcError.Message(), 500)
			return
		}
		client.Log.Info("login of user", grpcReq.Email)
		json.NewEncoder(w).Encode(grpcResp)
	}
}

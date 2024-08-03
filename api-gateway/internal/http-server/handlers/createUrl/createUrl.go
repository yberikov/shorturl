package createUrl

import (
	clientConn "apiGW/internal/http-server/client"
	"apiGW/internal/http-server/middleware"
	"context"
	"encoding/json"
	us "gitea.com/yberikov/us-protos/gen/us-service"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
)

type Request struct {
	OriginalUrl string `json:"original_url"`
}

func New(client *clientConn.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request
		// Parse request body and map to req
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		grpcReq := &us.ShortenUrlRequest{OriginalUrl: req.OriginalUrl, UserId: userID}
		log.Println(grpcReq)
		grpcResp, err := client.UrlShortenerClient.ShortenUrl(context.Background(), grpcReq)
		if err != nil {
			grpcError, _ := status.FromError(err)
			http.Error(w, grpcError.Message(), 500)
			return
		}
		client.Log.Info("creatingURl for:", req.OriginalUrl)
		json.NewEncoder(w).Encode(grpcResp)
	}
}

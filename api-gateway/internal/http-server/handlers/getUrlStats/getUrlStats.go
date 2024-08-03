package getUrlStats

import (
	clientConn "apiGW/internal/http-server/client"
	"context"
	"encoding/json"
	an "gitea.com/yberikov/us-protos/gen/analytics-service"
	"google.golang.org/grpc/status"
	"net/http"
)

type Request struct {
	ShortUrl string `json:"short_url"`
}

func New(client *clientConn.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request
		// Parse request body and map to req
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		grpcReq := &an.GetURLStatsRequest{Url: req.ShortUrl}

		grpcResp, err := client.AnalyticsClient.GetURLStats(context.Background(), grpcReq)
		if err != nil {
			grpcError, _ := status.FromError(err)
			http.Error(w, grpcError.Message(), 500)
			return
		}

		client.Log.Info("getting stats of:", grpcReq.Url)
		json.NewEncoder(w).Encode(grpcResp)
	}
}

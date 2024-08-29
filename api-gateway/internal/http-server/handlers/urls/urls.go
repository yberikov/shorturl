package urls

import (
	clientConn "apiGW/internal/http-server/client"
	"apiGW/internal/http-server/middleware"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	an "github.com/yberikov/us-protos/gen/analytics-microservice"
	us "github.com/yberikov/us-protos/gen/us-microservice"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
)

type RequestCreateUrl struct {
	OriginalUrl string `json:"original_url"`
}

func NewCreateUrl(client *clientConn.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RequestCreateUrl
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

func NewGetUrl(client *clientConn.ClientConn) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		shortUrl := chi.URLParam(r, "alias")
		grpcReq := &us.GetOriginalUrlRequest{ShortUrl: shortUrl}

		grpcResp, err := client.UrlShortenerClient.GetOriginalUrl(context.Background(), grpcReq)
		if err != nil {
			grpcError, _ := status.FromError(err)
			http.Error(w, grpcError.Message(), 500)
			return
		}

		client.Log.Info("redirecting to", grpcResp.OriginalUrl)
		w.Header().Set("Location", grpcResp.OriginalUrl)
		w.WriteHeader(http.StatusFound)
	}
}

type RequestGetUrlStats struct {
	ShortUrl string `json:"short_url"`
}

func NewGetUrlStats(client *clientConn.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RequestGetUrlStats
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

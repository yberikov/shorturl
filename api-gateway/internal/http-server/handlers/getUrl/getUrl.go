package getUrl

import (
	clientConn "apiGW/internal/http-server/client"
	"context"
	us "gitea.com/yberikov/us-protos/gen/us-service"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/status"
	"net/http"
)

type Request struct {
	ShortUrl string `json:"short_url"`
}

func New(client *clientConn.ClientConn) http.HandlerFunc {

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

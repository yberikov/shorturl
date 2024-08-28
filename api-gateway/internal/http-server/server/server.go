package server

import (
	"apiGW/internal/config"
	clientConn "apiGW/internal/http-server/client"
	"apiGW/internal/http-server/handlers/urls"
	"apiGW/internal/http-server/handlers/user"
	"apiGW/internal/http-server/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func New(cfg *config.Config, client *clientConn.ClientConn) *http.Server {
	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		r.Use(middleware.JwtMiddleware(client))

		r.HandleFunc("/createUrl", urls.NewCreateUrl(client))
		r.HandleFunc("/getUrlStats", urls.NewGetUrlStats(client))
	})

	router.HandleFunc("/login", user.NewLogin(client))
	router.HandleFunc("/register", user.NewRegister(client))
	router.HandleFunc("/{alias}", urls.NewGetUrl(client))

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}
}

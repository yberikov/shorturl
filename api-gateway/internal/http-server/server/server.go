package server

import (
	"apiGW/internal/config"
	clientConn "apiGW/internal/http-server/client"
	"apiGW/internal/http-server/handlers/createUrl"
	"apiGW/internal/http-server/handlers/getUrl"
	"apiGW/internal/http-server/handlers/getUrlStats"
	"apiGW/internal/http-server/handlers/login"
	"apiGW/internal/http-server/handlers/register"
	"apiGW/internal/http-server/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func New(cfg *config.Config, client *clientConn.ClientConn) *http.Server {
	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		r.Use(middleware.JwtMiddleware(client))

		r.HandleFunc("/createUrl", createUrl.New(client))
		r.HandleFunc("/getUrlStats", getUrlStats.New(client))
	})

	router.HandleFunc("/login", login.New(client))
	router.HandleFunc("/register", register.New(client))
	router.HandleFunc("/{alias}", getUrl.New(client))

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}
}

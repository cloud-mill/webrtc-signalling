package server

import (
	"net/http"
	"strings"

	"github.com/cloud-mill/webrtc-signalling/internal/config"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func NewRouter(
	authMiddleware func(next http.Handler, secretKey interface{}, authCfg AuthConfig) http.Handler,
) *mux.Router {
	c := cors.New(cors.Options{
		AllowedOrigins:   config.Config.AllowedOrigins,
		ExposedHeaders:   []string{config.Config.Auth.CsrfHeaderName},
		AllowedHeaders:   []string{"*"}, // TODO: hardening required
		AllowCredentials: true,
	})

	router := mux.NewRouter().StrictSlash(true)
	router.Use(c.Handler)

	for _, route := range OpenRoutes {
		methods := strings.Split(route.Method, ",")
		router.Methods(methods...).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	authCfg := AuthConfig{
		JwtCookieName:  config.Config.Auth.JwtCookieName,
		CsrfCookieName: config.Config.Auth.CsrfCookieName,
		CsrfHeaderName: config.Config.Auth.CsrfHeaderName,
	}

	for _, route := range ProtectedRoutes {
		methods := strings.Split(route.Method, ",")
		handler := authMiddleware(
			route.HandlerFunc,
			config.Config.Auth.AuthMiddlewareSecretKey,
			authCfg,
		)

		router.Methods(methods...).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

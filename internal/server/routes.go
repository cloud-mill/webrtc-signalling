package server

import (
	"net/http"
)

type Route struct {
	Name         string
	Description  string
	Method       string
	Pattern      string
	HandlerFunc  http.HandlerFunc
	Authenticate bool
}

type Routes []Route

var okHandler = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

var OpenRoutes = Routes{
	{
		Name:        "HealthCheck",
		Description: "Liveness health check endpoint",
		Method:      "GET,OPTIONS",
		Pattern:     "/healthz",
		HandlerFunc: okHandler,
	},
}

var ProtectedRoutes = Routes{
	{
		Name:         "connect",
		Description:  "connect to server",
		Method:       "GET,OPTIONS",
		Pattern:      "/connect",
		HandlerFunc:  AcceptConnection,
		Authenticate: true,
	},
}

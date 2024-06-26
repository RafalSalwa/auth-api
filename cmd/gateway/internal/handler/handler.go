package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

type (
	HandlerFunc     func(http.ResponseWriter, *http.Request)
	RouteRegisterer interface {
		RegisterRoutes(r *mux.Router, cfg interface{})
	}
)

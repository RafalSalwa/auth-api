package handler

import "net/http"

type HandlerFunc func(http.ResponseWriter, *http.Request)

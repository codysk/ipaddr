package http

import (
	"net/http"
)

func NewFileServer() http.Handler {
	server := http.FileServer(http.Dir("/app/public"))
	return server
}

package http

import (
	"ipprovider/pkg/addressmanager"
	"log"
	"net/http"
)

type Server struct {
	port string
	apiServer http.Handler
	fileServer http.Handler
}

func (server *Server) StartHttpServer() error {
	log.Println("http server started")
	http.Handle("/api/", server.apiServer)
	http.Handle("/", server.fileServer)
	return  http.ListenAndServe(server.port, nil)
}

func NewHttpServer(port string, manager *addressmanager.Manager) *Server {
	return &Server{
		port: port,
		apiServer: NewApiServer(manager),
		fileServer: NewFileServer(),
	}
}
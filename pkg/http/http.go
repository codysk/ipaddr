package http

import (
	"fmt"
	"net/http"
)

type Server struct {
	port string
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	//log.Print(writer, r)
	switch r.URL.Path {
	case "/api/getContainers":
		server.api_getContainers(writer, r)
		break
	default:
		server.notFound(writer, r)
	}
}

func (*Server) api_getContainers(writer http.ResponseWriter, _ *http.Request) {
	_, _ = writer.Write([]byte("hello containers"))
}

func (*Server) notFound(writer http.ResponseWriter, r *http.Request)  {
	writer.WriteHeader(404)
	_, _ = writer.Write([]byte(fmt.Sprintf("%s not found\n", r.URL.Path)))
}

func (server *Server) StartHttpServer() error {
	return  http.ListenAndServe(server.port, server)
}

func NewHttpServer(port string) *Server {
	return &Server{
		port: port,
	}
}
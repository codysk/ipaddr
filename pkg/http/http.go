package http

import (
	"fmt"
	"ipprovider/pkg/addressmanager"
	"log"
	"net"
	"net/http"
)

type Server struct {
	port string
	manager *addressmanager.Manager
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	//log.Print(writer, r)
	switch r.URL.Path {
	case "/api/getContainers":
		server.api_getContainers(writer, r)
		break
	case "/api/assignIPForContainer":
		server.api_assignIPForContainer(writer, r)
		break
	default:
		server.notFound(writer, r)
	}
}

func (*Server) api_getContainers(writer http.ResponseWriter, _ *http.Request) {
	_, _ = writer.Write([]byte("hello containers"))
}

func (server *Server) api_assignIPForContainer(writer http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("parse req error")
		log.Println(err)
		return
	}

	// log.Printf("Form: %v", r.Form)

	ipstr := r.FormValue("ip")
	ip := net.ParseIP(ipstr).To4()
	// log.Printf("form ip is: %s", ipstr)
	containerid := r.FormValue("container_id")
	err = server.manager.AssignIPForContainer(ip, containerid)
	if err != nil {
		_, _ = writer.Write([]byte(err.Error()))
		return
	}
	_, _ = writer.Write([]byte("done"))

}

func (*Server) notFound(writer http.ResponseWriter, r *http.Request)  {
	writer.WriteHeader(404)
	_, _ = writer.Write([]byte(fmt.Sprintf("%s not found\n", r.URL.Path)))
}

func (server *Server) StartHttpServer() error {
	log.Println("http server started")
	return  http.ListenAndServe(server.port, server)
}

func NewHttpServer(port string, manager *addressmanager.Manager) *Server {
	return &Server{
		port: port,
		manager: manager,
	}
}
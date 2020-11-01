package tcpserver

import (
	"io/ioutil"
	"log"
	"net"
	"strconv"
)

type Server struct {
	Addr string
	Port int

	listener *net.TCPListener
}

func (server *Server) Start() error {
	listenAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(server.Addr, strconv.Itoa(server.Port)))
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return err
	}
	server.listener = listener
	return nil
}

func (server *Server) GetListenerPort() int {
	return server.listener.Addr().(*net.TCPAddr).Port
}

func (server *Server) Serve() error {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			return err
		}
		go server.Handle(conn)
	}
}

func (server *Server) Handle(conn net.Conn) {
	defer conn.Close()
	req, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Printf("Failed to read from client after having received %d bytes: %v", len(req), err)
		return
	}
}

func (server *Server) Shutdown() {
	server.listener.Close()
	server.listener = nil
}

func RepeatedlyConnect(host string, port int) (int, error) {
	conns := make([]net.Conn, 0)
	defer func() {
		for _, conn := range conns {
			if err := conn.Close(); err != nil {
				log.Printf("failed to close client connection: %v", err)
			}
		}
	}()
	for i := 0; ; i++ {
		conn, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
		if err != nil {
			return i, err
		}
		conns = append(conns, conn)
	}
}

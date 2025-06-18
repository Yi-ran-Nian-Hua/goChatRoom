package main

import (
	"fmt"
	"net"
)

// Server 定义 server 结构体
type Server struct {
	IP   string // IP 地址
	Port int    // 端口号
}

// NewServer 用于创建一个新的 server 并返回
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:   ip,
		Port: port,
	}

	return server
}

func (s *Server) Handler(connection net.Conn) {
	// ... 链接当前的业务
	fmt.Println("连接建立成功")
}

// Start 启动 server
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("net.Listen Error: ", err)
	}
	// close listen socket
	defer listener.Close()
	for {
		// accept
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err: ", err)
			continue
		}
		// do handler
		go s.Handler(connection)
	}
}

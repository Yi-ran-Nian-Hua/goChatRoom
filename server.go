package main

import (
	"fmt"
	"net"
	"sync"
)

// Server 定义 server 结构体
type Server struct {
	IP   string // IP 地址
	Port int    // 端口号

	// 在线用户的列表, key为 string, value 为对应的 user
	OnlineMap map[string]*User
	// 为保证 map 读写正确而添加的读写锁
	mapLock sync.RWMutex

	// 消息广播的 channel
	Message chan string
}

// NewServer 用于创建一个新的 server 并返回
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

func (s *Server) Handler(connection net.Conn) {
	user := NewUser(connection)

	// 用户上线, 将其添加到 onlineMap 中
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	// 广播当前用户上线消息
	s.BroadCast(user, "已上线")

	// 当前 handler 阻塞
	select {}
}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	s.Message <- sendMsg
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

	// 启动监听 Message 的 goroutine
	go s.ListenMessager()

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

// ListenMessager 监听 Message 广播消息 channel 的 goroutine, 一旦有消息就发送给全部在线的 User
func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message

		// 将 msg 发送给全部的 user
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.Channel <- msg
		}
		s.mapLock.Unlock()
	}
}

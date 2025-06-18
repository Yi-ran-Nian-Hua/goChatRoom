package main

import "net"

type User struct {
	Name       string
	Addr       string
	Channel    chan string
	Connection net.Conn
}

// ListenMessage 监听当前 user channel 的方法, 一旦有消息就直接发送给对端客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.Channel
		u.Connection.Write([]byte(msg + "\n"))
	}
}

// NewUser 创建一个用户
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:       userAddr,
		Addr:       userAddr,
		Channel:    make(chan string),
		Connection: conn,
	}

	// 启动监听当前 user channel 消息的 goroutine
	go user.ListenMessage()

	return user
}

func (u *User) Online(s *Server) {
	// 用户上线, 将其添加到 onlineMap 中
	s.mapLock.Lock()
	s.OnlineMap[u.Name] = u
	s.mapLock.Unlock()

	// 广播当前用户上线消息
	s.BroadCast(u, "已上线")
}

func (u *User) Offline(s *Server) {
	s.BroadCast(u, "下线")
}

func (u *User) DoMessage(s *Server, message string) {
	s.BroadCast(u, message)
}

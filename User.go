package main

import "net"

type User struct {
	Name       string
	Addr       string
	Channel    chan string
	Connection net.Conn

	// 保存用户连接到了哪个服务器中
	Server *Server
}

// ListenMessage 监听当前 user channel 的方法, 一旦有消息就直接发送给对端客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.Channel
		u.Connection.Write([]byte(msg + "\n"))
	}
}

// NewUser 创建一个用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:       userAddr,
		Addr:       userAddr,
		Channel:    make(chan string),
		Connection: conn,
		Server:     server,
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

func (u *User) Offline() {
	// 首先将用户从在线列表中移除
	u.Server.mapLock.Lock()
	delete(u.Server.OnlineMap, u.Name)
	u.Server.mapLock.Unlock()
	// 之后广播下线
	u.Server.BroadCast(u, "下线")
}

func (u *User) DoMessage(message string) {

	if message == "who" {
		u.Server.mapLock.Lock()
		for _, value := range u.Server.OnlineMap {
			onlineUser := "[" + value.Addr + "]" + value.Name + ": 在线...\n"
			u.SendMsg(onlineUser)
		}
		u.Server.mapLock.Unlock()
	} else {
		u.Server.BroadCast(u, message)
	}

}

func (u *User) SendMsg(message string) {
	u.Connection.Write([]byte(message))
}

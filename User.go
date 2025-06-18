package main

import (
	"net"
	"strings"
)

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
	} else if len(message) > 7 && message[:7] == "rename|" {
		newName := strings.Split(message, "|")[1] // 保留右侧部分
		// 之后去 map 中查询当前用户名是否存在

		_, value := u.Server.OnlineMap[newName]
		if value {
			// 表示当前用户名已被使用
			u.SendMsg("当前用户名已被使用\n")
		} else {
			// 更新用户名
			// 首先删除原来的用户名
			u.Server.mapLock.Lock()
			delete(u.Server.OnlineMap, u.Name)

			// 之后添加新的用户名
			u.Server.OnlineMap[newName] = u

			u.Server.mapLock.Unlock()
			u.Name = newName
			u.SendMsg("您已更新用户名" + newName)

		}

	} else if len(message) > 4 && message[:3] == "to|" {
		// 消息格式: to|张三|消息内容

		// 获取对方的用户名
		remoteName := strings.Split(message, "|")[1]
		if remoteName == "" {
			u.SendMsg("消息格式不正确，请使用 \"to|张三|你好啊\"格式。\n")
			return
		}

		// 查找是否存在
		remoteUser, ok := u.Server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("要发送的用户不存在\n")
			return
		}

		// 获取消息的内容, 发送消息
		content := strings.Split(message, "|")[2]
		if content == "" {
			u.SendMsg("消息发送内容为空\n")
			return
		}

		remoteUser.SendMsg(u.Name + "对您说: " + content)
	} else {
		u.Server.BroadCast(u, message)
	}

}

func (u *User) SendMsg(message string) {
	u.Connection.Write([]byte(message))
}

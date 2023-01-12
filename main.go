package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lonnng/nano"
	"github.com/lonnng/nano/component"
	"github.com/lonnng/nano/serialize/json"
	"github.com/lonnng/nano/session"
)

type (
	// define Room component
	Room struct {
		component.Base
		group *nano.Group
	}

	// protocol messages
	UserMessages struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	NewUser struct {
		Content string `json:"content"`
	}

	AllMembers struct {
		Members []int64 `json:"members"`
	}

	JoinResponse struct {
		Code   int    `json:"code"`
		Result string `json:"result"`
	}
)

func NewRoom() *Room {
	return &Room{
		group: nano.NewGroup("room"),
	}
}

func (r *Room) AfterInit() {
	nano.OnSessionClosed(func(s *session.Session) {
		r.group.Leave(s)
	})
}

// Join into room
func (r *Room) Join(s *session.Session, msg []byte) error {
	s.Bind(s.ID()) // binding session uid
	s.Push("onMembers", &AllMembers{Members: r.group.Members()})

	// notify others
	r.group.Broadcast("onNewUser", &NewUser{Content: fmt.Sprintf("New user: %d", s.ID())})

	// new user join group
	r.group().Add(s) // add session to group

	return s.Response(&JoinResponse{Result: "Success"})
}

// Send message
func (r *Room) Message(s *session.Session, msg *UserMessages) error {
	return r.group.Broadcast("onMessage", msg)
}

func main() {
	components := &component.Components{}
	components.Register(NewRoom())

	log.SetFlags(log.LstdFlags | log.Llongfile)

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	nano.SetCheckOriginFunc(func(_ *http.Request) bool { return true })
	nano.Listen(":3250",
		nano.WithIsWebsocket(true),
		nano.WithCheckOriginFunc(func(_ *http.Request) bool { return true }),
		nano.WithWSPath("/ws"),
		nano.WithDebugMode(),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithComponents(components),
	)
}

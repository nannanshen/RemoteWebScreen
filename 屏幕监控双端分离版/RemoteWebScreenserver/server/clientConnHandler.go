package server
import (
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader2 = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var Clientconn *websocket.Conn

func ClientConnHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader2.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	Clientconn = conn


}
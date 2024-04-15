package main

import (
	"crypto/tls"
	"log"
	"RemoteWebScreenclient/server"
	"github.com/gorilla/websocket"
	"sync"
)

func main() {
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c, _, err := dialer.Dial("wss://localhost:443/clientConnHandler", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	var mu sync.Mutex

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)

			var response []byte
			if string(message) == "captureScreen" {
				imgBytes, err := server.CaptureScreen(100)
				if err != nil {
					//log.Printf("imgBytes, err := captureScreen(captureScreenquality, captureScreenscale) Error: %v", err)
					return
				}
				response = imgBytes
			} else {
				log.Printf("recv: %s", message)
				server.SimulateDesktopHDMessage(message)
				response = nil
			}
			mu.Lock()
			err = c.WriteMessage(websocket.TextMessage, response)
			mu.Unlock()
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}()

	for {

	}
}
package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var wg sync.WaitGroup
var connections = make(map[*websocket.Conn]bool)
var mu sync.Mutex
var mu2 sync.Mutex
var mu3 sync.Mutex

func ScreenshotHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		//log.Printf("Error screenshotHandler: %v", err)
		return
	}
	connections[conn] = true
	defer conn.Close()
	go func() {
		ticker := time.NewTicker(6 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				err := conn.WriteMessage(websocket.PingMessage, []byte{})
				mu.Unlock()
				if err != nil {
					return
				}
			}
		}
	}()
	wg.Add(1)
	go func() {
		ticker := time.NewTicker(33 * time.Millisecond)
		defer ticker.Stop()
		defer wg.Done()
		for {
			select {
			case <-ticker.C:
				mu2.Lock()
				err = Clientconn.WriteMessage(websocket.TextMessage, []byte("captureScreen"))
				if err != nil {
					log.Println(err)
					return
				}
				_, imgBytes, err := Clientconn.ReadMessage()
				mu2.Unlock()
				if err != nil {
					log.Println(err)
					return
				}
				//imgBytes, err := captureScreen(CaptureScreenquality)
				if err != nil {
					//log.Printf("imgBytes, err := captureScreen(captureScreenquality, captureScreenscale) Error: %v", err)
					return
				}
				if imgBytes == nil {
					continue
				}
				err = sendImage(conn, imgBytes)
				if err != nil {
					//log.Printf("err = sendImage(conn, imgBytes) Error: %v", err)
					return
				}
			}
		}
	}()
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				//log.Printf("_, msg, err := conn.ReadMessage Error: %v", err)
				break
			}
			var message map[string]interface{}
			if err := json.Unmarshal(msg, &message); err != nil {
				//log.Printf("if err := json.Unmarshal(msg, &message); err != nil Error: %v", err)
				return
			}

			switch messageType := message["type"].(string); messageType {
			case "10":
				conn.Close()
			}
			mu3.Lock()
			err = Clientconn.WriteMessage(websocket.TextMessage, msg)
			mu3.Unlock()
			if err != nil {
				log.Println(err)
				return
			}
		}
	}()
	defer func() {
		delete(connections, conn)
		conn.Close()
	}()
	wg.Wait()
}

// 关闭无用连接
func CleanupConnections() {
	for conn := range connections {
		mu.Lock()
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			conn.Close()
			delete(connections, conn)
		}
		mu.Unlock()
	}
}

func sendImage(conn *websocket.Conn, imgBytes []byte) error {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(imgBytes); err != nil {
		//log.Printf("Write Error sending image: %v", err)
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		//log.Printf("Close Error sending image: %v", err)
		return err
	}

	mu.Lock()
	err := conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
	mu.Unlock()
	return err
}

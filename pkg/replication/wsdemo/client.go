//go:build ignore
// +build ignore

package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"bufio"

	"github.com/gorilla/websocket"
)

const (
	maxBufSize = 1024*1024
	fileName = "sample-copy.db"
)

// read database
func readDB(dbPath string, conn *websocket.Conn) error {
	file, err := os.OpenFile(dbPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println("cannot able to read the file", err)
		return err
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	for {
		messageType, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return err
		}

		if messageType == websocket.BinaryMessage {
			n, err := writer.Write(messageBytes)
			if err != nil {
				return err
			}
			log.Println("Writen %d bytes", n)
		}
		
		if messageType == websocket.TextMessage {
			log.Printf("recv: %s", messageBytes)
			if string(messageBytes) == "EOF" {
				log.Printf("Done receiving database")
				break
			}
		}
	}
	return nil
}

func main() {
	address := "localhost:8080"
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	wsUrl := url.URL{Scheme: "ws", Host: address, Path: "/replicate"}
	log.Printf("connecting to %s", wsUrl.String())

	conn, _, err := websocket.DefaultDialer.Dial(wsUrl.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("basebackup")); err != nil {
		log.Fatal("Error writing", err)
	}
	if err := readDB(fileName, conn); err != nil {
		log.Fatal("Error reading db:", err)
	}
	
	// wait for queries
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Fatal("after read:", err)
		}
		log.Println("Execute query: %s", string(msg))
	}
}

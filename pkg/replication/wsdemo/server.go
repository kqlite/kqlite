//go:build ignore
// +build ignore


package main

import (
	"io"
	"os"
	"net"
	"net/http"
	"log"
	"bufio"
	"time"
	"sync"
	"fmt"

	"github.com/gorilla/websocket"
)

const (
	maxBufSize = 1024*1024
	fileName = "sample.db"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  maxBufSize,
	WriteBufferSize: maxBufSize,
}

var (
	chq chan string
	queue []string
	mu sync.Mutex
	replReady bool
)

func queueAdd(elem string) {
	mu.Lock()
	defer mu.Unlock()
	if elem != "" {
		queue = append(queue, elem)
	}
}

func queueNext() string {
	mu.Lock()
	defer mu.Unlock()
	var elem string
	
	if len(queue) != 0 {
		elem = queue[0]
		queue = queue[1:]
	}
	return elem
}

func queueEmpty() bool {
	mu.Lock()
	defer mu.Unlock()
	return len(queue) != 0
}

func sendDB(dbPath string, conn *websocket.Conn) error {
	file, err := os.Open(dbPath)
	if err != nil {
		log.Println("cannot able to read the file", err)
		return err
	}
	defer file.Close()

	buf := make([]byte, maxBufSize)
	reader := bufio.NewReader(file)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
			log.Println("Done sending database")
			// Send EOF to client
			if err := conn.WriteMessage(websocket.TextMessage, []byte("EOF")); err != nil {
				return err
			}
			return nil
		}

		log.Println("Read %d bytes, sending", n)
		if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
			return err
		}
		clear(buf)
	}

	return nil
}

func handleReplicate(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("upgrade:", err)
	}
	defer conn.Close()

	// Sync database with client
	msgType, messageBytes, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("read:", err)
	}
	
	// replicate database
	if msgType == websocket.TextMessage {
		if string(messageBytes) == "basebackup" {
			if err := sendDB(fileName, conn); err != nil {
				log.Fatal("sendDB error: %s", err.Error())
			}
		}
	}
	
	// check if there are any buffered queries before starting to sync realtime
	qlen := len(queue)
	if len(queue) != 0 {
		for _ = range qlen {
			query := queueNext()
			log.Println("Sending buffered query: ", query)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(query)); err != nil {
				log.Fatal("Got error %s", err.Error())
			}
		}
	}
	// wait for queries to send to the client
	for {
		// Signal done and start reading queries from channel
		replReady = true
		// Start waiting for incoming queries.
		query := <-chq
		log.Println("Sending query: ", query)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(query)); err != nil {
			log.Fatal("Got error %s", err.Error())
		}
	}
}

func sendQueries() {
	var incr int64
	go func() {
		for {
			incr++
			query := fmt.Sprintf("UPDATE customers set name = John-%d where customers.product = TV-%d", incr, incr)
			time.Sleep(500 * time.Millisecond)
			if !replReady {
				log.Println("not ready to accept queries, buffering")
				queueAdd(query)
				continue
			}
			chq <- query
		}
	}()
}

// Start server
func main() {
	addr := "localhost:8080"
	var httpServer http.Server
	chq = make(chan string, 1)
	queue = []string{}
	replReady = false

	log.SetFlags(0)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		os.Exit(1)
	}

	sendQueries()
	http.HandleFunc("/replicate", handleReplicate)
	httpServer.Serve(listener)
}


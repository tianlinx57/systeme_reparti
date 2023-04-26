package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"time"
)

type myData struct {
	Number string `json:"number"`
}

func handleWebSocket(conn *websocket.Conn) {
	defer conn.Close()

	// 从客户端接收消息
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				return
			}
			var data myData
			err = json.Unmarshal(message, &data)
			if err != nil {
				fmt.Println("unmarshal:", err)
				return
			}
			num, _ := strconv.Atoi(data.Number)
			fmt.Printf("Received message: %d\n", num)
		}
	}()

	// 向客户端发送数字
	num := 0
	for {
		num = num + 1
		msg := &myData{
			Number: strconv.Itoa(num),
		}
		err := conn.WriteJSON(msg)
		if err != nil {
			fmt.Println("write:", err)
			return
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
}

func main() {
	var port = flag.String("port", "4444", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")

	flag.Parse()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("upgrade:", err)
			return
		}
		fmt.Println("WebSocket connected")
		handleWebSocket(conn)
	})
	http.ListenAndServe(*addr+":"+*port, nil)
}

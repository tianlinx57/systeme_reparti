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

func receiveMsg(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrade:", err)
		return
	}
	defer conn.Close()

	fmt.Println("已经建立连接")
	for {
		_, message, err := conn.ReadMessage()
		var data myData
		err = json.Unmarshal(message, &data)
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		num, _ := strconv.Atoi(data.Number)

		// 输出结果
		fmt.Printf("num = %d\n", num)
	}
}

func sendMsg(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrade:", err)
		return
	}
	defer conn.Close()

	fmt.Println("已经建立连接")
	num := 0
	for {
		num = num + 1
		msg := &myData{
			Number: strconv.Itoa(num),
		}
		err := conn.WriteJSON(msg)
		time.Sleep(time.Duration(2) * time.Second)
		if err != nil {
			// 处理错误
			return
		}
	}
}

func main() {
	var port = flag.String("port", "4444", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")

	flag.Parse()
	http.HandleFunc("/ws1", sendMsg)
	http.HandleFunc("/ws2", receiveMsg)
	http.ListenAndServe(*addr+":"+*port, nil)

}

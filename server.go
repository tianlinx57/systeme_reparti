package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type myData struct {
	Number int `json:"number"`
}

func do_webserver(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Bonjour depuis le serveur web en Go !")
}

func do_websocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	cnn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrade:", err)
		return
	}

	for {
		_, message, err := cnn.ReadMessage()
		var data myData
		err = json.Unmarshal(message, &data)
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		num := data.Number

		// 输出结果
		fmt.Printf("num = %d\n", num)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// 处理错误
		return
	}
	defer conn.Close()
	num := 0
	for {
		num = num + 1
		msg := &myData{
			Number: num,
		}
		err = conn.WriteJSON(msg)
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

	http.HandleFunc("/", do_webserver)
	http.HandleFunc("/ws", do_websocket)
	http.HandleFunc("/test", wsHandler)
	http.ListenAndServe(*addr+":"+*port, nil)

}

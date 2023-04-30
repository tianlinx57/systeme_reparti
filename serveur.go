package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
)

var ws *websocket.Conn
var stderr = log.New(os.Stderr, "", 0)

func display_d(where string, what string) {
	stderr.Printf(" + %-8.8s : %s\n", where, what)
}

func do_websocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrade:", err)
		return
	}
	display_d("ws_create", "Client connecté")
	ws = conn
	go receive()
	go ws_receive()
}
func ws_send(text, mylock string) {
	msg := &myData{
		Number:       strconv.Itoa(stock),
		Text:         text,
		MyLock:       mylock,
		Horloge:      strconv.Itoa(horloge),
		Snapshot:     "",
		Snapshottime: "",
	}
	err := ws.WriteJSON(msg)
	if err != nil {
		display_d("write:", string(err.Error()))
		return
	}
}

func ws_receive() {
	defer ws_close()
	nom *= -1
	ws_send("start", status)
	//msg := &myData{
	//	Number:  strconv.Itoa(stock),
	//	Text:    "start",
	//	MyLock:  "unlocked",
	//	Horloge: strconv.Itoa(horloge),
	//}
	l := log.New(os.Stderr, "", 0)
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			l.Println("read:", err)
			return
		}
		mutex.Lock()
		var data myData
		err = json.Unmarshal(message, &data)
		if err != nil {
			l.Println("unmarshal:", err)
			return
		}
		//快照请求的判断
		//
		if data.Text == "demand snapshot" {
			msg_send(msg_format("receiver", strconv.Itoa(nom*(-1))) + msg_format("type", "demandeSnap") + msg_format("sender", strconv.Itoa(nom)) + msg_format("hlg", strconv.Itoa(horloge)))
			status = "locked"
			ws_send("生成快照中 请耐心等待", status)
			mutex.Unlock()
			continue
		}

		count, _ = strconv.Atoi(data.Number)
		l.Printf("Received message: %d\n", count)

		//排队中 请耐心等待
		//fmt.Printf("/=receiver=%d/=type=demandeSC/=sender=%d/=hlg=%d\n", nom*(-1), nom, 0)
		msg_send(msg_format("receiver", strconv.Itoa(nom*(-1))) + msg_format("type", "demandeSC") + msg_format("sender", strconv.Itoa(nom)) + msg_format("hlg", strconv.Itoa(horloge)))
		status = "locked"
		ws_send("排队中 请耐心等待", status)
		//msg := &myData{
		//	Number:  strconv.Itoa(stock),
		//	Text:    "排队中 请耐心等待",
		//	MyLock:  "locked",
		//	Horloge: strconv.Itoa(horloge),
		//}
		mutex.Unlock()
		// /=type=demandeSC/=sender=1/=hlg=0/=receiver=-1
		// /=type=demandeSC/=sender=1/=hlg=0/=receiver=-1
	}

}
func ws_close() {
	display_d("ws_close", "Fin des réceptions => fermeture de la websocket")
	ws.Close()
}

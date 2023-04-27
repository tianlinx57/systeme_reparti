package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type messageType int

const (
	updateSC messageType = iota
	permetSC
)

type message struct {
	msgType messageType
	count   int
}

type myData struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

func handleWebSocket(conn *websocket.Conn) {
	defer conn.Close()
	nom *= -1
	// 从客户端接收消息
	msg := &myData{
		Number: strconv.Itoa(stock),
	}
	err := conn.WriteJSON(msg)
	if err != nil {
		fmt.Println("write:", err)
		return
	}
	go func() {
		var rcvmsg string
		l := log.New(os.Stderr, "", 0)
		l.Printf(string(nom))
		var receiver int
		var msgType messageType
		var count int
		for {
			fmt.Scanln(&rcvmsg)
			msgType = -1
			mutex.Lock()
			l.Printf("%d message reçu : %s\n", nom, rcvmsg)
			tab_allkeyval := strings.Split(rcvmsg[1:], rcvmsg[0:1])
			//l.Printf("%q\n", tab_allkeyval)
			for _, keyval := range tab_allkeyval {
				tab_keyval := strings.Split(keyval[1:], keyval[0:1])
				//l.Printf("  %q\n", tab_keyval)
				//l.Printf("  key : %s  val : %s\n", tab_keyval[0], tab_keyval[1])
				//如果不是自己该收丢弃
				if tab_keyval[0] == "receiver" {
					receiver, _ = strconv.Atoi(tab_keyval[1])
					if receiver != nom {
						msgType = -1
						break
					}
				} else if tab_keyval[0] == "type" {
					switch tab_keyval[1] {
					case "updateSC":
						msgType = updateSC
					case "permetSC":
						msgType = permetSC
					default:
						msgType = -1
						l.Println("Invalid message type. Please try again.")
						break
					}
				} else if tab_keyval[0] == "count" {
					count, _ = strconv.Atoi(tab_keyval[1])
				}

			}
			msg := message{
				msgType: msgType,
				count:   count,
			}
			if msgType == -1 {
				rcvmsg = ""
				mutex.Unlock()
				continue
			}
			handleMessage(msg, conn)
			mutex.Unlock()
			rcvmsg = ""
		}
	}()
	l := log.New(os.Stderr, "", 0)
	for {
		_, message, err := conn.ReadMessage()
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
		count, _ = strconv.Atoi(data.Number)
		l.Printf("Received message: %d\n", count)

		//排队中 请耐心等待
		fmt.Printf("/=receiver=%d/=type=demandeSC/=sender=%d/=hlg=%d\n", nom*(-1), nom, 0)
		msg := &myData{
			Text:   "排队中 请耐心等待",
			Number: strconv.Itoa(stock),
		}
		err = conn.WriteJSON(msg)
		if err != nil {
			l.Println("write:", err)
			return
		}
		mutex.Unlock()
		// /=type=demandeSC/=sender=1/=hlg=0/=receiver=-1
		// /=type=demandeSC/=sender=1/=hlg=0/=receiver=-1
	}

}
func handleMessage(msg message, conn *websocket.Conn) {
	switch msg.msgType {
	case permetSC:
		//轮到您了 正在尝试抢购
		time.Sleep(time.Duration(2) * time.Second)
		if (stock-count) >= 0 && count > 0 {
			stock -= count
			//抢购成功 本次抢购 x 件
			fmt.Printf("/=receiver=%d/=type=finSC/=sender=%d/=hlg=%d/=count=%d\n", nom*-1, nom, 0, stock)
			msg := &myData{
				Text:   "抢购成功 本次抢购" + strconv.Itoa(count) + "件",
				Number: strconv.Itoa(stock),
			}
			err := conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("write:", err)
				return
			}
		} else {
			//抢购失败 没货了 感谢您的参与
			fmt.Printf("/=receiver=%d/=type=finSC/=sender=%d/=hlg=%d/=count=%d\n", nom*-1, nom, 0, stock)
			msg := &myData{
				Text:   "抢购失败 库存不足",
				Number: strconv.Itoa(stock),
			}
			err := conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("write:", err)
				return
			}
		}

		if stock == 0 {
			//抢购结束 感谢您的参与
			msg := &myData{
				Text:   "抢购结束 感谢您的参与",
				Number: strconv.Itoa(stock),
			}
			err := conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("write:", err)
				return
			}
			//conn.Close()
			return
		}
	case updateSC:
		stock = msg.count
		msg := &myData{
			Number: strconv.Itoa(stock),
		}
		err := conn.WriteJSON(msg)
		if err != nil {
			fmt.Println("write:", err)
			return
		}
		if stock == 0 {
			//抢购结束 感谢您的参与
			msg := &myData{
				Text:   "抢购结束 感谢您的参与",
				Number: strconv.Itoa(stock),
			}
			err = conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("write:", err)
				return
			}
			//conn.Close()
			return
		}
	}
}

var nom int
var stock = 10
var mutex = &sync.Mutex{}
var count int

func main() {
	var p = flag.String("p", "4444", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")
	flag.IntVar(&nom, "n", 1, "nom de site app")
	fmt.Printf(string(nom))
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
		//fmt.Println("WebSocket connected")
		handleWebSocket(conn)
	})
	http.ListenAndServe(*addr+":"+*p, nil)
}

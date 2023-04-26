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
	Text   string `json:"text"`
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
			count, _ := strconv.Atoi(data.Number)
			//fmt.Printf("Received message: %d\n", count)

			//排队中 请耐心等待
			fmt.Printf("/=type=demandeSC/=sender=%d/=hlg=%d/=receiver=%d", nom, 0, nom*-1)
			msg := &myData{
				Text: "排队中 请耐心等待",
			}
			err = conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("write:", err)
				return
			}
			//轮到您了 正在尝试抢购
			if (stock - count) > 0 {
				stock -= count
				//抢购成功 本次抢购 x 件
				fmt.Printf("/=type=finSC/=sender=%d/=hlg=%d/=receiver=%d/=count=%d", nom, 0, nom*-1, count)
				msg := &myData{
					Text: "抢购成功 本次抢购" + strconv.Itoa(count) + "件",
				}
				err = conn.WriteJSON(msg)
				if err != nil {
					fmt.Println("write:", err)
					return
				}
			} else {
				//抢购失败 没货了 感谢您的参与
				fmt.Printf("/=type=finSC/=sender=%d/=hlg=%d/=receiver=%d/=count=0", nom, 0, nom*-1)
				msg := &myData{
					Text: "抢购失败 没货了 感谢您的参与",
				}
				err = conn.WriteJSON(msg)
				if err != nil {
					fmt.Println("write:", err)
					return
				}
			}

		}
	}()

	// 向客户端发送数字
	for {
		msg := &myData{
			Number: strconv.Itoa(stock),
		}
		err := conn.WriteJSON(msg)
		if err != nil {
			fmt.Println("write:", err)
			return
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
}

var nom = -1
var stock = 10

func main() {
	var port = flag.String("port", "4444", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")
	flag.IntVar(&nom, "n", -1, "nom de site app")
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

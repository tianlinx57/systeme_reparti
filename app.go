package main

import (
	"flag"
	"fmt"
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
	updateHorloge
	donneSnap
)

type message struct {
	msgType       messageType
	count         int
	snapshot      string
	snapshot_time string
}

type myData struct {
	Number       string `json:"number"`
	Text         string `json:"text"`
	MyLock       string `json:"mylock"`
	Horloge      string `json:"horloge"`
	Snapshot     string `json:"snapshot"`
	Snapshottime string `json:"snapshot_time"`
}

// 标准收
func findval(msg string, key string) string {
	if len(msg) < 4 {
		return ""
	}

	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)

	for _, keyval := range tab_allkeyvals {
		//l := log.New(os.Stderr, "", 0)
		//l.Printf(keyval)
		if len(keyval) >= 4 {
			equ := keyval[0:1]
			tabkeyval := strings.Split(keyval[1:], equ)
			if tabkeyval[0] == key {
				return tabkeyval[1]
			}
		}
	}

	return ""
}

// 改进收发 标准函数
var fieldsep = "/"
var keyvalsep = "="

func msg_format(key string, val string) string {
	return fieldsep + keyvalsep + key + keyvalsep + val
}

func msg_send(msg string) {
	display_d("msg_send", "émission de "+msg)
	fmt.Print(msg + "\n")
}

func receive() {
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
		//tab_allkeyval := strings.Split(rcvmsg[1:], rcvmsg[0:1])
		//l.Printf("%q\n", tab_allkeyval)

		s_receiver := findval(rcvmsg, "receiver")
		if s_receiver != "" {
			receiver, _ = strconv.Atoi(s_receiver)
			if receiver != nom {
				msgType = -1
				mutex.Unlock()
				rcvmsg = ""
				continue
			}
		}

		s_type := findval(rcvmsg, "type")
		if s_type != "" {
			switch s_type {
			case "updateSC":
				msgType = updateSC
			case "permetSC":
				msgType = permetSC
			case "updateHorloge":
				msgType = updateHorloge
			case "donneSnap":
				msgType = donneSnap

			default:
				msgType = -1
				l.Println("Invalid message type. Please try again.")
			}
		}

		s_count := findval(rcvmsg, "count")
		if s_count != "" {
			count, _ = strconv.Atoi(s_count)
		}

		s_snapshot := findval(rcvmsg, "snapshot")
		s_snapshot_time := findval(rcvmsg, "snapshot_time")

		s_hlg := findval(rcvmsg, "hlg")
		if s_hlg != "" {
			horloge, _ = strconv.Atoi(s_hlg)
		}

		msg := message{
			msgType:       msgType,
			count:         count,
			snapshot:      s_snapshot,
			snapshot_time: s_snapshot_time,
		}
		if msgType == -1 {
			rcvmsg = ""
			mutex.Unlock()
			continue
		}
		handleMessage(msg)
		mutex.Unlock()
		rcvmsg = ""
	}
}

func handleMessage(msg message) {
	switch msg.msgType {
	case permetSC:
		//轮到您了 正在尝试抢购
		time.Sleep(time.Duration(2) * time.Second)
		if (stock-count) >= 0 && count > 0 {
			stock -= count
			//抢购成功 本次抢购 x 件
			//fmt.Printf("/=receiver=%d/=type=finSC/=sender=%d/=hlg=%d/=count=%d\n", nom*-1, nom, 0, stock)
			msg_send(msg_format("receiver", strconv.Itoa(nom*(-1))) + msg_format("type", "finSC") + msg_format("sender", strconv.Itoa(nom)) + msg_format("hlg", strconv.Itoa(horloge)) + msg_format("count", strconv.Itoa(stock)))
			status = "unlocked" //是否禁止前端输入
			ws_send("抢购成功 本次抢购"+strconv.Itoa(count)+"件", status)
			//msg := &myData{
			//	Number:  strconv.Itoa(stock),
			//	Text:    "抢购成功 本次抢购" + strconv.Itoa(count) + "件",
			//	MyLock:  "unlocked",
			//	Horloge: strconv.Itoa(horloge),
			//}
		} else {
			//抢购失败 没货了 感谢您的参与
			//fmt.Printf("/=receiver=%d/=type=finSC/=sender=%d/=hlg=%d/=count=%d\n", nom*-1, nom, 0, stock)
			msg_send(msg_format("receiver", strconv.Itoa(nom*(-1))) + msg_format("type", "finSC") + msg_format("sender", strconv.Itoa(nom)) + msg_format("hlg", strconv.Itoa(horloge)) + msg_format("count", strconv.Itoa(stock)))
			status = "unlocked" //是否禁止前端输入
			ws_send("抢购失败 库存不足", status)
			//msg := &myData{
			//	Number:  strconv.Itoa(stock),
			//	Text:    "抢购失败 库存不足",
			//	MyLock:  "unlocked",
			//	Horloge: strconv.Itoa(horloge),
			//}
		}

		if stock == 0 {
			//抢购结束 感谢您的参与
			status = "locked" //是否禁止前端输入
			ws_send("抢购结束 感谢您的参与", status)
			//msg := &myData{
			//	Number:  strconv.Itoa(stock),
			//	Text:    "抢购结束 感谢您的参与",
			//	MyLock:  "locked",
			//	Horloge: strconv.Itoa(horloge),
			//}

			//conn.Close()
			return
		}
	case updateSC:
		stock = msg.count
		ws_send("更新库存", status)
		//msg := &myData{
		//	Number: strconv.Itoa(stock),
		//}
		//}
		if stock == 0 {
			//抢购结束 感谢您的参与
			status = "locked" //是否禁止前端输入
			ws_send("抢购结束 感谢您的参与", status)
			//msg := &myData{
			//	Text:    "抢购结束 感谢您的参与",
			//	Number:  strconv.Itoa(stock),
			//	MyLock:  "locked",
			//	Horloge: strconv.Itoa(horloge),
			//}

			//conn.Close()
			return
		}
	case updateHorloge:
		//抢购结束 感谢您的参与
		ws_send("更新时钟", status)
		//msg := &myData{
		//	Text:    "更新时钟",
		//	MyLock:  "unlocked",
		//	Horloge: strconv.Itoa(horloge),
		//	Number:  strconv.Itoa(stock),
		//}

		//conn.Close()
		return

	case donneSnap:
		status = "unlocked"
		msg := &myData{
			Number:       strconv.Itoa(stock),
			Text:         "备份内容",
			MyLock:       "unlocked",
			Horloge:      strconv.Itoa(horloge),
			Snapshot:     msg.snapshot,
			Snapshottime: msg.snapshot_time,
		}
		err := ws.WriteJSON(msg)
		if err != nil {
			display_d("write:", string(err.Error()))
			return
		}
	}

}

var nom int
var stock = 10
var mutex = &sync.Mutex{}
var count int
var horloge = 0
var status = "unlocked" //是否禁止前端输入

func main() {
	var p = flag.String("p", "4444", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")
	flag.IntVar(&nom, "n", 1, "nom de site app")
	fmt.Printf(string(nom))
	flag.Parse()

	http.HandleFunc("/ws", do_websocket)
	http.ListenAndServe(*addr+":"+*p, nil)

}

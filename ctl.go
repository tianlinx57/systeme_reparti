package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

const N = 3

type messageType int

const (
	request messageType = iota
	release
	ack
	demandeSC
	finSC
	demandeSnap
	finSnap
)

type message struct {
	msgType     messageType
	logicalTime int
	sender      int
	receiver    int
	count       int
	h1          int
	h2          int
	h3          int
}

type site struct {
	id          int
	logicalTime int
	tab         [N + 1][2]int
}

// 改进收发 标准函数
var fieldsep = "/"
var keyvalsep = "="

func msg_format(key string, val string) string {
	return fieldsep + keyvalsep + key + keyvalsep + val
}

func msg_send(msg string) {
	fmt.Print(msg + "\n")
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

// 移除字符串中的不可打印字符
func removeUnprintableChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, s)
}

func (s *site) run() {
	var rcvmsg string
	l := log.New(os.Stderr, "", 0)
	var logicalTime int
	var sender int
	var receiver int
	var count int
	var h1 int
	var h2 int
	var h3 int
	var msgType messageType
	for {
		msgType = -1
		rcvmsg = ""
		fmt.Scanln(&rcvmsg)
		mutex.Lock()
		// 移除输入字符串中的不可打印字符 !!! 很关键 gpt4教的
		rcvmsg = removeUnprintableChars(rcvmsg)
		//

		//separator := rcvmsg[0:1]
		//tab_allkeyval := strings.Split(rcvmsg[1:], separator)
		////l.Printf("%q\n", tab_allkeyval)
		//for _, keyval := range tab_allkeyval {
		//	tab_keyval := strings.Split(keyval[1:], keyval[0:1])
		//	//l.Printf("  %q\n", tab_keyval)
		//	//l.Printf("  key : %s  val : %s\n", tab_keyval[0], tab_keyval[1])
		//	//如果不是自己该收的转发，应该发给app层的丢弃
		l.Printf("%d message reçu: %s\n", s.id, rcvmsg)
		s_receiver := findval(rcvmsg, "receiver")
		if s_receiver != "" {
			receiver, _ = strconv.Atoi(s_receiver)
			if receiver != s.id {
				if receiver > 0 {
					//l.Printf("zhuanfa")
					fmt.Println(rcvmsg)
				}
				rcvmsg = ""
				mutex.Unlock()
				continue
			}
		}

		s_type := findval(rcvmsg, "type")
		if s_type != "" {
			switch s_type {
			case "request":
				msgType = request
			case "release":
				msgType = release
			case "ack":
				msgType = ack
			case "demandeSC":
				msgType = demandeSC
			case
				"finSC":
				msgType = finSC
			case
				"demandeSnap":
				msgType = demandeSnap
			case
				"finSnap":
				msgType = finSnap
			default:
				msgType = -1
				//l.Println("Invalid message type. Please try again.")
			}
		}
		h1 = 0
		h2 = 0
		h3 = 0
		s_sender := findval(rcvmsg, "sender")
		if s_sender != "" {
			sender, _ = strconv.Atoi(s_sender)
		}

		s_hlg := findval(rcvmsg, "hlg")
		if s_hlg != "" {
			logicalTime, _ = strconv.Atoi(s_hlg)
		}

		s_h1 := findval(rcvmsg, "h1")
		if s_h1 != "" {
			h1, _ = strconv.Atoi(s_h1)
		}
		s_h2 := findval(rcvmsg, "h2")
		if s_h2 != "" {
			h2, _ = strconv.Atoi(s_h2)
		}
		s_h3 := findval(rcvmsg, "h3")
		if s_h3 != "" {
			h3, _ = strconv.Atoi(s_h3)
		}

		s_count := findval(rcvmsg, "count")
		if s_hlg != "" {
			count, _ = strconv.Atoi(s_count)
		}

		if msgType != release && msgType != finSC {
			count = 0
		}
		msg := message{
			msgType:     msgType,
			logicalTime: logicalTime,
			sender:      sender,
			receiver:    receiver,
			count:       count,
			h1:          h1,
			h2:          h2,
			h3:          h3,
		}
		if msgType == -1 {
			rcvmsg = ""
			mutex.Unlock()
			continue
		}
		s.handleMessage(msg)
		rcvmsg = ""
		mutex.Unlock()
	}
}

func (s *site) handleMessage(msg message) {
	switch msg.msgType {
	case request:
		s.logicalTime = recaler(s.logicalTime, msg.logicalTime)

		arr := []int{0, msg.h1, msg.h2, msg.h3}
		horloge_vec = calVec(horloge_vec, arr)
		horloge_vec[s.id] += 1

		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))
		s.tab[msg.sender][0] = 0
		s.tab[msg.sender][1] = msg.logicalTime
		//fmt.Printf("Sending ack from %d to %d with logical time %d\n", s.id, msg.sender, s.logicalTime)
		msg_send(msg_format("receiver", strconv.Itoa(msg.sender)) + msg_format("type", "ack") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("h1", strconv.Itoa(horloge_vec[1])) + msg_format("h2", strconv.Itoa(horloge_vec[2])) + msg_format("h3", strconv.Itoa(horloge_vec[3])))
		//fmt.Printf("/=receiver=%d/=type=ack/=sender=%d/=hlg=%d\n", msg.sender, s.id, s.logicalTime)
	case release:
		s.logicalTime = recaler(s.logicalTime, msg.logicalTime)

		arr := []int{0, msg.h1, msg.h2, msg.h3}
		horloge_vec = calVec(horloge_vec, arr)
		horloge_vec[s.id] += 1

		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))
		s.tab[msg.sender][0] = 1
		s.tab[msg.sender][1] = msg.logicalTime
		//如果不是最新的release就不要告诉app了（再看看逻辑）
		flag := true
		for i := 1; i <= N; i++ {
			if s.tab[i][1] > msg.logicalTime && s.tab[i][0] == 1 {
				flag = false
			}
		}
		if flag {
			msg_send(msg_format("receiver", strconv.Itoa(s.id*-1)) + msg_format("type", "updateSC") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("count", strconv.Itoa(msg.count)))
			//fmt.Printf("/=receiver=%d/=type=updateSC/=sender=%d/=hlg=%d/=count=%d\n", s.id*-1, s.id, s.logicalTime, msg.count)
		}
	case ack:
		s.logicalTime = recaler(s.logicalTime, msg.logicalTime)

		arr := []int{0, msg.h1, msg.h2, msg.h3}
		horloge_vec = calVec(horloge_vec, arr)
		horloge_vec[s.id] += 1

		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))
		if s.tab[msg.sender][0] != 0 {
			s.tab[msg.sender][0] = 2
			s.tab[msg.sender][1] = msg.logicalTime
		}
	case demandeSC:
		s.logicalTime = s.logicalTime + 1

		horloge_vec[s.id] += 1

		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))
		s.tab[s.id][0] = 0
		s.tab[s.id][1] = s.logicalTime
		for i := 1; i <= N; i++ {
			if i != s.id {
				//fmt.Printf("Sending request from %d to %d with logical time %d\n", s.id, i, s.logicalTime)
				//fmt.Printf("/=receiver=%d/=type=request/=sender=%d/=hlg=%d\n", i, s.id, s.logicalTime)
				msg_send(msg_format("receiver", strconv.Itoa(i)) + msg_format("type", "request") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("h1", strconv.Itoa(horloge_vec[1])) + msg_format("h2", strconv.Itoa(horloge_vec[2])) + msg_format("h3", strconv.Itoa(horloge_vec[3])))
				//l := log.New(os.Stderr, "", 0)
				//l.Printf("/=receiver=%d/=type=request/=sender=%d/=hlg=%d\n", i, s.id, s.logicalTime)
			}
		}
		inable = true
	case finSC:
		s.logicalTime = s.logicalTime + 1
		horloge_vec[s.id] += 1

		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))
		s.tab[s.id][0] = 1
		s.tab[s.id][1] = s.logicalTime
		count := last_stock - msg.count
		last_stock = msg.count
		snapshot = append(snapshot, ",horloge_vectorielle:["+strconv.Itoa(horloge_vec[1])+","+strconv.Itoa(horloge_vec[2])+","+strconv.Itoa(horloge_vec[3])+"],site:"+strconv.Itoa(s.id)+",nombre_achat:"+strconv.Itoa(count))
		for i := 1; i <= N; i++ {
			if i != s.id {
				//fmt.Printf("Sending release from %d to %d with logical time %d\n", s.id, i, s.logicalTime)
				//fmt.Printf("/=receiver=%d/=type=release/=sender=%d/=hlg=%d/=count=%d\n", i, s.id, s.logicalTime, msg.count)
				msg_send(msg_format("receiver", strconv.Itoa(i)) + msg_format("type", "release") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("count", strconv.Itoa(msg.count)) + msg_format("h1", strconv.Itoa(horloge_vec[1])) + msg_format("h2", strconv.Itoa(horloge_vec[2])) + msg_format("h3", strconv.Itoa(horloge_vec[3])))
			}
		}
	case demandeSnap:
		if msg.sender > 0 {
			s.logicalTime = recaler(s.logicalTime, msg.logicalTime)
			s.tab[msg.sender][1] = msg.logicalTime
			s.tab[s.id][1] = s.logicalTime

			arr := []int{0, msg.h1, msg.h2, msg.h3}
			horloge_vec = calVec(horloge_vec, arr)
			horloge_vec[s.id] += 1

		} else {
			s.logicalTime = s.logicalTime + 1
			s.tab[s.id][1] = s.logicalTime

			horloge_vec[s.id] += 1
		}
		if couleur == 0 {
			msg_send(msg_format("receiver", strconv.Itoa((s.id%N)+1)+msg_format("type", "demandeSnap")+msg_format("sender", strconv.Itoa(s.id))+msg_format("hlg", strconv.Itoa(s.logicalTime))+msg_format("h1", strconv.Itoa(horloge_vec[1]))+msg_format("h2", strconv.Itoa(horloge_vec[2]))+msg_format("h3", strconv.Itoa(horloge_vec[3]))))
			snapshot_time := "[" + strconv.Itoa(horloge_vec[1]) + "," + strconv.Itoa(horloge_vec[2]) + "," + strconv.Itoa(horloge_vec[3]) + "]"
			msg_send(msg_format("receiver", strconv.Itoa((s.id*(-1)))+msg_format("type", "donneSnap")+msg_format("sender", strconv.Itoa(s.id))+msg_format("hlg", strconv.Itoa(s.logicalTime))+msg_format("snapshot", strings.Join(snapshot, "@"))+msg_format("snapshot_time", snapshot_time)))
			couleur = 1
		} else if couleur == 1 {
			msg_send(msg_format("receiver", strconv.Itoa((s.id%N)+1)+msg_format("type", "finSnap")+msg_format("sender", strconv.Itoa(s.id))+msg_format("hlg", strconv.Itoa(s.logicalTime))+msg_format("h1", strconv.Itoa(horloge_vec[1]))+msg_format("h2", strconv.Itoa(horloge_vec[2]))+msg_format("h3", strconv.Itoa(horloge_vec[3]))))
			msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))
			couleur = 0
		}

	case finSnap:
		s.logicalTime = recaler(s.logicalTime, msg.logicalTime)
		s.tab[s.id][1] = s.logicalTime
		s.tab[msg.sender][1] = msg.logicalTime

		arr := []int{0, msg.h1, msg.h2, msg.h3}
		horloge_vec = calVec(horloge_vec, arr)
		horloge_vec[s.id] += 1

		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))
		if couleur == 1 {
			couleur = 0
			msg_send(msg_format("receiver", strconv.Itoa((s.id%N)+1)+msg_format("type", "finSnap")+msg_format("sender", strconv.Itoa(s.id))+msg_format("hlg", strconv.Itoa(s.logicalTime))+msg_format("h1", strconv.Itoa(horloge_vec[1]))+msg_format("h2", strconv.Itoa(horloge_vec[2]))+msg_format("h3", strconv.Itoa(horloge_vec[3]))))
		}

	}
	if msg.msgType != demandeSC && msg.msgType != finSC && msg.msgType != demandeSnap && msg.msgType != finSnap {
		s.checkCriticalSection()
	}
}

func (s *site) checkCriticalSection() {
	if s.tab[s.id][0] == 0 {
		isSmallest := true
		for k := 1; k <= N; k++ {
			if k != s.id && (s.tab[k][1] < s.tab[s.id][1] || (s.tab[k][1] == s.tab[s.id][1] && k < s.id)) {
				isSmallest = false
				break
			}
		}
		if isSmallest {
			//一次demande只要通知上层进去一次就好 原来算法的漏洞 傻逼玩意
			if inable {
				//fmt.Printf("Sending debutSC from %d to %d \n", s.id, s.id*-1)
				//fmt.Printf("/=receiver=%d/=type=permetSC/=sender=%d/=hlg=%d\n", -1*s.id, s.id, s.logicalTime)
				msg_send(msg_format("receiver", strconv.Itoa(-1*s.id)) + msg_format("type", "permetSC") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("h1", strconv.Itoa(horloge_vec[1])) + msg_format("h2", strconv.Itoa(horloge_vec[2])) + msg_format("h3", strconv.Itoa(horloge_vec[3])))
				inable = false
			}

		}
	}
}

func recaler(x, y int) int {
	if x < y {
		return y + 1
	}
	return x + 1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func calVec(x, y []int) []int {
	res := make([]int, 4)
	res[0] = 0
	res[1] = max(x[1], y[1])
	res[2] = max(x[2], y[2])
	res[3] = max(x[3], y[3])
	return res
}

var mutex = &sync.Mutex{}
var inable = true
var couleur = 0

// 因为我们每次传递的是剩余库存值，要和上次相减得出本次购买量
var last_stock = 10

// 快照
var snapshot = make([]string, 0)

// 向量时钟
var horloge_vec = make([]int, 4)

func main() {
	var nom = 1
	var period = 0

	flag.IntVar(&nom, "n", 1, "nom de site")
	flag.IntVar(&period, "t", 1, "timer")
	flag.Parse()

	var s site
	//fmt.Print("Enter the site ID: ")
	s.id = nom
	s.logicalTime = 0
	for i := 0; i < N+1; i++ {
		s.tab[i][0] = 1
		horloge_vec[i] = 1
	}

	s.run()
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const N = 3

type messageType int

const (
	request messageType = iota
	release
	ack
	demandeSC
	finSC
)

type message struct {
	msgType     messageType
	logicalTime int
	sender      int
	receiver    int
}

type site struct {
	id          int
	logicalTime int
	tab         [N][2]int
}

func (s *site) run() {
	var rcvmsg string
	l := log.New(os.Stderr, "", 0)
	var logicalTime int
	var sender int
	var receiver int
	var msgType messageType

	for {
		fmt.Scanln(&rcvmsg)
		mutex.Lock()
		l.Printf("message reçu : %s\n", rcvmsg)

		tab_allkeyval := strings.Split(rcvmsg[1:], rcvmsg[0:1])
		//l.Printf("%q\n", tab_allkeyval)
		for _, keyval := range tab_allkeyval {
			tab_keyval := strings.Split(keyval[1:], keyval[0:1])
			//l.Printf("  %q\n", tab_keyval)
			//l.Printf("  key : %s  val : %s\n", tab_keyval[0], tab_keyval[1])
			//如果不是自己该收的转发，应该发给app层的丢弃
			if tab_keyval[0] == "receiver" {
				receiver, _ = strconv.Atoi(tab_keyval[1])
				if receiver != s.id {
					if receiver > 0 {
						fmt.Println(rcvmsg)
					}
					msgType = -1
					break
				}
			} else if tab_keyval[0] == "type" {
				switch tab_keyval[1] {
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
				default:
					msgType = -1
					l.Println("Invalid message type. Please try again.")
					break
				}
			} else if tab_keyval[0] == "sender" {
				sender, _ = strconv.Atoi(tab_keyval[1])
			} else if tab_keyval[0] == "hlg" {
				logicalTime, _ = strconv.Atoi(tab_keyval[1])
			}

		}
		msg := message{
			msgType:     msgType,
			logicalTime: logicalTime,
			sender:      sender,
			receiver:    receiver,
		}
		if msgType == -1 {
			rcvmsg = ""
			continue
		}
		s.handleMessage(msg)
		mutex.Unlock()
		rcvmsg = ""
	}
}

func (s *site) handleMessage(msg message) {
	s.logicalTime = max(s.logicalTime, msg.logicalTime) + 1
	switch msg.msgType {
	case request:
		s.tab[msg.sender][0] = 0
		s.tab[msg.sender][1] = msg.logicalTime
		fmt.Printf("Sending ack from %d to %d with logical time %d\n", s.id, msg.sender, s.logicalTime)
	case release:
		s.tab[msg.sender][0] = 1
		s.tab[msg.sender][1] = msg.logicalTime
	case ack:
		if s.tab[msg.sender][0] != 0 {
			s.tab[msg.sender][0] = 2
			s.tab[msg.sender][1] = msg.logicalTime
		}
	case demandeSC:
		s.tab[s.id][0] = 0
		s.tab[s.id][1] = s.logicalTime
		for i := 0; i < N; i++ {
			if i != s.id {
				fmt.Printf("Sending request from %d to %d with logical time %d\n", s.id, i, s.logicalTime)
			}
		}
	case finSC:
		s.tab[s.id][0] = 1
		s.tab[s.id][1] = s.logicalTime
		for i := 0; i < N; i++ {
			if i != s.id {
				fmt.Printf("Sending release from %d to %d with logical time %d\n", s.id, i, s.logicalTime)
			}
		}
	}
	if msg.msgType != demandeSC && msg.msgType != finSC {
		s.checkCriticalSection()
	}
}

func (s *site) checkCriticalSection() {
	if s.tab[s.id][0] == 0 {
		isSmallest := true
		for k := 0; k < N; k++ {
			if k != s.id && (s.tab[k][1] < s.tab[s.id][1] || (s.tab[k][1] == s.tab[s.id][1] && k < s.id)) {
				isSmallest = false
				break
			}
		}
		if isSmallest {
			fmt.Printf("Sending debutSC from %d to %d \n", s.id, s.id*-1)
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var mutex = &sync.Mutex{}

func main() {
	var nom = 1
	var period = 0

	flag.IntVar(&nom, "n", 1, "nom de site")
	flag.IntVar(&period, "t", 1, "timer")
	flag.Parse()

	var s site
	fmt.Print("Enter the site ID: ")
	s.id = nom
	s.logicalTime = 0
	for i := 0; i < N; i++ {
		s.tab[i][0] = 1
	}

	s.run()
}

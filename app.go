package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func sendperiodic() {
	var sndmsg string
	for {
		mutex.Lock()
		hi = hi + 1
		sndmsg = ",=snd=" + nom + ",=hlg=" + strconv.Itoa(hi) + "\n"
		fmt.Print(sndmsg)
		mutex.Unlock()
		time.Sleep(time.Duration(period) * time.Second)
	}
}

func receive() {
	var rcvmsg string
	l := log.New(os.Stderr, "", 0)

	for {
		fmt.Scanln(&rcvmsg)
		mutex.Lock()
		l.Printf("%s", cyan)
		l.Printf("message reçu : %s hrg_local %d\n", rcvmsg, hi)

		tab_allkeyval := strings.Split(rcvmsg[1:], rcvmsg[0:1])
		//l.Printf("%q\n", tab_allkeyval)

		for _, keyval := range tab_allkeyval {
			tab_keyval := strings.Split(keyval[1:], keyval[0:1])
			//l.Printf("  %q\n", tab_keyval)
			//l.Printf("  key : %s  val : %s\n", tab_keyval[0], tab_keyval[1])
			if tab_keyval[0] == "hlg" {
				hm, _ := strconv.Atoi(tab_keyval[1])
				hi = Max(hi, hm) + 1
				l.Printf("  site : %s  horloge : %d\n", nom, hi)
			}
		}
		/*for i := 1; i < 5; i++ {
			//l.Println("traitement message", i)
			time.Sleep(time.Duration(1) * time.Second)
		}*/
		l.Printf("%s", raz)
		mutex.Unlock()
		rcvmsg = ""
	}
}

var mutex = &sync.Mutex{}
var hi = 0
var nom = ""
var period = 0
var cyan string = "\033[1;36m"
var raz string = "\033[0;00m"

func main() {
	flag.StringVar(&nom, "n", "A", "nom de site")
	flag.IntVar(&period, "t", 1, "timer")
	flag.Parse()
	go sendperiodic()
	go receive()
	for {
		time.Sleep(time.Duration(60) * time.Second)
	} // Pour attendre la fin des goroutines...
}

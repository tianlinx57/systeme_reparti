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

		// Analyser le message reçu et mettre à jour l'horloge locale
		tab_allkeyval := strings.Split(rcvmsg[1:], rcvmsg[0:1])

		for _, keyval := range tab_allkeyval {
			tab_keyval := strings.Split(keyval[1:], keyval[0:1])
			if tab_keyval[0] == "hlg" {
				hm, _ := strconv.Atoi(tab_keyval[1])
				hi = Max(hi, hm) + 1
				l.Printf("  site : %s  horloge : %d\n", nom, hi)
			}
		}

		// Réinitialiser le message reçu et déverrouiller le mutex
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
	// Analyser les arguments en ligne de commande pour définir le nom de site et le timer
	flag.StringVar(&nom, "n", "A", "nom de site")
	flag.IntVar(&period, "t", 1, "timer")
	flag.Parse()

	// Démarrer les goroutines pour envoyer et recevoir des messages périodiques
	go sendperiodic()
	go receive()

	// Boucle infinie pour empêcher le programme de se terminer prématurément
	for {
		time.Sleep(time.Duration(60) * time.Second)
	}
}

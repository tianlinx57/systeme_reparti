// Package principal
package main

// Importation des bibliothèques nécessaires
import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Déclaration des constantes pour les types de messages
type messageType int

const (
	updateSC messageType = iota
	permetSC
)

// Déclaration des variables globales
var nom int
var stock = 10
var mutex = &sync.Mutex{}
var count int

// Structure pour les messages
type message struct {
	msgType messageType
	count   int
}

// Structure pour les données
type myData struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

// Fonction pour gérer la connexion WebSocket
func handleWebSocket(conn *websocket.Conn) {
	// Fermeture de la connexion à la fin de la fonction
	defer conn.Close()
	nom *= -1

	// Envoi du message initial au client
	msg := &myData{
		Number: strconv.Itoa(stock),
	}
	err := conn.WriteJSON(msg)
	if err != nil {
		fmt.Println("write:", err)
		return
	}

	// Lancement d'une goroutine pour traiter les messages reçus
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
			for _, keyval := range tab_allkeyval {
				tab_keyval := strings.Split(keyval[1:], keyval[0:1])
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

	// Lecture des messages entrants et traitement des données
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

		fmt.Printf("/=receiver=%d/=type=demandeSC/=sender=%d/=hlg=%d\n", nom*(-1), nom, 0)
		msg := &myData{
			Text:   "Attendre en ligne s'il vous plaît soyez patient",
			Number: strconv.Itoa(stock),
		}
		err = conn.WriteJSON(msg)
		if err != nil {
			l.Println("write:", err)
			return
		}
		mutex.Unlock()
	}
}

// Fonction pour traiter les messages reçus
func handleMessage(msg message, conn *websocket.Conn) {
	switch msg.msgType {
	case permetSC:
		// Attendre avant de traiter le message
		time.Sleep(time.Duration(2) * time.Second)

		// Vérifier si l'achat est possible
		if (stock-count) >= 0 && count > 0 {
			// Mise à jour du stock
			stock -= count
			fmt.Printf("/=receiver=%d/=type=finSC/=sender=%d/=hlg=%d/=count=%d\n", nom*-1, nom, 0, stock)
			// Envoi du message de réussite d'achat
			msg := &myData{
				Text:   "L'achat est réussi, cet achat" + strconv.Itoa(count) + "foi",
				Number: strconv.Itoa(stock),
			}
			err := conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("write:", err)
				return
			}
		} else {
			// Envoi du message d'échec d'achat
			fmt.Printf("/=receiver=%d/=type=finSC/=sender=%d/=hlg=%d/=count=%d\n", nom*-1, nom, 0, stock)
			msg := &myData{
				Text:   "Échec de l'achat, rupture de stock",
				Number: strconv.Itoa(stock),
			}
			err := conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("write:", err)
				return
			}
		}

		// Envoi du message de fin d'achat si le stock est épuisé
		if stock == 0 {
			msg := &myData{
				Text:   "L'achat est terminé, merci pour votre participation",
				Number: strconv.Itoa(stock),
			}
			err := conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("write:", err)
				return
			}
			return
		}
	case updateSC:
		// Mise à jour du stock
		stock = msg.count
		msg := &myData{
			Number: strconv.Itoa(stock),
		}
		err := conn.WriteJSON(msg)
		if err != nil {
			fmt.Println("write:", err)
			return
		}

		// Envoi du message de fin d'achat si le stock est épuisé
		if stock == 0 {
			msg := &myData{
				Text:   "L'achat est terminé, merci pour votre participation",
				Number: strconv.Itoa(stock),
			}
			err = conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("write:", err)
				return
			}
			return
		}
	}
}

// Fonction principale
func main() {
	// Définition des options de ligne de commande
	var p = flag.String("p", "4444", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")
	flag.IntVar(&nom, "n", 1, "nom de site app")
	fmt.Printf(string(nom))
	flag.Parse()

	// Configuration du gestionnaire de WebSocket
	http.HandleFunc("/ws", func(w http.ResponseWriter, r


	// Configuration du gestionnaire de WebSocket
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Configuration de l'upgrader WebSocket
		var upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		// Mise à niveau de la connexion HTTP en connexion WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("upgrade:", err)
			return
		}
		// Gestion de la connexion WebSocket
		handleWebSocket(conn)
	})
	// Lancement du serveur HTTP
	http.ListenAndServe(*addr+":"+*p, nil)
}
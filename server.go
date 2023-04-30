package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// Définition de la structure de données pour le message JSON
type myData struct {
	Number string `json:"number"`
}

// Fonction pour gérer la connexion WebSocket
func handleWebSocket(conn *websocket.Conn) {
	defer conn.Close()

	// Recevoir des messages du client
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
			num, _ := strconv.Atoi(data.Number)
			fmt.Printf("Message reçu : %d\n", num)
		}
	}()

	// Envoyer des nombres au client
	num := 0
	for {
		num = num + 1
		msg := &myData{
			Number: strconv.Itoa(num),
		}
		err := conn.WriteJSON(msg)
		if err != nil {
			fmt.Println("write:", err)
			return
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
}

func main() {
	var port = flag.String("port", "4444", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")

	flag.Parse()

	// Définir la fonction de gestion pour l'URI "/ws"
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("upgrade:", err)
			return
		}
		fmt.Println("WebSocket connecté")
		handleWebSocket(conn)
	})

	// Démarrer le serveur HTTP
	http.ListenAndServe(*addr+":"+*port, nil)
}

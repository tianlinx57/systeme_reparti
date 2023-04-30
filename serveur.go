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

// Définition de la structure de données pour le message JSON
type myData struct {
	Number  string `json:"number"`
	Text    string `json:"text"`
	MyLock  string `json:"mylock"`
	Horloge string `json:"horloge"`
}

// Connexion WebSocket
var ws *websocket.Conn

// Log d'erreur standard
var stderr = log.New(os.Stderr, "", 0)

// Fonction pour afficher des informations de débogage
func display_d(where string, what string) {
	stderr.Printf(" + %-8.8s : %s\n", where, what)
}

// Fonction pour gérer une connexion WebSocket
func do_websocket(w http.ResponseWriter, r *http.Request) {
	// Configurer l'upgrader pour la connexion WebSocket
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrade:", err)
		return
	}
	// Afficher un message de connexion réussie
	display_d("ws_create", "Client connecté")
	ws = conn
	go receive()
	go ws_receive()
}

// Fonction pour envoyer des messages WebSocket
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

// Fonction pour recevoir des messages WebSocket
func ws_receive() {
	// Fermer la connexion WebSocket à la fin de la réception
	defer ws_close()
	// Inverser le signe du nom pour désigner le récepteur
	nom *= -1
	// Envoyer un message de démarrage au client
	ws_send("start", status)
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

		// Si le message est une demande de snapshot, envoyer une demande au serveur
		if data.Text == "demand snapshot" {
			msg_send(msg_format("receiver", strconv.Itoa(nom*(-1))) + msg_format("type", "demandeSnap") + msg_format("sender", strconv.Itoa(nom)) + msg_format("hlg", strconv.Itoa(horloge)))
			status = "locked"
			ws_send("L'instantané est en cours de génération, veuillez patienter", status)
			mutex.Unlock()
			continue
		}

		count, _ = strconv.Atoi(data.Number)
		l.Printf("Received message: %d\n", count)
		// Envoyer une demande d'exclusion mutuelle au serve
		msg_send(msg_format("receiver", strconv.Itoa(nom*(-1))) + msg_format("type", "demandeSC") + msg_format("sender", strconv.Itoa(nom)) + msg_format("hlg", strconv.Itoa(horloge)))
		status = "locked"
		ws_send("Attendre en ligne s'il vous plaît soyez patient", status)
		mutex.Unlock()
	}
}

// Fonction pour fermer la connexion WebSocket
func ws_close() {
	display_d("ws_close", "Fin des réceptions => fermeture de la websocket")
	ws.Close()
}

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

// Définition du type messageType
type messageType int

// Déclaration des constantes pour les différents types de messages
const (
	updateSC messageType = iota
	permetSC
	updateHorloge
	donneSnap
)

// Déclaration de la structure message
type message struct {
	msgType       messageType
	count         int
	snapshot      string
	snapshot_time string
}

// Déclaration de la structure myData pour stocker les données JSON
type myData struct {
	Number       string `json:"number"`
	Text         string `json:"text"`
	MyLock       string `json:"mylock"`
	Horloge      string `json:"horloge"`
	Snapshot     string `json:"snapshot"`
	Snapshottime string `json:"snapshot_time"`
}

// Fonction findval pour trouver la valeur correspondant à une clé dans un message
func findval(msg string, key string) string {
	if len(msg) < 4 {
		return ""
	}

	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)

	for _, keyval := range tab_allkeyvals {
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

// Déclaration des variables pour les séparateurs de champ et de clé/valeur
var fieldsep = "/"
var keyvalsep = "="

// Fonction msg_format pour formater un message clé-valeur
func msg_format(key string, val string) string {
	return fieldsep + keyvalsep + key + keyvalsep + val
}

// Fonction msg_send pour envoyer un message
func msg_send(msg string) {
	display_d("msg_send", "émission de "+msg)
	fmt.Print(msg + "\n")
}

// La fonction receive est utilisée pour recevoir des messages, les analyser et les traiter en fonction de leur type.
func receive() {
	// Déclaration des variables locales
	var rcvmsg string
	l := log.New(os.Stderr, "", 0)
	l.Printf(string(nom))
	var receiver int
	var msgType messageType
	var count int

	// Boucle infinie pour lire les messages en continu
	for {
		// Lecture du message entrant
		fmt.Scanln(&rcvmsg)
		msgType = -1
		// Verrouillage du mutex pour assurer la synchronisation
		mutex.Lock()
		// Affichage du message reçu
		l.Printf("%d message reçu : %s\n", nom, rcvmsg)

		// Récupération du champ 'receiver' du message et vérification de sa correspondance avec le nom
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

		// Récupération du type de message et traitement en conséquence
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

			// Si le type de message est invalide, on recommence la boucle
			default:
				msgType = -1
				l.Println("Invalid message type. Please try again.")
			}
		}

		// Récupération de la valeur du champ 'count' dans le message
		s_count := findval(rcvmsg, "count")
		if s_count != "" {
			count, _ = strconv.Atoi(s_count)
		}

		// Récupération des valeurs des champs 'snapshot' et 'snapshot_time' dans le message
		s_snapshot := findval(rcvmsg, "snapshot")
		s_snapshot_time := findval(rcvmsg, "snapshot_time")

		// Récupération de la valeur du champ 'hlg' (horloge) dans le message
		s_hlg := findval(rcvmsg, "hlg")
		if s_hlg != "" {
			horloge, _ = strconv.Atoi(s_hlg)
		}

		// Création d'un objet message avec les informations récupérées
		msg := message{
			msgType:       msgType,
			count:         count,
			snapshot:      s_snapshot,
			snapshot_time: s_snapshot_time,
		}

		// Si le type de message est invalide, on recommence la boucle
		if msgType == -1 {
			rcvmsg = ""
			mutex.Unlock()
			continue
		}

		// Traitement du message avec la fonction handleMessage
		handleMessage(msg)
		// Déverrouillage du mutex
		mutex.Unlock()
		// Réinitialisation de la variable rcvmsg pour le prochain tour de boucle
		rcvmsg = ""
	}
}

// La fonction handleMessage traite les messages reçus en fonction de leur type.
func handleMessage(msg message) {
	// Traitement en fonction du type de message
	switch msg.msgType {
	case permetSC:
		// Attendre 2 secondes avant de traiter le message
		time.Sleep(time.Duration(2) * time.Second)
		// Vérifier si le stock est suffisant pour l'achat
		if (stock-count) >= 0 && count > 0 {
			// Réduire le stock
			stock -= count
			// Envoyer un message pour signaler la fin de l'achat
			msg_send(msg_format("receiver", strconv.Itoa(nom*(-1))) + msg_format("type", "finSC") + msg_format("sender", strconv.Itoa(nom)) + msg_format("hlg", strconv.Itoa(horloge)) + msg_format("count", strconv.Itoa(stock)))
			status = "unlocked"
			ws_send("L'achat est réussi, cet achat"+strconv.Itoa(count)+"foi", status)
		} else {
			// Envoyer un message pour signaler l'échec de l'achat
			msg_send(msg_format("receiver", strconv.Itoa(nom*(-1))) + msg_format("type", "finSC") + msg_format("sender", strconv.Itoa(nom)) + msg_format("hlg", strconv.Itoa(horloge)) + msg_format("count", strconv.Itoa(stock)))
			status = "unlocked"
			ws_send("Échec de l'achat, rupture de stock", status)
		}

		// Si le stock est épuisé, verrouiller les achats et informer les utilisateurs
		if stock == 0 {
			status = "locked"
			ws_send("L'achat est terminé, merci pour votre participation", status)
			return
		}
	case updateSC:
		// Mettre à jour le stock avec la nouvelle valeur
		stock = msg.count
		ws_send("mettre à jour l'inventaire", status)
		// Si le stock est épuisé, verrouiller les achats et informer les utilisateurs
		if stock == 0 {
			status = "locked"
			ws_send("L'achat est terminé, merci pour votre participation", status)
			return
		}
	case updateHorloge:
		// Mettre à jour l'horloge et informer les utilisateurs
		ws_send("mettre à jour l'horloge", status)
		return

	case donneSnap:
		// Traiter les données de la sauvegarde
		status = "unlocked"
		msg := &myData{
			Number:       strconv.Itoa(stock),
			Text:         "sauvegarde",
			MyLock:       "unlocked",
			Horloge:      strconv.Itoa(horloge),
			Snapshot:     msg.snapshot,
			Snapshottime: msg.snapshot_time,
		}
		// Envoyer les données de la sauvegarde au client via WebSocket
		err := ws.WriteJSON(msg)
		if err != nil {
			display_d("write:", string(err.Error()))
			return
		}
	}
}

// Déclaration des variables globales
var nom int
var stock = 10
var mutex = &sync.Mutex{}
var count int
var horloge = 0
var status = "unlocked"

func main() {
	// Déclaration et initialisation des variables pour les arguments de la ligne de commande
	var p = flag.String("p", "4444", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")
	flag.IntVar(&nom, "n", 1, "nom de site app")
	fmt.Printf(string(nom))
	flag.Parse()

	// Configuration du gestionnaire de WebSocket
	http.HandleFunc("/ws", do_websocket)
	// Démarrage du serveur HTTP avec les adresses et ports spécifiés
	http.ListenAndServe(*addr+":"+*p, nil)
}

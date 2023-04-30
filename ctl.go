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

var fieldsep = "/"
var keyvalsep = "="

func msg_format(key string, val string) string {
	// Concatène les chaînes pour créer un message formaté à partir d'une clé et d'une valeur
	return fieldsep + keyvalsep + key + keyvalsep + val
}

func msg_send(msg string) {
	fmt.Print(msg + "\n")
}

func findval(msg string, key string) string {
	// Si la longueur du message est inférieure à 4, retourne une chaîne vide
	if len(msg) < 4 {
		return ""
	}

	// Récupère le séparateur à partir du message
	sep := msg[0:1]
	// Sépare le message en utilisant le séparateur
	tab_allkeyvals := strings.Split(msg[1:], sep)

	// Parcourt les paires clé-valeur
	for _, keyval := range tab_allkeyvals {
		// Si la longueur de la paire clé-valeur est supérieure ou égale à 4
		if len(keyval) >= 4 {
			// Récupère le signe égal à partir de la paire clé-valeur
			equ := keyval[0:1]
			// Sépare la paire clé-valeur en utilisant le signe égal
			tabkeyval := strings.Split(keyval[1:], equ)
			// Si la clé correspond à la clé recherchée, retourne la valeur correspondante
			if tabkeyval[0] == key {
				return tabkeyval[1]
			}
		}
	}

	// Si la clé n'est pas trouvée, retourne une chaîne vide
	return ""
}

func removeUnprintableChars(s string) string {
	// Supprime les caractères non imprimables de la chaîne d'entrée en utilisant une fonction de mappage
	return strings.Map(func(r rune) rune {
		// Si le caractère est imprimable, retourne le caractère
		if unicode.IsPrint(r) {
			return r
		}
		// Sinon, retourne -1 pour supprimer le caractère
		return -1
	}, s)
}

func (s *site) run() {
	// Initialisation des variables
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

	// Boucle infinie pour continuer à recevoir et traiter les messages
	for {
		// Réinitialisation des variables
		msgType = -1
		rcvmsg = ""

		// Lecture du message à partir de l'entrée standard
		fmt.Scanln(&rcvmsg)
		// Verrouillage du mutex pour assurer l'exclusion mutuelle
		mutex.Lock()
		// Suppression des caractères non imprimables du message
		rcvmsg = removeUnprintableChars(rcvmsg)
		// Journalisation du message reçu
		l.Printf("%d message reçu: %s\n", s.id, rcvmsg)
		// Recherche du destinataire dans le message
		s_receiver := findval(rcvmsg, "receiver")
		if s_receiver != "" {
			// Conversion du destinataire en entier
			receiver, _ = strconv.Atoi(s_receiver)
			// Si le destinataire n'est pas égal à l'ID du site, continuez
			if receiver != s.id {
				if receiver > 0 {
					fmt.Println(rcvmsg)
				}
				rcvmsg = ""
				mutex.Unlock()
				continue
			}
		}

		// Recherche du type de message
		s_type := findval(rcvmsg, "type")
		if s_type != "" {
			// Détermination du type de message
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
				//l.Println("Type de message non valide. Veuillez réessayer.")
			}
		}
		// Extraction et conversion des valeurs du message
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

		// Si le type de message n'est pas "release" ou "finSC", réinitialisez le compteur à 0
		if msgType != release && msgType != finSC {
			count = 0
		}
		// Créez un objet message avec les valeurs extraites
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
		// Si le type de message est -1, continuez avec la boucle suivante
		if msgType == -1 {
			rcvmsg = ""
			mutex.Unlock()
			continue
		}
		// Traitez le message avec la méthode handleMessage
		s.handleMessage(msg)
		// Réinitialisez le message reçu
		rcvmsg = ""
		// Déverrouillez le mutex
		mutex.Unlock()
	}
}

func (s *site) handleMessage(msg message) {
	// Traiter les messages en fonction de leur type
	switch msg.msgType {
	case request:
		// Mettre à jour l'horloge logique du site
		s.logicalTime = recaler(s.logicalTime, msg.logicalTime)

		// Mettre à jour l'horloge vectorielle
		arr := []int{0, msg.h1, msg.h2, msg.h3}
		horloge_vec = calVec(horloge_vec, arr)
		horloge_vec[s.id] += 1

		// Envoyer un message pour mettre à jour l'horloge
		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))

		// Mettre à jour l'état et l'horloge du demandeur
		s.tab[msg.sender][0] = 0
		s.tab[msg.sender][1] = msg.logicalTime

		// Envoyer un message de type "ack" au demandeur
		msg_send(msg_format("receiver", strconv.Itoa(msg.sender)) + msg_format("type", "ack") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("h1", strconv.Itoa(horloge_vec[1])) + msg_format("h2", strconv.Itoa(horloge_vec[2])) + msg_format("h3", strconv.Itoa(horloge_vec[3])))
	case release:
		// Mettre à jour l'horloge logique du site
		s.logicalTime = recaler(s.logicalTime, msg.logicalTime)

		// Mettre à jour l'horloge vectorielle
		arr := []int{0, msg.h1, msg.h2, msg.h3}
		horloge_vec = calVec(horloge_vec, arr)
		horloge_vec[s.id] += 1

		// Envoyer un message pour mettre à jour l'horloge
		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))

		// Mettre à jour l'état et l'horloge du demandeur
		s.tab[msg.sender][0] = 1
		s.tab[msg.sender][1] = msg.logicalTime

		// Vérifier si toutes les conditions sont remplies pour mettre à jour le nombre de billets vendus
		flag := true
		for i := 1; i <= N; i++ {
			if s.tab[i][1] > msg.logicalTime && s.tab[i][0] == 1 {
				flag = false
			}
		}
		// Mettre à jour le nombre de billets vendus si les conditions sont remplies
		if flag {
			msg_send(msg_format("receiver", strconv.Itoa(s.id*-1)) + msg_format("type", "updateSC") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("count", strconv.Itoa(msg.count)))
		}
	case ack:
		// Mettre à jour l'horloge logique du site
		s.logicalTime = recaler(s.logicalTime, msg.logicalTime)

		// Mettre à jour l'horloge vectorielle
		arr := []int{0, msg.h1, msg.h2, msg.h3}
		horloge_vec = calVec(horloge_vec, arr)
		horloge_vec[s.id] += 1

		// Envoyer un message pour mettre à jour l'horloge
		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))

		// Mettre à jour l'état et l'horloge du demandeur
		if s.tab[msg.sender][0] != 0 {
			s.tab[msg.sender][0] = 2
			s.tab[msg.sender][1] = msg.logicalTime
		}
	case demandeSC:
		// Incrémenter l'horloge logique du site
		s.logicalTime = s.logicalTime + 1

		// Mettre à jour l'horloge vectorielle
		horloge_vec[s.id] += 1

		// Envoyer un message pour mettre à jour l'horloge
		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))

		// Mettre à jour l'état et l'horloge du site
		s.tab[s.id][0] = 0
		s.tab[s.id][1] = s.logicalTime

		// Envoyer une demande de requête à tous les autres sites
		for i := 1; i <= N; i++ {
			if i != s.id {
				msg_send(msg_format("receiver", strconv.Itoa(i)) + msg_format("type", "request") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("h1", strconv.Itoa(horloge_vec[1])) + msg_format("h2", strconv.Itoa(horloge_vec[2])) + msg_format("h3", strconv.Itoa(horloge_vec[3])))
			}
		}
		inable = true
	case finSC:
		// Incrémenter l'horloge logique du site
		s.logicalTime = s.logicalTime + 1

		// Mettre à jour l'horloge vectorielle
		horloge_vec[s.id] += 1

		// Envoyer un message pour mettre à jour l'horloge
		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))

		// Mettre à jour l'état et l'horloge du site
		s.tab[s.id][0] = 1
		s.tab[s.id][1] = s.logicalTime

		// Calculer la différence de billets vendus
		count := last_stock - msg.count
		last_stock = msg.count
		snapshot = append(snapshot, ",horloge_vectorielle:["+strconv.Itoa(horloge_vec[1])+","+strconv.Itoa(horloge_vec[2])+","+strconv.Itoa(horloge_vec[3])+"],site:"+strconv.Itoa(s.id)+",nombre_achat:"+strconv.Itoa(count))
		// Envoyer un message de libération à tous les autres sites
		for i := 1; i <= N; i++ {
			if i != s.id {
				msg_send(msg_format("receiver", strconv.Itoa(i)) + msg_format("type", "release") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("count", strconv.Itoa(msg.count)) + msg_format("h1", strconv.Itoa(horloge_vec[1])) + msg_format("h2", strconv.Itoa(horloge_vec[2])) + msg_format("h3", strconv.Itoa(horloge_vec[3])))
			}
		}
	case demandeSnap:
		if msg.sender > 0 {
			// Mettre à jour l'horloge logique du site
			s.logicalTime = recaler(s.logicalTime, msg.logicalTime)

			// Mettre à jour l'horloge du site et de l'expéditeur
			s.tab[msg.sender][1] = msg.logicalTime
			s.tab[s.id][1] = s.logicalTime

			// Mettre à jour l'horloge vectorielle
			arr := []int{0, msg.h1, msg.h2, msg.h3}
			horloge_vec = calVec(horloge_vec, arr)
			horloge_vec[s.id] += 1

		} else {
			// Incrémenter l'horloge logique du site
			s.logicalTime = s.logicalTime + 1

			// Mettre à jour l'horloge du site
			s.tab[s.id][1] = s.logicalTime

			// Mettre à jour l'horloge vectorielle
			horloge_vec[s.id] += 1
		}

		// Traiter la demande de snapshot
		if couleur == 0 {
			// Envoyer un message de demande de snapshot au site suivant
			msg_send(msg_format("receiver", strconv.Itoa((s.id%N)+1)) + msg_format("type", "demandeSnap") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("h1", strconv.Itoa(horloge_vec[1])) + msg_format("h2", strconv.Itoa(horloge_vec[2])) + msg_format("h3", strconv.Itoa(horloge_vec[3])))

			// Enregistrer le snapshot et l'heure de l'horloge vectorielle
			snapshot_time := "[" + strconv.Itoa(horloge_vec[1]) + "," + strconv.Itoa(horloge_vec[2]) + "," + strconv.Itoa(horloge_vec[3]) + "]"
			msg_send(msg_format("receiver", strconv.Itoa((s.id*(-1)))+msg_format("type", "donneSnap")+msg_format("sender", strconv.Itoa(s.id))+msg_format("hlg", strconv.Itoa(s.logicalTime))+msg_format("snapshot", strings.Join(snapshot, "@"))+msg_format("snapshot_time", snapshot_time)))

			// Changer la couleur du site
			couleur = 1
		} else if couleur == 1 {
			// Envoyer un message de fin de snapshot au site suivant
			msg_send(msg_format("receiver", strconv.Itoa((s.id%N)+1)) + msg_format("type", "finSnap") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("h1", strconv.Itoa(horloge_vec[1])) + msg_format("h2", strconv.Itoa(horloge_vec[2])) + msg_format("h3", strconv.Itoa(horloge_vec[3])))
			// Mettre à jour l'horloge logique du site
			msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))

			// Changer la couleur du site
			couleur = 0
		}

	case finSnap:
		// Mettre à jour l'horloge logique du site
		s.logicalTime = recaler(s.logicalTime, msg.logicalTime)

		// Mettre à jour l'horloge du site et de l'expéditeur
		s.tab[s.id][1] = s.logicalTime
		s.tab[msg.sender][1] = msg.logicalTime

		// Mettre à jour l'horloge vectorielle
		arr := []int{0, msg.h1, msg.h2, msg.h3}
		horloge_vec = calVec(horloge_vec, arr)
		horloge_vec[s.id] += 1

		// Mettre à jour l'horloge logique du site
		msg_send(msg_format("receiver", strconv.Itoa(s.id*(-1))) + msg_format("type", "updateHorloge") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)))

		// Traiter la fin du snapshot
		if couleur == 1 {
			// Changer la couleur du site
			couleur = 0

			// Envoyer un message de fin de snapshot au site suivant
			msg_send(msg_format("receiver", strconv.Itoa((s.id%N)+1)) + msg_format("type", "finSnap") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("h1", strconv.Itoa(horloge_vec[1])) + msg_format("h2", strconv.Itoa(horloge_vec[2])) + msg_format("h3", strconv.Itoa(horloge_vec[3])))
		}
	}

	// Vérifier si le site peut entrer dans la section critique
	if msg.msgType != demandeSC && msg.msgType != finSC && msg.msgType != demandeSnap && msg.msgType != finSnap {
		s.checkCriticalSection()
	}
}

// La fonction checkCriticalSection vérifie si le site peut entrer dans la section critique
func (s site) checkCriticalSection() {
	// Si l'état du site est en demande de section critique (0)
	if s.tab[s.id][0] == 0 {
		isSmallest := true
		// Parcourir les autres sites pour vérifier si notre site a la priorité
		for k := 1; k <= N; k++ {
			// Si un autre site a une horloge logique plus petite ou égale avec un identifiant plus petit, notre site n'a pas la priorité
			if k != s.id && (s.tab[k][1] < s.tab[s.id][1] || (s.tab[k][1] == s.tab[s.id][1] && k < s.id)) {
				isSmallest = false
				break
			}
		}
		// Si notre site a la priorité
		if isSmallest {
			// Si le site est autorisé à entrer en section critique
			if inable {
				msg_send(msg_format("receiver", strconv.Itoa(-1*s.id)) + msg_format("type", "permetSC") + msg_format("sender", strconv.Itoa(s.id)) + msg_format("hlg", strconv.Itoa(s.logicalTime)) + msg_format("h1", strconv.Itoa(horloge_vec[1])) + msg_format("h2", strconv.Itoa(horloge_vec[2])) + msg_format("h3", strconv.Itoa(horloge_vec[3])))
				inable = false
			}
		}
	}
}

// La fonction recaler ajuste l'horloge logique en fonction des horloges des autres sites
func recaler(x, y int) int {
	if x < y {
		return y + 1
	}
	return x + 1
}

// La fonction max retourne le maximum entre deux entiers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// La fonction calVec calcule la nouvelle horloge vectorielle en comparant deux horloges vectorielles
func calVec(x, y []int) []int {
	res := make([]int, 4)
	res[0] = 0
	res[1] = max(x[1], y[1])
	res[2] = max(x[2], y[2])
	res[3] = max(x[3], y[3])
	return res
}

// Déclaration des variables globales
var mutex = &sync.Mutex{}        // Mutex pour protéger les accès concurrents aux ressources
var inable = true                // Indique si le site est autorisé à entrer en section critique
var couleur = 0                  // Couleur du site pour le snapshot
var last_stock = 10              // Dernier état du stock
var snapshot = make([]string, 0) // Tableau pour stocker les snapshots
var horloge_vec = make([]int, 4) // Horloge vectorielle pour synchroniser les sites

// Fonction principale du programme
func main() {
	var nom = 1    // Identifiant du site
	var period = 0 // Période pour l'envoi des messages

	flag.IntVar(&nom, "n", 1, "nom de site")
	flag.IntVar(&period, "t", 1, "timer")
	flag.Parse()

	// Initialisation du site
	var s site
	s.id = nom        // Définition de l'identifiant du site
	s.logicalTime = 0 // Initialisation de l'horloge logique du site
	for i := 0; i < N+1; i++ {
		s.tab[i][0] = 1    // Initialisation de l'état des autres sites
		horloge_vec[i] = 1 // Initialisation de l'horloge vectorielle
	}

	// Lancement du site
	s.run()
}

package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

//j'ai du refaire toute le code quasiment je vais faire un hangman irl la 
type HangmanGame struct {  
	RemainingAttempts int      `json:"remaining_attempts"`	//je définie une structure pour pouvoir communiquer avec l'html
	WordShown         []string `json:"word_shown"`
	GuessedLetters    []string `json:"guessed_letters"`
	TargetWord        string   `json:"target_word"`
	GameStatus        string   `json:"game_status"`
	Difficulty        string   `json:"difficulty"` 
}

var currentGame *HangmanGame 

func main() { //bah main
	server()
}

func initGame(difficulty string) *HangmanGame { //lance les paramètres du hangman
	var filepath string
	switch difficulty {  //jsp pourquoi je fait un commentaire c clair je pense
	case "facile":
		filepath = "dic/words1.txt"
	case "moyen":
		filepath = "dic/words2.txt"
	case "difficile":
		filepath = "dic/words3.txt"
	default:
		filepath = "dic/words.txt" 
	}

	fileIO, err := os.OpenFile(filepath, os.O_RDWR, 0600) //open le file words en question
	if err != nil {
		log.Println("Error opening words file:", err)
		return nil
	}
	defer fileIO.Close()

	rawBytes, err := io.ReadAll(fileIO) //read le fichier ouvers avant 
	if err != nil {
		log.Println("Error reading words file:", err)
		return nil
	}

	lines := strings.Split(string(rawBytes), "\n") //met un espace entre chaque lettre sinon c moche
	rdmnbr := rand.Intn(len(lines)) //prend un nombre aléatoire <= nombre de ligne 
	selecmot := strings.ToUpper(strings.TrimSpace(lines[rdmnbr])) //prend le mot aléatoire 

	game := &HangmanGame{ //donne des valeurs a la struct définie précédement 
		RemainingAttempts: 10,
		TargetWord:        selecmot,
		WordShown:         make([]string, len(selecmot)),
		GuessedLetters:    []string{},
		GameStatus:        "ongoing",
		Difficulty:        difficulty, 
	}

	for i := range game.WordShown { // transforme le mot en tiret
		game.WordShown[i] = "_"
	}

	for i := 0; i < len(selecmot)/2-1; i++ { //Révèle certaines lettres du mot 
		rdmindex := rand.Intn(len(selecmot))
		for game.WordShown[rdmindex] != "_" { //être sur que ce soit pas encore la même lettre
			rdmindex = rand.Intn(len(selecmot))
		}
		game.WordShown[rdmindex] = string(selecmot[rdmindex])
	}

	return game
}

func penduHandler(w http.ResponseWriter, r *http.Request) { //handler pendu
	switch r.Method {
	case "GET":
		difficulty := r.URL.Query().Get("difficulty") //prend la variable difficulté envoyer en post
		currentGame = initGame(difficulty) 

		tmpl, err := template.ParseFiles("./html/pendu.html")
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, currentGame); err != nil {
			http.Error(w, "Error la template", http.StatusInternalServerError)
			return
		}

	case "POST": //si une lettre est envoyé 
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		guess := strings.ToUpper(r.Form.Get("guess")) //prend la valeur de guess la met en majuscule pour la comparer 
		processGuess(guess)

		tmpl, err := template.ParseFiles("./html/pendu.html")
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, currentGame); err != nil {
			http.Error(w, "Error ici template", http.StatusInternalServerError)
			return
		}
	}
}

func processGuess(guess string) {
	if currentGame == nil || currentGame.GameStatus != "ongoing" { //vérifie que la game continue bien 
		return
	}

	if guess == currentGame.TargetWord {// si guess = le mot bah win
		currentGame.WordShown = strings.Split(currentGame.TargetWord, "")
		currentGame.GameStatus = "won"
		return
	}

	currentGame.GuessedLetters = append(currentGame.GuessedLetters, guess)

	found := false //check si la lettre est dans le mot 		
	for i, char := range currentGame.TargetWord {
		if string(char) == guess && currentGame.WordShown[i] == "_" {
			currentGame.WordShown[i] = guess
			found = true
		}
	}

	if !found {
		currentGame.RemainingAttempts--
		if len(guess) > 1{
			currentGame.RemainingAttempts--
		}
	}

	if currentGame.RemainingAttempts <= 0 {
		currentGame.GameStatus = "lost"
	}

	if strings.Join(currentGame.WordShown, "") == currentGame.TargetWord {
		currentGame.GameStatus = "won"
	}
}

func diffhandler(w http.ResponseWriter, r *http.Request) { //handler pour la difficulté
	if r.Method != http.MethodGet {
		http.Error(w, "non", http.StatusMethodNotAllowed) //autorise seulement les gets psk on est en cyber 
		return
	}

	tmpl, err := template.ParseFiles("./html/difficulté.html") 
	if err != nil {
		http.Error(w, "template non", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, currentGame); err != nil {
		http.Error(w, "template non", http.StatusInternalServerError)
		return
	}
}

func reglehandler(w http.ResponseWriter, r *http.Request) { //handler pour lhtml des regles 
	if r.Method != http.MethodGet {
		http.Error(w, "non", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.ParseFiles("./html/règles.html") //renvoi l'html
	if err != nil {
		http.Error(w, "template non", http.StatusInternalServerError) 
		return
	}

	if err := tmpl.Execute(w, currentGame); err != nil { //check les erreurs
		http.Error(w, "template non", http.StatusInternalServerError)
		return
	}
}

func server() { //gère les chemins du serveur et les urls 
	fileServer := http.FileServer(http.Dir("./html"))
	http.Handle("/", fileServer)

	http.HandleFunc("/pendu", penduHandler)
	http.HandleFunc("/diff", diffhandler)
	http.HandleFunc("/règles", reglehandler)

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs)) //la triche pour le css qui marhce pas parce que c de la merde

	musique := http.FileServer(http.Dir("./musique"))
	http.Handle("/musique/", http.StripPrefix("/musique/", musique))

	fmt.Println("Server running at http://localhost:7080/") 
	if err := http.ListenAndServe(":7080", nil); err != nil {
		log.Fatal(err)
	}
}

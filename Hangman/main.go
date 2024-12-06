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

type HangmanGame struct {
	RemainingAttempts int      `json:"remaining_attempts"`
	WordShown         []string `json:"word_shown"`
	GuessedLetters    []string `json:"guessed_letters"`
	TargetWord        string   `json:"target_word"`
	GameStatus        string   `json:"game_status"`
}

var currentGame *HangmanGame

func main() {
	server()
}

func initGame() *HangmanGame {
	fileIO, err := os.OpenFile("dic/words.txt", os.O_RDWR, 0600)
	if err != nil {
		log.Println("Error opening words file:", err)
		return nil
	}
	defer fileIO.Close()

	rawBytes, err := io.ReadAll(fileIO)
	if err != nil {
		log.Println("Error reading words file:", err)
		return nil
	}

	lines := strings.Split(string(rawBytes), "\n")
	rdmnbr := rand.Intn(len(lines))
	selecmot := strings.ToUpper(strings.TrimSpace(lines[rdmnbr]))

	game := &HangmanGame{
		RemainingAttempts: 10,
		TargetWord:        selecmot,
		WordShown:         make([]string, len(selecmot)),
		GuessedLetters:    []string{},
		GameStatus:        "ongoing",
	}

	for i := range game.WordShown {
		game.WordShown[i] = "_"
	}

	for i := 0; i < len(selecmot)/2-1; i++ {
		rdmindex := rand.Intn(len(selecmot))
		for game.WordShown[rdmindex] != "_" {
			rdmindex = rand.Intn(len(selecmot))
		}
		game.WordShown[rdmindex] = string(selecmot[rdmindex])
	}

	return game
}

func penduHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		currentGame = initGame()
		tmpl, err := template.ParseFiles("./html/pendu.html")
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, currentGame); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}

	case "POST":
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		guess := strings.ToUpper(r.Form.Get("guess"))
		processGuess(guess)
		tmpl, err := template.ParseFiles("./html/pendu.html")
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, currentGame); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}

func processGuess(guess string) {
	if currentGame == nil || currentGame.GameStatus != "ongoing" {
		return
	}

	if guess == currentGame.TargetWord {
		currentGame.WordShown = strings.Split(currentGame.TargetWord, "")
		currentGame.GameStatus = "won"
		return
	}

	currentGame.GuessedLetters = append(currentGame.GuessedLetters, guess)

	found := false
	for i, char := range currentGame.TargetWord {
		if string(char) == guess && currentGame.WordShown[i] == "_" {
			currentGame.WordShown[i] = guess
			found = true
		}
	}

	if !found {
		currentGame.RemainingAttempts--
	}

	if currentGame.RemainingAttempts <= 0 {
		currentGame.GameStatus = "lost"
	}

	if strings.Join(currentGame.WordShown, "") == currentGame.TargetWord {
		currentGame.GameStatus = "won"
	}
}

func server() {
	fileServer := http.FileServer(http.Dir("./html"))
	http.Handle("/", fileServer)

	http.HandleFunc("/pendu", penduHandler)

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	musique := http.FileServer(http.Dir("./musique"))
	http.Handle("/musique/", http.StripPrefix("/musique/", musique))

	fmt.Println("Server running at http://localhost:7080/")
	if err := http.ListenAndServe(":7080", nil); err != nil {
		log.Fatal(err)
	}
}

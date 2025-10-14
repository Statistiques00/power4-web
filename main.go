package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Gravity int

const (
	GravityDown Gravity = iota
	GravityUp
)

// Ajoute un champ Mode à Game pour retenir le mode de jeu
type Game struct {
	Board         [][]int
	Rows, Cols    int
	CurrentPlayer int
	Winner        int
	GameOver      bool
	LastRow       int
	LastCol       int
	TurnCount     int
	Gravity       Gravity
	Difficulty    string
	Username      string
	Mode          string // "normal" ou "inverse"
}

var (
	game  *Game
	mutex sync.Mutex
)

func NewGame(rows, cols, prefill int, difficulty, username, mode string) *Game {
	board := make([][]int, rows)
	for i := range board {
		board[i] = make([]int, cols)
	}
	// Prefill random cells
	rand.Seed(time.Now().UnixNano())
	for n := 0; n < prefill; {
		r := rand.Intn(rows)
		c := rand.Intn(cols)
		if board[r][c] == 0 {
			board[r][c] = rand.Intn(2) + 1
			n++
		}
	}
	gravity := GravityDown
	if mode == "inverse" {
		gravity = GravityUp
	} else {
		gravity = GravityDown
	}
	return &Game{
		Board:         board,
		Rows:          rows,
		Cols:          cols,
		CurrentPlayer: 1,
		Winner:        0,
		GameOver:      false,
		LastRow:       -1,
		LastCol:       -1,
		TurnCount:     0,
		Gravity:       gravity,
		Difficulty:    difficulty,
		Username:      username,
		Mode:          mode,
	}
}

// DropToken now supports gravity direction and increments turn count.
func (g *Game) DropToken(col int) bool {
	if col < 0 || col >= g.Cols || g.GameOver {
		return false
	}
	var row int
	if g.Gravity == GravityDown {
		for row = g.Rows - 1; row >= 0; row-- {
			if g.Board[row][col] == 0 {
				break
			}
		}
	} else {
		for row = 0; row < g.Rows; row++ {
			if g.Board[row][col] == 0 {
				break
			}
		}
	}
	if row < 0 || row >= g.Rows || g.Board[row][col] != 0 {
		return false
	}
	g.Board[row][col] = g.CurrentPlayer
	g.LastRow = row
	g.LastCol = col
	g.TurnCount++
	// Gravity reversal every 5 turns
	if g.TurnCount%5 == 0 {
		if g.Gravity == GravityDown {
			g.Gravity = GravityUp
		} else {
			g.Gravity = GravityDown
		}
	}
	if g.checkWin(row, col) {
		g.Winner = g.CurrentPlayer
		g.GameOver = true
	} else if g.isDraw() {
		g.GameOver = true
	}
	g.CurrentPlayer = 3 - g.CurrentPlayer
	return true
}

// checkWin vérifie si le dernier coup joué (row, col) crée un alignement de 4 jetons de même couleur.
func (g *Game) checkWin(row, col int) bool {
	player := g.Board[row][col]
	dirs := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}
	for _, d := range dirs {
		count := 1
		for i := 1; i < 4; i++ {
			r := row + d[0]*i
			c := col + d[1]*i
			if r >= 0 && r < g.Rows && c >= 0 && c < g.Cols && g.Board[r][c] == player {
				count++
			} else {
				break
			}
		}
		for i := 1; i < 4; i++ {
			r := row - d[0]*i
			c := col - d[1]*i
			if r >= 0 && r < g.Rows && c >= 0 && c < g.Cols && g.Board[r][c] == player {
				count++
			} else {
				break
			}
		}
		if count >= 4 {
			return true
		}
	}
	return false
}

// isDraw vérifie si le plateau est plein (aucune case vide en haut de chaque colonne).
func (g *Game) isDraw() bool {
	for c := 0; c < g.Cols; c++ {
		if g.Board[0][c] == 0 {
			return false
		}
	}
	return true
}

// getWinningPositions retourne les positions des 4 jetons gagnants si victoire, sinon nil.
func (g *Game) getWinningPositions() [][2]int {
	player := g.Winner
	if player == 0 {
		return nil
	}
	dirs := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}
	for r := 0; r < g.Rows; r++ {
		for c := 0; c < g.Cols; c++ {
			if g.Board[r][c] != player {
				continue
			}
			for _, d := range dirs {
				positions := [][2]int{{r, c}}
				for i := 1; i < 4; i++ {
					r2 := r + d[0]*i
					c2 := c + d[1]*i
					if r2 >= 0 && r2 < g.Rows && c2 >= 0 && c2 < g.Cols && g.Board[r2][c2] == player {
						positions = append(positions, [2]int{r2, c2})
					} else {
						break
					}
				}
				if len(positions) == 4 {
					return positions
				}
			}
		}
	}
	return nil
}

// renderBoard génère le HTML du plateau et des boutons de jeu selon l'état de la partie.
func renderBoard(g *Game) template.HTML {
	playerClass := "p1"
	if g.CurrentPlayer == 2 {
		playerClass = "p2"
	}
	gravityArrow := "↓"
	if g.Gravity == GravityUp {
		gravityArrow = "↑"
	}
	winning := map[[2]int]bool{}
	if g.GameOver && g.Winner != 0 {
		for _, pos := range g.getWinningPositions() {
			winning[pos] = true
		}
	}
	html := "<form method='POST' id='board-form'><input type='hidden' name='col' id='col-input'/>\n"
	html += "<div class='board-wrap " + playerClass
	if g.Gravity == GravityUp {
		html += " gravity-up"
	} else {
		html += " gravity-down"
	}
	html += "' id='board-wrap' style='overflow-x:auto; max-width:100vw;'>\n"
	html += "<table class='board' id='board' data-gameover='"
	if g.GameOver {
		html += "1'"
	} else {
		html += "0'"
	}
	html += " data-current='" + strconv.Itoa(g.CurrentPlayer) + "' style='margin:auto;'>\n"

	// Ligne de sélection alignée avec les colonnes, flèche directionnelle
	html += "<tr>"
	for c := 0; c < g.Cols; c++ {
		if !g.GameOver {
			html += "<td style='padding:0; border:none; background:none; text-align:center;'>"
			html += "<div class='selector-token' data-col='" + strconv.Itoa(c) + "' title='Jouer colonne " + strconv.Itoa(c+1) + "'>"
			html += "<span class='selector-arrow'>" + gravityArrow + "</span>"
			html += "</div></td>"
		} else {
			html += "<td style='padding:0; border:none; background:none; text-align:center;'><div class='selector-token disabled'><span class='selector-arrow'>" + gravityArrow + "</span></div></td>"
		}
	}
	html += "</tr>"

	// Plateau de jeu
	for r := 0; r < g.Rows; r++ {
		html += "<tr>"
		for c := 0; c < g.Cols; c++ {
			cell := ""
			tokenCls := ""
			wrapCls := ""
			if winning[[2]int{r, c}] {
				tokenCls = " winner-token"
			}
			if g.LastRow == r && g.LastCol == c {
				wrapCls = " just-played"
			}
			switch g.Board[r][c] {
			case 1:
				cell = "<div class='token-wrap" + wrapCls + "'><div class='token red" + tokenCls + "'></div></div>"
			case 2:
				cell = "<div class='token-wrap" + wrapCls + "'><div class='token yellow" + tokenCls + "'></div></div>"
			}
			html += "<td data-col='" + strconv.Itoa(c) + "'>" + cell + "</td>"
		}
		html += "</tr>"
	}
	html += "</table>\n"
	html += "<div id='selector' class='selector' style='display:none'></div>\n"
	html += "</div>" // end board-wrap
	html += "<div class='controls'><button name='reset' value='1'>Nouvelle partie</button>"
	if g.GameOver {
		html += "<button name='rematch' value='1'>Revanche</button>"
	}
	html += "</div></form>"

	// JS pour gérer le clic sur les "jetons" de sélection
	html += `<script>
	document.querySelectorAll('.selector-token:not(.disabled)').forEach(function(div) {
		div.addEventListener('click', function() {
			document.getElementById('col-input').value = div.getAttribute('data-col');
			document.getElementById('board-form').submit();
		});
	});
	</script>`

	return template.HTML(html)
}

// --- Template loading ---
var (
	pageTmpl      *template.Template
	startTmpl     *template.Template
	winTmpl       *template.Template
	loseTmpl      *template.Template
	modeTmpl      *template.Template
)

func loadTemplates() error {
	var err error
	pageTmpl, err = template.ParseFiles("templates/game.html")
	if err != nil {
		return err
	}
	startTmpl, err = template.ParseFiles("templates/start.html")
	if err != nil {
		return err
	}
	winTmpl, err = template.ParseFiles("templates/win.html")
	if err != nil {
		return err
	}
	loseTmpl, err = template.ParseFiles("templates/lose.html")
	if err != nil {
		return err
	}
	modeTmpl, err = template.ParseFiles("templates/mode.html")
	return err
}

// --- Nouveau handler pour choisir le mode ---
func modeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		mode := r.FormValue("mode")
		username := r.FormValue("username")
		difficulty := r.FormValue("difficulty")
		http.Redirect(w, r, "/connect4?username="+username+"&difficulty="+difficulty+"&mode="+mode, http.StatusSeeOther)
		return
	}
	// On récupère username et difficulty pour les garder dans le formulaire
	username := r.URL.Query().Get("username")
	difficulty := r.URL.Query().Get("difficulty")
	modeTmpl.Execute(w, map[string]interface{}{
		"Username":   username,
		"Difficulty": difficulty,
	})
}

// --- Modifie startHandler pour rediriger vers /mode ---
func startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		difficulty := r.FormValue("difficulty")
		http.Redirect(w, r, "/mode?username="+username+"&difficulty="+difficulty, http.StatusSeeOther)
		return
	}
	startTmpl.Execute(w, nil)
}

// --- Modifie handler pour prendre en compte le mode ---
func handler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	username := r.URL.Query().Get("username")
	difficulty := r.URL.Query().Get("difficulty")
	mode := r.URL.Query().Get("mode")
	if mode != "inverse" {
		mode = "normal"
	}

	rows, cols, prefill := 6, 7, 0
	switch difficulty {
	case "easy":
		rows, cols, prefill = 6, 7, 3
	case "normal":
		rows, cols, prefill = 7, 8, 5
	case "hard":
		rows, cols, prefill = 8, 10, 7
	}

	if game == nil || (username != "" && (game.Username != username || game.Difficulty != difficulty || game.Mode != mode)) {
		game = NewGame(rows, cols, prefill, difficulty, username, mode)
	}

	if r.Method == "POST" {
		r.ParseForm()
		if r.FormValue("reset") == "1" {
			game = nil
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if r.FormValue("rematch") == "1" {
			game = NewGame(rows, cols, prefill, difficulty, username, mode)
		} else if colStr := r.FormValue("col"); colStr != "" {
			col, err := strconv.Atoi(colStr)
			if err == nil {
				game.DropToken(col)
			}
		}
	}

	// Prépare le message de fin si besoin
	endMessage := ""
	if game.GameOver {
		if game.Winner == 1 {
			endMessage = "🎉 Victoire !"
		} else if game.Winner == 2 {
			endMessage = "💀 Défaite !"
		} else {
			endMessage = "Match nul !"
		}
	}

	data := struct {
		BoardHTML     template.HTML
		CurrentPlayer int
		Winner        int
		GameOver      bool
		Gravity       Gravity
		Username      string
		Difficulty    string
		Rows          int
		Cols          int
		Mode          string
		EndMessage    string
	}{
		BoardHTML:     renderBoard(game),
		CurrentPlayer: game.CurrentPlayer,
		Winner:        game.Winner,
		GameOver:      game.GameOver,
		Gravity:       game.Gravity,
		Username:      game.Username,
		Difficulty:    game.Difficulty,
		Rows:          game.Rows,
		Cols:          game.Cols,
		Mode:          game.Mode,
		EndMessage:    endMessage,
	}
	pageTmpl.Execute(w, data)
}

func main() {
	if err := loadTemplates(); err != nil {
		panic("Erreur chargement templates: " + err.Error())
	}
	http.HandleFunc("/", startHandler)
	http.HandleFunc("/mode", modeHandler)
	http.HandleFunc("/connect4", handler)
	http.Handle("/style.css", http.FileServer(http.Dir(".")))
	http.Handle("/favicon.svg", http.FileServer(http.Dir(".")))
	fmt.Println("Serveur Puissance 4 Go sur http://localhost:8080/")
	http.ListenAndServe(":8080", nil)
}
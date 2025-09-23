package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"
)

const (
	ROWS = 6
	COLS = 7
)

type Game struct {
	Board         [ROWS][COLS]int
	CurrentPlayer int
	Winner        int
	GameOver      bool
}

var (
	game  = NewGame()
	mutex sync.Mutex
)

// NewGame crée une nouvelle partie avec un plateau vide et le joueur 1 qui commence.
func NewGame() *Game {
	return &Game{
		CurrentPlayer: 1,
		Winner:        0,
		GameOver:      false,
	}
}

// DropToken place un jeton dans la colonne choisie pour le joueur courant.
// Retourne true si le coup est valide, false sinon (colonne pleine ou partie finie).
func (g *Game) DropToken(col int) bool {
	if col < 0 || col >= COLS || g.GameOver {
		return false
	}
	for row := ROWS - 1; row >= 0; row-- {
		if g.Board[row][col] == 0 {
			g.Board[row][col] = g.CurrentPlayer
			if g.checkWin(row, col) {
				g.Winner = g.CurrentPlayer
				g.GameOver = true
			} else if g.isDraw() {
				g.GameOver = true
			}
			g.CurrentPlayer = 3 - g.CurrentPlayer
			return true
		}
	}
	return false
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
			if r >= 0 && r < ROWS && c >= 0 && c < COLS && g.Board[r][c] == player {
				count++
			} else {
				break
			}
		}
		for i := 1; i < 4; i++ {
			r := row - d[0]*i
			c := col - d[1]*i
			if r >= 0 && r < ROWS && c >= 0 && c < COLS && g.Board[r][c] == player {
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
	for c := 0; c < COLS; c++ {
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
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			if g.Board[r][c] != player {
				continue
			}
			for _, d := range dirs {
				positions := [][2]int{{r, c}}
				for i := 1; i < 4; i++ {
					r2 := r + d[0]*i
					c2 := c + d[1]*i
					if r2 >= 0 && r2 < ROWS && c2 >= 0 && c2 < COLS && g.Board[r2][c2] == player {
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
	tokenClass := "col-btn-red"
	if g.CurrentPlayer == 2 {
		tokenClass = "col-btn-yellow"
	}
	winning := map[[2]int]bool{}
	if g.GameOver && g.Winner != 0 {
		for _, pos := range g.getWinningPositions() {
			winning[pos] = true
		}
	}
	html := "<form method='POST'><table class='board'>\n"
	// Plateau
	for r := 0; r < ROWS; r++ {
		html += "<tr>"
		for c := 0; c < COLS; c++ {
			cell := ""
			cls := ""
			if winning[[2]int{r, c}] {
				cls = " winner-token"
			}
			switch g.Board[r][c] {
			case 1:
				cell = "<div class='token red" + cls + "'></div>"
			case 2:
				cell = "<div class='token yellow" + cls + "'></div>"
			}
			html += "<td>" + cell + "</td>"
		}
		html += "</tr>"
	}
	// Ligne de boutons sous le plateau uniquement si la partie n'est pas finie
	if !g.GameOver {
		html += "<tr>"
		for c := 0; c < COLS; c++ {
			html += "<td><button class='col-btn " + tokenClass + "' name='col' value='" + strconv.Itoa(c) + "'>&#8593;</button></td>"
		}
		html += "</tr>"
	}

	html += "</table>"
	html += "<div class='controls'><button name='reset' value='1'>Nouvelle partie</button></div></form>"
	return template.HTML(html)
}

// loadTemplate charge le template HTML principal depuis le fichier template.html.
var pageTmpl *template.Template

func loadTemplate() error {
	var err error
	pageTmpl, err = template.ParseFiles("template.html")
	return err
}

// handler gère les requêtes HTTP, traite les coups, le reset et affiche la page du jeu.
func handler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	if r.Method == "POST" {
		err := r.ParseForm()
		if err == nil {
			if r.FormValue("reset") == "1" {
				game = NewGame()
			} else if colStr := r.FormValue("col"); colStr != "" {
				col, err := strconv.Atoi(colStr)
				if err == nil {
					game.DropToken(col)
				}
			}
		}
	}
	data := struct {
		BoardHTML     template.HTML
		CurrentPlayer int
		Winner        int
		GameOver      bool
	}{
		BoardHTML:     renderBoard(game),
		CurrentPlayer: game.CurrentPlayer,
		Winner:        game.Winner,
		GameOver:      game.GameOver,
	}
	pageTmpl.Execute(w, data)
}

// main démarre le serveur web, configure les routes et sert le CSS statique.
func main() {
	if err := loadTemplate(); err != nil {
		panic("Erreur chargement template: " + err.Error())
	}
	http.HandleFunc("/connect4", handler)
	http.Handle("/style.css", http.FileServer(http.Dir(".")))
	fmt.Println("Serveur Puissance 4 Go sur http://localhost:8080/connect4")
	http.ListenAndServe(":8080", nil)
}

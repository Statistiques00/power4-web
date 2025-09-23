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

func NewGame() *Game {
	return &Game{
		CurrentPlayer: 1,
		Winner:        0,
		GameOver:      false,
	}
}

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

func (g *Game) isDraw() bool {
	for c := 0; c < COLS; c++ {
		if g.Board[0][c] == 0 {
			return false
		}
	}
	return true
}

func renderBoard(g *Game) template.HTML {
	tokenClass := "col-btn-red"
	if g.CurrentPlayer == 2 {
		tokenClass = "col-btn-yellow"
	}
	html := "<form method='POST'><table class='board'>\n"
	// Plateau
	for r := 0; r < ROWS; r++ {
		html += "<tr>"
		for c := 0; c < COLS; c++ {
			cell := ""
			switch g.Board[r][c] {
			case 1:
				cell = "<div class='token red'></div>"
			case 2:
				cell = "<div class='token yellow'></div>"
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

var pageTmpl *template.Template

func loadTemplate() error {
	var err error
	pageTmpl, err = template.ParseFiles("template.html")
	return err
}

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

func main() {
	if err := loadTemplate(); err != nil {
		panic("Erreur chargement template: " + err.Error())
	}
	http.HandleFunc("/", handler)
	http.Handle("/style.css", http.FileServer(http.Dir(".")))
	fmt.Println("Serveur Puissance 4 Go sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

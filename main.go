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
	html := "<table class='board'>"
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
	html += "</table>"
	return template.HTML(html)
}

var pageTmpl = template.Must(template.New("page").Parse(`
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>Puissance 4 en Go</title>
    <style>
        body { background: #181c24; color: #fff; font-family: Arial, sans-serif; text-align: center; }
        .board { margin: 30px auto; border-spacing: 5px; }
        .board td { width: 60px; height: 60px; background: #222; border-radius: 50%; position: relative; }
        .token { width: 50px; height: 50px; border-radius: 50%; margin: 5px auto; }
        .red { background: linear-gradient(135deg, #ff4b2b, #ff416c); box-shadow: 0 0 10px #ff416c; }
        .yellow { background: linear-gradient(135deg, #ffe259, #ffa751); box-shadow: 0 0 10px #ffe259; }
        .controls { margin: 20px; }
        .controls button { padding: 10px 20px; font-size: 1.2em; margin: 0 5px; border-radius: 10px; border: none; background: #333; color: #fff; cursor: pointer; }
        .controls button:hover { background: #444; }
        .winner { font-size: 2em; color: #ffe259; margin: 20px; }
        .draw { font-size: 2em; color: #aaa; margin: 20px; }
    </style>
</head>
<body>
    <h1>Puissance 4 (Go Only)</h1>
    <form method="POST" class="controls">
        {{if not .GameOver}}
            <label>Joueur {{.CurrentPlayer}} ({{if eq .CurrentPlayer 1}}🔴{{else}}🟡{{end}}) : Choisissez une colonne</label><br>
            {{range $i, $v := .Cols}}
                <button name="col" value="{{$i}}">{{$i | add1}}</button>
            {{end}}
        {{end}}
        <button name="reset" value="1">Nouvelle partie</button>
    </form>
    <div>{{.BoardHTML}}</div>
    {{if .Winner}}
        <div class="winner">🎉 Joueur {{.Winner}} a gagné !</div>
    {{else if .GameOver}}
        <div class="draw">Match nul !</div>
    {{end}}
</body>
</html>
`))

func add1(i int) int { return i + 1 }

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
		Cols          []int
	}{
		BoardHTML:     renderBoard(game),
		CurrentPlayer: game.CurrentPlayer,
		Winner:        game.Winner,
		GameOver:      game.GameOver,
		Cols:          []int{0, 1, 2, 3, 4, 5, 6},
	}
	pageTmpl.Funcs(template.FuncMap{"add1": add1}).Execute(w, data)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Serveur Puissance 4 Go sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

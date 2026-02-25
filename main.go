package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rikdc/pong-go/game"
	"golang.org/x/term"
)

const tickRate = 50 * time.Millisecond

func render(g *game.Game) {
	// Move cursor to top-left without clearing (reduces flicker).
	fmt.Print("\033[H")

	width := g.Width
	height := g.Height

	// Build frame as a 2D grid of runes.
	buf := make([][]rune, height)
	for y := range buf {
		buf[y] = make([]rune, width)
		for x := range buf[y] {
			buf[y][x] = ' '
		}
	}

	// Borders.
	for x := 0; x < width; x++ {
		buf[0][x] = '-'
		buf[height-1][x] = '-'
	}
	// Centre dashes.
	for y := 1; y < height-1; y++ {
		if y%2 == 0 {
			buf[y][width/2] = '|'
		}
	}

	// Paddles.
	drawPaddle(buf, g.Player, height)
	drawPaddle(buf, g.AI, height)

	// Ball.
	bx := int(g.Ball.X + 0.5)
	by := int(g.Ball.Y + 0.5)
	if bx >= 0 && bx < width && by >= 0 && by < height {
		buf[by][bx] = 'O'
	}

	// Render each row.
	for _, row := range buf {
		fmt.Println(string(row))
	}

	// Score line.
	fmt.Printf("  Player: %d    AI: %d\n", g.Player.Score, g.AI.Score)
	fmt.Println("  W/S to move | Q to quit")
}

func drawPaddle(buf [][]rune, p game.Paddle, height int) {
	for i := 0; i < p.Height; i++ {
		y := p.Y + i
		if y >= 0 && y < height {
			buf[y][p.X] = '|'
		}
	}
}

func renderGameOver(g *game.Game) {
	fmt.Print("\033[H\033[2J")
	if g.State == game.StatePlayerWon {
		fmt.Println("*** YOU WIN! ***")
	} else {
		fmt.Println("*** AI WINS! ***")
	}
	fmt.Printf("Final score — Player: %d  AI: %d\n", g.Player.Score, g.AI.Score)
	fmt.Println("Press any key to exit.")
}

func main() {
	// Put terminal into raw mode so we can read keystrokes without Enter.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to set terminal to raw mode:", err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Hide cursor.
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	// Clear screen.
	fmt.Print("\033[2J")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	g := game.New(game.DefaultWidth, game.DefaultHeight)

	input := make(chan byte, 8)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				return
			}
			input <- buf[0]
		}
	}()

	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	render(g)

loop:
	for {
		select {
		case <-sig:
			break loop
		case ch := <-input:
			switch ch {
			case 'q', 'Q', 3: // 3 = Ctrl-C
				break loop
			case 'w', 'W':
				g.MovePlayer(game.DirUp)
			case 's', 'S':
				g.MovePlayer(game.DirDown)
			}
			if g.IsOver() {
				break loop
			}
		case <-ticker.C:
			g.Update()
			if g.IsOver() {
				renderGameOver(g)
				// Wait for a final keypress.
				buf := make([]byte, 1)
				os.Stdin.Read(buf) //nolint:errcheck
				break loop
			}
			render(g)
		}
	}
}

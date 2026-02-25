package game

import "math"

const (
	DefaultWidth      = 80
	DefaultHeight     = 24
	PaddleHeight      = 4
	PaddleWidth       = 1
	BallSize          = 1
	WinScore          = 7
	PaddleSpeed       = 1
	DefaultBallSpeedX = 1.0
	DefaultBallSpeedY = 0.5
)

type Direction int

const (
	DirNone Direction = iota
	DirUp
	DirDown
)

type Paddle struct {
	X, Y   int
	Height int
	Score  int
}

type Ball struct {
	X, Y   float64
	VX, VY float64
}

type GameState int

const (
	StateRunning GameState = iota
	StatePlayerWon
	StateAIWon
)

type Game struct {
	Width  int
	Height int
	Player Paddle
	AI     Paddle
	Ball   Ball
	State  GameState
}

// New creates a new Game with default dimensions.
func New(width, height int) *Game {
	g := &Game{
		Width:  width,
		Height: height,
	}
	g.reset()
	return g
}

func (g *Game) reset() {
	midY := g.Height / 2
	g.Player = Paddle{
		X:      1,
		Y:      midY - PaddleHeight/2,
		Height: PaddleHeight,
	}
	g.AI = Paddle{
		X:      g.Width - 2,
		Y:      midY - PaddleHeight/2,
		Height: PaddleHeight,
	}
	g.Ball = Ball{
		X:  float64(g.Width) / 2,
		Y:  float64(g.Height) / 2,
		VX: DefaultBallSpeedX,
		VY: DefaultBallSpeedY,
	}
	g.State = StateRunning
}

// MovePlayer moves the player paddle in the given direction.
func (g *Game) MovePlayer(dir Direction) {
	switch dir {
	case DirUp:
		if g.Player.Y > 1 {
			g.Player.Y -= PaddleSpeed
		}
	case DirDown:
		if g.Player.Y+g.Player.Height < g.Height-1 {
			g.Player.Y += PaddleSpeed
		}
	}
}

// Update advances the game by one tick.
func (g *Game) Update() {
	if g.State != StateRunning {
		return
	}

	g.moveAI()

	g.Ball.X += g.Ball.VX
	g.Ball.Y += g.Ball.VY

	// Bounce off top/bottom walls (leaving row 0 and row Height-1 as borders).
	if g.Ball.Y <= 1 {
		g.Ball.Y = 1
		g.Ball.VY = math.Abs(g.Ball.VY)
	}
	if g.Ball.Y >= float64(g.Height)-2 {
		g.Ball.Y = float64(g.Height) - 2
		g.Ball.VY = -math.Abs(g.Ball.VY)
	}

	g.handlePaddleCollision(&g.Player)
	g.handlePaddleCollision(&g.AI)

	// Ball exits left side — AI scores.
	if g.Ball.X < 0 {
		g.AI.Score++
		if g.AI.Score >= WinScore {
			g.State = StateAIWon
		} else {
			g.resetBall(-DefaultBallSpeedX)
		}
	}

	// Ball exits right side — Player scores.
	if g.Ball.X >= float64(g.Width) {
		g.Player.Score++
		if g.Player.Score >= WinScore {
			g.State = StatePlayerWon
		} else {
			g.resetBall(DefaultBallSpeedX)
		}
	}
}

func (g *Game) resetBall(vx float64) {
	g.Ball = Ball{
		X:  float64(g.Width) / 2,
		Y:  float64(g.Height) / 2,
		VX: vx,
		VY: DefaultBallSpeedY,
	}
}

func (g *Game) handlePaddleCollision(p *Paddle) {
	bx := int(math.Round(g.Ball.X))
	by := int(math.Round(g.Ball.Y))

	// Check whether the ball overlaps the paddle column.
	if bx != p.X {
		return
	}
	if by < p.Y || by >= p.Y+p.Height {
		return
	}

	// Reflect horizontally and add slight angle variation based on hit position.
	g.Ball.VX = -g.Ball.VX
	relPos := float64(by-p.Y) / float64(p.Height)
	g.Ball.VY = (relPos - 0.5) * 2.0
}

// moveAI performs simple AI tracking of the ball.
func (g *Game) moveAI() {
	paddleMid := g.AI.Y + g.AI.Height/2
	ballY := int(math.Round(g.Ball.Y))

	if paddleMid < ballY && g.AI.Y+g.AI.Height < g.Height-1 {
		g.AI.Y += PaddleSpeed
	} else if paddleMid > ballY && g.AI.Y > 1 {
		g.AI.Y -= PaddleSpeed
	}
}

// IsOver returns true when the game has ended.
func (g *Game) IsOver() bool {
	return g.State != StateRunning
}

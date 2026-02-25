package game

import (
	"math"
	"testing"
)

func newTestGame() *Game {
	return New(40, 20)
}

// --- Construction ---

func TestNew_InitialState(t *testing.T) {
	g := newTestGame()

	if g.State != StateRunning {
		t.Fatalf("expected StateRunning, got %v", g.State)
	}
	if g.Player.Score != 0 || g.AI.Score != 0 {
		t.Fatalf("scores should start at 0")
	}
	if g.Ball.X == 0 && g.Ball.Y == 0 {
		t.Fatal("ball should not start at origin")
	}
}

func TestNew_PaddlePositions(t *testing.T) {
	g := newTestGame()

	if g.Player.X != 1 {
		t.Errorf("player X = %d, want 1", g.Player.X)
	}
	if g.AI.X != g.Width-2 {
		t.Errorf("AI X = %d, want %d", g.AI.X, g.Width-2)
	}
}

// --- Player movement ---

func TestMovePlayer_Up(t *testing.T) {
	g := newTestGame()
	startY := g.Player.Y
	g.MovePlayer(DirUp)
	if g.Player.Y != startY-PaddleSpeed {
		t.Errorf("expected Y=%d, got %d", startY-PaddleSpeed, g.Player.Y)
	}
}

func TestMovePlayer_Down(t *testing.T) {
	g := newTestGame()
	startY := g.Player.Y
	g.MovePlayer(DirDown)
	if g.Player.Y != startY+PaddleSpeed {
		t.Errorf("expected Y=%d, got %d", startY+PaddleSpeed, g.Player.Y)
	}
}

func TestMovePlayer_ClampTop(t *testing.T) {
	g := newTestGame()
	g.Player.Y = 1 // already at the top boundary
	g.MovePlayer(DirUp)
	if g.Player.Y != 1 {
		t.Errorf("paddle should not move above row 1, got %d", g.Player.Y)
	}
}

func TestMovePlayer_ClampBottom(t *testing.T) {
	g := newTestGame()
	g.Player.Y = g.Height - 1 - g.Player.Height // bottom boundary
	g.MovePlayer(DirDown)
	if g.Player.Y+g.Player.Height > g.Height-1 {
		t.Errorf("paddle should not extend past row %d", g.Height-1)
	}
}

func TestMovePlayer_NoneDoesNothing(t *testing.T) {
	g := newTestGame()
	startY := g.Player.Y
	g.MovePlayer(DirNone)
	if g.Player.Y != startY {
		t.Errorf("DirNone should not change Y, got %d", g.Player.Y)
	}
}

// --- Ball wall bouncing ---

func TestUpdate_BallBouncesOffTop(t *testing.T) {
	g := newTestGame()
	g.Ball.X = float64(g.Width) / 2
	g.Ball.Y = 1
	g.Ball.VY = -0.5

	g.Update()

	if g.Ball.VY <= 0 {
		t.Errorf("ball VY should be positive after top bounce, got %f", g.Ball.VY)
	}
}

func TestUpdate_BallBouncesOffBottom(t *testing.T) {
	g := newTestGame()
	g.Ball.X = float64(g.Width) / 2
	g.Ball.Y = float64(g.Height) - 2
	g.Ball.VY = 0.5

	g.Update()

	if g.Ball.VY >= 0 {
		t.Errorf("ball VY should be negative after bottom bounce, got %f", g.Ball.VY)
	}
}

// --- Scoring ---

func TestUpdate_AIScoresWhenBallExitsLeft(t *testing.T) {
	g := newTestGame()
	// Place ball just off the left edge on the next tick.
	g.Ball.X = 0.4
	g.Ball.Y = float64(g.Height) / 2
	g.Ball.VX = -1.0

	g.Update()

	if g.AI.Score != 1 {
		t.Errorf("AI score = %d, want 1", g.AI.Score)
	}
	if g.Player.Score != 0 {
		t.Errorf("Player score should remain 0, got %d", g.Player.Score)
	}
}

func TestUpdate_PlayerScoresWhenBallExitsRight(t *testing.T) {
	g := newTestGame()
	g.Ball.X = float64(g.Width) - 0.4
	g.Ball.Y = float64(g.Height) / 2
	g.Ball.VX = 1.0

	g.Update()

	if g.Player.Score != 1 {
		t.Errorf("Player score = %d, want 1", g.Player.Score)
	}
	if g.AI.Score != 0 {
		t.Errorf("AI score should remain 0, got %d", g.AI.Score)
	}
}

// --- Win conditions ---

func TestUpdate_PlayerWins(t *testing.T) {
	g := newTestGame()
	g.Player.Score = WinScore - 1
	g.Ball.X = float64(g.Width) - 0.4
	g.Ball.Y = float64(g.Height) / 2
	g.Ball.VX = 1.0

	g.Update()

	if g.State != StatePlayerWon {
		t.Errorf("expected StatePlayerWon, got %v", g.State)
	}
}

func TestUpdate_AIWins(t *testing.T) {
	g := newTestGame()
	g.AI.Score = WinScore - 1
	g.Ball.X = 0.4
	g.Ball.Y = float64(g.Height) / 2
	g.Ball.VX = -1.0

	g.Update()

	if g.State != StateAIWon {
		t.Errorf("expected StateAIWon, got %v", g.State)
	}
}

func TestUpdate_NoProgressWhenGameOver(t *testing.T) {
	g := newTestGame()
	g.State = StatePlayerWon
	ballX := g.Ball.X

	g.Update()

	if g.Ball.X != ballX {
		t.Error("ball should not move after game over")
	}
}

func TestIsOver(t *testing.T) {
	g := newTestGame()
	if g.IsOver() {
		t.Error("game should not be over initially")
	}
	g.State = StateAIWon
	if !g.IsOver() {
		t.Error("game should be over when AI has won")
	}
}

// --- Paddle collision ---

func TestHandlePaddleCollision_ReflectsBall(t *testing.T) {
	g := newTestGame()
	// Position ball exactly on the player paddle.
	g.Ball.X = float64(g.Player.X)
	g.Ball.Y = float64(g.Player.Y + g.Player.Height/2)
	g.Ball.VX = -1.0
	origVX := g.Ball.VX

	g.handlePaddleCollision(&g.Player)

	if math.Signbit(g.Ball.VX) == math.Signbit(origVX) {
		t.Errorf("VX should reverse sign after paddle collision; orig=%f new=%f", origVX, g.Ball.VX)
	}
}

func TestHandlePaddleCollision_NoCollision(t *testing.T) {
	g := newTestGame()
	g.Ball.X = float64(g.Player.X + 5) // far from paddle
	g.Ball.VX = -1.0
	origVX := g.Ball.VX

	g.handlePaddleCollision(&g.Player)

	if g.Ball.VX != origVX {
		t.Error("VX should not change when ball is not on paddle column")
	}
}

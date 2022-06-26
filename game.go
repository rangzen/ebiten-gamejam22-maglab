package main

import (
	"github.com/jakecoffman/cp"
)

const (
	preparationDuration = 5    // Time for preparing, before game starts
	scoreStartingPoints = 6128 // Max score for 1 minute
	scoreMalusMagnet    = 42   // How many malus points for each magnet
)

type GameState int

const (
	GameInitialising GameState = iota + 1
	GameReady
	GamePreparing
	GameRunning
	GameEnded
)

var (
	gamestate            = GameInitialising
	gameLevel            int
	gameLevels           = []Level{LevelS{}}
	gamePreparingTimeout float64
	magnetCounter        = 0
	score                int
)

type Game struct {
	space *cp.Space
	ball  *cp.Body
	mags  []Mag
	bell  *cp.Body
	time  float64
}

func NewGame() *Game {
	game := &Game{}

	// Chipmunk Space
	game.space = cp.NewSpace()

	createWalls(game.space)

	gameLevels[gameLevel].StartPosition(game)

	game.mags = make([]Mag, 0, 10)

	game.space.NewCollisionHandler(CollisionBall, CollisionBell).PreSolveFunc = collisionBallBellCallback()

	gamestate = GameReady

	return game
}

type Mag struct {
	pos       cp.Vector
	timeToDie float64
}

// newMag creates a new mag at a specific location.
func (g *Game) newMag(x, y int) {
	g.mags = append(g.mags, Mag{
		pos:       cp.Vector{X: float64(x), Y: float64(y)},
		timeToDie: g.time + magTTL,
	})
}

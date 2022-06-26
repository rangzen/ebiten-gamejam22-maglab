package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp"
)

func createWalls(space *cp.Space) {
	topWall := cp.NewBox2(space.StaticBody, cp.BB{L: -10, B: 0, R: screenWidth + 10, T: -10}, 0)
	topWall.SetFriction(0)
	topWall.SetElasticity(.7)
	space.AddShape(topWall)
	bottomWall := cp.NewBox2(space.StaticBody, cp.BB{L: -10, B: screenHeight + 10, R: screenWidth + 10, T: screenHeight}, 0)
	bottomWall.SetFriction(0)
	bottomWall.SetElasticity(.7)
	space.AddShape(bottomWall)
	leftWall := cp.NewBox2(space.StaticBody, cp.BB{L: -10, B: screenHeight + 10, R: 0, T: -10}, 0)
	leftWall.SetFriction(0)
	leftWall.SetElasticity(.7)
	space.AddShape(leftWall)
	rightWall := cp.NewBox2(space.StaticBody, cp.BB{L: screenWidth, B: screenHeight + 10, R: screenWidth + 10, T: -10}, 0)
	rightWall.SetFriction(0)
	rightWall.SetElasticity(.7)
	space.AddShape(rightWall)
	topRightTwoThirdWall := cp.NewBox2(space.StaticBody, cp.BB{
		L: screenWidth / 3,
		B: screenHeight/3 + 2,
		R: screenWidth + 10,
		T: screenHeight/3 - 2,
	}, 0)
	topRightTwoThirdWall.SetFriction(0)
	topRightTwoThirdWall.SetElasticity(.7)
	space.AddShape(topRightTwoThirdWall)
	bottomLeftTwoThirdWall := cp.NewBox2(space.StaticBody, cp.BB{
		L: -10,
		B: screenHeight/3*2 + 2,
		R: screenWidth / 3 * 2,
		T: screenHeight/3*2 - 2,
	}, 0)
	bottomLeftTwoThirdWall.SetFriction(0)
	bottomLeftTwoThirdWall.SetElasticity(.7)
	space.AddShape(bottomLeftTwoThirdWall)
}

func drawWalls(screen *ebiten.Image) {
	// Top
	ebitenutil.DrawLine(screen, 0, 0, screenWidth, 0, color.White)
	// Bottom
	ebitenutil.DrawLine(screen, 0, screenHeight-1, screenWidth, screenHeight-1, color.White)
	// Left
	ebitenutil.DrawLine(screen, 1, 0, 1, screenHeight, color.White)
	// Right
	ebitenutil.DrawLine(screen, screenWidth, 0, screenWidth, screenHeight, color.White)
}

type Level interface {
	Name() string
	StartPosition(game *Game)
	Draw(screen *ebiten.Image)
}

type LevelS struct{}

func (l LevelS) Name() string {
	return "Snake"
}

func (l LevelS) StartPosition(game *Game) {
	var mass = 1.

	space := game.space

	// Ball
	ballMoment := cp.MomentForCircle(mass, 0, ballSize, cp.Vector{})
	ballBody := space.AddBody(cp.NewBody(mass, ballMoment))
	ballBody.SetPosition(cp.Vector{X: screenWidth / 6 * 5, Y: screenHeight / 6})
	ballShape := space.AddShape(cp.NewCircle(ballBody, ballSize, cp.Vector{}))
	ballShape.SetFriction(0.7)
	ballShape.SetElasticity(.7)
	ballShape.SetCollisionType(CollisionBall)
	game.ball = ballBody

	// Bell
	bellMoment := cp.MomentForCircle(mass, 0, bellSize, cp.Vector{})
	bellBody := space.AddBody(cp.NewBody(mass, bellMoment))
	bellBody.SetPosition(cp.Vector{X: screenWidth / 6, Y: screenHeight / 6 * 5})
	bellShape := space.AddShape(cp.NewCircle(bellBody, bellSize, cp.Vector{}))
	bellShape.SetFriction(0.7)
	bellShape.SetElasticity(.7)
	bellShape.SetCollisionType(CollisionBell)
	game.bell = bellBody
}

func (l LevelS) Draw(screen *ebiten.Image) {
	// Top right 2/3
	ebitenutil.DrawLine(screen, screenWidth/3, screenHeight/3, screenWidth, screenHeight/3, color.White)
	// Bottom left 2/3
	ebitenutil.DrawLine(screen, 0, screenHeight/3*2, screenWidth/3*2, screenHeight/3*2, color.White)
}

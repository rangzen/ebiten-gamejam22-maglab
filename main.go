package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jakecoffman/cp"
	"golang.org/x/image/colornames"

	"log"
)

const (
	// Window
	title        = "MagLab - Ebitengine Game Jam 22"
	screenWidth  = 800
	screenHeight = 600

	// GUI
	ballSize     = 5
	magSize      = 5
	drawVelocity = false

	// Physics
	magTTL    = 10    // Time to live of a mag in seconds
	maxLength = 300.  // Distance max between ball and mag to have effect
	maxForce  = 3000. // Max force to apply to ball
)

var (
	ballImage = ebiten.NewImage(ballSize, ballSize)
	magImage  = ebiten.NewImage(magSize, magSize)
	colorBlue = color.RGBA{B: 255, A: 1}
)

func init() {
	ballImage.Fill(color.White)
	magImage.Fill(colorBlue)
}

func main() {
	log.Println(title)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle(title)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

type Mag struct {
	pos       cp.Vector
	timeToDie float64
}

type Game struct {
	space *cp.Space
	ball  *cp.Body
	mags  []Mag
	time  float64
}

func NewGame() *Game {
	// Chipmunk Space
	space := cp.NewSpace()

	// S wall
	topWall := cp.NewBox2(space.StaticBody, cp.BB{L: -10, B: 0, R: screenWidth + 10, T: -10}, 0)
	topWall.SetFriction(1)
	topWall.SetElasticity(.7)
	space.AddShape(topWall)
	bottomWall := cp.NewBox2(space.StaticBody, cp.BB{L: -10, B: screenHeight + 10, R: screenWidth + 10, T: screenHeight}, 0)
	bottomWall.SetFriction(1)
	bottomWall.SetElasticity(.7)
	space.AddShape(bottomWall)
	leftWall := cp.NewBox2(space.StaticBody, cp.BB{L: -10, B: screenHeight + 10, R: 0, T: -10}, 0)
	leftWall.SetFriction(1)
	leftWall.SetElasticity(.7)
	space.AddShape(leftWall)
	rightWall := cp.NewBox2(space.StaticBody, cp.BB{L: screenWidth, B: screenHeight + 10, R: screenWidth + 10, T: -10}, 0)
	rightWall.SetFriction(1)
	rightWall.SetElasticity(.7)
	space.AddShape(rightWall)
	topRightTwoThirdWall := cp.NewBox2(space.StaticBody, cp.BB{
		L: screenWidth / 3,
		B: screenHeight/3 + 2,
		R: screenWidth + 10,
		T: screenHeight/3 - 2,
	}, 0)
	topRightTwoThirdWall.SetFriction(1)
	topRightTwoThirdWall.SetElasticity(.7)
	space.AddShape(topRightTwoThirdWall)
	bottomLeftTwoThirdWall := cp.NewBox2(space.StaticBody, cp.BB{
		L: -10,
		B: screenHeight/3*2 + 2,
		R: screenWidth / 3 * 2,
		T: screenHeight/3*2 - 2,
	}, 0)
	bottomLeftTwoThirdWall.SetFriction(1)
	bottomLeftTwoThirdWall.SetElasticity(.7)
	space.AddShape(bottomLeftTwoThirdWall)

	// Ball
	var radius float64 = ballSize
	var mass = 1.
	moment := cp.MomentForCircle(mass, 0, radius, cp.Vector{})
	ballBody := space.AddBody(cp.NewBody(mass, moment))
	ballBody.SetPosition(cp.Vector{X: screenWidth / 6 * 5, Y: screenHeight / 6})
	ballShape := space.AddShape(cp.NewCircle(ballBody, radius, cp.Vector{}))
	ballShape.SetFriction(0.7)
	ballShape.SetElasticity(.7)

	// Mags
	mags := make([]Mag, 0, 10)
	// mags = append(mags, Mag{pos: cp.Vector{X: screenWidth / 4, Y: screenHeight / 4}, enabled: true})
	// mags = append(mags, Mag{pos: cp.Vector{X: screenWidth / 4 * 3, Y: screenHeight / 4}, enabled: true})
	// mags = append(mags, Mag{pos: cp.Vector{X: screenWidth / 4 * 3, Y: screenHeight / 2}, enabled: true})

	return &Game{
		space: space,
		ball:  ballBody,
		mags:  mags,
	}
}

func (g *Game) Update() error {
	// Kill the first old mag
	for i, mag := range g.mags {
		if mag.timeToDie <= g.time {
			g.mags = append(g.mags[:i], g.mags[i+1:]...)
			break
		}
	}

	// New mag
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		g.mags = append(g.mags, Mag{
			pos:       cp.Vector{X: float64(mx), Y: float64(my)},
			timeToDie: g.time + magTTL,
		})
	}

	// Reset mags
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.mags = make([]Mag, 0, 10)
	}

	// Update force from mags
	for _, mag := range g.mags {
		v := cp.Vector{
			X: mag.pos.X - g.ball.Position().X,
			Y: mag.pos.Y - g.ball.Position().Y,
		}
		l := v.Length()
		if l > maxLength {
			continue
		}
		f := math.Pow(maxLength-l, 2)
		vnm := v.Normalize().Mult(f)
		g.ball.ApplyForceAtWorldPoint(vnm, cp.Vector{})
	}

	// Apply max force to ball
	if g.ball.Force().Length() > maxForce {
		g.ball.SetForce(g.ball.Force().Normalize().Mult(maxForce))
	}

	timeStep := 1.0 / float64(ebiten.MaxTPS())
	g.time += timeStep
	g.space.Step(timeStep)

	// Simulate friction
	g.ball.SetVelocityVector(g.ball.Velocity().Mult(.98))

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(colornames.Black)

	// S wall
	// Top
	ebitenutil.DrawLine(screen, 0, 0, screenWidth, 0, color.White)
	// Bottom
	ebitenutil.DrawLine(screen, 0, screenHeight-1, screenWidth, screenHeight-1, color.White)
	// Left
	ebitenutil.DrawLine(screen, 1, 0, 1, screenHeight, color.White)
	// Right
	ebitenutil.DrawLine(screen, screenWidth, 0, screenWidth, screenHeight, color.White)
	// Top right 2/3
	ebitenutil.DrawLine(screen, screenWidth/3, screenHeight/3, screenWidth, screenHeight/3, color.White)
	// Bottom left 2/3
	ebitenutil.DrawLine(screen, 0, screenHeight/3*2, screenWidth/3*2, screenHeight/3*2, color.White)

	// Ball
	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(200.0/255.0, 200.0/255.0, 200.0/255.0, 1)
	op.GeoM.Translate(g.ball.Position().X, g.ball.Position().Y)
	screen.DrawImage(ballImage, op)
	// Ball Velocity
	if drawVelocity {
		f := .1
		ebitenutil.DrawLine(screen,
			g.ball.Position().X, g.ball.Position().Y,
			g.ball.Position().X+g.ball.Velocity().X*f, g.ball.Position().Y+g.ball.Velocity().Y*f,
			color.White)
	}

	// Mags
	for _, m := range g.mags {
		op.GeoM.Reset()
		op.GeoM.Translate(m.pos.X, m.pos.Y)
		screen.DrawImage(magImage, op)
	}

	// Debug
	pos := g.ball.Position()
	vel := g.ball.Velocity()
	force := g.ball.Force()
	ebitenutil.DebugPrint(
		screen,
		fmt.Sprintf(
			"Time is %5.1f\nPosition (%4.1f, %4.1f)\nVelocity (%4.1f, %4.1f)\nForce    (%4.1f, %4.1f)",
			g.time,
			pos.X, pos.Y,
			vel.X, vel.Y,
			force.X, force.Y,
		))
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

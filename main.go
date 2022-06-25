package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/jakecoffman/cp"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"log"
)

const (
	// Window
	title        = "MagLab - Ebitengine Game Jam 22"
	screenFactor = 3
	screenWidth  = 108 * screenFactor
	screenHeight = 234 * screenFactor

	// GUI
	ballSize     = 5
	magSize      = 5
	bellSize     = 20
	scoreMargin  = 5
	drawVelocity = false

	// Physics
	magTTL    = 10    // Time to live of a mag in seconds
	maxLength = 300.  // Distance max between ball and mag to have effect
	maxForce  = 3000. // Max force to apply to ball

	// Game Settings
	preparationDuration = 5    // Time for preparing, before game starts
	scoreStartingPoints = 6128 // Max score for 1 minute
	scoreMalusMagnet    = 42   // How many malus points for each magnet
)

const (
	CollisionUnknown = iota
	CollisionBall
	CollisionBell
	collisionEnd
)

type GameState int

const (
	GameUnknown GameState = iota
	GameInitialising
	GameReady
	GamePreparing
	GameRunning
	GameEnded
	gameEnd
)

var gamestate = GameInitialising
var gamePreparingTimeout float64
var magnetCounter = 0
var score int

var (
	ballImage = ebiten.NewImage(ballSize, ballSize)
	magImage  = ebiten.NewImage(magSize, magSize)
	bellImage = ebiten.NewImage(bellSize, bellSize)
	colorBlue = color.RGBA{B: 255, A: 1}
	colorRed  = color.RGBA{R: 255, A: 1}

	titleArcadeFont font.Face
	arcadeFont      font.Face
	smallArcadeFont font.Face
	fontSize        = 24
	titleFontSize   = int(float64(fontSize) * 1.5)
	smallFontSize   = fontSize / 2
)

func init() {
	// Images
	ballImage.Fill(color.White)
	magImage.Fill(colorBlue)
	bellImage.Fill(colorRed)

	// Fonts
	tt, err := opentype.Parse(fonts.PressStart2P_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	titleArcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    float64(titleFontSize),
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	arcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    float64(fontSize),
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	smallArcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    float64(smallFontSize),
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
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
	bell  *cp.Body
	time  float64
}

func NewGame() *Game {
	// Chipmunk Space
	space := cp.NewSpace()

	createWalls(space)

	var mass = 1.

	// Ball
	ballMoment := cp.MomentForCircle(mass, 0, ballSize, cp.Vector{})
	ballBody := space.AddBody(cp.NewBody(mass, ballMoment))
	ballBody.SetPosition(cp.Vector{X: screenWidth / 6 * 5, Y: screenHeight / 6})
	ballShape := space.AddShape(cp.NewCircle(ballBody, ballSize, cp.Vector{}))
	ballShape.SetFriction(0.7)
	ballShape.SetElasticity(.7)
	ballShape.SetCollisionType(CollisionBall)

	// Mags
	mags := make([]Mag, 0, 10)

	// Bell
	bellMoment := cp.MomentForCircle(mass, 0, bellSize, cp.Vector{})
	bellBody := space.AddBody(cp.NewBody(mass, bellMoment))
	bellBody.SetPosition(cp.Vector{X: screenWidth / 6, Y: screenHeight / 6 * 5})
	bellShape := space.AddShape(cp.NewCircle(bellBody, bellSize, cp.Vector{}))
	bellShape.SetFriction(0.7)
	bellShape.SetElasticity(.7)
	bellShape.SetCollisionType(CollisionBell)

	game := &Game{
		space: space,
		ball:  ballBody,
		mags:  mags,
		bell:  bellBody,
	}

	space.NewCollisionHandler(CollisionBall, CollisionBell).PreSolveFunc = collisionBallBellCallback()

	gamestate = GameReady

	return game
}

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

func collisionBallBellCallback() cp.CollisionPreSolveFunc {
	return func(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
		if gamestate == GameRunning {
			gamestate = GameEnded
			return true
		}
		return arb.Ignore()
	}
}

func (g *Game) Update() error {
	// Score
	if gamestate == GameRunning {
		score = scoreStartingPoints - int((g.time-gamePreparingTimeout)*10) - magnetCounter*scoreMalusMagnet
	}

	// Kill the first old mag
	if gamestate >= GameRunning {
		for i, mag := range g.mags {
			if mag.timeToDie <= g.time {
				g.mags = append(g.mags[:i], g.mags[i+1:]...)
				break
			}
		}
	}

	// Starting the game after the preparation delay
	if gamestate == GamePreparing && g.time > gamePreparingTimeout {
		gamestate = GameRunning
	}

	// Start the preparation phase
	if gamestate == GameReady &&
		(inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
			len(inpututil.AppendJustPressedTouchIDs(nil)) > 0) {
		gamestate = GamePreparing
		gamePreparingTimeout = g.time + preparationDuration
	}

	// New mag
	if gamestate >= GamePreparing && gamestate < GameEnded && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		g.newMag(mx, my)
		magnetCounter++
	}

	// Reset mags
	if gamestate == GamePreparing && inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.mags = make([]Mag, 0, 10)
	}

	// Update force from mags
	if gamestate >= GameRunning {
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
	}

	timeStep := 1.0 / float64(ebiten.MaxTPS())
	g.time += timeStep
	g.space.Step(timeStep)

	// Simulate friction
	g.ball.SetVelocityVector(g.ball.Velocity().Mult(.98))

	return nil
}

// newMag creates a new mag at a specific location.
func (g *Game) newMag(x, y int) {
	g.mags = append(g.mags, Mag{
		pos:       cp.Vector{X: float64(x), Y: float64(y)},
		timeToDie: g.time + magTTL,
	})
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
	op.GeoM.Translate(g.ball.Position().X-ballSize/2, g.ball.Position().Y-ballSize/2)
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

	// Bell
	op.GeoM.Reset()
	op.GeoM.Translate(g.bell.Position().X-bellSize/2, g.bell.Position().Y-bellSize/2)
	screen.DrawImage(bellImage, op)

	// Texts
	// Ref: https://github.com/hajimehoshi/ebiten/blob/main/examples/flappy/main.go
	var titleTexts []string
	var texts []string
	switch gamestate {
	case GameReady:
		titleTexts = []string{"MagLab"}
		texts = []string{"", "", "", "", "", "", "", "", "",
			"PRESS SPACE,", "CLICK,", "OR TOUCH", "TO START", "PREPARATION", "PERIOD",
			fmt.Sprintf("(%d seconds)", preparationDuration)}
	case GamePreparing:
		t := gamePreparingTimeout - g.time
		if t > 0.1 {
			texts = []string{"", "", "", "", "", "", "", "", "", "", "", "", "", "", fmt.Sprintf("%.1f", t)}
		}
	case GameEnded:
		texts = []string{"", "GAME OVER!"}
	}
	for i, l := range titleTexts {
		x := (screenWidth - len(l)*titleFontSize) / 2
		text.Draw(screen, l, titleArcadeFont, x, (i+4)*titleFontSize, color.White)
	}
	for i, l := range texts {
		x := (screenWidth - len(l)*fontSize) / 2
		text.Draw(screen, l, arcadeFont, x, (i+4)*fontSize, color.White)
	}

	if gamestate <= GameReady {
		msg := []string{
			"MagLab by rangzen is",
			"licenced under CC BY 3.0.",
		}
		for i, l := range msg {
			x := (screenWidth - len(l)*smallFontSize) / 2
			text.Draw(screen, l, smallArcadeFont, x, screenHeight-4+(i-1)*smallFontSize, color.White)
		}
	}

	// Score
	if gamestate >= GameRunning {
		scoreStr := fmt.Sprintf("Score: %d", score)
		text.Draw(screen, scoreStr, arcadeFont, screenWidth-len(scoreStr)*fontSize-scoreMargin, screenHeight-fontSize+3*scoreMargin, color.White)
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

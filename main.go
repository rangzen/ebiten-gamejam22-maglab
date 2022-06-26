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
)

const (
	CollisionBall = iota + 1
	CollisionBell
)

var (
	// Images
	ballImage = ebiten.NewImage(ballSize, ballSize)
	magImage  = ebiten.NewImage(magSize, magSize)
	bellImage = ebiten.NewImage(bellSize, bellSize)

	// Colors
	colorBlue = color.RGBA{B: 255, A: 1}
	colorRed  = color.RGBA{R: 255, A: 1}

	// Fonts
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
	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if gamestate == GameReady &&
		(inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
			len(touchIDs) > 0) {
		gamestate = GamePreparing
		gamePreparingTimeout = g.time + preparationDuration
	}

	// New mag from mouse
	if gamestate >= GamePreparing && gamestate < GameEnded && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		g.newMag(mx, my)
		magnetCounter++
	}

	// New mag from touch
	if gamestate >= GamePreparing && gamestate < GameEnded && len(touchIDs) > 0 {
		for _, t := range touchIDs {
			mx, my := ebiten.TouchPosition(t)
			g.newMag(mx, my)
			magnetCounter++
		}
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

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(colornames.Black)

	drawWalls(screen)
	gameLevels[gameLevel].Draw(screen)

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
		texts = []string{"", "", "", "", "", "", "", "",
			"PRESS SPACE,", "CLICK,", "OR TOUCH", "TO START", "PREPARATION", "PERIOD",
			fmt.Sprintf("(%d seconds)", preparationDuration)}
	case GamePreparing:
		t := gamePreparingTimeout - g.time
		if t > 0.1 {
			texts = []string{"", "", "", "", "", "", "", "",
				"ADD SOME", "MAGS", "TO BRING", "THE BALL", "TO THE BELL",
				"", fmt.Sprintf("%.1f", t)}
		}
	case GameEnded:
		texts = []string{"", "", "", "", "", "", "", "", "", "", "", "GAME OVER!"}
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

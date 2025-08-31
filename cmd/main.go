package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth  = 1200
	screenHeight = 1000

	stepIdle = iota
	stepRequesting
	stepResponding
	stepJoining
	stepFinished

	textScale = 24.0 / 13.0
)

type User struct {
	UserID int
	Name   string
}

type Order struct {
	OrderID int
	UserID  int
	Item    string
}

type JoinedData struct {
	User  User
	Order Order
}

type Game struct {
	Users  []User
	Orders []Order
	Joined []JoinedData

	animationStep                int
	currentUserIndex             int
	animationTimer               *time.Ticker
	packetX, packetY             float32
	packetTargetX, packetTargetY float32
	packetSpeed                  float32
	showJoined                   bool
}

func NewGame() *Game {
	g := &Game{
		Users: []User{
			{UserID: 1, Name: "Alice"},
			{UserID: 2, Name: "Bob"},
			{UserID: 3, Name: "Charlie"},
			{UserID: 4, Name: "David"},
			{UserID: 5, Name: "Eve"},
			{UserID: 6, Name: "Frank"},
			{UserID: 7, Name: "Grace"},
			{UserID: 8, Name: "Heidi"},
			{UserID: 9, Name: "Ivan"},
			{UserID: 10, Name: "Judy"},
		},
		Orders: []Order{
			{OrderID: 101, UserID: 2, Item: "Book"},
			{OrderID: 102, UserID: 1, Item: "Pen"},
			{OrderID: 103, UserID: 3, Item: "Note"},
			{OrderID: 104, UserID: 4, Item: "Laptop"},
			{OrderID: 105, UserID: 5, Item: "Mouse"},
			{OrderID: 106, UserID: 6, Item: "Keyboard"},
			{OrderID: 107, UserID: 7, Item: "Monitor"},
			{OrderID: 108, UserID: 8, Item: "Webcam"},
			{OrderID: 109, UserID: 9, Item: "HDMI Cable"},
			{OrderID: 110, UserID: 10, Item: "USB Hub"},
		},
		animationStep:    stepIdle,
		currentUserIndex: -1,
		packetSpeed:      5,
	}
	return g
}

func (g *Game) startAnimation() {
	g.animationStep = stepRequesting
	g.currentUserIndex = 0
	g.Joined = []JoinedData{}
	g.animationTimer = time.NewTicker(1 * time.Second)
	g.setPacketStartPosition()
}

func (g *Game) setPacketStartPosition() {
	userY := 110 + g.currentUserIndex*30 + 12
	g.packetX = 60
	g.packetY = float32(userY)
	g.packetTargetX = 510
	g.packetTargetY = float32(userY)
}

func (g *Game) Update() error {
	if g.animationStep == stepIdle && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.startAnimation()
	}

	if g.animationStep == stepRequesting {
		g.packetX += g.packetSpeed
		if g.packetX >= g.packetTargetX {
			g.animationStep = stepResponding
			// Find matching order to set response packet path
			for i, o := range g.Orders {
				if o.UserID == g.Users[g.currentUserIndex].UserID {
					g.packetY = float32(110 + i*30 + 12)
					break
				}
			}
			g.packetTargetX = 60
		}
	} else if g.animationStep == stepResponding {
		g.packetX -= g.packetSpeed
		if g.packetX <= g.packetTargetX {
			g.animationStep = stepJoining
			g.showJoined = true
			// Perform the join
			for _, o := range g.Orders {
				if o.UserID == g.Users[g.currentUserIndex].UserID {
					g.Joined = append(g.Joined, JoinedData{User: g.Users[g.currentUserIndex], Order: o})
					break
				}
			}
		}
	} else if g.animationStep == stepJoining {
		select {
		case <-g.animationTimer.C:
			g.currentUserIndex++
			if g.currentUserIndex >= len(g.Users) {
				g.animationStep = stepFinished
				g.animationTimer.Stop()
			} else {
				g.animationStep = stepRequesting
				g.setPacketStartPosition()
			}
		default:
		}
	} else if g.animationStep == stepFinished {
		g.startAnimation()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawTables(screen)
	g.drawJoinedTable(screen)

	if g.animationStep == stepRequesting || g.animationStep == stepResponding {
		vector.DrawFilledCircle(screen, g.packetX, g.packetY, 5, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
	}

	if g.animationStep == stepIdle {
		g.drawScaledText(screen, "Press Space to Start Animation", 393, screenHeight-40, color.White)
	}
}

func (g *Game) drawScaledText(screen *ebiten.Image, str string, x, y int, clr color.Color) {
	bounds := text.BoundString(basicfont.Face7x13, str)
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return
	}
	offscreen := ebiten.NewImage(w, h)
	text.Draw(offscreen, str, basicfont.Face7x13, -bounds.Min.X, -bounds.Min.Y, clr)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(textScale, textScale)
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(offscreen, op)
}

func (g *Game) drawTables(screen *ebiten.Image) {
	// Draw User Table
	vector.DrawFilledRect(screen, 50, 50, 400, 450, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
	g.drawScaledText(screen, "Machine 1: User Table", 60, 60, color.White)
	for i, u := range g.Users {
		var c color.Color = color.White
		if g.animationStep > stepIdle && g.currentUserIndex == i {
			c = color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff} // Yellow
		}
		g.drawScaledText(screen, fmt.Sprintf("UserID: %d, Name: %s", u.UserID, u.Name), 60, 110+i*30, c)
	}

	// Draw Order Table
	vector.DrawFilledRect(screen, 500, 50, 650, 450, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
	g.drawScaledText(screen, "Machine 2: Order Table", 510, 60, color.White)
	for i, o := range g.Orders {
		var c color.Color = color.White
		if g.currentUserIndex >= 0 && g.currentUserIndex < len(g.Users) && (g.animationStep == stepResponding || g.animationStep == stepJoining) && g.Users[g.currentUserIndex].UserID == o.UserID {
			c = color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff} // Yellow
		}
		g.drawScaledText(screen, fmt.Sprintf("OrderID: %d, UserID: %d, Item: %s", o.OrderID, o.UserID, o.Item), 510, 110+i*30, c)
	}
}

func (g *Game) drawJoinedTable(screen *ebiten.Image) {
	if !g.showJoined {
		return
	}
	vector.DrawFilledRect(screen, 50, 520, 1100, 450, color.RGBA{R: 0x30, G: 0x30, B: 0x60, A: 0xff}, false)
	g.drawScaledText(screen, "JOIN Result", 60, 530, color.White)
	for i, j := range g.Joined {
		g.drawScaledText(screen, fmt.Sprintf("UserID: %d, Name: %s, OrderID: %d, Item: %s", j.User.UserID, j.User.Name, j.Order.OrderID, j.Order.Item), 60, 580+i*30, color.White)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Spanner Distributed JOIN Animation")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

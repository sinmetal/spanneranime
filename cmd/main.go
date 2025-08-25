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
	screenWidth  = 900
	screenHeight = 600

	stepIdle = iota
	stepRequesting
	stepResponding
	stepJoining
	stepFinished
)

type User struct {
	ID   int
	Name string
}

type Order struct {
	ID     int
	UserID int
	Item   string
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
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
			{ID: 3, Name: "Charlie"},
		},
		Orders: []Order{
			{ID: 101, UserID: 2, Item: "Book"},
			{ID: 102, UserID: 1, Item: "Pen"},
			{ID: 103, UserID: 3, Item: "Note"},
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
	userY := 100 + g.currentUserIndex*20
	g.packetX = 200
	g.packetY = float32(userY)
	g.packetTargetX = 460
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
				if o.UserID == g.Users[g.currentUserIndex].ID {
					g.packetY = float32(100 + i*20)
					break
				}
			}
			g.packetTargetX = 200
		}
	} else if g.animationStep == stepResponding {
		g.packetX -= g.packetSpeed
		if g.packetX <= g.packetTargetX {
			g.animationStep = stepJoining
			g.showJoined = true
			// Perform the join
			for _, o := range g.Orders {
				if o.UserID == g.Users[g.currentUserIndex].ID {
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
		text.Draw(screen, "Press Space to Start Animation", basicfont.Face7x13, screenWidth/2-120, screenHeight-30, color.White)
	}
}

func (g *Game) drawTables(screen *ebiten.Image) {
	// Draw User Table
	vector.DrawFilledRect(screen, 50, 50, 300, 400, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
	text.Draw(screen, "Machine 1: User Table", basicfont.Face7x13, 60, 70, color.White)
	for i, u := range g.Users {
		var c color.Color = color.White
		if g.animationStep > stepIdle && g.currentUserIndex == i {
			c = color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff} // Yellow
		}
		text.Draw(screen, fmt.Sprintf("ID: %d, Name: %s", u.ID, u.Name), basicfont.Face7x13, 60, 100+i*20, c)
	}

	// Draw Order Table
	vector.DrawFilledRect(screen, 450, 50, 300, 400, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
	text.Draw(screen, "Machine 2: Order Table", basicfont.Face7x13, 460, 70, color.White)
	for i, o := range g.Orders {
		var c color.Color = color.White
		if g.currentUserIndex >= 0 && g.currentUserIndex < len(g.Users) && (g.animationStep == stepResponding || g.animationStep == stepJoining) && g.Users[g.currentUserIndex].ID == o.UserID {
			c = color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff} // Yellow
		}
		text.Draw(screen, fmt.Sprintf("ID: %d, UserID: %d, Item: %s", o.ID, o.UserID, o.Item), basicfont.Face7x13, 460, 100+i*20, c)
	}
}

func (g *Game) drawJoinedTable(screen *ebiten.Image) {
	if !g.showJoined {
		return
	}
	vector.DrawFilledRect(screen, 50, 460, 700, 130, color.RGBA{R: 0x30, G: 0x30, B: 0x60, A: 0xff}, false)
	text.Draw(screen, "JOIN Result", basicfont.Face7x13, 60, 480, color.White)
	for i, j := range g.Joined {
		text.Draw(screen, fmt.Sprintf("UserID: %d, Name: %s, Item: %s", j.User.ID, j.User.Name, j.Order.Item), basicfont.Face7x13, 60, 510+i*20, color.White)
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

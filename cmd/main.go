package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth  = 1600
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
	Users         []User
	Orders        []Order
	UserMachines  [2][]User
	OrderMachines [2][]Order
	Joined        []JoinedData

	animationStep                int
	currentUserIndex             int
	currentUserMachineIndex      int
	currentOrderMachineIndex     int
	currentOrderIndex            int
	animationTimer               *time.Ticker
	packetX, packetY             float32
	packetStartX, packetStartY   float32
	packetTargetX, packetTargetY float32
	packetSpeedX, packetSpeedY   float32
	packetSpeed                  float32
	showJoined                   bool
	AnimationType                string
}

func NewGame(animationType string) *Game {
	if animationType == "JOIN2" {
		return NewGameJOIN2(animationType)
	}
	return NewGameJOIN1(animationType)
}

func NewGameJOIN1(animationType string) *Game {
	users := []User{
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
	}

	orders := []Order{
		{OrderID: 101, Item: "Book"},
		{OrderID: 102, Item: "Pen"},
		{OrderID: 103, Item: "Note"},
		{OrderID: 104, Item: "Laptop"},
		{OrderID: 105, Item: "Mouse"},
		{OrderID: 106, Item: "Keyboard"},
		{OrderID: 107, Item: "Monitor"},
		{OrderID: 108, Item: "Webcam"},
		{OrderID: 109, Item: "HDMI Cable"},
		{OrderID: 110, Item: "USB Hub"},
	}

	userIDs := make([]int, len(users))
	for i := range users {
		userIDs[i] = users[i].UserID
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(userIDs), func(i, j int) {
		userIDs[i], userIDs[j] = userIDs[j], userIDs[i]
	})

	for i := range orders {
		orders[i].UserID = userIDs[i]
	}

	g := &Game{
		Users:            users,
		Orders:           orders,
		animationStep:    stepIdle,
		currentUserIndex: -1,
		packetSpeed:      5,
		AnimationType:    animationType,
	}
	return g
}

func NewGameJOIN2(animationType string) *Game {
	userMachines := [2][]User{}
	userMachines[0] = []User{
		{UserID: 1, Name: "Alice"},
		{UserID: 2, Name: "Bob"},
		{UserID: 3, Name: "Charlie"},
		{UserID: 4, Name: "David"},
		{UserID: 5, Name: "Eve"},
	}
	userMachines[1] = []User{
		{UserID: 6, Name: "Frank"},
		{UserID: 7, Name: "Grace"},
		{UserID: 8, Name: "Heidi"},
		{UserID: 9, Name: "Ivan"},
		{UserID: 10, Name: "Judy"},
	}

	orderMachines := [2][]Order{}
	orderMachines[0] = []Order{
		{OrderID: 101, Item: "Book"},
		{OrderID: 102, Item: "Pen"},
		{OrderID: 103, Item: "Note"},
		{OrderID: 104, Item: "Laptop"},
		{OrderID: 105, Item: "Mouse"},
	}
	orderMachines[1] = []Order{
		{OrderID: 106, Item: "Keyboard"},
		{OrderID: 107, Item: "Monitor"},
		{OrderID: 108, Item: "Webcam"},
		{OrderID: 109, Item: "HDMI Cable"},
		{OrderID: 110, Item: "USB Hub"},
	}

	userIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(userIDs), func(i, j int) {
		userIDs[i], userIDs[j] = userIDs[j], userIDs[i]
	})

	for i := 0; i < 5; i++ {
		orderMachines[0][i].UserID = userIDs[i]
		orderMachines[1][i].UserID = userIDs[i+5]
	}

	g := &Game{
		UserMachines:     userMachines,
		OrderMachines:    orderMachines,
		animationStep:    stepIdle,
		currentUserIndex: -1,
		packetSpeed:      10,
		AnimationType:    animationType,
	}
	return g
}

func (g *Game) startAnimation() {
	g.animationStep = stepRequesting
	g.currentUserIndex = 0
	g.currentUserMachineIndex = 0
	g.Joined = []JoinedData{}
	g.animationTimer = time.NewTicker(500 * time.Millisecond)
	g.setPacketStartPosition()
}

func (g *Game) setPacketStartPosition() {
	if g.AnimationType == "JOIN2" {
		g.setPacketStartPositionJOIN2()
		return
	}
	userY := 110 + g.currentUserIndex*30 + 12
	g.packetX = 60
	g.packetY = float32(userY)
	g.packetTargetX = 510
	g.packetTargetY = float32(userY)
}

func (g *Game) Update() error {
	if g.AnimationType == "JOIN2" {
		return g.updateJOIN2()
	}
	return g.updateJOIN1()
}

func (g *Game) updateJOIN1() error {
	if g.animationStep == stepIdle {
		if g.AnimationType == "JOIN1" {
			g.startAnimation()
		} else if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.startAnimation()
		}
	}

	if g.animationStep == stepRequesting {
		g.packetX += g.packetSpeed
		if g.packetX >= g.packetTargetX {
			g.animationStep = stepResponding
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
	if g.AnimationType == "JOIN2" {
		g.drawJOIN2(screen)
	} else {
		g.drawJOIN1(screen)
	}
}

func (g *Game) drawJOIN1(screen *ebiten.Image) {
	g.drawTables(screen)
	g.drawJoinedTable(screen)

	if g.animationStep == stepRequesting || g.animationStep == stepResponding {
		vector.DrawFilledCircle(screen, g.packetX, g.packetY, 5, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
	}

	if g.animationStep == stepIdle && g.AnimationType != "JOIN1" {
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

// JOIN2 specific functions

func (g *Game) setPacketStartPositionJOIN2() {
	userTableYOffset := float32(50)
	if g.currentUserMachineIndex == 1 {
		userTableYOffset = 350
	}
	userY := userTableYOffset + 60 + float32(g.currentUserIndex*30) + 12

	g.packetX = 450 // Right edge of user tables
	g.packetY = userY
	g.packetStartX = g.packetX
	g.packetStartY = g.packetY
}

func (g *Game) updateJOIN2() error {
	if g.animationStep == stepIdle {
		if g.AnimationType == "JOIN2" {
			g.startAnimation()
		} else if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.startAnimation()
		}
	}

	if g.animationStep == stepRequesting {
		// Find a match
		foundMatch := false
		currentUser := g.UserMachines[g.currentUserMachineIndex][g.currentUserIndex]
		for orderMachineIndex, orderMachine := range g.OrderMachines {
			for orderIndex, order := range orderMachine {
				if currentUser.UserID == order.UserID {
					g.currentOrderMachineIndex = orderMachineIndex
					g.currentOrderIndex = orderIndex
					orderTableYOffset := float32(50)
					if orderMachineIndex == 1 {
						orderTableYOffset = 350
					}
					g.packetTargetX = 750 // Left edge of order tables
					g.packetTargetY = orderTableYOffset + 60 + float32(orderIndex*30) + 12

					deltaX := g.packetTargetX - g.packetStartX
					deltaY := g.packetTargetY - g.packetStartY
					distance := float32(math.Sqrt(float64(deltaX*deltaX + deltaY*deltaY)))
					g.packetSpeedX = g.packetSpeed * deltaX / distance
					g.packetSpeedY = g.packetSpeed * deltaY / distance

					foundMatch = true
					break
				}
			}
			if foundMatch {
				break
			}
		}
		g.animationStep = stepResponding

	} else if g.animationStep == stepResponding {
		g.packetX += g.packetSpeedX
		g.packetY += g.packetSpeedY

		dx1 := g.packetX - g.packetStartX
		dy1 := g.packetY - g.packetStartY
		dx2 := g.packetTargetX - g.packetStartX
		dy2 := g.packetTargetY - g.packetStartY

		if dx1*dx1+dy1*dy1 >= dx2*dx2+dy2*dy2 {
			g.animationStep = stepJoining
			g.showJoined = true
			currentUser := g.UserMachines[g.currentUserMachineIndex][g.currentUserIndex]
			for _, orderMachine := range g.OrderMachines {
				for _, order := range orderMachine {
					if currentUser.UserID == order.UserID {
						g.Joined = append(g.Joined, JoinedData{User: currentUser, Order: order})
						break
					}
				}
			}
		}
	} else if g.animationStep == stepJoining {
		select {
		case <-g.animationTimer.C:
			g.currentUserIndex++
			if g.currentUserIndex >= len(g.UserMachines[g.currentUserMachineIndex]) {
				g.currentUserIndex = 0
				g.currentUserMachineIndex++
				if g.currentUserMachineIndex >= len(g.UserMachines) {
					g.animationStep = stepFinished
					g.animationTimer.Stop()
				} else {
					g.animationStep = stepRequesting
					g.setPacketStartPosition()
				}
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

func (g *Game) drawJOIN2(screen *ebiten.Image) {
	g.drawTablesJOIN2(screen)
	g.drawJoinedTableJOIN2(screen)

	if g.animationStep == stepResponding {
		vector.DrawFilledCircle(screen, g.packetX, g.packetY, 5, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
	}

	if g.animationStep == stepIdle && g.AnimationType != "JOIN2" {
		g.drawScaledText(screen, "Press Space to Start Animation", 393, screenHeight-40, color.White)
	}
}

func (g *Game) drawTablesJOIN2(screen *ebiten.Image) {
	// User Tables
	vector.DrawFilledRect(screen, 50, 50, 400, 250, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
	g.drawScaledText(screen, "User Table 1", 60, 60, color.White)
	vector.DrawFilledRect(screen, 50, 350, 400, 250, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
	g.drawScaledText(screen, "User Table 2", 60, 360, color.White)

	for machineIndex, machine := range g.UserMachines {
		tableYOffset := 50
		if machineIndex == 1 {
			tableYOffset = 350
		}
		for i, u := range machine {
			var c color.Color = color.White
			if g.animationStep > stepIdle && g.currentUserMachineIndex == machineIndex && g.currentUserIndex == i {
				c = color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff} // Yellow
			}
			g.drawScaledText(screen, fmt.Sprintf("UserID: %d, Name: %s", u.UserID, u.Name), 60, tableYOffset+60+i*30, c)
		}
	}

	// Order Tables
	vector.DrawFilledRect(screen, 750, 50, 550, 250, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
	g.drawScaledText(screen, "Order Table 1", 760, 60, color.White)
	vector.DrawFilledRect(screen, 750, 350, 550, 250, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
	g.drawScaledText(screen, "Order Table 2", 760, 360, color.White)

	for machineIndex, machine := range g.OrderMachines {
		tableYOffset := 50
		if machineIndex == 1 {
			tableYOffset = 350
		}
		for i, o := range machine {
			var c color.Color = color.White
			if g.animationStep > stepRequesting && g.currentOrderMachineIndex == machineIndex && g.currentOrderIndex == i {
				c = color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff} // Yellow
			}
			g.drawScaledText(screen, fmt.Sprintf("OrderID: %d, UserID: %d, Item: %s", o.OrderID, o.UserID, o.Item), 760, tableYOffset+60+i*30, c)
		}
	}
}

func (g *Game) drawJoinedTableJOIN2(screen *ebiten.Image) {
	if !g.showJoined {
		return
	}
	vector.DrawFilledRect(screen, 50, 650, 1500, 300, color.RGBA{R: 0x30, G: 0x30, B: 0x60, A: 0xff}, false)
	g.drawScaledText(screen, "JOIN Result", 60, 660, color.White)
	for i, j := range g.Joined {
		row := i / 2
		col := i % 2
		x := 60 + col*770
		y := 710 + row*30
		g.drawScaledText(screen, fmt.Sprintf("UserID: %d, Name: %s, OrderID: %d, Item: %s", j.User.UserID, j.User.Name, j.Order.OrderID, j.Order.Item), x, y, color.White)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Spanner Distributed JOIN Animation")
	animationType := ""
	if len(os.Args) > 1 {
		animationType = os.Args[1]
	}
	if err := ebiten.RunGame(NewGame(animationType)); err != nil {
		log.Fatal(err)
	}
}

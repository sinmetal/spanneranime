package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
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

	// JOIN3 specific steps
	stepUserToIndexRequest
	stepUserToIndexResponse
	stepIndexToOrderRequest
	stepIndexToOrderResponse

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

type IndexEntry struct {
	UserID  int
	OrderID int
}

type JoinedData struct {
	User  User
	Order Order
}

type Game struct {
	// Common
	AnimationType  string
	animationStep  int
	showJoined     bool
	animationTimer *time.Ticker
	Joined         []JoinedData
	packetSpeed    float32

	// Data stores
	Users         []User
	Orders        []Order
	UserMachines  [2][]User
	OrderMachines [2][]Order
	IndexMachines [2][]IndexEntry

	// State
	currentUserIndex         int
	currentOrderMachineIndex [2]int
	currentOrderIndex        [2]int
	currentIndexMachineIndex [2]int
	currentIndexIndex        [2]int

	// Packets (up to 2)
	packetX, packetY             [2]float32
	packetStartX, packetStartY   [2]float32
	packetTargetX, packetTargetY [2]float32
	packetSpeedX, packetSpeedY   [2]float32
}

// --- Game Setup ---

func NewGame(animationType string) *Game {
	switch animationType {
	case "JOIN2":
		return NewGameJOIN2(animationType)
	case "JOIN3":
		return NewGameJOIN3(animationType)
	default:
		return NewGameJOIN1(animationType)
	}
}

func NewGameJOIN1(animationType string) *Game {
	users := []User{
		{UserID: 1, Name: "Alice"}, {UserID: 2, Name: "Bob"}, {UserID: 3, Name: "Charlie"}, {UserID: 4, Name: "David"}, {UserID: 5, Name: "Eve"},
		{UserID: 6, Name: "Frank"}, {UserID: 7, Name: "Grace"}, {UserID: 8, Name: "Heidi"}, {UserID: 9, Name: "Ivan"}, {UserID: 10, Name: "Judy"},
	}
	orders := make([]Order, 10)
	userIDs := make([]int, len(users))
	for i, u := range users {
		orders[i] = Order{OrderID: 101 + i, Item: fmt.Sprintf("Item%d", 101+i)}
		userIDs[i] = u.UserID
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(userIDs), func(i, j int) { userIDs[i], userIDs[j] = userIDs[j], userIDs[i] })
	for i := range orders {
		orders[i].UserID = userIDs[i]
	}
	g := &Game{
		Users:         users,
		Orders:        orders,
		animationStep: stepIdle,
		packetSpeed:   15,
		AnimationType: animationType,
	}
	return g
}

func NewGameJOIN2(animationType string) *Game {
	userMachines := [2][]User{}
	userMachines[0] = []User{
		{UserID: 1, Name: "Alice"}, {UserID: 2, Name: "Bob"}, {UserID: 3, Name: "Charlie"}, {UserID: 4, Name: "David"}, {UserID: 5, Name: "Eve"},
	}
	userMachines[1] = []User{
		{UserID: 6, Name: "Frank"}, {UserID: 7, Name: "Grace"}, {UserID: 8, Name: "Heidi"}, {UserID: 9, Name: "Ivan"}, {UserID: 10, Name: "Judy"},
	}

	orderMachines := [2][]Order{}
	orderMachines[0] = make([]Order, 5)
	orderMachines[1] = make([]Order, 5)
	userIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(userIDs), func(i, j int) { userIDs[i], userIDs[j] = userIDs[j], userIDs[i] })
	for i := 0; i < 5; i++ {
		orderMachines[0][i] = Order{OrderID: 101 + i, UserID: userIDs[i], Item: fmt.Sprintf("Item%d", 101+i)}
		orderMachines[1][i] = Order{OrderID: 106 + i, UserID: userIDs[i+5], Item: fmt.Sprintf("Item%d", 106+i)}
	}

	g := &Game{
		UserMachines:  userMachines,
		OrderMachines: orderMachines,
		animationStep: stepIdle,
		packetSpeed:   10,
		AnimationType: animationType,
	}
	return g
}

func NewGameJOIN3(animationType string) *Game {
	userMachines := [2][]User{}
	userMachines[0] = []User{
		{UserID: 1, Name: "Alice"}, {UserID: 2, Name: "Bob"}, {UserID: 3, Name: "Charlie"}, {UserID: 4, Name: "David"}, {UserID: 5, Name: "Eve"},
	}
	userMachines[1] = []User{
		{UserID: 6, Name: "Frank"}, {UserID: 7, Name: "Grace"}, {UserID: 8, Name: "Heidi"}, {UserID: 9, Name: "Ivan"}, {UserID: 10, Name: "Judy"},
	}

	orderMachines := [2][]Order{}
	orderMachines[0] = make([]Order, 5)
	orderMachines[1] = make([]Order, 5)
	userIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(userIDs), func(i, j int) { userIDs[i], userIDs[j] = userIDs[j], userIDs[i] })

	fullOrderList := make([]Order, 0, 10)
	for i := 0; i < 5; i++ {
		orderMachines[0][i] = Order{OrderID: 101 + i, UserID: userIDs[i], Item: fmt.Sprintf("Item%d", 101+i)}
		orderMachines[1][i] = Order{OrderID: 106 + i, UserID: userIDs[i+5], Item: fmt.Sprintf("Item%d", 106+i)}
		fullOrderList = append(fullOrderList, orderMachines[0][i], orderMachines[1][i])
	}

	index := make([]IndexEntry, 10)
	for i, o := range fullOrderList {
		index[i] = IndexEntry{UserID: o.UserID, OrderID: o.OrderID}
	}
	sort.Slice(index, func(i, j int) bool { return index[i].UserID < index[j].UserID })

	indexMachines := [2][]IndexEntry{}
	indexMachines[0] = index[0:5]
	indexMachines[1] = index[5:10]

	g := &Game{
		UserMachines:  userMachines,
		OrderMachines: orderMachines,
		IndexMachines: indexMachines,
		animationStep: stepIdle,
		packetSpeed:   15,
		AnimationType: animationType,
	}
	return g
}

// --- Core Logic ---

func (g *Game) startAnimation() {
	if g.AnimationType == "JOIN3" {
		g.animationStep = stepUserToIndexRequest
	} else {
		g.animationStep = stepRequesting
	}
	g.currentUserIndex = 0
	g.Joined = []JoinedData{}
	g.animationTimer = time.NewTicker(500 * time.Millisecond)
	g.setPacketStartPosition()
}

func (g *Game) Update() error {
	switch g.AnimationType {
	case "JOIN2":
		return g.updateJOIN2()
	case "JOIN3":
		return g.updateJOIN3()
	default: // JOIN1 and empty
		return g.updateJOIN1()
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.AnimationType {
	case "JOIN2":
		g.drawJOIN2(screen)
	case "JOIN3":
		g.drawJOIN3(screen)
	default: // JOIN1 and empty
		g.drawJOIN1(screen)
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

// --- Update Logic ---

func (g *Game) updateJOIN1() error {
	if g.animationStep == stepIdle {
		if g.AnimationType == "JOIN1" || g.AnimationType == "" {
			if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				g.startAnimation()
			}
		}
	}
	if g.animationStep == stepRequesting {
		found := false
		for i, o := range g.Orders {
			if o.UserID == g.Users[g.currentUserIndex].UserID {
				g.packetTargetX[0] = 550
				g.packetTargetY[0] = 110 + float32(i*30) + 12
				g.setupPacket(0)
				found = true
				break
			}
		}
		if found {
			g.animationStep = stepResponding
		}
	} else if g.animationStep == stepResponding {
		if g.movePacket(0) {
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

func (g *Game) updateJOIN2() error {
	if g.animationStep == stepIdle {
		if g.AnimationType == "JOIN2" {
			g.startAnimation()
		}
	}
	if g.animationStep == stepRequesting {
		for i := 0; i < 2; i++ {
			foundMatch := false
			currentUser := g.UserMachines[i][g.currentUserIndex]
			for orderMachineIndex, orderMachine := range g.OrderMachines {
				for orderIndex, order := range orderMachine {
					if currentUser.UserID == order.UserID {
						g.currentOrderMachineIndex[i] = orderMachineIndex
						g.currentOrderIndex[i] = orderIndex
						orderTableYOffset := float32(50)
						if orderMachineIndex == 1 {
							orderTableYOffset = 350
						}
						g.packetTargetX[i] = 750
						g.packetTargetY[i] = orderTableYOffset + 60 + float32(orderIndex*30) + 12
						g.setupPacket(i)
						foundMatch = true
						break
					}
				}
				if foundMatch {
					break
				}
			}
		}
		g.animationStep = stepResponding
	} else if g.animationStep == stepResponding {
		packetsFinished := 0
		for i := 0; i < 2; i++ {
			if g.movePacket(i) {
				packetsFinished++
			}
		}
		if packetsFinished == 2 {
			g.animationStep = stepJoining
			g.showJoined = true
			for i := 0; i < 2; i++ {
				currentUser := g.UserMachines[i][g.currentUserIndex]
				order := g.OrderMachines[g.currentOrderMachineIndex[i]][g.currentOrderIndex[i]]
				g.Joined = append(g.Joined, JoinedData{User: currentUser, Order: order})
			}
		}
	} else if g.animationStep == stepJoining {
		select {
		case <-g.animationTimer.C:
			g.currentUserIndex++
			if g.currentUserIndex >= len(g.UserMachines[0]) {
				g.animationStep = stepFinished
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

func (g *Game) updateJOIN3() error {
	if g.animationStep == stepIdle {
		if g.AnimationType == "JOIN3" {
			g.startAnimation()
		}
	}

	switch g.animationStep {
	case stepUserToIndexRequest:
		for i := 0; i < 2; i++ {
			currentUser := g.UserMachines[i][g.currentUserIndex]
			found := false
			for j, machine := range g.IndexMachines {
				for k, entry := range machine {
					if entry.UserID == currentUser.UserID {
						g.currentIndexMachineIndex[i] = j
						g.currentIndexIndex[i] = k
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			indexY := float32(50 + 60 + g.currentIndexIndex[i]*30 + 12)
			if g.currentIndexMachineIndex[i] == 1 {
				indexY += 300
			}
			g.packetTargetX[i] = 550
			g.packetTargetY[i] = indexY
			g.setupPacket(i)
		}
		g.animationStep = stepUserToIndexResponse

	case stepUserToIndexResponse:
		packetsFinished := 0
		for i := 0; i < 2; i++ {
			if g.movePacket(i) {
				packetsFinished++
			}
		}
		if packetsFinished == 2 {
			g.animationStep = stepIndexToOrderRequest
		}

	case stepIndexToOrderRequest:
		for i := 0; i < 2; i++ {
			indexEntry := g.IndexMachines[g.currentIndexMachineIndex[i]][g.currentIndexIndex[i]]
			found := false
			for j, machine := range g.OrderMachines {
				for k, order := range machine {
					if order.OrderID == indexEntry.OrderID {
						g.currentOrderMachineIndex[i] = j
						g.currentOrderIndex[i] = k
						found = true
						break
					}
				}
				if found {
					break
				}
			}

			indexY := float32(50 + 60 + g.currentIndexIndex[i]*30 + 12)
			if g.currentIndexMachineIndex[i] == 1 {
				indexY += 300
			}
			g.packetStartX[i] = 950
			g.packetStartY[i] = indexY
			g.packetX[i], g.packetY[i] = g.packetStartX[i], g.packetStartY[i]

			orderY := float32(50 + 60 + g.currentOrderIndex[i]*30 + 12)
			if g.currentOrderMachineIndex[i] == 1 {
				orderY += 300
			}
			g.packetTargetX[i] = 1050
			g.packetTargetY[i] = orderY
			g.setupPacket(i)
		}
		g.animationStep = stepIndexToOrderResponse

	case stepIndexToOrderResponse:
		packetsFinished := 0
		for i := 0; i < 2; i++ {
			if g.movePacket(i) {
				packetsFinished++
			}
		}
		if packetsFinished == 2 {
			for i := 0; i < 2; i++ {
				user := g.UserMachines[i][g.currentUserIndex]
				order := g.OrderMachines[g.currentOrderMachineIndex[i]][g.currentOrderIndex[i]]
				g.Joined = append(g.Joined, JoinedData{User: user, Order: order})
			}
			g.showJoined = true
			g.animationStep = stepJoining
		}

	case stepJoining:
		select {
		case <-g.animationTimer.C:
			g.currentUserIndex++
			if g.currentUserIndex >= len(g.UserMachines[0]) {
				g.animationStep = stepFinished
			} else {
				g.animationStep = stepUserToIndexRequest
				g.setPacketStartPosition()
			}
		default:
		}

	case stepFinished:
		g.startAnimation()
	}
	return nil
}

// --- Draw Logic ---

func (g *Game) drawJOIN1(screen *ebiten.Image) {
	g.drawTablesJOIN1(screen)
	g.drawJoinedTable(screen, 520)
	if g.animationStep == stepResponding {
		vector.DrawFilledCircle(screen, g.packetX[0], g.packetY[0], 5, color.RGBA{R: 0xff, A: 0xff}, false)
	}
	if g.animationStep == stepIdle {
		g.drawScaledText(screen, "Press Space to Start Animation", 393, screenHeight-40, color.White)
	}
}

func (g *Game) drawJOIN2(screen *ebiten.Image) {
	g.drawTablesJOIN2(screen)
	g.drawJoinedTable(screen, 650)
	if g.animationStep == stepResponding {
		for i := 0; i < 2; i++ {
			vector.DrawFilledCircle(screen, g.packetX[i], g.packetY[i], 5, color.RGBA{R: 0xff, A: 0xff}, false)
		}
	}
}

func (g *Game) drawJOIN3(screen *ebiten.Image) {
	g.drawTablesJOIN3(screen)
	g.drawJoinedTable(screen, 650)
	if g.animationStep == stepUserToIndexResponse || g.animationStep == stepIndexToOrderResponse {
		for i := 0; i < 2; i++ {
			vector.DrawFilledCircle(screen, g.packetX[i], g.packetY[i], 5, color.RGBA{R: 0xff, A: 0xff}, false)
		}
	}
}

func (g *Game) drawTablesJOIN1(screen *ebiten.Image) {
	// User Table
	vector.DrawFilledRect(screen, 50, 50, 400, 450, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
	g.drawScaledText(screen, "User Table", 60, 60, color.White)
	for i, u := range g.Users {
		var c color.Color = color.White
		if g.animationStep > stepIdle && g.currentUserIndex == i {
			c = color.RGBA{R: 0xff, G: 0xff, A: 0xff}
		}
		g.drawScaledText(screen, fmt.Sprintf("UserID: %d, Name: %s", u.UserID, u.Name), 60, 110+i*30, c)
	}

	// Order Table
	vector.DrawFilledRect(screen, 550, 50, 500, 450, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
	g.drawScaledText(screen, "Order Table", 560, 60, color.White)
	for i, o := range g.Orders {
		var c color.Color = color.White
		if g.animationStep == stepResponding || g.animationStep == stepJoining {
			if o.UserID == g.Users[g.currentUserIndex].UserID {
				c = color.RGBA{R: 0xff, G: 0xff, A: 0xff}
			}
		}
		g.drawScaledText(screen, fmt.Sprintf("OrderID: %d, UserID: %d, Item: %s", o.OrderID, o.UserID, o.Item), 560, 110+i*30, c)
	}
}

func (g *Game) drawTablesJOIN2(screen *ebiten.Image) {
	// User Machines
	for i := 0; i < 2; i++ {
		yOffset := float32(50 + i*300)
		vector.DrawFilledRect(screen, 50, yOffset, 400, 250, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
		g.drawScaledText(screen, fmt.Sprintf("User Machine %d", i+1), 60, int(yOffset)+10, color.White)
		for j, u := range g.UserMachines[i] {
			var c color.Color = color.White
			if g.animationStep > stepIdle && g.currentUserIndex == j {
				c = color.RGBA{R: 0xff, G: 0xff, A: 0xff}
			}
			g.drawScaledText(screen, fmt.Sprintf("UserID: %d, Name: %s", u.UserID, u.Name), 60, int(yOffset)+60+j*30, c)
		}
	}

	// Order Machines
	for i := 0; i < 2; i++ {
		yOffset := float32(50 + i*300)
		vector.DrawFilledRect(screen, 750, yOffset, 550, 250, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
		g.drawScaledText(screen, fmt.Sprintf("Order Machine %d", i+1), 760, int(yOffset)+10, color.White)
		for j, o := range g.OrderMachines[i] {
			var c color.Color = color.White
			if g.animationStep > stepRequesting {
				if (g.currentOrderMachineIndex[0] == i && g.currentOrderIndex[0] == j) || (g.currentOrderMachineIndex[1] == i && g.currentOrderIndex[1] == j) {
					c = color.RGBA{R: 0xff, G: 0xff, A: 0xff} // Yellow
				}
			}
			g.drawScaledText(screen, fmt.Sprintf("OrderID: %d, UserID: %d, Item: %s", o.OrderID, o.UserID, o.Item), 760, int(yOffset)+60+j*30, c)
		}
	}
}

func (g *Game) drawTablesJOIN3(screen *ebiten.Image) {
	// User Machines
	for i := 0; i < 2; i++ {
		yOffset := float32(50 + i*300)
		vector.DrawFilledRect(screen, 50, yOffset, 400, 250, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
		g.drawScaledText(screen, fmt.Sprintf("User Machine %d", i+1), 60, int(yOffset)+10, color.White)
		for j, u := range g.UserMachines[i] {
			var c color.Color = color.White
			if g.animationStep >= stepUserToIndexRequest && g.currentUserIndex == j {
				c = color.RGBA{R: 0xff, G: 0xff, A: 0xff}
			}
			g.drawScaledText(screen, fmt.Sprintf("UserID: %d, Name: %s", u.UserID, u.Name), 60, int(yOffset)+60+j*30, c)
		}
	}

	// Index Machines
	for i := 0; i < 2; i++ {
		yOffset := float32(50 + i*300)
		vector.DrawFilledRect(screen, 550, yOffset, 400, 250, color.RGBA{R: 0x30, G: 0x30, B: 0x60, A: 0xff}, false)
		g.drawScaledText(screen, fmt.Sprintf("Index Machine %d", i+1), 560, int(yOffset)+10, color.White)
		for j, entry := range g.IndexMachines[i] {
			var c color.Color = color.White
			if g.animationStep >= stepUserToIndexResponse {
				if (g.currentIndexMachineIndex[0] == i && g.currentIndexIndex[0] == j) || (g.currentIndexMachineIndex[1] == i && g.currentIndexIndex[1] == j) {
					c = color.RGBA{R: 0xff, G: 0xff, A: 0xff}
				}
			}
			g.drawScaledText(screen, fmt.Sprintf("UserID: %d, OrderID: %d", entry.UserID, entry.OrderID), 560, int(yOffset)+60+j*30, c)
		}
	}

	// Order Machines
	for i := 0; i < 2; i++ {
		yOffset := float32(50 + i*300)
		vector.DrawFilledRect(screen, 1050, yOffset, 500, 250, color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)
		g.drawScaledText(screen, fmt.Sprintf("Order Machine %d", i+1), 1060, int(yOffset)+10, color.White)
		for j, o := range g.OrderMachines[i] {
			var c color.Color = color.White
			if g.animationStep == stepIndexToOrderResponse {
				if (g.currentOrderMachineIndex[0] == i && g.currentOrderIndex[0] == j) || (g.currentOrderMachineIndex[1] == i && g.currentOrderIndex[1] == j) {
					c = color.RGBA{R: 0xff, G: 0xff, A: 0xff}
				}
			}
			g.drawScaledText(screen, fmt.Sprintf("OrderID: %d, UserID: %d, Item: %s", o.OrderID, o.UserID, o.Item), 1060, int(yOffset)+60+j*30, c)
		}
	}
}

func (g *Game) drawJoinedTable(screen *ebiten.Image, yPos float32) {
	if !g.showJoined {
		return
	}
	vector.DrawFilledRect(screen, 50, yPos, 1500, 300, color.RGBA{R: 0x30, G: 0x30, B: 0x60, A: 0xff}, false)
	g.drawScaledText(screen, "JOIN Result", 60, int(yPos)+10, color.White)
	for i, j := range g.Joined {
		row := i / 2
		col := i % 2
		x := 60 + col*770
		y := int(yPos) + 60 + row*30
		g.drawScaledText(screen, fmt.Sprintf("UserID: %d, Name: %s, OrderID: %d, Item: %s", j.User.UserID, j.User.Name, j.Order.OrderID, j.Order.Item), x, y, color.White)
	}
}

// --- Helpers ---

func (g *Game) setPacketStartPosition() {
	switch g.AnimationType {
	case "JOIN2":
		g.setPacketStartPositionJOIN2()
	case "JOIN3":
		g.setPacketStartPositionJOIN3()
	default: // JOIN1
		userY := 110 + float32(g.currentUserIndex*30) + 12
		g.packetX[0] = 450
		g.packetY[0] = userY
		g.packetStartX[0], g.packetStartY[0] = g.packetX[0], g.packetY[0]
	}
}

func (g *Game) setPacketStartPositionJOIN2() {
	for i := 0; i < 2; i++ {
		userTableYOffset := float32(50)
		if i == 1 {
			userTableYOffset = 350
		}
		userY := userTableYOffset + 60 + float32(g.currentUserIndex*30) + 12
		g.packetX[i] = 450
		g.packetY[i] = userY
		g.packetStartX[i], g.packetStartY[i] = g.packetX[i], g.packetY[i]
	}
}

func (g *Game) setPacketStartPositionJOIN3() {
	for i := 0; i < 2; i++ {
		userY := float32(50 + 60 + g.currentUserIndex*30 + 12)
		if i == 1 {
			userY += 300
		}
		g.packetStartX[i] = 450
		g.packetStartY[i] = userY
		g.packetX[i], g.packetY[i] = g.packetStartX[i], g.packetStartY[i]
	}
}

func (g *Game) setupPacket(i int) {
	deltaX := g.packetTargetX[i] - g.packetStartX[i]
	deltaY := g.packetTargetY[i] - g.packetStartY[i]
	distance := float32(math.Sqrt(float64(deltaX*deltaX + deltaY*deltaY)))
	if distance > 0 {
		g.packetSpeedX[i] = g.packetSpeed * deltaX / distance
		g.packetSpeedY[i] = g.packetSpeed * deltaY / distance
	}
}

func (g *Game) movePacket(i int) bool {
	g.packetX[i] += g.packetSpeedX[i]
	g.packetY[i] += g.packetSpeedY[i]
	dx1 := g.packetX[i] - g.packetStartX[i]
	dy1 := g.packetY[i] - g.packetStartY[i]
	dx2 := g.packetTargetX[i] - g.packetStartX[i]
	dy2 := g.packetTargetY[i] - g.packetStartY[i]
	return dx1*dx1+dy1*dy1 >= dx2*dx2+dy2*dy2
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

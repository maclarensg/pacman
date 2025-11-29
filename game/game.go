package game

import (
	"fmt"
	"image"
	"os"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/mattn/go-sixel"
	"golang.org/x/term"
)

type Game struct {
	renderer   *Renderer
	maze       *Maze
	pacman     *Pacman
	ghosts     []*Ghost
	state      GameState
	score      int
	lives      int
	level      int
	scale      int
	width      int
	height     int
	frame      int
	fruit      *Fruit
	fruitTimer int
	lastFrameTime time.Time
	fps        float64
}

// GhostPos stores a ghost's previous position for collision detection
type GhostPos struct {
	prevX, prevY int
}

func NewGame() (*Game, error) {
	// Initialize keyboard
	if err := keyboard.Open(); err != nil {
		return nil, err
	}

	// Calculate scale
	scale := calculateScale()

	game := &Game{
		renderer:   NewRenderer(scale),
		maze:       NewMaze(),
		pacman:     NewPacman(),
		state:      StateStart,
		score:      0,
		lives:      InitialLives,
		level:      1,
		scale:      scale,
		width:      BaseWidth * scale,
		height:     BaseHeight * scale,
		frame:      0,
		fruit:      nil,
		fruitTimer: 0,
	}

	// Create ghosts at starting positions (all outside house, spread out)
	game.ghosts = []*Ghost{
		NewGhost(GhostBlinky, 14, 11), // Center
		NewGhost(GhostPinky, 12, 11),  // Left
		NewGhost(GhostInky, 16, 11),   // Right
		NewGhost(GhostClyde, 14, 9),   // Above
	}

	return game, nil
}

func calculateScale() int {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 2
	}
	pixelWidth := width * 8
	pixelHeight := height * 16
	scaleWidth := pixelWidth / BaseWidth
	scaleHeight := pixelHeight / BaseHeight
	scale := scaleWidth
	if scaleHeight < scaleWidth {
		scale = scaleHeight
	}
	if scale < 1 {
		scale = 1
	}
	if scale > 5 {
		scale = 5
	}
	return scale
}

func (g *Game) Run() {
	// Setup terminal - alternate screen buffer and hide cursor
	fmt.Print("\033[?1049h") // Enter alternate screen
	fmt.Print("\033[?25l")   // Hide cursor
	fmt.Print("\033[2J")     // Clear screen

	// Render start screen once
	g.renderStartScreen()

	// Game loop timing
	ticker := time.NewTicker(time.Second / time.Duration(TicksPerSecond))
	defer ticker.Stop()

	// Keyboard input goroutine
	keyChan := make(chan keyboard.KeyEvent, 100) // Larger buffer for better responsiveness
	go func() {
		for {
			char, key, err := keyboard.GetKey()
			if err == nil {
				keyChan <- keyboard.KeyEvent{Rune: char, Key: key}
			}
		}
	}()

	for {
		select {
		case ev := <-keyChan:
			if g.handleInput(ev) {
				return // Quit
			}
		case <-ticker.C:
			// Process ALL pending input events first (drain the channel)
			draining := true
			for draining {
				select {
				case ev := <-keyChan:
					if g.handleInput(ev) {
						return // Quit
					}
				default:
					draining = false
				}
			}

			// Calculate FPS
			now := time.Now()
			if !g.lastFrameTime.IsZero() {
				frameTime := now.Sub(g.lastFrameTime).Seconds()
				if frameTime > 0 {
					g.fps = 1.0 / frameTime
				}
			}
			g.lastFrameTime = now

			g.update()
			g.render()
			g.frame++
		}
	}
}

func (g *Game) handleInput(ev keyboard.KeyEvent) bool {
	// Check for quit keys
	if ev.Key == keyboard.KeyEsc || ev.Key == keyboard.KeyCtrlC || ev.Rune == 'q' {
		return true // Quit
	}

	// Handle start screen
	if g.state == StateStart {
		if ev.Key == keyboard.KeySpace {
			g.state = StatePlaying
		}
		return false
	}

	// Handle game over state
	if g.state == StateGameOver {
		if ev.Rune == 'r' || ev.Rune == 'R' {
			g.Reset() // Retry
			return false
		}
		return false // Stay in game over until quit or retry
	}

	// Handle playing state
	if g.state == StatePlaying {
		switch ev.Key {
		case keyboard.KeyArrowUp:
			g.pacman.SetDirection(DirUp)
		case keyboard.KeyArrowDown:
			g.pacman.SetDirection(DirDown)
		case keyboard.KeyArrowLeft:
			g.pacman.SetDirection(DirLeft)
		case keyboard.KeyArrowRight:
			g.pacman.SetDirection(DirRight)
		}
	}
	return false
}

func (g *Game) update() {
	if g.state != StatePlaying {
		return
	}

	// Spawn fruit periodically
	g.fruitTimer++
	if g.fruit == nil && g.fruitTimer > 600 { // Every 10 seconds
		g.fruit = NewFruit(g.level)
		g.fruitTimer = 0
	}

	// Despawn fruit after 8 seconds
	if g.fruit != nil && g.fruit.Active && !g.fruit.Eaten {
		g.fruit.SpawnTime++
		if g.fruit.SpawnTime > 480 { // 8 seconds
			g.fruit = nil
			g.fruitTimer = 0
		}
	}

	// Store Pac-Man's previous position for collision detection
	prevPacX, prevPacY := g.pacman.X, g.pacman.Y

	// Update Pacman
	g.pacman.Update(g.maze)

	// Check pellet eating
	isPower, ate := g.maze.EatPellet(g.pacman.X, g.pacman.Y)
	if ate {
		if isPower {
			g.score += PowerPelletScore
			g.pacman.ActivatePowerMode(g.level)
		} else {
			g.score += PelletScore
		}
	}

	// Check fruit eating
	if g.fruit != nil && g.fruit.Active && !g.fruit.Eaten {
		if g.pacman.X == g.fruit.X && g.pacman.Y == g.fruit.Y {
			g.score += FruitScores[g.fruit.Type]
			g.fruit.Eaten = true
			g.fruit = nil
		}
	}

	// Update ghosts (pass level for difficulty scaling)
	// Store previous positions for collision prevention
	prevPositions := make([]GhostPos, len(g.ghosts))
	for i, ghost := range g.ghosts {
		prevPositions[i] = GhostPos{ghost.X, ghost.Y}
		ghost.Update(g.maze, g.pacman, g.level)
	}

	// Prevent ghosts from overlapping - if two ghosts are at same position,
	// move the second one back to its previous position
	for i := 0; i < len(g.ghosts); i++ {
		for j := i + 1; j < len(g.ghosts); j++ {
			if g.ghosts[i].X == g.ghosts[j].X && g.ghosts[i].Y == g.ghosts[j].Y {
				// Collision! Move the second ghost back
				g.ghosts[j].X = prevPositions[j].prevX
				g.ghosts[j].Y = prevPositions[j].prevY
			}
		}
	}

	// Check collisions (with position swap detection)
	g.checkCollisions(prevPacX, prevPacY, prevPositions)

	// Check win (level complete)
	if g.maze.RemainingPellets == 0 {
		g.NextLevel()
	}
}

func (g *Game) NextLevel() {
	g.level++
	g.maze.Reset()

	// Reset positions
	g.pacman.X = 14
	g.pacman.Y = 23
	g.pacman.Dir = DirNone
	g.pacman.NextDir = DirNone
	g.pacman.PowerMode = false

	// Reset ghosts (all outside house, spread out)
	g.ghosts[0].X, g.ghosts[0].Y = 14, 11 // Blinky center
	g.ghosts[1].X, g.ghosts[1].Y = 12, 11 // Pinky left
	g.ghosts[2].X, g.ghosts[2].Y = 16, 11 // Inky right
	g.ghosts[3].X, g.ghosts[3].Y = 14, 9  // Clyde above
	for _, ghost := range g.ghosts {
		ghost.Mode = ModeScatter
		ghost.Dir = DirLeft
	}

	g.state = StatePlaying
}

func (g *Game) Reset() {
	g.score = 0
	g.lives = InitialLives
	g.level = 1
	g.maze.Reset()

	// Reset positions
	g.pacman.X = 14
	g.pacman.Y = 23
	g.pacman.Dir = DirNone
	g.pacman.NextDir = DirNone
	g.pacman.PowerMode = false

	// Reset ghosts (all outside house, spread out)
	g.ghosts[0].X, g.ghosts[0].Y = 14, 11 // Blinky center
	g.ghosts[1].X, g.ghosts[1].Y = 12, 11 // Pinky left
	g.ghosts[2].X, g.ghosts[2].Y = 16, 11 // Inky right
	g.ghosts[3].X, g.ghosts[3].Y = 14, 9  // Clyde above
	for _, ghost := range g.ghosts {
		ghost.Mode = ModeScatter
		ghost.Dir = DirLeft
	}

	g.state = StatePlaying
}

func (g *Game) checkCollisions(prevPacX, prevPacY int, prevGhostPos []GhostPos) {
	for i, ghost := range g.ghosts {
		// Check for collision - either at same position OR position swap (passing through each other)
		collision := false

		// Case 1: Same position now
		if g.pacman.X == ghost.X && g.pacman.Y == ghost.Y {
			collision = true
		}

		// Case 2: Position swap - Pac-Man moved to where ghost was, ghost moved to where Pac-Man was
		if g.pacman.X == prevGhostPos[i].prevX && g.pacman.Y == prevGhostPos[i].prevY &&
			ghost.X == prevPacX && ghost.Y == prevPacY {
			collision = true
		}

		if collision {
			if g.pacman.PowerMode && ghost.Mode == ModeFrightened {
				// Eat ghost - becomes eyes and returns to ghost house ENTRANCE
				g.score += GhostScore
				ghost.Mode = ModeEaten
				ghost.TargetX = 14
				ghost.TargetY = 11 // Target entrance above ghost house, not inside
			} else if ghost.Mode != ModeFrightened && ghost.Mode != ModeEaten {
				// Pac-Man dies
				g.lives--
				g.pacman.X = 14
				g.pacman.Y = 23
				g.pacman.Dir = DirNone
				if g.lives <= 0 {
					g.state = StateGameOver
				}
			}
		}
	}
}

func (g *Game) render() {
	// Don't re-render start screen (already rendered once at startup)
	if g.state == StateStart {
		return
	}

	// Create image buffer
	screen := image.NewRGBA(image.Rect(0, 0, g.width, g.height))

	// Fill background with black (CRITICAL - test files do this!)
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			screen.Set(x, y, ColorBlack)
		}
	}

	// Render maze (with level-based color)
	g.renderer.RenderMaze(g.maze, screen, g.frame, g.level)

	// Render Pacman
	g.renderer.RenderPacman(screen, g.pacman.X, g.pacman.Y, g.pacman.Dir, g.pacman.AnimFrame)

	// Render fruit
	if g.fruit != nil && g.fruit.Active && !g.fruit.Eaten {
		g.renderer.RenderFruit(screen, g.fruit)
	}

	// Render ghosts
	for _, ghost := range g.ghosts {
		frightened := ghost.Mode == ModeFrightened
		blinking := false
		isEyes := ghost.Mode == ModeEaten

		// Check if power mode is ending (last ~2 seconds)
		if frightened && g.pacman.PowerMode {
			// Power mode lasts 48 frames (~6 seconds at 8 FPS)
			// Blink during last ~2 seconds = last 16 frames
			remainingFrames := g.pacman.PowerTimeLeft()
			if remainingFrames < 16 {
				blinking = true
			}
		}

		color := ghost.GetColor()
		g.renderer.RenderGhost(screen, ghost.X, ghost.Y, color, frightened, blinking, isEyes, g.frame)
	}

	// Render HUD
	g.renderer.RenderHUD(screen, g.score, g.lives, g.level)

	// Render game over message
	switch g.state {
	case StateGameOver:
		g.renderer.RenderGameOver(screen, false)
	case StateWin:
		g.renderer.RenderGameOver(screen, true)
	}

	// Output to terminal using Sixel
	fmt.Print("\033[H") // Move cursor to home
	enc := sixel.NewEncoder(os.Stdout)
	if err := enc.Encode(screen); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode screen: %v\n", err)
	}

	// Print HUD below the game screen
	powerInfo := ""
	if g.pacman.PowerMode {
		powerInfo = fmt.Sprintf(" POWER: %d ", g.pacman.PowerTicks)
	}
	fmt.Printf("\n\033[1;33m SCORE: %-8d LIVES: %d    LEVEL: %d %s FPS: %.1f \033[0m\n", g.score, g.lives, g.level, powerInfo, g.fps)

	// Print game state messages
	switch g.state {
	case StateGameOver:
		fmt.Print("\033[1;31m")
		fmt.Println("\n ═══════════════════════════════════════")
		fmt.Println("         GAME OVER!")
		fmt.Println(" ═══════════════════════════════════════")
		fmt.Print("\033[1;37m")
		fmt.Println("   Press R to Retry  |  Q to Quit")
		fmt.Print("\033[0m")
	case StateWin:
		fmt.Print("\033[1;32m")
		fmt.Println("\n ═══════════════════════════════════════")
		fmt.Printf("      LEVEL %d COMPLETE!\n", g.level)
		fmt.Println(" ═══════════════════════════════════════")
		fmt.Print("\033[1;37m")
		fmt.Println("      Starting next level...")
		fmt.Print("\033[0m")
	}
}

func (g *Game) Cleanup() {
	// Restore terminal state
	fmt.Print("\033[?25h")   // Show cursor
	fmt.Print("\033[?1049l") // Exit alternate screen
	if err := keyboard.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to close keyboard: %v\n", err)
	}
}

func (g *Game) renderStartScreen() {
	fmt.Print("\033[H") // Move cursor to home

	fmt.Print("\033[1;36m") // Cyan color
	fmt.Println(`
    ██████╗  █████╗  ██████╗    ███╗   ███╗ █████╗ ███╗   ██╗
    ██╔══██╗██╔══██╗██╔════╝    ████╗ ████║██╔══██╗████╗  ██║
    ██████╔╝███████║██║         ██╔████╔██║███████║██╔██╗ ██║
    ██╔═══╝ ██╔══██║██║         ██║╚██╔╝██║██╔══██║██║╚██╗██║
    ██║     ██║  ██║╚██████╗    ██║ ╚═╝ ██║██║  ██║██║ ╚████║
    ╚═╝     ╚═╝  ╚═╝ ╚═════╝    ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝
	`)
	fmt.Print("\033[0m")

	fmt.Println()
	fmt.Print("\033[1;33m") // Yellow
	fmt.Println("          🍒 CLASSIC 1980s ARCADE EDITION 🍒")
	fmt.Print("\033[0m")

	fmt.Println()
	fmt.Println()
	fmt.Print("\033[1;37m") // White
	fmt.Println("  CONTROLS:")
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   ↑ ↓ ← →  : Move Pac-Man")
	fmt.Println("   Q / ESC  : Quit Game")
	fmt.Println("   R        : Retry (Game Over)")
	fmt.Print("\033[0m")

	fmt.Println()
	fmt.Println()
	fmt.Print("\033[1;32m") // Green
	fmt.Println("  SCORING:")
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   Pellet        :    10 points")
	fmt.Println("   Power Pellet  :    50 points")
	fmt.Println("   Ghost         :   200 points")
	fmt.Println("   Fruit         : 100-5000 points")
	fmt.Print("\033[0m")

	fmt.Println()
	fmt.Println()
	fmt.Print("\033[1;35m") // Magenta
	fmt.Println("  ═══════════════════════════════════════════")
	fmt.Println()
	fmt.Println("       Press SPACE to Start!")
	fmt.Println()
	fmt.Println("  ═══════════════════════════════════════════")
	fmt.Print("\033[0m")
}

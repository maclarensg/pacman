package game

import "image/color"

// Direction represents movement direction
type Direction int

const (
	DirNone Direction = iota
	DirUp
	DirDown
	DirLeft
	DirRight
)

// Position in pixel coordinates
type Position struct {
	X, Y int
}

// Colors matching original arcade Pac-Man
var (
	ColorBlack        = color.RGBA{0, 0, 0, 255}
	ColorPacmanYellow = color.RGBA{255, 255, 0, 255}
	ColorBlinkyRed    = color.RGBA{255, 0, 0, 255}
	ColorPinkyPink    = color.RGBA{255, 184, 255, 255}
	ColorInkyCyan     = color.RGBA{0, 255, 255, 255}
	ColorClydeOrange  = color.RGBA{255, 184, 82, 255}
	ColorMazeBlue     = color.RGBA{33, 33, 255, 255}
	ColorPellet       = color.RGBA{255, 183, 174, 255}
	ColorWhite        = color.RGBA{255, 255, 255, 255}
	ColorFrightened   = color.RGBA{33, 33, 255, 255}
	ColorEyeWhite     = color.RGBA{255, 255, 255, 255}
	ColorEyeBlue      = color.RGBA{33, 33, 255, 255}
)

// Game constants
const (
	// Pixel-based measurements
	TileSize        = 8  // Each tile is 8x8 pixels
	MazeWidthTiles  = 28 // Maze is 28 tiles wide
	MazeHeightTiles = 31 // Maze is 31 tiles tall
	BaseWidth       = MazeWidthTiles * TileSize  // 224 pixels
	BaseHeight      = MazeHeightTiles * TileSize // 248 pixels

	// Game speed (base values, adjusted by level)
	TicksPerSecond        = 30 // Reduced from 60 for better Sixel performance
	BaseMovementSpeed     = 2 // Move every 2 ticks (30 moves per second)
	BaseGhostSpeed        = 2 // Ghosts same speed
	AnimationSpeed        = 2 // Animation frame change (every 2 frames = 30fps)

	// Computed values (compatibility)
	MovementSpeed      = BaseMovementSpeed
	GhostMovementSpeed = BaseGhostSpeed

	// Game mechanics
	InitialLives     = 3
	PelletScore      = 10
	PowerPelletScore = 50
	GhostScore       = 200
	PowerDuration    = 48 // ticks (~6 seconds at 8 FPS)
)

// Ghost types
type GhostType int

const (
	GhostBlinky GhostType = iota
	GhostPinky
	GhostInky
	GhostClyde
)

// Ghost modes
type GhostMode int

const (
	ModeScatter GhostMode = iota
	ModeChase
	ModeFrightened
	ModeEaten
)

// Game state
type GameState int

const (
	StateStart GameState = iota
	StatePlaying
	StateGameOver
	StateWin
)

// Fruit types
type FruitType int

const (
	FruitCherry FruitType = iota
	FruitStrawberry
	FruitOrange
	FruitApple
	FruitMelon
	FruitGalaxian
	FruitBell
	FruitKey
)

// Fruit scores
var FruitScores = map[FruitType]int{
	FruitCherry:     100,
	FruitStrawberry: 300,
	FruitOrange:     500,
	FruitApple:      700,
	FruitMelon:      1000,
	FruitGalaxian:   2000,
	FruitBell:       3000,
	FruitKey:        5000,
}

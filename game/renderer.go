package game

import (
	"image"
	"image/color"

	"pacman/sprites"
)

// Renderer handles drawing to the screen using Sixel graphics
type Renderer struct {
	scale int
}

func NewRenderer(scale int) *Renderer {
	return &Renderer{scale: scale}
}

// RenderMaze renders the entire maze to an image (with level-based color)
func (r *Renderer) RenderMaze(maze *Maze, img *image.RGBA, frame int, level int) {
	// Power pellet blinks same speed as Pac-Man mouth (every 2 frames = 30 times per second)
	showPowerPellet := (frame/2)%2 == 0

	// Change maze color based on level
	mazeColor := r.getMazeColor(level)

	for y := 0; y < maze.Height; y++ {
		for x := 0; x < maze.Width; x++ {
			cell := maze.GetCell(x, y)
			tileX := x * TileSize * r.scale
			tileY := y * TileSize * r.scale

			switch cell {
			case CellWall:
				// Wall tile with level-based color (scaled)
				for py := 0; py < TileSize*r.scale; py++ {
					for px := 0; px < TileSize*r.scale; px++ {
						img.Set(tileX+px, tileY+py, mazeColor)
					}
				}
			case CellPellet:
				// Small pellet (scaled 2x2 → 4x4 at scale 2)
				for py := 3 * r.scale; py < 5*r.scale; py++ {
					for px := 3 * r.scale; px < 5*r.scale; px++ {
						img.Set(tileX+px, tileY+py, ColorPellet)
					}
				}
			case CellPowerPellet:
				// Power pellet BLINKS (scaled 4x4 → 8x8 at scale 2)
				if showPowerPellet {
					for py := 2 * r.scale; py < 6*r.scale; py++ {
						for px := 2 * r.scale; px < 6*r.scale; px++ {
							img.Set(tileX+px, tileY+py, ColorPellet)
						}
					}
				}
			case CellGhostDoor:
				// Ghost door (pink line, scaled)
				for py := 0; py < r.scale; py++ {
					for px := 0; px < TileSize*r.scale; px++ {
						img.Set(tileX+px, tileY+TileSize*r.scale/2+py, ColorPinkyPink)
					}
				}
			}
		}
	}
}

// RenderPacman renders Pac-Man with pixel art
func (r *Renderer) RenderPacman(img *image.RGBA, x, y int, dir Direction, animFrame int) {
	// Determine which Pac-Man sprite to use based on animation frame
	animCycle := (animFrame / AnimationSpeed) % 2
	var pacmanSprite [][]int

	if animCycle == 0 {
		pacmanSprite = sprites.PacmanClosed
	} else {
		// Rotate sprite based on direction
		switch dir {
		case DirRight:
			pacmanSprite = sprites.PacmanOpenRight
		case DirLeft:
			pacmanSprite = r.flipHorizontal(sprites.PacmanOpenRight)
		case DirUp:
			pacmanSprite = r.rotate90CCW(sprites.PacmanOpenRight)
		case DirDown:
			pacmanSprite = r.rotate90CW(sprites.PacmanOpenRight)
		default:
			pacmanSprite = sprites.PacmanOpenRight
		}
	}

	// NO OFFSET - render exactly like test files
	r.renderSprite(img, x*TileSize*r.scale, y*TileSize*r.scale, pacmanSprite, sprites.ColorPacmanYellow)
}

// RenderGhost renders a ghost with pixel art
func (r *Renderer) RenderGhost(img *image.RGBA, x, y int, bodyColor color.RGBA, frightened bool, blinking bool, isEyes bool, frame int) {
	// If eaten, show only eyes
	if isEyes {
		r.renderGhostEyes(img, x*TileSize*r.scale, y*TileSize*r.scale)
		return
	}

	// If frightened and blinking, alternate between blue and white
	if frightened && blinking {
		// Blink every 3 frames (~20 times per second at 60 FPS)
		if ((frame / 3) % 2) == 0 {
			bodyColor = color.RGBA{33, 33, 255, 255} // Blue
		} else {
			bodyColor = color.RGBA{255, 255, 255, 255} // White
		}
	} else if frightened {
		bodyColor = ColorFrightened
	}

	// NO OFFSET - render exactly like test files
	r.renderGhostSprite(img, x*TileSize*r.scale, y*TileSize*r.scale, bodyColor)
}

// RenderHUD renders score, lives, etc. (uses terminal output instead of image)
func (r *Renderer) RenderHUD(img *image.RGBA, score, lives, level int) {
	// HUD is rendered using terminal text after the Sixel image
}

// RenderGameOver renders game over message (uses terminal output)
func (r *Renderer) RenderGameOver(img *image.RGBA, won bool) {
	// Game over is rendered using terminal text after the Sixel image
}

// Helper function to render a sprite
func (r *Renderer) renderSprite(img *image.RGBA, x, y int, sprite [][]int, bodyColor color.RGBA) {
	pixelSize := r.scale / 2
	if pixelSize < 1 {
		pixelSize = 1
	}

	for py := 0; py < len(sprite); py++ {
		for px := 0; px < len(sprite[py]); px++ {
			if sprite[py][px] == 1 {
				for dy := 0; dy < pixelSize; dy++ {
					for dx := 0; dx < pixelSize; dx++ {
						img.Set(x+px*pixelSize+dx, y+py*pixelSize+dy, bodyColor)
					}
				}
			}
		}
	}
}

// Helper function to render a ghost sprite with eyes
func (r *Renderer) renderGhostSprite(img *image.RGBA, x, y int, bodyColor color.RGBA) {
	pixelSize := r.scale / 2
	if pixelSize < 1 {
		pixelSize = 1
	}

	for py := 0; py < len(sprites.Ghost); py++ {
		for px := 0; px < len(sprites.Ghost[py]); px++ {
			var c color.RGBA
			switch sprites.Ghost[py][px] {
			case 1:
				c = bodyColor
			case 2:
				c = sprites.ColorEyeWhite
			case 3:
				c = sprites.ColorEyeBlue
			default:
				continue
			}
			for dy := 0; dy < pixelSize; dy++ {
				for dx := 0; dx < pixelSize; dx++ {
					img.Set(x+px*pixelSize+dx, y+py*pixelSize+dy, c)
				}
			}
		}
	}
}

// Sprite rotation helpers
func (r *Renderer) flipHorizontal(sprite [][]int) [][]int {
	height := len(sprite)
	width := len(sprite[0])
	result := make([][]int, height)
	for y := 0; y < height; y++ {
		result[y] = make([]int, width)
		for x := 0; x < width; x++ {
			result[y][x] = sprite[y][width-1-x]
		}
	}
	return result
}

func (r *Renderer) rotate90CW(sprite [][]int) [][]int {
	height := len(sprite)
	width := len(sprite[0])
	result := make([][]int, width)
	for y := 0; y < width; y++ {
		result[y] = make([]int, height)
		for x := 0; x < height; x++ {
			result[y][x] = sprite[height-1-x][y]
		}
	}
	return result
}

func (r *Renderer) rotate90CCW(sprite [][]int) [][]int {
	height := len(sprite)
	width := len(sprite[0])
	result := make([][]int, width)
	for y := 0; y < width; y++ {
		result[y] = make([]int, height)
		for x := 0; x < height; x++ {
			result[y][x] = sprite[x][width-1-y]
		}
	}
	return result
}

// Get maze color based on level
func (r *Renderer) getMazeColor(level int) color.RGBA {
	colors := []color.RGBA{
		{33, 33, 255, 255},    // Blue (level 1)
		{0, 255, 0, 255},      // Green (level 2)
		{255, 0, 255, 255},    // Magenta (level 3)
		{255, 165, 0, 255},    // Orange (level 4)
		{255, 0, 0, 255},      // Red (level 5)
		{0, 255, 255, 255},    // Cyan (level 6)
		{255, 255, 0, 255},    // Yellow (level 7)
		{128, 0, 128, 255},    // Purple (level 8+)
	}
	idx := (level - 1) % len(colors)
	return colors[idx]
}

// Render ghost eyes only (when eaten)
func (r *Renderer) renderGhostEyes(img *image.RGBA, x, y int) {
	pixelSize := r.scale / 2
	if pixelSize < 1 {
		pixelSize = 1
	}

	// Render just the eyes from the ghost sprite
	for py := 0; py < len(sprites.Ghost); py++ {
		for px := 0; px < len(sprites.Ghost[py]); px++ {
			var c color.RGBA
			switch sprites.Ghost[py][px] {
			case 2:
				c = sprites.ColorEyeWhite
			case 3:
				c = sprites.ColorEyeBlue
			default:
				continue // Skip body pixels
			}
			for dy := 0; dy < pixelSize; dy++ {
				for dx := 0; dx < pixelSize; dx++ {
					img.Set(x+px*pixelSize+dx, y+py*pixelSize+dy, c)
				}
			}
		}
	}
}

// Render bonus fruit with pixel art
func (r *Renderer) RenderFruit(img *image.RGBA, fruit *Fruit) {
	if fruit == nil || !fruit.Active {
		return
	}

	tileX := fruit.X * TileSize * r.scale
	tileY := fruit.Y * TileSize * r.scale
	fruitColor := fruit.GetColor()
	stemColor := color.RGBA{139, 69, 19, 255} // Brown stem

	// Simple cherry sprite (8x8 pixels)
	// Pattern: two cherries with stems
	cherryPattern := [][]int{
		{0, 0, 1, 1, 1, 0, 0, 0}, // Stem
		{0, 0, 0, 1, 0, 0, 0, 0},
		{0, 2, 2, 0, 2, 2, 0, 0}, // Two cherries
		{0, 2, 2, 2, 2, 2, 0, 0},
		{0, 2, 2, 2, 2, 2, 0, 0},
		{0, 0, 2, 2, 2, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}

	pixelSize := r.scale / 2
	if pixelSize < 1 {
		pixelSize = 1
	}

	for py := 0; py < 8; py++ {
		for px := 0; px < 8; px++ {
			var c color.RGBA
			switch cherryPattern[py][px] {
			case 1:
				c = stemColor
			case 2:
				c = fruitColor
			default:
				continue
			}
			for dy := 0; dy < pixelSize; dy++ {
				for dx := 0; dx < pixelSize; dx++ {
					img.Set(tileX+px*pixelSize+dx, tileY+py*pixelSize+dy, c)
				}
			}
		}
	}
}

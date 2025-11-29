package game

import (
	"image/color"
	"math"
)

// Pacman entity
type Pacman struct {
	X, Y       int // Tile position
	Dir        Direction
	NextDir    Direction
	AnimFrame  int
	MoveTick   int
	PowerMode  bool
	PowerTicks int
}

func NewPacman() *Pacman {
	return &Pacman{
		X:       14,
		Y:       23,
		Dir:     DirNone,
		NextDir: DirNone,
	}
}

func (p *Pacman) Update(maze *Maze) {
	// Update power mode timer
	if p.PowerMode {
		p.PowerTicks--
		if p.PowerTicks <= 0 {
			p.PowerMode = false
		}
	}

	// Update animation
	p.AnimFrame++

	// Movement with speed control - FASTER during power mode
	speed := MovementSpeed
	if p.PowerMode {
		speed = MovementSpeed - 1 // Faster during power mode
		if speed < 1 {
			speed = 1
		}
	}

	p.MoveTick++
	if p.MoveTick < speed {
		return // Don't move yet
	}
	p.MoveTick = 0

	// Try buffered direction change
	if p.NextDir != DirNone {
		nx, ny := p.getNextPos(p.NextDir)
		if maze.IsWalkable(nx, ny) {
			p.Dir = p.NextDir
			p.NextDir = DirNone
		}
	}

	// Move in current direction
	if p.Dir != DirNone {
		nx, ny := p.getNextPos(p.Dir)

		// Tunnel wraparound
		if nx < 0 {
			nx = maze.Width - 1
		} else if nx >= maze.Width {
			nx = 0
		}

		if maze.IsWalkable(nx, ny) {
			p.X = nx
			p.Y = ny
		}
	}
}

func (p *Pacman) getNextPos(dir Direction) (int, int) {
	nx, ny := p.X, p.Y
	switch dir {
	case DirUp:
		ny--
	case DirDown:
		ny++
	case DirLeft:
		nx--
	case DirRight:
		nx++
	}
	return nx, ny
}

func (p *Pacman) SetDirection(dir Direction) {
	p.NextDir = dir
}

func (p *Pacman) ActivatePowerMode(level int) {
	p.PowerMode = true
	// Power duration decreases with level (like original Pac-Man)
	switch {
	case level == 1:
		p.PowerTicks = 48 // ~6 seconds at 8 FPS
	case level == 2:
		p.PowerTicks = 40 // ~5 seconds
	case level <= 4:
		p.PowerTicks = 32 // ~4 seconds
	case level <= 8:
		p.PowerTicks = 24 // ~3 seconds
	default:
		p.PowerTicks = 16 // ~2 seconds for high levels
	}
}

func (p *Pacman) PowerTimeLeft() int {
	return p.PowerTicks
}

// Ghost entity
type Ghost struct {
	Type         GhostType
	X, Y         int // Tile position
	Dir          Direction
	Mode         GhostMode
	MoveTick     int
	AnimFrame    int
	TargetX      int // Target position when eaten
	TargetY      int
	RespawnTimer int // Immunity after respawning
}

func NewGhost(gtype GhostType, x, y int) *Ghost {
	return &Ghost{
		Type: gtype,
		X:    x,
		Y:    y,
		Dir:  DirLeft,
		Mode: ModeScatter,
	}
}

func (g *Ghost) Update(maze *Maze, pacman *Pacman, level int) {
	g.AnimFrame++

	// Handle eaten mode - return to ghost house
	if g.Mode == ModeEaten {
		// Check if close to target entrance (within 3 tiles - more forgiving)
		dx := g.X - g.TargetX
		if dx < 0 {
			dx = -dx
		}
		dy := g.Y - g.TargetY
		if dy < 0 {
			dy = -dy
		}

		// Respawn when close to ghost house entrance
		if dx <= 2 && dy <= 2 {
			// Teleport inside ghost house and respawn
			g.X = 14
			g.Y = 14
			g.Mode = ModeScatter
			g.RespawnTimer = 16 // ~2 second immunity at 8 FPS
			return
		}
		// Eyes move faster
		g.MoveTick++
		if g.MoveTick < 1 { // Very fast
			return
		}
		g.MoveTick = 0
		g.Dir = g.chooseReturnDirection(maze, g.TargetX, g.TargetY)
		nx, ny := g.getNextPos(g.Dir)
		if maze.IsWalkableForGhost(nx, ny) {
			g.X = nx
			g.Y = ny
		}
		return
	}

	// Decrement respawn immunity timer
	if g.RespawnTimer > 0 {
		g.RespawnTimer--
	}

	// Update mode based on pacman power mode (but not if eaten)
	// Note: RespawnTimer doesn't prevent becoming frightened, only prevents immediate re-frightening
	if pacman.PowerMode && g.Mode != ModeFrightened && g.Mode != ModeEaten {
		g.Mode = ModeFrightened
	} else if !pacman.PowerMode && g.Mode == ModeFrightened {
		g.Mode = ModeChase
	}

	// Movement with speed control - SLOWER when frightened, FASTER at higher levels
	speed := BaseGhostSpeed
	if level > 1 {
		speed = speed - (level-1)/3 // Get faster every 3 levels
		if speed < 1 {
			speed = 1
		}
	}
	if g.Mode == ModeFrightened {
		speed = speed + 2 // Slower when frightened
	}

	g.MoveTick++
	if g.MoveTick < speed {
		return
	}
	g.MoveTick = 0

	// Choose direction with simple AI
	g.Dir = g.chooseDirection(maze, pacman)

	// Move
	nx, ny := g.getNextPos(g.Dir)

	// Tunnel wraparound
	if nx < 0 {
		nx = maze.Width - 1
	} else if nx >= maze.Width {
		nx = 0
	}

	if maze.IsWalkableForGhost(nx, ny) {
		g.X = nx
		g.Y = ny
	}
}

func (g *Ghost) chooseDirection(maze *Maze, pacman *Pacman) Direction {
	var targetX, targetY int

	if g.Mode == ModeFrightened {
		// Run away from pacman
		targetX = g.X*2 - pacman.X
		targetY = g.Y*2 - pacman.Y
	} else {
		// Chase pacman with ghost-specific behavior
		targetX = pacman.X
		targetY = pacman.Y

		switch g.Type {
		case GhostPinky:
			// Target ahead of pacman
			switch pacman.Dir {
			case DirUp:
				targetY -= 4
			case DirDown:
				targetY += 4
			case DirLeft:
				targetX -= 4
			case DirRight:
				targetX += 4
			}
		case GhostInky:
			// Flanking behavior
			targetX = 2*pacman.X - g.X
			targetY = 2*pacman.Y - g.Y
		case GhostClyde:
			// Random when close
			dist := math.Abs(float64(g.X-pacman.X)) + math.Abs(float64(g.Y-pacman.Y))
			if dist < 8 {
				targetX = 0
				targetY = maze.Height - 1
			}
		}
	}

	// Find best direction (prefer not to reverse, but allow it if stuck)
	bestDir := DirNone
	bestDist := math.MaxFloat64
	reverseDir := DirNone
	reverseDist := math.MaxFloat64

	for _, dir := range []Direction{DirUp, DirDown, DirLeft, DirRight} {
		nx, ny := g.X, g.Y
		switch dir {
		case DirUp:
			ny--
		case DirDown:
			ny++
		case DirLeft:
			nx--
		case DirRight:
			nx++
		}

		if !maze.IsWalkableForGhost(nx, ny) {
			continue
		}

		dist := math.Abs(float64(nx-targetX)) + math.Abs(float64(ny-targetY))

		if dir == oppositeDir(g.Dir) {
			// Save reverse as backup
			if dist < reverseDist {
				reverseDist = dist
				reverseDir = dir
			}
		} else {
			// Prefer non-reverse directions
			if dist < bestDist {
				bestDist = dist
				bestDir = dir
			}
		}
	}

	// If no valid direction found, use reverse (better than getting stuck)
	if bestDir == DirNone {
		if reverseDir != DirNone {
			return reverseDir
		}
		// Complete stuck - keep current direction
		return g.Dir
	}

	return bestDir
}

func (g *Ghost) getNextPos(dir Direction) (int, int) {
	nx, ny := g.X, g.Y
	switch dir {
	case DirUp:
		ny--
	case DirDown:
		ny++
	case DirLeft:
		nx--
	case DirRight:
		nx++
	}
	return nx, ny
}

func (g *Ghost) GetColor() color.RGBA {
	switch g.Type {
	case GhostBlinky:
		return ColorBlinkyRed
	case GhostPinky:
		return ColorPinkyPink
	case GhostInky:
		return ColorInkyCyan
	case GhostClyde:
		return ColorClydeOrange
	}
	return ColorWhite
}

// BFS pathfinding for returning to ghost house - finds actual shortest path
func (g *Ghost) chooseReturnDirection(maze *Maze, targetX, targetY int) Direction {
	// BFS to find shortest path
	type Node struct {
		x, y     int
		firstDir Direction // First direction taken from start
	}

	queue := []Node{{g.X, g.Y, DirNone}}
	visited := make(map[[2]int]bool)
	visited[[2]int{g.X, g.Y}] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		// Check if we reached target (or very close)
		if curr.x == targetX && curr.y == targetY {
			if curr.firstDir != DirNone {
				return curr.firstDir
			}
			break // Already at target
		}

		// Explore all 4 directions
		for _, dir := range []Direction{DirUp, DirDown, DirLeft, DirRight} {
			nx, ny := curr.x, curr.y
			switch dir {
			case DirUp:
				ny--
			case DirDown:
				ny++
			case DirLeft:
				nx--
			case DirRight:
				nx++
			}

			// Handle tunnel wraparound
			if nx < 0 {
				nx = maze.Width - 1
			} else if nx >= maze.Width {
				nx = 0
			}

			key := [2]int{nx, ny}
			if visited[key] || !maze.IsWalkableForGhost(nx, ny) {
				continue
			}

			visited[key] = true
			firstDir := curr.firstDir
			if firstDir == DirNone {
				firstDir = dir // Record first move from start
			}
			queue = append(queue, Node{nx, ny, firstDir})
		}
	}

	// No path found or at target - try any valid direction
	for _, dir := range []Direction{DirUp, DirDown, DirLeft, DirRight} {
		nx, ny := g.X, g.Y
		switch dir {
		case DirUp:
			ny--
		case DirDown:
			ny++
		case DirLeft:
			nx--
		case DirRight:
			nx++
		}
		if maze.IsWalkableForGhost(nx, ny) {
			return dir
		}
	}

	return g.Dir // Keep current direction if stuck
}

func oppositeDir(dir Direction) Direction {
	switch dir {
	case DirUp:
		return DirDown
	case DirDown:
		return DirUp
	case DirLeft:
		return DirRight
	case DirRight:
		return DirLeft
	}
	return DirNone
}

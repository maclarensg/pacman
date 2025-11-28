package game

// Maze layout (28x31 tiles) - COMPLETE Pac-Man maze with ALL pellets
// W=wall, .=pellet, O=power pellet, _=empty, G=ghost door
var mazeLayout = []string{
	"WWWWWWWWWWWWWWWWWWWWWWWWWWWW",
	"W............WW............W",
	"W.WWWW.WWWWW.WW.WWWWW.WWWW.W",
	"WOWWWW.WWWWW.WW.WWWWW.WWWWOW",
	"W.WWWW.WWWWW.WW.WWWWW.WWWW.W",
	"W..........................W",
	"W.WWWW.WW.WWWWWWWW.WW.WWWW.W",
	"W.WWWW.WW.WWWWWWWW.WW.WWWW.W",
	"W......WW....WW....WW......W",
	"WWWWWW.WWWWW.WW.WWWWW.WWWWWW",
	"WWWWWW.WWWWW.WW.WWWWW.WWWWWW",
	"WWWWWW.WW..........WW.WWWWWW",
	"WWWWWW.WW.WWWGGWWW.WW.WWWWWW",
	"WWWWWW.WW.W______W.WW.WWWWWW",
	"..........W______W..........",
	"WWWWWW.WW.W______W.WW.WWWWWW",
	"WWWWWW.WW.WWWWWWWW.WW.WWWWWW",
	"WWWWWW.WW..........WW.WWWWWW",
	"WWWWWW.WW.WWWWWWWW.WW.WWWWWW",
	"WWWWWW.WW.WWWWWWWW.WW.WWWWWW",
	"W............WW............W",
	"W.WWWW.WWWWW.WW.WWWWW.WWWW.W",
	"W.WWWW.WWWWW.WW.WWWWW.WWWW.W",
	"WO..WW................WW..OW",
	"WWW.WW.WW.WWWWWWWW.WW.WW.WWW",
	"WWW.WW.WW.WWWWWWWW.WW.WW.WWW",
	"W......WW....WW....WW......W",
	"W.WWWWWWWWWW.WW.WWWWWWWWWW.W",
	"W.WWWWWWWWWW.WW.WWWWWWWWWW.W",
	"W..........................W",
	"WWWWWWWWWWWWWWWWWWWWWWWWWWWW",
}

type CellType byte

const (
	CellWall        CellType = 'W'
	CellPellet      CellType = '.'
	CellPowerPellet CellType = 'O'
	CellEmpty       CellType = '_'
	CellGhostDoor   CellType = 'G'
)

type Maze struct {
	Width            int
	Height           int
	Cells            [][]CellType
	TotalPellets     int
	RemainingPellets int
}

func NewMaze() *Maze {
	height := len(mazeLayout)
	width := len(mazeLayout[0])

	m := &Maze{
		Width:  width,
		Height: height,
		Cells:  make([][]CellType, height),
	}

	pelletCount := 0
	for y := 0; y < height; y++ {
		m.Cells[y] = make([]CellType, width)
		for x := 0; x < width; x++ {
			cell := CellType(mazeLayout[y][x])
			m.Cells[y][x] = cell
			if cell == CellPellet || cell == CellPowerPellet {
				pelletCount++
			}
		}
	}

	m.TotalPellets = pelletCount
	m.RemainingPellets = pelletCount

	return m
}

func (m *Maze) GetCell(x, y int) CellType {
	if y < 0 || y >= m.Height || x < 0 || x >= m.Width {
		return CellWall
	}
	return m.Cells[y][x]
}

func (m *Maze) SetCell(x, y int, cell CellType) {
	if y >= 0 && y < m.Height && x >= 0 && x < m.Width {
		m.Cells[y][x] = cell
	}
}

func (m *Maze) IsWalkable(x, y int) bool {
	cell := m.GetCell(x, y)
	return cell != CellWall && cell != CellGhostDoor
}

func (m *Maze) IsWalkableForGhost(x, y int) bool {
	cell := m.GetCell(x, y)
	return cell != CellWall
}

func (m *Maze) EatPellet(x, y int) (isPowerPellet bool, ate bool) {
	cell := m.GetCell(x, y)
	if cell == CellPellet {
		m.SetCell(x, y, CellEmpty)
		m.RemainingPellets--
		return false, true
	} else if cell == CellPowerPellet {
		m.SetCell(x, y, CellEmpty)
		m.RemainingPellets--
		return true, true
	}
	return false, false
}

func (m *Maze) Reset() {
	*m = *NewMaze()
}

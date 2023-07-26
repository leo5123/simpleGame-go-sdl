package game

import (
	"bufio"
	"fmt"
	"os"
)

type Tile rune

type GameUI interface {
	Draw(*Level)
	GetInput() *Input
}

const (
	GrassCliff Tile = '#'
	Grass      Tile = '.'
	ClosedDoor Tile = 'X'
	OpenDoor   Tile = 'x'
	Blank      Tile = 0
	Pending    Tile = -1
	Test       Tile = 't'
)

type InputType int

const (
	None InputType = iota
	Up
	Down
	Left
	Right
	Quit
	Action
)

type Input struct {
	Typ InputType
}

type Entity struct {
	X, Y int
}

type Player struct {
	Entity
}

type Level struct {
	Map [][]Tile
	Player
}

func loadLevelFromFile(fileName string) *Level {
	levelLines := make([]string, 0)

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err, "Error reading file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	longestRow := 0
	index := 0
	for scanner.Scan() {
		levelLines = append(levelLines, scanner.Text())
		if len(levelLines[index]) > longestRow {
			longestRow = len(levelLines[index])
		}
		index++
	}
	level := &Level{}
	level.Map = make([][]Tile, len(levelLines))
	for i := range level.Map {
		level.Map[i] = make([]Tile, longestRow)
	}

	for y := 0; y < len(level.Map); y++ {
		line := levelLines[y]
		for x, c := range line {
			var t Tile
			switch c {
			case ' ', '\t', '\n', '\r':
				t = Blank
			case '#':
				t = GrassCliff
			case '.':
				t = Grass
			case 'X':
				t = ClosedDoor
			case 'x':
				t = OpenDoor
			case 'P':
				level.Player.X = x
				level.Player.Y = y
				t = Pending
			case 't':
				t = Test
			default:
				panic("Invalid character in map")
			}
			level.Map[y][x] = t
		}
	}

	for y, row := range level.Map {
		for x, tile := range row {
			if tile == Pending {
			SearchLoop:
				for searchX := x - 1; searchX < x+1; searchX++ {
					for searchY := y - 1; searchY <= y+1; searchY++ {
						searchTile := level.Map[searchY][searchX]
						switch searchTile {
						case Grass:
							level.Map[y][x] = Grass
							break SearchLoop
						}
					}
				}
			}
		}
	}

	return level
}

func canWalk(level *Level, x, y int) bool {
	t := level.Map[x][y]
	switch t {
	case GrassCliff, ClosedDoor, Blank:
		return false
	default:
		return true
	}
}

func isDoor(level *Level, x, y int, input InputType) {
	t := level.Map[x][y]
	if t == OpenDoor {
		level.Map[x][y] = ClosedDoor
		return
	}
	if t == ClosedDoor {
		level.Map[x][y] = OpenDoor
		return
	}
}

func handleInput(level *Level, input *Input) {
	p := level.Player
	switch input.Typ {
	case Up:
		if canWalk(level, p.Y-1, p.X) {
			level.Player.Y--
		}
	case Down:
		if canWalk(level, p.Y+1, p.X) {
			level.Player.Y++
		}
	case Left:
		if canWalk(level, p.Y, p.X-1) {
			level.Player.X--
		}
	case Right:
		if canWalk(level, p.Y, p.X+1) {
			level.Player.X++
		}
	case Action:
		isDoor(level, p.Y, p.X+1, Action)
		isDoor(level, p.Y, p.X-1, Action)
		isDoor(level, p.Y+1, p.X, Action)
		isDoor(level, p.Y-1, p.X, Action)
	}
}

func Run(ui GameUI) {
	level := loadLevelFromFile("game/maps/level1.map")
	for {
		ui.Draw(level)
		input := ui.GetInput()

		if input != nil && input.Typ == Quit {
			return
		}

		handleInput(level, input)

	}
}

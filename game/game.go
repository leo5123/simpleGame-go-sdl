package game

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"time"
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
	Search //
)

type Input struct {
	Typ InputType
}

type Pos struct {
	X, Y int
}

type Entity struct {
	Pos
}

type Player struct {
	Entity
}

type Level struct {
	Map [][]Tile
	Player
	Debug map[Pos]bool
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
				for searchX := x - 1; searchX <= x+1; searchX++ {
					for searchY := y - 1; searchY <= y+1; searchY++ {
						searchTile := level.Map[searchY][searchY]
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

func canWalk(level *Level, pos Pos) bool {
	t := level.Map[pos.Y][pos.X]
	switch t {
	case GrassCliff, ClosedDoor, Blank:
		return false
	default:
		return true
	}
}

func isDoor(level *Level, pos Pos, input InputType) {
	t := level.Map[pos.Y][pos.X]
	if t == OpenDoor {
		level.Map[pos.Y][pos.X] = ClosedDoor
		return
	}
	if t == ClosedDoor {
		level.Map[pos.Y][pos.X] = OpenDoor
		return
	}
}

func handleInput(ui GameUI, level *Level, input *Input) {
	p := level.Player
	switch input.Typ {
	case Up:
		if canWalk(level, Pos{Y: p.Y - 1, X: p.X}) {
			level.Player.Y--
		}
	case Down:
		if canWalk(level, Pos{Y: p.Y + 1, X: p.X}) {
			level.Player.Y++
		}
	case Left:
		if canWalk(level, Pos{Y: p.Y, X: p.X - 1}) {
			level.Player.X--
		}
	case Right:
		if canWalk(level, Pos{Y: p.Y, X: p.X + 1}) {
			level.Player.X++
		}
	case Search:
		// bfs(ui, level, level.Player.Pos)
		astar(ui, level, level.Player.Pos, Pos{X: 66, Y: 7})
		fmt.Println(level.Player.X, level.Player.Y)
	case Action:
		isDoor(level, Pos{Y: p.Y, X: p.X + 1}, Action)
		isDoor(level, Pos{Y: p.Y, X: p.X - 1}, Action)
		isDoor(level, Pos{Y: p.Y + 1, X: p.X}, Action)
		isDoor(level, Pos{Y: p.Y - 1, X: p.X}, Action)
	}
}

func getNeighbors(level *Level, pos Pos) []Pos {
	neighbors := make([]Pos, 0, 4)

	left := Pos{X: pos.X - 1, Y: pos.Y}
	right := Pos{X: pos.X + 1, Y: pos.Y}
	down := Pos{X: pos.X, Y: pos.Y - 1}
	up := Pos{X: pos.X, Y: pos.Y + 1}

	if canWalk(level, right) {
		neighbors = append(neighbors, right)
	}
	if canWalk(level, left) {
		neighbors = append(neighbors, left)
	}
	if canWalk(level, down) {
		neighbors = append(neighbors, down)
	}
	if canWalk(level, up) {
		neighbors = append(neighbors, up)
	}

	return neighbors
}

func bfs(ui GameUI, level *Level, start Pos) {
	frontier := make([]Pos, 0, 8)
	frontier = append(frontier, start)
	visited := make(map[Pos]bool)
	visited[start] = true
	level.Debug = visited

	for len(frontier) > 0 {
		current := frontier[0]
		frontier = frontier[1:]
		for _, next := range getNeighbors(level, current) {
			if !visited[next] {
				frontier = append(frontier, next)
				visited[next] = true
				ui.Draw(level)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func astar(ui GameUI, level *Level, start Pos, goal Pos) []Pos {
	frontier := make(pqueue, 0, 4)
	frontier = frontier.push(start, 1)
	cameFrom := make(map[Pos]Pos)
	costSoFar := make(map[Pos]int, 0)
	costSoFar[start] = 0

	level.Debug = make(map[Pos]bool)

	var current Pos

	for len(frontier) > 0 {
		frontier, current = frontier.pop()

		if current == goal {
			path := make([]Pos, 0)
			p := current
			for p != start {
				path = append(path, p)
				p = cameFrom[p]
			}
			path = append(path, p)
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			for _, pos := range path {
				level.Debug[pos] = true
				ui.Draw(level)
				time.Sleep(100 * time.Millisecond)
			}
			return path
		}
		for _, next := range getNeighbors(level, current) {
			newCost := costSoFar[current] + 1
			_, exists := costSoFar[next]
			if !exists || newCost < costSoFar[next] {
				costSoFar[next] = newCost
				xDist := int(math.Abs(float64(goal.X - next.X)))
				yDist := int(math.Abs(float64(goal.X - next.X)))
				priority := newCost + xDist + yDist
				frontier = frontier.push(next, priority)
				cameFrom[next] = current
			}
		}
	}

	return nil
}

func Run(ui GameUI) {
	level := loadLevelFromFile("game/maps/level1.map")
	for {
		ui.Draw(level)
		input := ui.GetInput()

		if input != nil && input.Typ == Quit {
			return
		}

		handleInput(ui, level, input)
	}
}

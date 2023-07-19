package ui2d

import (
	"bufio"
	"fmt"
	"image/png"
	"math/rand"
	"os"
	"rpg/game"
	"strconv"
	"strings"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type UI2d struct {
}

const winWidth, winHeight = 1280, 720

var (
	renderer          *sdl.Renderer
	textureAtlas      *sdl.Texture
	textureIndex      map[game.Tile][]sdl.Rect
	keyboardState     []uint8
	prevKeyboardState []uint8
)

func loadTextureIndex() {
	textureIndex = make(map[game.Tile][]sdl.Rect)
	infile, err := os.Open("ui2d/assets/atlas-index.txt")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := game.Tile(line[0])
		xy := line[1:]
		xy = strings.TrimSpace(xy)
		splitXYC := strings.Split(xy, ",")
		x, err := strconv.ParseInt(splitXYC[0], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(splitXYC[1], 10, 64)
		if err != nil {
			panic(err)
		}
		variationCount, err := strconv.ParseInt(strings.TrimSpace(splitXYC[2]), 10, 64)

		if err != nil {
			panic(err)
		}

		var rects []sdl.Rect
		for i := 0; i < int(variationCount); i++ {
			rects = append(rects, sdl.Rect{
				X: int32(x * 32),
				Y: int32(y * 32),
				W: 32,
				H: 32,
			},
			)
			x++
			if x > 30 {
				x = 0
				y++
			}
		}
		//rectangle for tile rune
		textureIndex[tileRune] = rects

	}
}

func imgFileToTexture(filename string) *sdl.Texture {
	infile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}

	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	pixels := make([]byte, w*h*4)
	bIndex := 0

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[bIndex] = byte(r / 256)
			bIndex++
			pixels[bIndex] = byte(g / 256)
			bIndex++
			pixels[bIndex] = byte(b / 256)
			bIndex++
			pixels[bIndex] = byte(a / 256)
			bIndex++
		}
	}

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, unsafe.Pointer(&pixels[100000]), w*4)

	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	return tex
}

func init() {
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err, "SDL ERROR")
		return
	}

	window, err := sdl.CreateWindow("RPG", 200, 200,
		int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err, "SDL ERROR")
		return
	}

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err, "SDL ERROR")
		return
	}

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	textureAtlas = imgFileToTexture("ui2d/assets/gfx/32x32_map_tile.png")
	loadTextureIndex()

	keyboardState = sdl.GetKeyboardState()
	prevKeyboardState = make([]uint8, len(keyboardState))
	for i, v := range keyboardState {
		prevKeyboardState[i] = v
	}
}

func (ui *UI2d) Draw(level *game.Level) {
	rand.Seed(1)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile != game.Blank {
				srcRects := textureIndex[tile]
				srcRect := srcRects[rand.Intn(len(srcRects))]
				dstRect := sdl.Rect{int32(x * 32), int32(y * 32), 32, 32}
				renderer.Copy(textureAtlas, &srcRect, &dstRect)
			}
		}
	}
	renderer.Copy(textureAtlas, &sdl.Rect{
		X: 72 * 32,
		Y: 34 * 32,
		W: 32,
		H: 32,
	}, &sdl.Rect{
		X: int32(level.Player.X) * 32,
		Y: int32(level.Player.Y) * 32,
		W: 32,
		H: 32,
	})
	renderer.Present()
}

func (ui *UI2d) GetInput() *game.Input {
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return &game.Input{game.Quit}

			}
		}
		var input game.Input
		if keyboardState[sdl.SCANCODE_UP] == 0 && prevKeyboardState[sdl.SCANCODE_UP] != 0 {
			input.Typ = game.Up
		}
		if keyboardState[sdl.SCANCODE_DOWN] == 0 && prevKeyboardState[sdl.SCANCODE_DOWN] != 0 {
			input.Typ = game.Down
		}
		if keyboardState[sdl.SCANCODE_LEFT] == 0 && prevKeyboardState[sdl.SCANCODE_LEFT] != 0 {
			input.Typ = game.Left
		}
		if keyboardState[sdl.SCANCODE_RIGHT] == 0 && prevKeyboardState[sdl.SCANCODE_RIGHT] != 0 {
			input.Typ = game.Right
		}
		for i, v := range keyboardState {
			prevKeyboardState[i] = v
		}
		if input.Typ != game.None {
			return &input
		}
	}
}

package main

import (
	"fmt"
	"sort"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const SCREEN_WIDTH = 800
const SCREEN_HEIGHT = 700

const MAP_WIDTH = 16
const MAP_HEIGHT = 16

const MAP_SCREEN_WIDTH = 160
const MAP_SCREEN_HEIGHT = 160

const TEX_WIDTH = 64
const TEX_HEIGHT = 64

const FLOOR_TEX_WIDTH = 64
const FLOOR_TEX_HEIGHT = 64

const PIXEL_SIZE = 2
const SPRITE_COUNT = 2

var MAP [MAP_WIDTH][MAP_HEIGHT]int = [MAP_WIDTH][MAP_HEIGHT]int{
	{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4},
	{4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},
	{4, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},
	{4, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},
	{4, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},
	{4, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},
	{4, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},
	{4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},

	{4, 0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 0, 0, 0, 0, 4},
	{4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},
	{4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},
	{4, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 2, 2, 0, 0, 4},
	{4, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 2, 2, 0, 0, 4},
	{4, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 4},
	{4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},
	{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4},
}

func assert(condition bool, msg string) {
	if !condition {
		panic(msg)
	}
}

type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(that Vec2) Vec2 {
	return Vec2{v.X + that.X, v.Y + that.Y}
}

func (v Vec2) Sub(that Vec2) Vec2 {
	return Vec2{v.X - that.X, v.Y - that.Y}
}

func (v Vec2) Mul(n float64) Vec2 {
	return Vec2{v.X * n, v.Y * n}
}

type Sprite struct {
	Tex rl.Texture2D
	Pos Vec2
}
var sprites [SPRITE_COUNT]Sprite
var zDistance [SCREEN_WIDTH]float64

var Pos = Vec2{2.6, 4.8}
var Dir = Vec2{1, 0}
var Plane = Vec2{0, 0.66}

var screenBuffer rl.RenderTexture2D

func RenderFloor() {
	var y float64
	for y = SCREEN_HEIGHT / 2; y < SCREEN_HEIGHT; y++ {
		var floorBuffer []rl.Color = make([]rl.Color, SCREEN_WIDTH)
		// var ceilBuffer []rl.Color = make([]rl.Color, SCREEN_WIDTH)
		rayBegin := Dir.Sub(Plane)
		rayEnd := Dir.Add(Plane)

		p := y - SCREEN_HEIGHT/2
		posZ := float64(0.5 * SCREEN_HEIGHT)
		rowDistance := posZ / p
		stepSizeX := rowDistance * (rayEnd.X - rayBegin.X) / SCREEN_WIDTH
		stepSizeY := rowDistance * (rayEnd.Y - rayBegin.Y) / SCREEN_WIDTH

		wallX := Pos.X + rowDistance*rayBegin.X
		wallY := Pos.Y + rowDistance*rayBegin.Y

		for x := range SCREEN_WIDTH {
			cellX := math.Floor(wallX)
			cellY := math.Floor(wallY)

			texX := int32(FLOOR_TEX_WIDTH*(wallX-cellX)) & (FLOOR_TEX_WIDTH - 1)
			texY := int32(FLOOR_TEX_HEIGHT*(wallY-cellY)) & (FLOOR_TEX_HEIGHT - 1)

			wallX += stepSizeX
			wallY += stepSizeY

			floorColor := floorTexture[int(texY)*int(FLOOR_TEX_WIDTH)+int(texX)]
			floorBuffer[x] = floorColor
		}
		rl.UpdateTextureRec(
			screenBuffer.Texture, 
			rl.NewRectangle(0, float32(SCREEN_HEIGHT/2 - (y - SCREEN_HEIGHT/2)), // Y is flipped in screenBuffer
			SCREEN_WIDTH, 1), 
			floorBuffer,
		)
	}
}

func RenderWall() {
	var x float64
	for x = 0; x < SCREEN_WIDTH; x++ {
		// var wallBuffer []rl.Color = make([]rl.Color, SCREEN_HEIGHT)
		cameraX := ((2 * x) / float64(SCREEN_WIDTH)) - 1 // cameraX : -1...0...1
		rayDir := Dir.Add(Plane.Mul(cameraX))

		mapX := int32(Pos.X)
		mapY := int32(Pos.Y)

		var deltaDist Vec2

		if rayDir.X == 0 {
			deltaDist.X = math.Inf(1)
		} else {
			deltaDist.X = math.Abs(1 / rayDir.X)
		}
		if rayDir.Y == 0 {
			deltaDist.Y = math.Inf(1)
		} else {
			deltaDist.Y = math.Abs(1 / rayDir.Y)
		}

		var sideDist Vec2
		var stepX int32
		var stepY int32

		if rayDir.X > 0 {
			stepX = 1
			sideDist.X = (float64(mapX+1) - Pos.X) * deltaDist.X
		} else {
			stepX = -1
			sideDist.X = (Pos.X - float64(mapX)) * deltaDist.X
		}
		if rayDir.Y > 0 {
			stepY = 1
			sideDist.Y = (float64(mapY+1) - Pos.Y) * deltaDist.Y
		} else {
			stepY = -1
			sideDist.Y = (Pos.Y - float64(mapY)) * deltaDist.Y
		}

		var side int8 = -1
		var hittedWall int

		for true {
			if sideDist.X < sideDist.Y {
				sideDist.X += deltaDist.X
				mapX += stepX
				side = 0
			} else {
				drawRay(rayDir.Mul(sideDist.Y))
				sideDist.Y += deltaDist.Y
				mapY += stepY
				side = 1
			}
			if MAP[mapY][mapX] > 0 {
				hittedWall = MAP[mapY][mapX]
				break
			}
		}

		var perpWallDist float64
		if side == 0 {
			perpWallDist = sideDist.X - deltaDist.X
		} else {
			perpWallDist = sideDist.Y - deltaDist.Y
		}

		var factor float64 = 1.25
		if perpWallDist > 0 {
			res := 1.75 / perpWallDist
			if res < factor {
				factor = res
			}
		}

		wallHeight := SCREEN_HEIGHT / perpWallDist
		wallStart := (-wallHeight / 2) + SCREEN_HEIGHT/2
		wallEnd := (wallHeight / 2) + SCREEN_HEIGHT/2

		if wallStart < 0 {
			wallStart = 0
		}
		if wallEnd >= SCREEN_HEIGHT {
			wallEnd = SCREEN_HEIGHT - 1
		}

		var wallX float64
		if side == 0 {
			wallX = Pos.Y + perpWallDist*rayDir.Y
		} else {
			wallX = Pos.X + perpWallDist*rayDir.X
		}
		wallX -= math.Floor(wallX)
		texX := int(wallX * float64(TEX_WIDTH))

		step := 1.0 * TEX_HEIGHT / wallHeight
		texPos := (wallStart - float64(SCREEN_HEIGHT)/2 + wallHeight/2) * step
		selectedTex := textures[hittedWall]
		texY := int(texPos) & (TEX_HEIGHT - 1)

		tintVal := 750 / perpWallDist
		tintVal = math.Max(tintVal, 50)
		tintVal = math.Min(tintVal, 255)
		tint := rl.NewColor(uint8(tintVal), uint8(tintVal), uint8(tintVal), 255)

		rl.DrawTexturePro(
			selectedTex,
			rl.NewRectangle(float32(texX), float32(texY), 1, TEX_HEIGHT),
			rl.NewRectangle(float32(x), float32(wallStart), 1, float32(wallHeight)),
			rl.Vector2{X: 0, Y: 0},
			0.0,
			tint,
		)
		drawRay(rayDir)
	}

}

func RenderScene() {
	RenderFloor()
	RenderWall()
}

func mapCoordToPixel(vec Vec2) Vec2 {
	ratioWidth := float64(MAP_SCREEN_WIDTH) / float64(MAP_WIDTH)
	ratioHeight := float64(MAP_SCREEN_HEIGHT) / float64(MAP_HEIGHT)
	return Vec2{vec.X * ratioWidth, vec.Y * ratioHeight}
}

func drawRay(rayDir Vec2) {
	rayEnd := Pos.Add(rayDir)
	rayEndPixel := mapCoordToPixel(rayEnd)

	posPixel := mapCoordToPixel(Pos)
	rl.DrawLine(
		int32(posPixel.X),
		int32(posPixel.Y),
		int32(rayEndPixel.X),
		int32(rayEndPixel.Y),
		rl.Red,
	)
}

func DrawMap() {
	w := float64(MAP_SCREEN_WIDTH / MAP_WIDTH)
	h := float64(MAP_SCREEN_HEIGHT / MAP_HEIGHT)

	for y := range int32(MAP_HEIGHT) {
		for x := range int32(MAP_WIDTH) {
			var color rl.Color
			if MAP[y][x] > 0 {
				color = rl.Gray
			} else {
				color = rl.NewColor(245, 245, 254, 255)
			}
			rl.DrawRectangle(x*int32(w), y*int32(h), int32(w), int32(h), color)
			rl.DrawRectangleLines(x*int32(w), y*int32(h), int32(w), int32(h), rl.Gray)
		}
	}

	px := mapCoordToPixel(Pos)
	rl.DrawCircle(int32(px.X), int32(px.Y), 4, rl.Magenta)
	drawRay(Dir)
}

func DrawFPS() {
	text := fmt.Sprintf("FPS: %d", rl.GetFPS())
	fontSize := int32(18)
	width := rl.MeasureText(text, fontSize)
	xPos := int32(SCREEN_WIDTH - width)
	yPos := int32(0)
	rl.DrawText(text, xPos, yPos, fontSize, rl.White)
}

func RotateRight(angle float64) {
	{
		x := Dir.X*math.Cos(angle) - Dir.Y*math.Sin(angle)
		y := Dir.X*math.Sin(angle) + Dir.Y*math.Cos(angle)
		Dir.X = x
		Dir.Y = y
	}
	// Camera Plane
	{
		x := Plane.X*math.Cos(angle) - Plane.Y*math.Sin(angle)
		y := Plane.X*math.Sin(angle) + Plane.Y*math.Cos(angle)
		Plane.X = x
		Plane.Y = y
	}
}

func RotateLeft(angle float64) {
	{
		x := Dir.X*math.Cos(angle) + Dir.Y*math.Sin(angle)
		y := -Dir.X*math.Sin(angle) + Dir.Y*math.Cos(angle)
		Dir.X = x
		Dir.Y = y
	}
	// Camera Plane
	{
		x := Plane.X*math.Cos(angle) + Plane.Y*math.Sin(angle)
		y := -Plane.X*math.Sin(angle) + Plane.Y*math.Cos(angle)
		Plane.X = x
		Plane.Y = y
	}
}

func MoveForward(speed float64) {
	newPos := Pos.Add(Dir.Mul(speed))
	newX := newPos.X
	newY := newPos.Y
	if (newX < 0 || newY < 0) ||
		(newX >= MAP_WIDTH || newY >= MAP_HEIGHT) ||
		(MAP[int32(newY)][int32(newX)]) > 0 {
		return
	}
	Pos = newPos
}

func MoveBackward(speed float64) {
	newPos := Pos.Sub(Dir.Mul(speed))
	newX := newPos.X
	newY := newPos.Y
	if (newX < 0 || newY < 0) ||
		(newX >= MAP_WIDTH || newY >= MAP_HEIGHT) ||
		(MAP[int32(newY)][int32(newX)]) > 0 {
		return
	}
	Pos = newPos
}

var xDir int32 = 0
var yDir int32 = 0

func ListenKeyDown() {
	frameTime := rl.GetFrameTime()
	speed := float64(frameTime * 2) // 2 unit per sec
	rad := float64(frameTime * 3)   // 3 rad per sec

	if rl.IsKeyPressed(rl.KeyRight) {
		xDir = 1
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		xDir = -1
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		yDir = -1
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		yDir = 1
	}

	if rl.IsKeyDown(rl.KeyRight) && xDir == 1 {
		RotateRight(rad)
	}
	if rl.IsKeyDown(rl.KeyLeft) && xDir == -1 {
		RotateLeft(rad)
	}
	if rl.IsKeyDown(rl.KeyDown) && yDir == 1 {
		MoveBackward(speed)
	}
	if rl.IsKeyDown(rl.KeyUp) && yDir == -1 {
		MoveForward(speed)
	}
}

type TexType int

const (
	TEX_1 TexType = iota + 1
	TEX_2
	TEX_3
	TEX_4
	TEX_5
)
const TEX_COUNT = 20

var floorTexture []color.RGBA
var bgTexture rl.Texture2D

var textures []rl.Texture2D = make([]rl.Texture2D, TEX_COUNT)

func InitTexture() {
	wall1 := rl.LoadTexture("./assets/aot/wall1.png")
	wall2 := rl.LoadTexture("./assets/aot/wall2.png")
	wall3 := rl.LoadTexture("./assets/aot/wall3.png")
	wall4 := rl.LoadTexture("./assets/aot/wall4.png")

	textures[TEX_1] = wall1
	textures[TEX_2] = wall2
	textures[TEX_3] = wall3
	textures[TEX_4] = wall4

	floorTexture = rl.LoadImageColors(rl.LoadImageFromTexture(rl.LoadTexture("./assets/aot/floor.png")))
	// ceilTexture = rl.LoadImageColors(rl.LoadImageFromTexture(wood))
	bgTexture = rl.LoadTexture("./assets/aot/bg.png")
	screenBuffer = rl.LoadRenderTexture(SCREEN_WIDTH, SCREEN_HEIGHT)
}

func InitSprite() {
	barrel := rl.LoadTexture("./assets/sprite/barrel.png")
	pillar := rl.LoadTexture("./assets/sprite/pillar.png")

	sprites[0] = Sprite{Pos: Vec2{0, 1}, Tex: barrel}
	sprites[1] = Sprite{Pos: Vec2{0, 1}, Tex: pillar}
}

func main() {
	rl.InitWindow(SCREEN_WIDTH, SCREEN_HEIGHT, "Raycasting From Scratch")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)
	rl.SetWindowSize(SCREEN_WIDTH, SCREEN_HEIGHT)

	InitTexture()
	InitSprite()

	for !rl.WindowShouldClose() {
		ListenKeyDown()

		rl.BeginTextureMode(screenBuffer)
			rl.ClearBackground(rl.NewColor(0, 0, 104, 255))
			RenderScene()
			DrawFPS()
			DrawMap()
		rl.EndTextureMode()

		rl.BeginDrawing()
			rl.DrawTextureRec(
				screenBuffer.Texture, 
				rl.NewRectangle(0, 0, float32(screenBuffer.Texture.Width), float32(-screenBuffer.Texture.Height)), 
				rl.NewVector2(0, 0), 
				rl.White,
		)
		rl.EndDrawing()
	}
}

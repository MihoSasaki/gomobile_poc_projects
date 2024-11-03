package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"

	"log"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

var (
	images   *glutil.Images
	fps      *debug.FPS
	program  gl.Program
	buf      gl.Buffer

	width  float32
	height float32
)

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		var sz size.Event
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					onStop(glctx)
					glctx = nil
				}
			case size.Event:
				sz = e
				width = float32(sz.WidthPx / 2)
				height = float32(sz.HeightPx / 2)
			case paint.Event:
				println("paint event")
				if glctx == nil || e.External {
					// As we are actively painting as fast as
					// we can (usually 60 FPS), skip any paint
					// events sent by the system.
					continue
				}

				onPaint(glctx, sz)
				a.Publish()
			case touch.Event:
				// touchX = e.X
				// touchY = e.Y
				fmt.Println(e.X)
				fmt.Println(e.Y)
			}
		}
	})
}

func onStart(glctx gl.Context) {
	var err error
	program, err = glutil.CreateProgram(glctx, vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	images = glutil.NewImages(glctx)
	fps = debug.NewFPS(images)
}

func onStop(glctx gl.Context) {
	glctx.DeleteProgram(program)
	glctx.DeleteBuffer(buf)
	fps.Release()
	images.Release()
}

func onPaint(glctx gl.Context, sz size.Event) {
	// set background color
	glctx.ClearColor(1, 1, 1, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)

	glctx.UseProgram(program)

	image := loadImage()
	applyImage(sz, image)

	// Draw fps box
	fps.Draw(sz)
}

func applyImage(sizeEvent size.Event, targetImg image.Image) {
	bounds := targetImg.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()

	ratio := float64(dy) / float64(dx)
	fmt.Println(ratio)
	img := images.NewImage(dx, dy)

	draw.Draw(img.RGBA, targetImg.Bounds(), targetImg, targetImg.Bounds().Min, draw.Src)
	img.Upload()
	img.Draw(sizeEvent, geom.Point{X: 0, Y: 0}, geom.Point{X: geom.Pt(float32(200)), Y: 0}, geom.Point{X: 0, Y: geom.Pt(float32(200 * ratio))}, targetImg.Bounds())
}

func loadImage() image.Image {
	a, err := asset.Open("cat.png")
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	m, err := png.Decode(a)
	if err != nil {
		log.Fatal(err)
	}

	return m
}

const vertexShader = `#version 100
uniform vec2 offset;

attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position + offset4;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`

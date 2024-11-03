package main

import (
	"time"

	"github.com/MihoSasaki/gomobile_poc_projects/minigame/game"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
	"golang.org/x/mobile/exp/sprite/glsprite"
	"golang.org/x/mobile/gl"
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
					onStop()
					glctx = nil
				}
			case size.Event:
				sz = e
				width = float32(sz.WidthPt)
				height = float32(sz.HeightPt)
			case paint.Event:
				if glctx == nil || e.External {
					continue
				}
				onPaint(glctx, sz)
				a.Publish()
				a.Send(paint.Event{}) // keep animating
			case touch.Event:
				cursor := g.CheckTouchIsCursor(int(e.X))
				condition := g.CheckTouchIsUp(int(e.Y))
				switch e.Type {
				case touch.TypeBegin:
					g.Press(true, condition, cursor)
				case touch.TypeEnd:
					g.Press(false, condition, cursor)
				}
			case key.Event:
				if down := e.Direction == key.DirPress; down || e.Direction == key.DirRelease {
					switch e.Code {
					case key.CodeUpArrow:
						g.Press(down, game.TouchUp, true)
					case key.CodeDownArrow:
						g.Press(down, game.TouchDown, true)
					}

					if down && e.Code == key.CodeSpacebar {
						g.PopNewBomb()
					}
				}
			}
		}
	})
}

var (
	startTime = time.Now()
	images    *glutil.Images
	eng       sprite.Engine
	scene     *sprite.Node
	g         *game.Game
	width     float32
	height    float32
)

func onStart(glctx gl.Context) {
	images = glutil.NewImages(glctx)
	eng = glsprite.Engine(images)
	g = game.NewGame(eng, width, height)
	scene = g.Scene()
}

func onStop() {
	eng.Release()
	images.Release()
	g = nil
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(0, 0, 128, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)
	now := clock.Time(time.Since(startTime) * 60 / time.Second)
	g.Update(now)
	eng.Render(scene, now, sz)
}

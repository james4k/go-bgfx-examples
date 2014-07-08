package main

import (
	"github.com/james4k/go-bgfx"
	"github.com/james4k/go-bgfx-examples/example"
)

func main() {
	app := example.Open()
	defer app.Close()
	bgfx.Init()
	defer bgfx.Shutdown()
	bgfx.Reset(app.Width, app.Height, bgfx.ResetVSync)
	bgfx.SetDebug(bgfx.DebugText)
	bgfx.SetViewClear(
		0,
		bgfx.ClearColor|bgfx.ClearDepth,
		0x303030ff,
		1.0,
		0,
	)
	for app.Continue() {
		bgfx.SetViewRect(0, 0, 0, app.Width, app.Height)
		bgfx.Submit(0)
		bgfx.DebugTextClear()
		bgfx.DebugTextPrintf(0, 1, 0x4f, app.Title)
		bgfx.DebugTextPrintf(0, 2, 0x6f, "Description: Initialization and debug text.")
		bgfx.Frame()
	}
}

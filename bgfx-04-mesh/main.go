package main

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/james4k/go-bgfx"
	"github.com/james4k/go-bgfx-examples/assets"
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

	uTime := bgfx.CreateUniform("u_time", bgfx.Uniform1f, 1)
	defer bgfx.DestroyUniform(uTime)

	prog, err := assets.LoadProgram("vs_mesh", "fs_mesh")
	if err != nil {
		log.Fatalln(err)
	}
	defer bgfx.DestroyProgram(prog)

	mesh := assets.LoadMesh("bunny")
	defer mesh.Unload()

	for app.Continue() {
		bgfx.SetViewRect(0, 0, 0, app.Width, app.Height)
		bgfx.DebugTextClear()
		bgfx.DebugTextPrintf(0, 1, 0x4f, app.Title)
		bgfx.DebugTextPrintf(0, 2, 0x6f, "Description: Loading meshes.")
		bgfx.DebugTextPrintf(0, 3, 0x0f, "Frame: % 7.3f[ms]", app.DeltaTime*1000.0)
		bgfx.Submit(0)

		bgfx.SetUniform(uTime, &app.Time, 1)

		var (
			eye = mgl32.Vec3{0, 1, -2.5}
			at  = mgl32.Vec3{0, 1, 0}
			up  = mgl32.Vec3{0, 1, 0}
		)
		view := [16]float32(mgl32.LookAtV(eye, at, up))
		proj := [16]float32(mgl32.Perspective(
			mgl32.DegToRad(60.0),
			float32(app.Width)/float32(app.Height),
			0.1, 100,
		))
		bgfx.SetViewTransform(0, view, proj)

		mtx := mgl32.HomogRotate3DY(app.Time * 0.37)
		mesh.Submit(0, prog, mtx)

		bgfx.Frame()
	}
}

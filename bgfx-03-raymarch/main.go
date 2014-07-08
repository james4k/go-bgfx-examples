package main

import (
	"log"
	"runtime"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/james4k/go-bgfx"
	"github.com/james4k/go-bgfx-examples/assets"
	"github.com/james4k/go-bgfx-examples/example"
)

type PosColorTexcoord0Vertex struct {
	X, Y, Z float32
	ABGR    uint32
	U, V    float32
}

func renderScreenSpaceQuad(view bgfx.ViewID, prog bgfx.Program, decl bgfx.VertexDecl, x, y, width, height float32) {
	var (
		verts []PosColorTexcoord0Vertex
		idxs  []uint16
	)
	tvb, tib, ok := bgfx.AllocTransientBuffers(&verts, &idxs, decl, 4, 6)
	if !ok {
		return
	}
	const (
		z    = 0.0
		minu = -1.0
		minv = -1.0
		maxu = 1.0
		maxv = 1.0
	)
	var (
		minx = x
		maxx = x + width
		miny = y
		maxy = y + height
	)
	verts[0].X = minx
	verts[0].Y = miny
	verts[0].Z = z
	verts[0].ABGR = 0xff0000ff
	verts[0].U = minu
	verts[0].V = minv

	verts[1].X = maxx
	verts[1].Y = miny
	verts[1].Z = z
	verts[1].ABGR = 0xff00ff00
	verts[1].U = maxu
	verts[1].V = minv

	verts[2].X = maxx
	verts[2].Y = maxy
	verts[2].Z = z
	verts[2].ABGR = 0xffff0000
	verts[2].U = maxu
	verts[2].V = maxv

	verts[3].X = minx
	verts[3].Y = maxy
	verts[3].Z = z
	verts[3].ABGR = 0xffffffff
	verts[3].U = minu
	verts[3].V = maxv

	idxs[0] = 0
	idxs[1] = 2
	idxs[2] = 1
	idxs[3] = 0
	idxs[4] = 3
	idxs[5] = 2

	bgfx.SetProgram(prog)
	bgfx.SetState(bgfx.StateDefault)
	bgfx.SetTransientIndexBuffer(tib, 0, 6)
	bgfx.SetTransientVertexBuffer(tvb, 0, 4)
	bgfx.Submit(view)
}

func init() {
	runtime.LockOSThread()
}

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
	bgfx.SetViewRect(0, 0, 0, app.Width, app.Height)
	bgfx.Submit(0)

	var vd bgfx.VertexDecl
	vd.Begin()
	vd.Add(bgfx.AttribPosition, 3, bgfx.AttribTypeFloat, false, false)
	vd.Add(bgfx.AttribColor0, 4, bgfx.AttribTypeUint8, true, false)
	vd.Add(bgfx.AttribTexcoord0, 2, bgfx.AttribTypeFloat, true, false)
	vd.End()

	uTime := bgfx.CreateUniform("u_time", bgfx.Uniform1f, 1)
	uMtx := bgfx.CreateUniform("u_mtx", bgfx.Uniform4x4fv, 1)
	uLightDir := bgfx.CreateUniform("u_lightDir", bgfx.Uniform3fv, 1)
	defer bgfx.DestroyUniform(uTime)
	defer bgfx.DestroyUniform(uMtx)
	defer bgfx.DestroyUniform(uLightDir)

	prog, err := assets.LoadProgram("vs_raymarching", "fs_raymarching")
	if err != nil {
		log.Fatalln(err)
	}
	defer bgfx.DestroyProgram(prog)

	for app.Continue() {
		bgfx.DebugTextClear()
		bgfx.DebugTextPrintf(0, 1, 0x4f, app.Title)
		bgfx.DebugTextPrintf(0, 2, 0x6f, "Description: Updating shader uniforms.")
		bgfx.DebugTextPrintf(0, 3, 0x0f, "Frame: % 7.3f[ms]", app.DeltaTime*1000.0)

		var (
			eye = mgl32.Vec3{0, 0, -15.0}
			at  = mgl32.Vec3{0, 0, 0}
			up  = mgl32.Vec3{1, 0, 0}
		)
		view := mgl32.LookAtV(eye, at, up)
		proj := mgl32.Perspective(
			mgl32.DegToRad(60.0),
			float32(app.Width)/float32(app.Height),
			0.1, 100.0,
		)
		bgfx.SetViewRect(0, 0, 0, app.Width, app.Height)
		bgfx.SetViewTransform(0, [16]float32(view), [16]float32(proj))
		bgfx.Submit(0)

		ortho := mgl32.Ortho(0, float32(app.Width), float32(app.Height), 0, 0, 100)
		bgfx.SetViewRect(1, 0, 0, app.Width, app.Height)
		bgfx.SetViewTransform(1, [16]float32(mgl32.Ident4()), [16]float32(ortho))

		viewProj := proj.Mul4(view)
		mtx := mgl32.HomogRotate3DX(app.Time).Mul4(mgl32.HomogRotate3DY(app.Time * 0.37))
		invMtx := mtx.Inv()
		lightDirModel := mgl32.Vec3{-0.4, -0.5, -1.0}
		lightDirModelN := lightDirModel.Normalize()
		lightDir := invMtx.Mul4x1(lightDirModelN.Vec4(0))
		invMvp := viewProj.Mul4(mtx).Inv()

		bgfx.SetUniform(uTime, &app.Time, 1)
		bgfx.SetUniform(uLightDir, &lightDir, 1)
		bgfx.SetUniform(uMtx, &invMvp, 1)

		renderScreenSpaceQuad(1, prog, vd, 0, 0, float32(app.Width), float32(app.Height))

		bgfx.Frame()
	}
}

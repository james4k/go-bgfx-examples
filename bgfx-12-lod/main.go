package main

import (
	"math"

	"github.com/james4k/go-bgfx"
	"github.com/james4k/go-bgfx-examples/assets"
	"github.com/james4k/go-bgfx-examples/example"
	"j4k.co/cgm"
	"j4k.co/cgm/mat4"
)

type knightPos struct {
	X, Y int32
}

var knightTour = []knightPos{
	{0, 0}, {1, 2}, {3, 3}, {4, 1}, {5, 3}, {7, 2}, {6, 0}, {5, 2},
	{7, 3}, {6, 1}, {4, 0}, {3, 2}, {2, 0}, {0, 1}, {1, 3}, {2, 1},
	{0, 2}, {1, 0}, {2, 2}, {0, 3}, {1, 1}, {3, 0}, {4, 2}, {5, 0},
	{7, 1}, {6, 3}, {5, 1}, {7, 0}, {6, 2}, {4, 3}, {3, 1}, {2, 3},
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

	var (
		uTexColor   = bgfx.CreateUniform("u_texColor", bgfx.Uniform1iv, 1)
		uStipple    = bgfx.CreateUniform("u_stipple", bgfx.Uniform3fv, 1)
		uTexStipple = bgfx.CreateUniform("u_texStipple", bgfx.Uniform1iv, 1)
	)
	defer bgfx.DestroyUniform(uTexColor)
	defer bgfx.DestroyUniform(uStipple)
	defer bgfx.DestroyUniform(uTexStipple)

	prog := assets.LoadProgram("vs_tree", "fs_tree")
	defer bgfx.DestroyProgram(prog)

	textureLeafs := assets.LoadTexture("leafs1.dds", 0)
	textureBark := assets.LoadTexture("bark1.dds", 0)
	defer bgfx.DestroyTexture(textureLeafs)
	defer bgfx.DestroyTexture(textureBark)

	stippleData := make([]byte, 8*4)
	for i, v := range knightTour {
		stippleData[(v.Y*8 + v.X)] = byte(i * 4)
	}
	textureStipple := bgfx.CreateTexture2D(8, 4, 1, bgfx.TextureFormatR8,
		bgfx.TextureMinPoint|bgfx.TextureMagPoint, stippleData)
	defer bgfx.DestroyTexture(textureStipple)

	meshTop := [3]assets.Mesh{
		assets.LoadMesh("tree1b_lod0_1"),
		assets.LoadMesh("tree1b_lod1_1"),
		assets.LoadMesh("tree1b_lod2_1"),
	}
	meshTrunk := [3]assets.Mesh{
		assets.LoadMesh("tree1b_lod0_2"),
		assets.LoadMesh("tree1b_lod1_2"),
		assets.LoadMesh("tree1b_lod2_2"),
	}
	for _, m := range meshTop {
		defer m.Unload()
	}
	for _, m := range meshTrunk {
		defer m.Unload()
	}

	var (
		stateCommon      = bgfx.StateRGBWrite | bgfx.StateAlphaWrite | bgfx.StateDepthTestLess | bgfx.StateCullCCW | bgfx.StateMSAA
		stateTransparent = stateCommon | bgfx.StateBlendAlpha()
		stateOpaque      = stateCommon | bgfx.StateDepthWrite

		transitions     = true
		transitionFrame = 0
		currLOD         = 0
		targetLOD       = 0
	)

	for app.Continue() {
		bgfx.SetViewRect(0, 0, 0, app.Width, app.Height)
		bgfx.Submit(0)
		bgfx.DebugTextClear()
		bgfx.DebugTextPrintf(0, 1, 0x4f, app.Title)
		bgfx.DebugTextPrintf(0, 2, 0x6f, "Description: Mesh LOD transitions.")
		bgfx.DebugTextPrintf(0, 3, 0x0f, "Frame: % 7.3f[ms]", app.DeltaTime*1000.0)

		var (
			currentLODframe = 32
			mainLOD         = targetLOD
			distance        = float32(math.Cos(float64(app.Time))*2 + 4.0)
		)
		if transitions {
			currentLODframe = 32 - transitionFrame
			mainLOD = currLOD
		}
		stipple := [3]float32{
			0,
			-1,
			(float32(currentLODframe) * 4 / 255) - (1.0 / 255),
		}
		stippleInv := [3]float32{
			31.0 * 4 / 255,
			1,
			(float32(transitionFrame) * 4 / 255) - (1.0 / 255),
		}

		var (
			eye = [3]float32{0, 1, -distance}
			at  = [3]float32{0, 1, 0}
			up  = [3]float32{0, 1, 0}
		)
		view := mat4.LookAtLH(eye, at, up)
		proj := mat4.PerspectiveLH(
			cgm.Degrees(60.0).ToRadians(),
			float32(app.Width)/float32(app.Height),
			0.1, 100,
		)
		mtx := mat4.Scale(0.1, 0.1, 0.1)
		bgfx.SetViewTransform(0, view, proj)

		bgfx.SetTexture(0, uTexColor, textureBark)
		bgfx.SetTexture(1, uTexStipple, textureStipple)
		bgfx.SetUniform(uStipple, &stipple, 1)
		meshTrunk[mainLOD].Submit(0, prog, mtx, stateOpaque)

		bgfx.SetTexture(0, uTexColor, textureLeafs)
		bgfx.SetTexture(1, uTexStipple, textureStipple)
		bgfx.SetUniform(uStipple, &stipple, 1)
		meshTop[mainLOD].Submit(0, prog, mtx, stateTransparent)

		if transitions && transitionFrame != 0 {
			bgfx.SetTexture(0, uTexColor, textureBark)
			bgfx.SetTexture(1, uTexStipple, textureStipple)
			bgfx.SetUniform(uStipple, &stippleInv, 1)
			meshTrunk[targetLOD].Submit(0, prog, mtx, stateOpaque)

			bgfx.SetTexture(0, uTexColor, textureLeafs)
			bgfx.SetTexture(1, uTexStipple, textureStipple)
			bgfx.SetUniform(uStipple, &stippleInv, 1)
			meshTop[targetLOD].Submit(0, prog, mtx, stateTransparent)
		}

		lod := 0
		if distance > 2.5 {
			lod = 1
		}
		if distance > 5.0 {
			lod = 2
		}
		if targetLOD != lod && targetLOD == currLOD {
			targetLOD = lod
		}
		if currLOD != targetLOD {
			transitionFrame++
		}
		if transitionFrame > 32 {
			currLOD = targetLOD
			transitionFrame = 0
		}

		bgfx.Frame()
	}
}

package main

import (
	"encoding/binary"
	"math"

	"github.com/james4k/go-bgfx"
	"github.com/james4k/go-bgfx-examples/assets"
	"github.com/james4k/go-bgfx-examples/example"
	"j4k.co/cgm"
	"j4k.co/cgm/mat4"
)

type PosNormalTangentTexcoordVertex struct {
	X, Y, Z         float32
	Normal, Tangent [4]uint8
	U, V            int16
}

func packF4u(x, y, z float32) [4]uint8 {
	return [4]uint8{
		uint8(x*127 + 128),
		uint8(y*127 + 128),
		uint8(z*127 + 128),
		0,
	}
}

var vertices = []PosNormalTangentTexcoordVertex{
	{-1.0, 1.0, 1.0, packF4u(0.0, 0.0, 1.0), [4]uint8{}, 0x0000, 0x0000},
	{1.0, 1.0, 1.0, packF4u(0.0, 0.0, 1.0), [4]uint8{}, 0x7fff, 0x0000},
	{-1.0, -1.0, 1.0, packF4u(0.0, 0.0, 1.0), [4]uint8{}, 0x0000, 0x7fff},
	{1.0, -1.0, 1.0, packF4u(0.0, 0.0, 1.0), [4]uint8{}, 0x7fff, 0x7fff},
	{-1.0, 1.0, -1.0, packF4u(0.0, 0.0, -1.0), [4]uint8{}, 0x0000, 0x0000},
	{1.0, 1.0, -1.0, packF4u(0.0, 0.0, -1.0), [4]uint8{}, 0x7fff, 0x0000},
	{-1.0, -1.0, -1.0, packF4u(0.0, 0.0, -1.0), [4]uint8{}, 0x0000, 0x7fff},
	{1.0, -1.0, -1.0, packF4u(0.0, 0.0, -1.0), [4]uint8{}, 0x7fff, 0x7fff},
	{-1.0, 1.0, 1.0, packF4u(0.0, 1.0, 0.0), [4]uint8{}, 0x0000, 0x0000},
	{1.0, 1.0, 1.0, packF4u(0.0, 1.0, 0.0), [4]uint8{}, 0x7fff, 0x0000},
	{-1.0, 1.0, -1.0, packF4u(0.0, 1.0, 0.0), [4]uint8{}, 0x0000, 0x7fff},
	{1.0, 1.0, -1.0, packF4u(0.0, 1.0, 0.0), [4]uint8{}, 0x7fff, 0x7fff},
	{-1.0, -1.0, 1.0, packF4u(0.0, -1.0, 0.0), [4]uint8{}, 0x0000, 0x0000},
	{1.0, -1.0, 1.0, packF4u(0.0, -1.0, 0.0), [4]uint8{}, 0x7fff, 0x0000},
	{-1.0, -1.0, -1.0, packF4u(0.0, -1.0, 0.0), [4]uint8{}, 0x0000, 0x7fff},
	{1.0, -1.0, -1.0, packF4u(0.0, -1.0, 0.0), [4]uint8{}, 0x7fff, 0x7fff},
	{1.0, -1.0, 1.0, packF4u(1.0, 0.0, 0.0), [4]uint8{}, 0x0000, 0x0000},
	{1.0, 1.0, 1.0, packF4u(1.0, 0.0, 0.0), [4]uint8{}, 0x7fff, 0x0000},
	{1.0, -1.0, -1.0, packF4u(1.0, 0.0, 0.0), [4]uint8{}, 0x0000, 0x7fff},
	{1.0, 1.0, -1.0, packF4u(1.0, 0.0, 0.0), [4]uint8{}, 0x7fff, 0x7fff},
	{-1.0, -1.0, 1.0, packF4u(-1.0, 0.0, 0.0), [4]uint8{}, 0x0000, 0x0000},
	{-1.0, 1.0, 1.0, packF4u(-1.0, 0.0, 0.0), [4]uint8{}, 0x7fff, 0x0000},
	{-1.0, -1.0, -1.0, packF4u(-1.0, 0.0, 0.0), [4]uint8{}, 0x0000, 0x7fff},
	{-1.0, 1.0, -1.0, packF4u(-1.0, 0.0, 0.0), [4]uint8{}, 0x7fff, 0x7fff},
}

var indices = []uint16{
	0, 2, 1,
	1, 2, 3,
	4, 5, 6,
	5, 7, 6,

	8, 10, 9,
	9, 10, 11,
	12, 13, 14,
	13, 15, 14,

	16, 18, 17,
	17, 18, 19,
	20, 21, 22,
	21, 23, 22,
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

	instancingSupported := bgfx.Caps().Supported&bgfx.CapsInstancing != 0

	var vd bgfx.VertexDecl
	vd.Begin()
	vd.Add(bgfx.AttribPosition, 3, bgfx.AttribTypeFloat, false, false)
	vd.Add(bgfx.AttribNormal, 4, bgfx.AttribTypeUint8, true, true)
	vd.Add(bgfx.AttribTangent, 4, bgfx.AttribTypeUint8, true, true)
	vd.Add(bgfx.AttribTexcoord0, 2, bgfx.AttribTypeInt16, true, true)
	vd.End()
	example.CalculateTangents(vertices, len(vertices), vd, indices)

	vb := bgfx.CreateVertexBuffer(vertices, vd)
	defer bgfx.DestroyVertexBuffer(vb)
	ib := bgfx.CreateIndexBuffer(indices)
	defer bgfx.DestroyIndexBuffer(ib)

	const numLights = 4
	uTexColor := bgfx.CreateUniform("u_texColor", bgfx.Uniform1iv, 1)
	uTexNormal := bgfx.CreateUniform("u_texNormal", bgfx.Uniform1iv, 1)
	uLightPosRadius := bgfx.CreateUniform("u_lightPosRadius", bgfx.Uniform4fv, numLights)
	uLightRgbInnerR := bgfx.CreateUniform("u_lightRgbInnerR", bgfx.Uniform4fv, numLights)

	vsbump := "vs_bump"
	if instancingSupported {
		vsbump = "vs_bump_instanced"
	}
	prog := assets.LoadProgram(vsbump, "fs_bump")
	defer bgfx.DestroyProgram(prog)

	textureColor := assets.LoadTexture("fieldstone-rgba.dds", 0)
	textureNormal := assets.LoadTexture("fieldstone-n.dds", 0)

	for app.Continue() {
		var (
			eye = [3]float32{0, 0, -7.0}
			at  = [3]float32{0, 0, 0}
			up  = [3]float32{1, 0, 0}
		)
		view := mat4.LookAtLH(eye, at, up)
		proj := mat4.PerspectiveLH(
			cgm.Degrees(60).ToRadians(),
			float32(app.Width)/float32(app.Height),
			0.1, 100.0,
		)
		bgfx.SetViewTransform(0, view, proj)
		bgfx.SetViewRect(0, 0, 0, app.Width, app.Height)
		bgfx.DebugTextClear()
		bgfx.DebugTextPrintf(0, 1, 0x4f, app.Title)
		bgfx.DebugTextPrintf(0, 2, 0x6f, "Description: Loading textures.")
		bgfx.DebugTextPrintf(0, 3, 0x0f, "Frame: % 7.3f[ms]", app.DeltaTime*1000.0)
		bgfx.Submit(0)

		const halfPi = math.Pi / 2
		var lightPosRadius [4][4]float32
		for i := 0; i < numLights; i++ {
			fi := float32(i)
			lightPosRadius[i][0] = float32(math.Sin(float64(app.Time*(0.1+fi*0.17)+fi*halfPi*1.37)) * 3.0)
			lightPosRadius[i][1] = float32(math.Cos(float64(app.Time*(0.2+fi*0.29)+fi*halfPi*1.49)) * 3.0)
			lightPosRadius[i][2] = -2.5
			lightPosRadius[i][3] = 3.0
		}
		bgfx.SetUniform(uLightPosRadius, &lightPosRadius, numLights)

		lightRgbInnerR := [4][4]float32{
			{1.0, 0.7, 0.2, 0.8},
			{0.7, 0.2, 1.0, 0.8},
			{0.2, 1.0, 0.7, 0.8},
			{1.0, 0.4, 0.2, 0.8},
		}
		bgfx.SetUniform(uLightRgbInnerR, &lightRgbInnerR, numLights)

		if instancingSupported {
			for y := 0; y < 3; y++ {
				const numInstances = 3
				const instanceStride = 64
				idb := bgfx.AllocInstanceDataBuffer(numInstances, instanceStride)
				for x := 0; x < 3; x++ {
					mtx := mat4.RotateXYZ(
						cgm.Radians(app.Time)*0.023+cgm.Radians(x)*0.21,
						cgm.Radians(app.Time)*0.03+cgm.Radians(y)*0.37,
						0,
					)
					mtx[12] = -3 + float32(x)*3
					mtx[13] = -3 + float32(y)*3
					mtx[14] = 0
					binary.Write(&idb, binary.LittleEndian, mtx)
				}
				bgfx.SetInstanceDataBuffer(idb)
				bgfx.SetProgram(prog)
				bgfx.SetVertexBuffer(vb)
				bgfx.SetIndexBuffer(ib)
				bgfx.SetTexture(0, uTexColor, textureColor)
				bgfx.SetTexture(1, uTexNormal, textureNormal)
				bgfx.SetState(bgfx.StateDefault & (^bgfx.StateCullMask))
				bgfx.Submit(0)
			}
		} else {
			for y := 0; y < 3; y++ {
				for x := 0; x < 3; x++ {
					mtx := mat4.RotateXYZ(
						cgm.Radians(app.Time)*0.023+cgm.Radians(x)*0.21,
						cgm.Radians(app.Time)*0.03+cgm.Radians(y)*0.37,
						0,
					)
					mtx[12] = -3 + float32(x)*3
					mtx[13] = -3 + float32(y)*3
					mtx[14] = 0

					bgfx.SetTransform([16]float32(mtx))
					bgfx.SetProgram(prog)
					bgfx.SetVertexBuffer(vb)
					bgfx.SetIndexBuffer(ib)
					bgfx.SetTexture(0, uTexColor, textureColor)
					bgfx.SetTexture(1, uTexNormal, textureNormal)
					bgfx.SetState(bgfx.StateDefault & (^bgfx.StateCullMask))
					bgfx.Submit(0)
				}
			}
		}
		bgfx.Frame()
	}
}

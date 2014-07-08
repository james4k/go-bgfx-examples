package main

import (
	"encoding/binary"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/james4k/go-bgfx"
	"github.com/james4k/go-bgfx-examples/assets"
	"github.com/james4k/go-bgfx-examples/example"
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
	calcTangents(vertices, len(vertices), vd, indices)

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
	prog, err := loadProgram(vsbump, "fs_bump")
	if err != nil {
		log.Fatalln(err)
	}
	defer bgfx.DestroyProgram(prog)

	textureColor, err := loadTexture("fieldstone-rgba.dds")
	if err != nil {
		log.Fatalln(err)
	}
	textureNormal, err := loadTexture("fieldstone-n.dds")
	if err != nil {
		log.Fatalln(err)
	}

	for app.Continue() {
		var (
			eye = mgl32.Vec3{0, 0, -7.0}
			at  = mgl32.Vec3{0, 0, 0}
			up  = mgl32.Vec3{1, 0, 0}
		)
		view := [16]float32(mgl32.LookAtV(eye, at, up))
		proj := [16]float32(mgl32.Perspective(
			mgl32.DegToRad(60.0),
			float32(app.Width)/float32(app.Height),
			0.1, 100.0,
		))
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
					mtx := mgl32.HomogRotate3DX(app.Time*0.023 + float32(x)*0.21)
					mtx = mtx.Mul4(mgl32.HomogRotate3DY(app.Time*0.03 + float32(y)*0.37))
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
					mtx := mgl32.HomogRotate3DX(app.Time*0.023 + float32(x)*0.21)
					mtx = mtx.Mul4(mgl32.HomogRotate3DY(app.Time*0.03 + float32(y)*0.37))
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

func loadProgram(vsh, fsh string) (bgfx.Program, error) {
	v, err := loadShader(vsh)
	if err != nil {
		return bgfx.Program{}, err
	}
	f, err := loadShader(fsh)
	if err != nil {
		return bgfx.Program{}, err
	}
	return bgfx.CreateProgram(v, f, true), nil
}

func loadShader(name string) (bgfx.Shader, error) {
	f, err := assets.Open(filepath.Join("shaders/glsl", name+".bin"))
	if err != nil {
		return bgfx.Shader{}, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return bgfx.Shader{}, err
	}
	return bgfx.CreateShader(data), nil
}

func loadTexture(name string) (bgfx.Texture, error) {
	f, err := assets.Open(filepath.Join("textures", name))
	if err != nil {
		return bgfx.Texture{}, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return bgfx.Texture{}, err
	}
	tex, _ := bgfx.CreateTexture(data, 0, 0)
	return tex, nil
}

func calcTangents(vertices interface{}, numVertices int, decl bgfx.VertexDecl, indices []uint16) {
	type posTexcoord struct {
		pos [4]float32
		uv  [4]float32
	}
	type tangent struct {
		u, v [3]float32
	}
	var v0, v1, v2 posTexcoord
	tangents := make([]tangent, numVertices)
	for i := 0; i < len(indices); i += 3 {
		var (
			i0 = int(indices[i])
			i1 = int(indices[i+1])
			i2 = int(indices[i+2])
		)
		v0.pos = bgfx.VertexUnpack(bgfx.AttribPosition, decl, vertices, i0)
		v0.uv = bgfx.VertexUnpack(bgfx.AttribTexcoord0, decl, vertices, i0)
		v1.pos = bgfx.VertexUnpack(bgfx.AttribPosition, decl, vertices, i1)
		v1.uv = bgfx.VertexUnpack(bgfx.AttribTexcoord0, decl, vertices, i1)
		v2.pos = bgfx.VertexUnpack(bgfx.AttribPosition, decl, vertices, i2)
		v2.uv = bgfx.VertexUnpack(bgfx.AttribTexcoord0, decl, vertices, i2)
		var (
			bax = v1.pos[0] - v0.pos[0]
			bay = v1.pos[1] - v0.pos[1]
			baz = v1.pos[2] - v0.pos[2]
			bau = v1.uv[0] - v0.uv[0]
			bav = v1.uv[1] - v0.uv[1]
			cax = v2.pos[0] - v0.pos[0]
			cay = v2.pos[1] - v0.pos[1]
			caz = v2.pos[2] - v0.pos[2]
			cau = v2.uv[0] - v0.uv[0]
			cav = v2.uv[1] - v0.uv[1]
		)
		var (
			invDet = 1.0 / (bau*cav - bav*cau)
			tx     = (bax*cav - cax*bav) * invDet
			ty     = (bay*cav - cay*bav) * invDet
			tz     = (baz*cav - caz*bav) * invDet
			bx     = (cax*bau - bax*cau) * invDet
			by     = (cay*bau - bay*cau) * invDet
			bz     = (caz*bau - baz*cau) * invDet
		)
		for j := 0; j < 3; j++ {
			tan := &tangents[indices[i+j]]
			tan.u[0] += tx
			tan.u[1] += ty
			tan.u[2] += tz
			tan.v[0] += bx
			tan.v[1] += by
			tan.v[2] += bz
		}
	}
	for i := 0; i < numVertices; i++ {
		tan := tangents[i]
		tanu := mgl32.Vec3(tan.u)
		tanv := mgl32.Vec3(tan.v)
		normal := mgl32.Vec4(
			bgfx.VertexUnpack(bgfx.AttribNormal, decl, vertices, i),
		).Vec3()
		ndt := normal.Dot(tanu)
		nxt := normal.Cross(tanu)
		tangent := tanu.Sub(normal.Mul(ndt)).Normalize().Vec4(1.0)
		if nxt.Dot(tanv) < 0.0 {
			tangent[3] = -1.0
		}
		bgfx.VertexPack(tangent, true, bgfx.AttribTangent, decl, vertices, i)
	}
}

package main

import (
	"math"

	"github.com/james4k/go-bgfx"
	"github.com/james4k/go-bgfx-examples/example"
	"j4k.co/cgm"
	"j4k.co/cgm/mat4"
	"j4k.co/cgm/vec3"
)

type PosNormalColorVertex struct {
	Position [3]float32
	Normal   [3]float32
	ABGR     uint32
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

	var vd bgfx.VertexDecl
	vd.Begin()
	vd.Add(bgfx.AttribPosition, 3, bgfx.AttribTypeFloat, false, false)
	vd.Add(bgfx.AttribNormal, 3, bgfx.AttribTypeFloat, false, false)
	vd.Add(bgfx.AttribColor0, 4, bgfx.AttribTypeUint8, true, false)
	vd.End()
	vsh := bgfx.CreateShader(vs_metaballs_glsl)
	fsh := bgfx.CreateShader(fs_metaballs_glsl)
	prog := bgfx.CreateProgram(vsh, fsh, true)
	defer bgfx.DestroyProgram(prog)

	const dim = 32
	const ypitch = dim
	const zpitch = dim * dim
	const invdim = 1.0 / (dim - 1)
	var grid [dim * dim * dim]cell

	for app.Continue() {
		var (
			eye = [3]float32{0, 0, -50.0}
			at  = [3]float32{0, 0, 0}
			up  = [3]float32{0, 1, 0}
		)
		view := mat4.LookAtLH(eye, at, up)
		proj := mat4.PerspectiveLH(
			cgm.ToRadians(60),
			float32(app.Width)/float32(app.Height),
			0.1, 100,
		)
		bgfx.SetViewTransform(0, view, proj)
		bgfx.SetViewRect(0, 0, 0, app.Width, app.Height)
		bgfx.Submit(0)

		// 32k vertices
		const maxVertices = (32 << 10)
		var vertices []PosNormalColorVertex
		tvb := bgfx.AllocTransientVertexBuffer(&vertices, maxVertices, vd)

		const numSpheres = 16
		var spheres [numSpheres][4]float32
		for i := 0; i < numSpheres; i++ {
			x := float64(i)
			t := float64(app.Time)
			spheres[i][0] = float32(math.Sin(t*(x*0.21)+x*0.37) * (dim*0.5 - 8.0))
			spheres[i][1] = float32(math.Sin(t*(x*0.37)+x*0.67) * (dim*0.5 - 8.0))
			spheres[i][2] = float32(math.Cos(t*(x*0.11)+x*0.13) * (dim*0.5 - 8.0))
			spheres[i][3] = 1.0 / (2.0 + float32(math.Sin(t*(x*0.13))*0.5+0.5)*2.0)
		}

		profUpdate := app.HighFreqTime()
		for z := 0; z < dim; z++ {
			fz := float32(z)
			for y := 0; y < dim; y++ {
				fy := float32(y)
				offset := (z*dim + y) * dim
				for x := 0; x < dim; x++ {
					var (
						fx      = float32(x)
						dist    float32
						prod    float32 = 1.0
						xoffset         = offset + x
					)
					for i := 0; i < numSpheres; i++ {
						pos := &spheres[i]
						dx := pos[0] - (-dim*0.5 + fx)
						dy := pos[1] - (-dim*0.5 + fy)
						dz := pos[2] - (-dim*0.5 + fz)
						invr := pos[3]
						dot := dx*dx + dy*dy + dz*dz
						dot *= invr * invr
						dist *= dot
						dist += prod
						prod *= dot
					}
					grid[xoffset].val = dist/prod - 1.0
				}
			}
		}
		profUpdate = app.HighFreqTime() - profUpdate

		profNormal := app.HighFreqTime()
		for z := 1; z < dim-1; z++ {
			for y := 1; y < dim-1; y++ {
				offset := (z*dim + y) * dim
				for x := 1; x < dim-1; x++ {
					xoffset := offset + x
					grid[xoffset].normal = vec3.Normal([3]float32{
						grid[xoffset-1].val - grid[xoffset+1].val,
						grid[xoffset-ypitch].val - grid[xoffset+ypitch].val,
						grid[xoffset-zpitch].val - grid[xoffset+zpitch].val,
					})
				}
			}
		}
		profNormal = app.HighFreqTime() - profNormal

		profTriangulate := app.HighFreqTime()
		numVertices := 0
		for z := 0; z < dim-1 && numVertices+12 < maxVertices; z++ {
			var (
				rgb [6]float32
				pos [3]float32
				val [8]*cell
			)
			rgb[2] = float32(z) * invdim
			rgb[5] = float32(z+1) * invdim
			for y := 0; y < dim-1 && numVertices+12 < maxVertices; y++ {
				offset := (z*dim + y) * dim
				rgb[1] = float32(y) * invdim
				rgb[4] = float32(y+1) * invdim
				for x := 0; x < dim-1 && numVertices+12 < maxVertices; x++ {
					xoffset := offset + x
					rgb[0] = float32(x) * invdim
					rgb[3] = float32(x+1) * invdim
					pos[0] = -dim*0.5 + float32(x)
					pos[1] = -dim*0.5 + float32(y)
					pos[2] = -dim*0.5 + float32(z)
					val[0] = &grid[xoffset+zpitch+ypitch]
					val[1] = &grid[xoffset+zpitch+ypitch+1]
					val[2] = &grid[xoffset+ypitch+1]
					val[3] = &grid[xoffset+ypitch]
					val[4] = &grid[xoffset+zpitch]
					val[5] = &grid[xoffset+zpitch+1]
					val[6] = &grid[xoffset+1]
					val[7] = &grid[xoffset]
					num := triangulate(vertices[numVertices:], rgb[:], pos[:], val[:], 0.5)
					numVertices += num
				}
			}
		}
		profTriangulate = app.HighFreqTime() - profTriangulate

		bgfx.DebugTextClear()
		bgfx.DebugTextPrintf(0, 1, 0x4f, app.Title)
		bgfx.DebugTextPrintf(0, 2, 0x6f, "Description: Rendering with transient buffers and embedded shaders.")
		bgfx.DebugTextPrintf(0, 4, 0x0f, "    Vertices: %d (%.2f%%)", numVertices, float32(numVertices*100)/maxVertices)
		bgfx.DebugTextPrintf(0, 5, 0x0f, "      Update: % 7.3f[ms]", profUpdate*1000.0)
		bgfx.DebugTextPrintf(0, 6, 0x0f, "Calc normals: % 7.3f[ms]", profNormal*1000.0)
		bgfx.DebugTextPrintf(0, 7, 0x0f, " Triangulate: % 7.3f[ms]", profTriangulate*1000.0)
		bgfx.DebugTextPrintf(0, 8, 0x0f, "       Frame: % 7.3f[ms]", app.DeltaTime*1000.0)

		mtx := mat4.RotateXYZ(
			cgm.Radians(app.Time)*0.67,
			cgm.Radians(app.Time),
			0,
		)
		bgfx.SetTransform(mtx)
		bgfx.SetProgram(prog)
		bgfx.SetTransientVertexBuffer(tvb, 0, numVertices)
		bgfx.SetState(bgfx.StateDefault)
		bgfx.Submit(0)

		bgfx.Frame()
	}
}

package main

import (
	"encoding"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/james4k/go-bgfx"
	"github.com/james4k/go-bgfx-examples/assets"
	"github.com/james4k/go-bgfx/window/bgfx_glfw"
)

func main() {
	runtime.LockOSThread()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var (
		width  = 1280
		height = 720
		title  = filepath.Base(os.Args[0])
	)
	glfw.SetErrorCallback(func(err glfw.ErrorCode, desc string) {
		log.Printf("glfw: %s\n", desc)
	})
	if !glfw.Init() {
		os.Exit(1)
	}
	defer glfw.Terminate()
	// for now, fized size window. bgfx currently breaks glfw events
	// because it overrides the NSWindow's content view
	glfw.WindowHint(glfw.Resizable, 0)
	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	bgfx_glfw.SetWindow(window)
	bgfx.Init()
	defer bgfx.Shutdown()

	bgfx.Reset(width, height, bgfx.ResetVSync)
	bgfx.SetDebug(bgfx.DebugText)
	bgfx.SetViewClear(
		0,
		bgfx.ClearColor|bgfx.ClearDepth,
		0x303030ff,
		1.0,
		0,
	)

	/*
		var vd bgfx.VertexDecl
		vd.Begin()
		vd.Add(bgfx.AttribPosition, 3, bgfx.AttribTypeFloat, false, false)
		vd.Add(bgfx.AttribColor0, 4, bgfx.AttribTypeUint8, true, false)
		vd.End()
		vb := bgfx.CreateVertexBuffer(vertices, vd)
		defer bgfx.DestroyVertexBuffer(vb)
		ib := bgfx.CreateIndexBuffer(indices)
		defer bgfx.DestroyIndexBuffer(ib)
	*/

	uTime := bgfx.CreateUniform("u_time", bgfx.Uniform1f, 1)
	defer bgfx.DestroyUniform(uTime)

	prog, err := loadProgram("vs_mesh", "fs_mesh")
	if err != nil {
		log.Fatalln(err)
	}
	defer bgfx.DestroyProgram(prog)

	mesh := loadMesh("bunny")
	defer unloadMesh(mesh)

	var last float32
	for !window.ShouldClose() {
		now := float32(glfw.GetTime())
		delta := now - last
		last = now
		bgfx.SetUniform(uTime, &now, 1)

		width, height = window.GetSize()
		bgfx.SetViewRect(0, 0, 0, width, height)
		bgfx.DebugTextClear()
		bgfx.DebugTextPrintf(0, 1, 0x4f, title)
		bgfx.DebugTextPrintf(0, 2, 0x6f, "Description: Loading meshes.")
		bgfx.DebugTextPrintf(0, 3, 0x0f, "Frame: % 7.3f[ms]", delta*1000.0)
		bgfx.Submit(0)

		var (
			eye = mgl32.Vec3{0, 1, -2.5}
			at  = mgl32.Vec3{0, 1, 0}
			up  = mgl32.Vec3{0, 1, 0}
		)
		view := [16]float32(mgl32.LookAtV(eye, at, up))
		proj := [16]float32(mgl32.Perspective(
			mgl32.DegToRad(60.0),
			float32(width)/float32(height),
			0.1, 100,
		))
		bgfx.SetViewTransform(0, view, proj)

		mtx := mgl32.HomogRotate3DY(now * 0.37)
		submitMesh(mesh, 0, prog, mtx)

		bgfx.Frame()
		glfw.PollEvents()
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

type Bounds struct {
	Sphere Sphere
	AABB   AABB
	OBB    OBB
}

type Sphere struct {
	Center [3]float32
	Radius float32
}

type AABB struct {
	Min, Max [3]float32
}

type OBB struct {
	Matrix [16]float32
}

type primitive struct {
	StartIndex  uint32
	NumIndices  uint32
	StartVertex uint32
	NumVertices uint32
	Bounds
}

type group struct {
	VB bgfx.VertexBuffer
	IB bgfx.IndexBuffer
	Bounds
	Prims []primitive
}

type mesh struct {
	groups []group
}

func loadMesh(name string) mesh {
	f, err := assets.Open(filepath.Join("meshes", name+".bin"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	return readMesh(f)
}

func read(r io.Reader, dest interface{}) {
	b, ok := dest.(encoding.BinaryUnmarshaler)
	if ok {
		size := reflect.TypeOf(dest).Elem().Size()
		buf := make([]byte, int(size))
		_, err := r.Read(buf)
		if err != nil {
			panic(err)
		}
		err = b.UnmarshalBinary(buf)
		if err != nil {
			panic(err)
		}
	} else {
		err := binary.Read(r, binary.LittleEndian, dest)
		if err != nil {
			panic(err)
		}
	}
}

func readMesh(r io.ReadSeeker) mesh {
	const (
		ChunkMagicVB  = 0x00204256 // fourcc "VB \x00"
		ChunkMagicIB  = 0x00204249 // fourcc "IB \x00"
		ChunkMagicPRI = 0x00495250 // fourcc "PRI\x00"
	)
	var (
		gg   []group
		g    group
		prim primitive
		decl bgfx.VertexDecl
	)
	for {
		var chunk uint32
		err := binary.Read(r, binary.LittleEndian, &chunk)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		switch chunk {
		case ChunkMagicVB:
			var numVertices uint16
			read(r, &g.Bounds)
			read(r, &decl)
			read(r, &numVertices)
			buf := make([]byte, int(numVertices)*decl.Stride())
			read(r, buf)
			g.VB = bgfx.CreateVertexBuffer(buf, decl)
		case ChunkMagicIB:
			var numIndices uint32
			read(r, &numIndices)
			buf := make([]uint16, int(numIndices))
			read(r, buf)
			g.IB = bgfx.CreateIndexBuffer(buf)
		case ChunkMagicPRI:
			var size, num uint16
			read(r, &size)
			r.Seek(int64(size), 1) // skip name, unused
			read(r, &num)
			for i := uint16(0); i < num; i++ {
				read(r, &size)
				r.Seek(int64(size), 1) // skip name, unused
				read(r, &prim)
				g.Prims = append(g.Prims, prim)
			}
			gg = append(gg, g)
			g = group{}
		default:
			n, _ := r.Seek(0, 1)
			log.Fatalf("mesh file: unknown chunk 0x%08x at %d\n", chunk, n)
		}
	}
	return mesh{
		groups: gg,
	}
}

func submitMesh(m mesh, view bgfx.ViewID, prog bgfx.Program, mtx [16]float32) {
	state := bgfx.StateDefault | bgfx.StateCullCCW
	state &= ^bgfx.StateCullCW
	for _, g := range m.groups {
		bgfx.SetTransform(mtx)
		bgfx.SetProgram(prog)
		bgfx.SetIndexBuffer(g.IB)
		bgfx.SetVertexBuffer(g.VB)
		bgfx.SetState(state)
		bgfx.Submit(view)
	}
}

func unloadMesh(m mesh) {
	for _, g := range m.groups {
		bgfx.DestroyVertexBuffer(g.VB)
		bgfx.DestroyIndexBuffer(g.IB)
	}
	m.groups = nil
}

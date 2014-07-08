package assets

import (
	"encoding"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/james4k/go-bgfx"
)

var dirs = filepath.SplitList(os.Getenv("GOPATH"))

func init() {
	const assets = "src/github.com/james4k/go-bgfx-examples/assets"
	for i := range dirs {
		dirs[i] = filepath.Join(dirs[i], assets)
	}
}

func Open(name string) (f *os.File, err error) {
	for _, dir := range dirs {
		f, err = os.Open(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		return
	}
	return
}

func LoadProgram(vsh, fsh string) (bgfx.Program, error) {
	v, err := LoadShader(vsh)
	if err != nil {
		return bgfx.Program{}, err
	}
	f, err := LoadShader(fsh)
	if err != nil {
		return bgfx.Program{}, err
	}
	return bgfx.CreateProgram(v, f, true), nil
}

func LoadShader(name string) (bgfx.Shader, error) {
	f, err := Open(filepath.Join("shaders/glsl", name+".bin"))
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

func LoadTexture(name string) (bgfx.Texture, error) {
	f, err := Open(filepath.Join("textures", name))
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

type Mesh struct {
	groups []group
}

func LoadMesh(name string) Mesh {
	f, err := Open(filepath.Join("meshes", name+".bin"))
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
		_, err := io.ReadFull(r, buf)
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

func readMesh(r io.ReadSeeker) Mesh {
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
	return Mesh{
		groups: gg,
	}
}

func (m Mesh) Submit(view bgfx.ViewID, prog bgfx.Program, mtx [16]float32) {
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

func (m Mesh) Unload() {
	for _, g := range m.groups {
		bgfx.DestroyVertexBuffer(g.VB)
		bgfx.DestroyIndexBuffer(g.IB)
	}
	m.groups = nil
}

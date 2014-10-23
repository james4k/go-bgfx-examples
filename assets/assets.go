package assets

import (
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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

func LoadProgram(vsh, fsh string) bgfx.Program {
	v, err := loadShader(vsh)
	if err != nil {
		log.Fatalln(err)
	}
	f, err := loadShader(fsh)
	if err != nil {
		log.Fatalln(err)
	}
	return bgfx.CreateProgram(v, f, true)
}

func loadShader(name string) (bgfx.Shader, error) {
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

func LoadTexture(name string, flags bgfx.TextureFlags) bgfx.Texture {
	f, err := Open(filepath.Join("textures", name))
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	tex, _ := bgfx.CreateTexture(data, flags, 0)
	return tex
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
	switch v := dest.(type) {
	case *bgfx.VertexDecl:
		readDecl(r, v)
	case *bool:
		var b uint8
		err := binary.Read(r, binary.LittleEndian, &b)
		if err != nil {
			panic(err)
		}
		*v = (b != 0)
	default:
		err := binary.Read(r, binary.LittleEndian, dest)
		if err != nil {
			panic(err)
		}
	}
}

var attribIdTable = map[uint16]bgfx.Attrib{
	0x01: bgfx.AttribPosition,
	0x02: bgfx.AttribNormal,
	0x03: bgfx.AttribTangent,
	0x04: bgfx.AttribBitangent,
	0x05: bgfx.AttribColor0,
	0x06: bgfx.AttribColor1,
	0x0e: bgfx.AttribIndices,
	0x0f: bgfx.AttribWeight,
	0x10: bgfx.AttribTexcoord0,
	0x11: bgfx.AttribTexcoord1,
	0x12: bgfx.AttribTexcoord2,
	0x13: bgfx.AttribTexcoord3,
	0x14: bgfx.AttribTexcoord4,
	0x15: bgfx.AttribTexcoord5,
	0x16: bgfx.AttribTexcoord6,
	0x17: bgfx.AttribTexcoord7,
}

func idToAttrib(id uint16) (bgfx.Attrib, bool) {
	a, ok := attribIdTable[id]
	return a, ok
}

func idToAttribType(id uint16) (bgfx.AttribType, bool) {
	switch id {
	case 0x01:
		return bgfx.AttribTypeUint8, true
	case 0x02:
		return bgfx.AttribTypeInt16, true
	case 0x03:
		return bgfx.AttribTypeHalf, true
	case 0x04:
		return bgfx.AttribTypeFloat, true
	default:
		return 0, false
	}
}

func readDecl(r io.Reader, decl *bgfx.VertexDecl) {
	var (
		nattrs uint8
		stride uint16
	)
	read(r, &nattrs)
	read(r, &stride)
	decl.Begin()
	defer decl.End()
	for i := uint8(0); i < nattrs; i++ {
		var (
			offset       uint16
			attribID     uint16
			num          uint8
			attribTypeID uint16
			normalized   bool
			asInt        bool
		)
		read(r, &offset)
		read(r, &attribID)
		read(r, &num)
		read(r, &attribTypeID)
		read(r, &normalized)
		read(r, &asInt)
		attr, ok := idToAttrib(attribID)
		if !ok {
			continue
		}
		typ, ok := idToAttribType(attribTypeID)
		if !ok {
			continue
		}
		decl.Add(attr, num, typ, normalized, asInt)
		// TODO: set offset.. with unsafe i guess?
	}
	// TODO: set stride.. with unsafe i guess?
}

func readMesh(r io.ReadSeeker) Mesh {
	const (
		ChunkMagicVB  = 0x01204256 // fourcc "VB \x01"
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

func (m Mesh) Submit(view bgfx.ViewID, prog bgfx.Program, mtx [16]float32, state bgfx.State) {
	if state == 0 {
		state = bgfx.StateDefault | bgfx.StateCullCCW
		state &= ^bgfx.StateCullCW
	}
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

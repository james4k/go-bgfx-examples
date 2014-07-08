package example

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/james4k/go-bgfx"
)

func CalculateTangents(vertices interface{}, numVertices int, decl bgfx.VertexDecl, indices []uint16) {
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

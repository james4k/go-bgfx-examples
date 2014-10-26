package main

import (
	"github.com/james4k/go-bgfx"
	"github.com/james4k/go-bgfx-examples/assets"
	"github.com/james4k/go-bgfx-examples/example"
	"j4k.co/cgm"
	"j4k.co/cgm/mat4"
)

type PosColorTexcoord0Vertex struct {
	X, Y, Z float32
	ABGR    uint32
	U, V    float32
}

func screenSpaceQuad(
	decl bgfx.VertexDecl,
	textureWidth,
	textureHeight float32,
	originBottomLeft bool,
) {
	var vertices []PosColorTexcoord0Vertex
	vb := bgfx.AllocTransientVertexBuffer(&vertices, 3, decl)
	const (
		z    = 0
		minx = -1.0
		maxx = 1.0
		miny = 0
		maxy = 1.0 * 2
	)
	var (
		texelHalfW = texelHalf / textureWidth
		texelHalfH = texelHalf / textureHeight

		minu = -1.0 + texelHalfW
		maxu = 1.0 + texelHalfH
		minv = texelHalfH
		maxv = 2.0 + texelHalfH
	)
	if originBottomLeft {
		minv, maxv = maxv-1, minv-1
	}

	vertices[0].X = minx
	vertices[0].Y = miny
	vertices[0].Z = z
	vertices[0].ABGR = 0xffffffff
	vertices[0].U = minu
	vertices[0].V = minv

	vertices[1].X = maxx
	vertices[1].Y = miny
	vertices[1].Z = z
	vertices[1].ABGR = 0xffffffff
	vertices[1].U = maxu
	vertices[1].V = minv

	vertices[2].X = maxx
	vertices[2].Y = maxy
	vertices[2].Z = z
	vertices[2].ABGR = 0xffffffff
	vertices[2].U = maxu
	vertices[2].V = maxv

	bgfx.SetTransientVertexBuffer(vb, 0, 3)
}

func setOffsets2x2Lum(uniform bgfx.Uniform, w, h int) {
	var (
		offsets [16][4]float32
		du      = 1.0 / float32(w)
		dv      = 1.0 / float32(h)
		num     = 0
	)
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			offsets[num][0] = (float32(x) - texelHalf) * du
			offsets[num][1] = (float32(y) - texelHalf) * dv
			num++
		}
	}
	bgfx.SetUniform(uniform, &offsets, num)
}

func setOffsets4x4Lum(uniform bgfx.Uniform, w, h int) {
	var (
		offsets [16][4]float32
		du      = 1.0 / float32(w)
		dv      = 1.0 / float32(h)
		num     = 0
	)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			offsets[num][0] = (float32(x) - 1 - texelHalf) * du
			offsets[num][1] = (float32(y) - 1 - texelHalf) * dv
			num++
		}
	}
	bgfx.SetUniform(uniform, &offsets, num)
}

var texelHalf float32

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

	caps := bgfx.Caps()
	originBottomLeft := false
	switch caps.RendererType {
	case bgfx.RendererTypeDirect3D9:
		texelHalf = 0.5
	case bgfx.RendererTypeOpenGL, bgfx.RendererTypeOpenGLES:
		originBottomLeft = true
	}

	var decl bgfx.VertexDecl
	decl.Begin()
	decl.Add(bgfx.AttribPosition, 3, bgfx.AttribTypeFloat, false, false)
	decl.Add(bgfx.AttribColor0, 4, bgfx.AttribTypeUint8, true, false)
	decl.Add(bgfx.AttribTexcoord0, 2, bgfx.AttribTypeFloat, false, false)
	decl.End()

	var (
		skyProg     = assets.LoadProgram("vs_hdr_skybox", "fs_hdr_skybox")
		lumProg     = assets.LoadProgram("vs_hdr_lum", "fs_hdr_lum")
		lumAvgProg  = assets.LoadProgram("vs_hdr_lumavg", "fs_hdr_lumavg")
		blurProg    = assets.LoadProgram("vs_hdr_blur", "fs_hdr_blur")
		brightProg  = assets.LoadProgram("vs_hdr_bright", "fs_hdr_bright")
		meshProg    = assets.LoadProgram("vs_hdr_mesh", "fs_hdr_mesh")
		tonemapProg = assets.LoadProgram("vs_hdr_tonemap", "fs_hdr_tonemap")
	)
	defer bgfx.DestroyProgram(skyProg)
	defer bgfx.DestroyProgram(lumProg)
	defer bgfx.DestroyProgram(lumAvgProg)
	defer bgfx.DestroyProgram(blurProg)
	defer bgfx.DestroyProgram(brightProg)
	defer bgfx.DestroyProgram(meshProg)
	defer bgfx.DestroyProgram(tonemapProg)

	var (
		uTime     = bgfx.CreateUniform("u_time", bgfx.Uniform1f, 1)
		uTexCube  = bgfx.CreateUniform("u_texCube", bgfx.Uniform1i, 1)
		uTexColor = bgfx.CreateUniform("u_texColor", bgfx.Uniform1i, 1)
		uTexLum   = bgfx.CreateUniform("u_texLum", bgfx.Uniform1i, 1)
		uTexBlur  = bgfx.CreateUniform("u_texBlur", bgfx.Uniform1i, 1)
		uMtx      = bgfx.CreateUniform("u_mtx", bgfx.Uniform4x4fv, 1)
		uTonemap  = bgfx.CreateUniform("u_tonemap", bgfx.Uniform4fv, 1)
		uOffset   = bgfx.CreateUniform("u_offset", bgfx.Uniform4fv, 16)
	)
	defer bgfx.DestroyUniform(uTime)
	defer bgfx.DestroyUniform(uTexCube)
	defer bgfx.DestroyUniform(uTexColor)
	defer bgfx.DestroyUniform(uTexLum)
	defer bgfx.DestroyUniform(uTexBlur)
	defer bgfx.DestroyUniform(uMtx)
	defer bgfx.DestroyUniform(uTonemap)
	defer bgfx.DestroyUniform(uOffset)

	mesh := assets.LoadMesh("bunny")
	defer mesh.Unload()

	uffizi := assets.LoadTexture("uffizi.dds", bgfx.TextureUClamp|bgfx.TextureVClamp|bgfx.TextureWClamp)
	defer bgfx.DestroyTexture(uffizi)

	fbtextures := []bgfx.Texture{
		bgfx.CreateTexture2D(app.Width, app.Height, 1, bgfx.TextureFormatBGRA8, bgfx.TextureRT|bgfx.TextureUClamp|bgfx.TextureVClamp, nil),
		bgfx.CreateTexture2D(app.Width, app.Height, 1, bgfx.TextureFormatD16, bgfx.TextureRTBufferOnly, nil),
	}
	fb := bgfx.CreateFrameBufferFromTextures(fbtextures, true)
	lum := [5]bgfx.FrameBuffer{
		bgfx.CreateFrameBuffer(128, 128, bgfx.TextureFormatBGRA8, 0),
		bgfx.CreateFrameBuffer(64, 64, bgfx.TextureFormatBGRA8, 0),
		bgfx.CreateFrameBuffer(16, 16, bgfx.TextureFormatBGRA8, 0),
		bgfx.CreateFrameBuffer(4, 4, bgfx.TextureFormatBGRA8, 0),
		bgfx.CreateFrameBuffer(1, 1, bgfx.TextureFormatBGRA8, 0),
	}
	bright := bgfx.CreateFrameBuffer(app.Width/2, app.Height/2, bgfx.TextureFormatBGRA8, 0)
	blur := bgfx.CreateFrameBuffer(app.Width/8, app.Height/8, bgfx.TextureFormatBGRA8, 0)
	// defer in closure to capture these by reference since we destroy
	// and recreate these when the window resizes.
	defer func() {
		for _, l := range lum {
			bgfx.DestroyFrameBuffer(l)
		}
		bgfx.DestroyFrameBuffer(fb)
		bgfx.DestroyFrameBuffer(bright)
		bgfx.DestroyFrameBuffer(blur)
	}()

	const (
		speed      = 0.37
		middleGray = 0.18
		white      = 1.1
		threshold  = 1.5
	)
	var (
		prevWidth  = app.Width
		prevHeight = app.Height
	)
	for app.Continue() {
		if prevWidth != app.Width || prevHeight != app.Height {
			prevWidth = app.Width
			prevHeight = app.Height
			bgfx.DestroyFrameBuffer(fb)
			bgfx.DestroyFrameBuffer(blur)
			bgfx.DestroyFrameBuffer(bright)
			fbtextures[0] = bgfx.CreateTexture2D(app.Width, app.Height, 1, bgfx.TextureFormatBGRA8, bgfx.TextureRT|bgfx.TextureUClamp|bgfx.TextureVClamp, nil)
			fbtextures[1] = bgfx.CreateTexture2D(app.Width, app.Height, 1, bgfx.TextureFormatD16, bgfx.TextureRTBufferOnly, nil)
			fb = bgfx.CreateFrameBufferFromTextures(fbtextures, true)
			bright = bgfx.CreateFrameBuffer(app.Width/2, app.Height/2, bgfx.TextureFormatBGRA8, 0)
			blur = bgfx.CreateFrameBuffer(app.Width/8, app.Height/8, bgfx.TextureFormatBGRA8, 0)
		}

		bgfx.DebugTextClear()
		bgfx.DebugTextPrintf(0, 1, 0x4f, app.Title)
		bgfx.DebugTextPrintf(0, 2, 0x6f, "Description: Using multiple views and render targets.")
		bgfx.DebugTextPrintf(0, 3, 0x0f, "Frame: % 7.3f[ms]", app.DeltaTime*1000.0)

		bgfx.SetUniform(uTime, &app.Time, 1)

		for i := 0; i < 6; i++ {
			bgfx.SetViewRect(bgfx.ViewID(i), 0, 0, app.Width, app.Height)
		}
		bgfx.SetViewFrameBuffer(0, fb)
		bgfx.SetViewFrameBuffer(1, fb)

		bgfx.SetViewRect(2, 0, 0, 128, 128)
		bgfx.SetViewFrameBuffer(2, lum[0])
		bgfx.SetViewRect(3, 0, 0, 64, 64)
		bgfx.SetViewFrameBuffer(3, lum[1])
		bgfx.SetViewRect(4, 0, 0, 16, 16)
		bgfx.SetViewFrameBuffer(4, lum[2])
		bgfx.SetViewRect(5, 0, 0, 4, 4)
		bgfx.SetViewFrameBuffer(5, lum[3])
		bgfx.SetViewRect(6, 0, 0, 1, 1)
		bgfx.SetViewFrameBuffer(6, lum[4])

		bgfx.SetViewRect(7, 0, 0, app.Width/2, app.Height/2)
		bgfx.SetViewFrameBuffer(7, bright)

		bgfx.SetViewRect(8, 0, 0, app.Width/8, app.Height/8)
		bgfx.SetViewFrameBuffer(8, blur)

		bgfx.SetViewRect(9, 0, 0, app.Width, app.Height)

		view := mat4.Identity()
		proj := mat4.OrthoLH(0, 1, 1, 0, 0, 100)
		for i := 0; i < 10; i++ {
			bgfx.SetViewTransform(bgfx.ViewID(i), view, proj)
		}

		var (
			eye = [3]float32{0, 1, -2.5}
			at  = [3]float32{0, 1, 0}
			up  = [3]float32{0, 1, 0}
			mtx = mat4.RotateXYZ(0, cgm.Radians(app.Time)*0.37, 0)
		)
		eye = mat4.Mul3(mtx, eye)
		view = mat4.LookAtLH(eye, at, up)
		proj = mat4.PerspectiveLH(
			cgm.ToRadians(60),
			float32(app.Width)/float32(app.Height),
			0.1, 100,
		)
		bgfx.SetViewTransform(1, view, proj)
		bgfx.SetUniform(uMtx, &mtx, 1)

		// Render skybox into view 0
		bgfx.SetTexture(0, uTexCube, uffizi)
		bgfx.SetProgram(skyProg)
		bgfx.SetState(bgfx.StateRGBWrite | bgfx.StateAlphaWrite)
		screenSpaceQuad(decl, float32(app.Width), float32(app.Height), true)
		bgfx.Submit(0)

		// Render mesh into view 1
		bgfx.SetTexture(0, uTexCube, uffizi)
		mesh.Submit(1, meshProg, mat4.Identity(), 0)

		// Calculate luminance.
		setOffsets2x2Lum(uOffset, 128, 128)
		bgfx.SetTexture(0, uTexColor, fbtextures[0])
		bgfx.SetProgram(lumProg)
		bgfx.SetState(bgfx.StateRGBWrite | bgfx.StateAlphaWrite)
		screenSpaceQuad(decl, 128, 128, originBottomLeft)
		bgfx.Submit(2)

		// Downscale luminance 0.
		setOffsets4x4Lum(uOffset, 128, 128)
		bgfx.SetTextureFromFrameBuffer(0, uTexColor, lum[0])
		bgfx.SetProgram(lumAvgProg)
		bgfx.SetState(bgfx.StateRGBWrite | bgfx.StateAlphaWrite)
		screenSpaceQuad(decl, 64, 64, originBottomLeft)
		bgfx.Submit(3)

		// Downscale luminance 1.
		setOffsets4x4Lum(uOffset, 64, 64)
		bgfx.SetTextureFromFrameBuffer(0, uTexColor, lum[1])
		bgfx.SetProgram(lumAvgProg)
		bgfx.SetState(bgfx.StateRGBWrite | bgfx.StateAlphaWrite)
		screenSpaceQuad(decl, 16, 16, originBottomLeft)
		bgfx.Submit(4)

		// Downscale luminance 2.
		setOffsets4x4Lum(uOffset, 16, 16)
		bgfx.SetTextureFromFrameBuffer(0, uTexColor, lum[2])
		bgfx.SetProgram(lumAvgProg)
		bgfx.SetState(bgfx.StateRGBWrite | bgfx.StateAlphaWrite)
		screenSpaceQuad(decl, 4, 4, originBottomLeft)
		bgfx.Submit(5)

		// Downscale luminance 3.
		setOffsets4x4Lum(uOffset, 4, 4)
		bgfx.SetTextureFromFrameBuffer(0, uTexColor, lum[3])
		bgfx.SetProgram(lumAvgProg)
		bgfx.SetState(bgfx.StateRGBWrite | bgfx.StateAlphaWrite)
		screenSpaceQuad(decl, 1, 1, originBottomLeft)
		bgfx.Submit(6)

		tonemap := [4]float32{middleGray, white * white, threshold, 0}
		bgfx.SetUniform(uTonemap, &tonemap, 1)

		// Bright pass threshold is tonemap[3]
		setOffsets4x4Lum(uOffset, app.Width/2, app.Height/2)
		bgfx.SetTexture(0, uTexColor, fbtextures[0])
		bgfx.SetTextureFromFrameBuffer(1, uTexLum, lum[4])
		bgfx.SetProgram(brightProg)
		bgfx.SetState(bgfx.StateRGBWrite | bgfx.StateAlphaWrite)
		screenSpaceQuad(decl, float32(app.Width/2), float32(app.Height/2), originBottomLeft)
		bgfx.Submit(7)

		// Blur pass vertically
		bgfx.SetTextureFromFrameBuffer(0, uTexColor, bright)
		bgfx.SetProgram(blurProg)
		bgfx.SetState(bgfx.StateRGBWrite | bgfx.StateAlphaWrite)
		screenSpaceQuad(decl, float32(app.Width/8), float32(app.Height/8), originBottomLeft)
		bgfx.Submit(8)

		// Blur pass horizontally, do tonemapping and combine.
		bgfx.SetTexture(0, uTexColor, fbtextures[0])
		bgfx.SetTextureFromFrameBuffer(1, uTexLum, lum[4])
		bgfx.SetTextureFromFrameBuffer(2, uTexBlur, blur)
		bgfx.SetProgram(tonemapProg)
		bgfx.SetState(bgfx.StateRGBWrite | bgfx.StateAlphaWrite)
		screenSpaceQuad(decl, float32(app.Width), float32(app.Height), originBottomLeft)
		bgfx.Submit(9)

		bgfx.Frame()
	}
}

/*
Package example is used by all of the go-bgfx-examples as a basic
framework for windowing, user input, and graphical utilities. The file
formats used are based on the formats used by the original bgfx
examples.
*/
package example

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	glfw "github.com/go-gl/glfw3"
	"github.com/james4k/go-bgfx/window/bgfx_glfw"
)

func init() {
	// Lock the main goroutine to the main thread. See the comment in
	// runtime·main() at http://golang.org/src/pkg/runtime/proc.c#L221
	//
	// "Lock the main goroutine onto this, the main OS thread,
	// during initialization.  Most programs won't care, but a few
	// do require certain calls to be made by the main thread.
	// Those can arrange for main.main to run in the main thread
	// by calling runtime.LockOSThread during initialization
	// to preserve the lock."
	//
	// This is needed because on some platforms, we need the main
	// thread, or at least a consistent thread, to make OS or window
	// system calls.
	runtime.LockOSThread()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type Application struct {
	window *glfw.Window

	Title         string
	Width, Height int

	Time      float32
	DeltaTime float32
}

// Open opens a new example app window, and must be called from the main
// goroutine. May only be called once.
func Open() *Application {
	app := &Application{}
	app.init()
	return app
}

func (a *Application) glfwError(errno glfw.ErrorCode, desc string) {
	log.Printf("glfw: %s\n", desc)
}

func (a *Application) init() {
	glfw.SetErrorCallback(a.glfwError)
	if !glfw.Init() {
		os.Exit(1)
	}

	a.Width = 1280
	a.Height = 720
	a.Title = filepath.Base(os.Args[0])

	// For now, force a fixed size window. bgfx currently breaks glfw
	// events on OS X because it overrides the NSWindow's content view.
	glfw.WindowHint(glfw.Resizable, 0)
	var err error
	a.window, err = glfw.CreateWindow(a.Width, a.Height, a.Title, nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	bgfx_glfw.SetWindow(a.window)
}

func (a *Application) Continue() bool {
	glfw.PollEvents()
	if a.window.ShouldClose() {
		return false
	}
	a.update()
	return true
}

func (a *Application) HighFreqTime() float64 {
	// TODO: cgo call overhead probably makes this less useful...
	return glfw.GetTime()
}

func (a *Application) Close() error {
	glfw.Terminate()
	return nil
}

func (a *Application) update() {
	a.Width, a.Height = a.window.GetSize()
	now := float32(glfw.GetTime())
	a.DeltaTime = now - a.Time
	a.Time = now
}

package schedd

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	// "ulambda/debug"
	"ulambda/fslib"
)

type Tstate struct {
	*fslib.FsLib
	t *testing.T
	s *fslib.System
}

func makeTstate(t *testing.T) *Tstate {
	ts := &Tstate{}

	bin := "../bin"
	s, err := fslib.Boot(bin)
	if err != nil {
		t.Fatalf("Boot %v\n", err)
	}
	ts.s = s

	ts.FsLib = fslib.MakeFsLib("schedl")
	ts.t = t

	return ts
}

func TestContinue(t *testing.T) {
	ts := makeTstate(t)

	pid := fslib.GenPid()
	a := &fslib.Attr{pid, "../bin/schedl", "", []string{"1", "name/out", ""}, nil, nil, nil}
	err := ts.Spawn(a)
	assert.Nil(t, err, "Spawn")

	log.Printf("Spawn %v\n", a)

	ts.Wait(pid)

	b, err := ts.ReadFile("name/out")
	assert.Nil(t, err, "ReadFile")
	assert.Equal(t, "hello again", string(b), "Output")

	ts.s.Shutdown(ts.FsLib)
}

func TestWait(t *testing.T) {
	ts := makeTstate(t)

	// debug.SetDebug(true)

	pid := fslib.GenPid()
	a := &fslib.Attr{pid, "../bin/schedl", "", []string{"0", "name/out", ""}, nil, nil, nil}
	err := ts.Spawn(a)
	assert.Nil(t, err, "Spawn")

	log.Printf("Spawn %v\n", a)

	ts.Wait(pid)

	b, err := ts.ReadFile("name/out")
	assert.Nil(t, err, "ReadFile")
	assert.Equal(t, string(b), "hello", "Output")

	ts.s.Shutdown(ts.FsLib)
}

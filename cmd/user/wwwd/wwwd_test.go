package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"ulambda/kernel"
	"ulambda/proc"
)

type Tstate struct {
	*kernel.System
	t   *testing.T
	pid string
}

func spawn(t *testing.T, ts *Tstate) string {
	a := proc.MakeProc("bin/user/wwwd", []string{""})
	err := ts.Spawn(a)
	assert.Nil(t, err, "Spawn")
	return a.Pid
}

func makeTstate(t *testing.T) *Tstate {
	var err error
	ts := &Tstate{}
	ts.t = t
	ts.System = kernel.MakeSystemAll("wwwd_test", "../../../")
	ts.pid = spawn(t, ts)

	err = ts.WaitStart(ts.pid)
	assert.Equal(t, nil, err)

	// ts.Exited(proc.GetPid(), "OK")

	return ts
}

func (ts *Tstate) waitWww() {
	ch := make(chan error)
	go func() {
		_, err := exec.Command("wget", "-qO-", "http://localhost:8080/exit/").Output()
		ch <- err
	}()

	status, err := ts.WaitExit(ts.pid)
	assert.Nil(ts.t, err, "WaitExit error")
	assert.True(ts.t, status.IsStatusEvicted(), "Exit status wrong")

	r := <-ch
	assert.NotEqual(ts.t, nil, r)

	ts.Shutdown()
}

func TestSandbox(t *testing.T) {
	ts := makeTstate(t)
	ts.waitWww()
}

func TestStatic(t *testing.T) {
	ts := makeTstate(t)

	out, err := exec.Command("wget", "-qO-", "http://localhost:8080/static/hello.html").Output()
	assert.Equal(t, nil, err)
	assert.Contains(t, string(out), "hello")

	out, err = exec.Command("wget", "-qO-", "http://localhost:8080/static/nonexist.html").Output()
	assert.NotEqual(t, nil, err) // wget return error because of HTTP not found

	ts.waitWww()
}

func TestView(t *testing.T) {
	ts := makeTstate(t)

	out, err := exec.Command("wget", "-qO-", "http://localhost:8080/book/view/").Output()
	assert.Equal(t, nil, err)
	assert.Contains(t, string(out), "Homer")

	ts.waitWww()
}

func TestEdit(t *testing.T) {
	ts := makeTstate(t)

	out, err := exec.Command("wget", "-qO-", "http://localhost:8080/book/edit/Odyssey").Output()
	assert.Equal(t, nil, err)
	assert.Contains(t, string(out), "Odyssey")

	ts.waitWww()
}

func TestSave(t *testing.T) {
	ts := makeTstate(t)

	out, err := exec.Command("wget", "-qO-", "--post-data", "title=Odyssey", "http://localhost:8080/book/save/Odyssey").Output()
	assert.Equal(t, nil, err)
	assert.Contains(t, string(out), "Homer")

	ts.waitWww()
}

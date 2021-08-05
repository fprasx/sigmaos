package gg

import (
	"fmt"
	"log"
	"path"
	//	"runtime/debug"

	"ulambda/fslib"
)

// Given a PID, create a no-op which waits on that Pid
func spawnNoOp(launch ExecutorLauncher, waitPid string) string {
	noOpPid := noOpPid(waitPid)
	exitDep := []string{waitPid}
	err := launch.SpawnNoOp(noOpPid, exitDep)
	if err != nil {
		log.Fatalf("Error spawning noop [%v]: %v\n", noOpPid, err)
	}
	return noOpPid
}

func spawnOrigDirUploader(launch ExecutorLauncher, dir string, subDir string) string {
	a := fslib.Attr{}
	a.Pid = origDirUploaderPid(subDir)
	a.Program = "bin/fsdiruploader"
	a.Args = []string{
		ggOrig(dir, subDir, ""),
		ggRemote(subDir, ""),
		"",
	}
	a.Env = []string{}
	a.StartDep = map[string]bool{}
	a.ExitDep = map[string]bool{}
	err := launch.Spawn(&a)
	if err != nil {
		log.Fatalf("Error spawning orig dir upload worker [%v/%v]: %v\n", dir, subDir, err)
	}
	return a.Pid
}

func spawnReductionWriter(launch ExecutorLauncher, target string, targetReduction string, dstDir string, subDir string, deps []string) string {
	a := fslib.Attr{}
	a.Pid = reductionWriterPid(dstDir, subDir, target)
	a.Program = "bin/gg-target-writer"
	a.Args = []string{
		path.Join(dstDir, subDir),
		target,
		targetReduction,
	}
	a.Env = []string{}
	a.StartDep = map[string]bool{}
	reductionPid := outputHandlerPid(targetReduction)
	noOpReductionPid := noOpPid(reductionPid)
	deps = append(deps, noOpReductionPid)
	exitDepMap := map[string]bool{}
	for _, dep := range deps {
		exitDepMap[dep] = false
	}
	a.ExitDep = exitDepMap
	err := launch.Spawn(&a)
	if err != nil {
		log.Fatalf("Error spawning target writer [%v]: %v\n", target, err)
	}
	return a.Pid
}

func spawnExecutor(launch ExecutorLauncher, targetHash string, depPids []string) (string, error) {
	a := fslib.Attr{}
	a.Pid = executorPid(targetHash)
	a.Program = "bin/gg-executor"
	a.Args = []string{
		targetHash,
	}
	a.Dir = ""
	a.StartDep = map[string]bool{}
	exitDepMap := map[string]bool{}
	for _, dep := range depPids {
		exitDepMap[dep] = false
	}
	a.ExitDep = exitDepMap
	err := launch.Spawn(&a)
	if err != nil {
		return a.Pid, fmt.Errorf("Error spawning executor [%v]: %v\n", targetHash, err)
	}
	return a.Pid, nil
}

func spawnThunkOutputHandler(launch ExecutorLauncher, deps []string, thunkHash string, outputFiles []string) string {
	a := fslib.Attr{}
	a.Pid = outputHandlerPid(thunkHash)
	a.Program = "bin/gg-thunk-output-handler"
	a.Args = []string{
		thunkHash,
	}
	a.Args = append(a.Args, outputFiles...)
	a.Env = []string{}
	a.StartDep = map[string]bool{}
	exitDepMap := map[string]bool{}
	for _, dep := range deps {
		exitDepMap[dep] = false
	}
	a.ExitDep = exitDepMap
	err := launch.Spawn(&a)
	if err != nil {
		log.Fatalf("Error spawning output handler [%v]: %v\n", thunkHash, err)
	}
	return a.Pid
}

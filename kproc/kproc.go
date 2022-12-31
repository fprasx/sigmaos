package kproc

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"sigmaos/container"
	db "sigmaos/debug"
	"sigmaos/proc"
)

// To run kernel procs and mark them LC
func RunKernelProc(p *proc.Proc, namedAddr []string, contain bool) (*exec.Cmd, error) {
	db.DPrintf(db.KERNEL, "RunKernelProc %v %v\n", p, namedAddr)
	p.FinalizeEnv("NONE")
	env := p.GetEnv()
	env = append(env, "NAMED="+strings.Join(namedAddr, ","))
	env = append(env, "SIGMAPROGRAM="+p.Program)

	cmd := exec.Command(p.Program, p.Args...)
	// Create a process group ID to kill all children if necessary.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), env...)
	if contain {
		if err := container.RunKernelContainer(cmd); err != nil {
			return nil, err
		}
	} else {
		if err := cmd.Start(); err != nil {
			return nil, err
		}
	}
	return cmd, nil
}
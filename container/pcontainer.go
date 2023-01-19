package container

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	//"path"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	db "sigmaos/debug"
	"sigmaos/rand"
	// "sigmaos/seccomp"
	"sigmaos/proc"
)

const (
	UBIN = "/bin"
)

func MakeProcContainer(cmd *exec.Cmd, realmid string) error {
	// // Set up new namespaces
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	Cloneflags: syscall.CLONE_NEWUTS |
	// 		syscall.CLONE_NEWNS |
	// 		syscall.CLONE_NEWIPC |
	// 		syscall.CLONE_NEWPID |
	// 		syscall.CLONE_NEWUSER,
	// 	UidMappings: []syscall.SysProcIDMap{
	// 		{
	// 			ContainerID: 0,
	// 			HostID:      os.Getuid(),
	// 			Size:        1,
	// 		},
	// 	},
	// 	GidMappings: []syscall.SysProcIDMap{
	// 		{
	// 			ContainerID: 0,
	// 			HostID:      os.Getgid(),
	// 			Size:        1,
	// 		},
	// 	},
	// }

	pn, err := exec.LookPath("exec-container")
	if err != nil {
		return fmt.Errorf("LookPath: %v", err)
	}
	cmd.Path = pn
	cmd.Args = append([]string{PROC, realmid}, cmd.Args...)

	db.DPrintf(db.CONTAINER, "Contain proc cmd %v os env %v p\n", cmd, os.Environ())
	return nil
}

func MkContainer(p *proc.Proc, realm string) (*Container, error) {
	db.DPrintf(db.CONTAINER, "dockerContainer %v\n", realm)
	image := "sigmauser"
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	// cmd := append([]string{"exec-container", PROC, "rootrealm", uproc.Program}, uproc.Args...)
	cmd := append([]string{p.Program}, p.Args...)
	db.DPrintf(db.CONTAINER, "ContainerCreate %v %v\n", cmd, p.GetEnv())
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   cmd, //AttachStdout: true,
		// AttachStderr: true,
		Tty: false,
		Env: p.GetEnv(),
	}, nil, nil, nil, "")
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		db.DPrintf(db.CONTAINER, "ContainerCreate err %v\n", err)
		return nil, err
	}
	return &Container{ctx, cli, resp.ID, ""}, nil
}

func Pexec(uproc *proc.Proc) error {
	db.DPrintf(db.CONTAINER, "proc: %v\n", uproc)

	//wl, err := seccomp.ReadWhiteList("./whitelist.yml")
	//if err != nil {
	//	return err
	//}
	// seccomp.LoadFilter(wl)

	ip, err := LocalIP()
	if err != nil {
		return err
	}
	db.DPrintf(db.CONTAINER, "Uproc ip %v", ip)

	cmd := exec.Command(uproc.Program, uproc.Args...)
	os.Setenv("PATH", "/home/sigmaos/bin/user")
	db.DPrintf(db.UPROCSRV, "exec %v\n", cmd)
	cmd.Env = uproc.GetEnv()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

// For debugging
func ls(dir string) error {
	log.Printf("== ls %s\n", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, file := range files {
		log.Println(file.Name(), file.IsDir())
	}
	return nil
}

// XXX pair down what is being mounted; exec needs a lot, but maybe
// not all of it (e.g., usr? and only some subdirectories)
func setupFs(newRoot string) error {
	oldRootMnt := "old_root" + rand.String(8)

	// xnewRoot := newRoot
	newRoot = newRoot + "/rootrealm"

	log.Printf("new root %v\n", newRoot)

	// Mount new file system as a mount point so we can pivot_root to it later
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND, ""); err != nil {
		log.Printf("failed to mount new root filesystem: %v", err)
		return err
	}

	// Chdir to new root
	if err := syscall.Chdir(newRoot); err != nil {
		log.Printf("failed to chdir to /: %v", err)
		return err
	}

	// Make dir for oldMount
	if err := syscall.Mkdir(oldRootMnt, 0700); err != nil {
		log.Printf("failed to mkdir: %v", err)
		return err
	}

	ls(".")

	// // Mount /sys for /sys/devices/system/cpu/online; XXX exclude
	// // /sys/firmware; others?
	// if err := syscall.Mount("/sys", path.Join(newRoot, "sys"), "sysfs", syscall.MS_BIND, ""); err != nil {
	// 	log.Printf("failed to mount /sys err %v", err)
	// 	return err
	// }

	// // Mount /dev/urandom
	// if err := syscall.Mount("/dev", "dev", "none", syscall.MS_BIND|syscall.MS_RDONLY, ""); err != nil {
	// 	log.Printf("failed to mount /dev: %v", err)
	// 	return err
	// }

	// Mount /usr
	if err := syscall.Mount("/usr", "usr", "none", syscall.MS_BIND|syscall.MS_RDONLY, ""); err != nil {
		log.Printf("failed to mount /dev/usr: %v", err)
		return err
	}

	// Mount /lib
	if err := syscall.Mount("/lib", "lib", "none", syscall.MS_BIND, ""); err != nil {
		log.Printf("failed to mount /dev/lib: %v", err)
		return err
	}

	// Mount /lib
	if err := syscall.Mount("/lib64", "lib64", "none", syscall.MS_BIND, ""); err != nil {
		log.Printf("failed to mount /dev/lib64: %v", err)
		return err
	}

	// Mount /etc
	if err := syscall.Mount("/etc", "etc", "none", syscall.MS_BIND, ""); err != nil {
		log.Printf("failed to mount /etc: %v", err)
		return err
	}

	// // Mount bin/user on /bin so that user procs can run only programs from /bin/user
	// if err := syscall.Mount(path.Join(xnewRoot)+"/bin/user", path.Join(newRoot, UBIN), "none", syscall.MS_BIND, ""); err != nil {
	// 	log.Printf("failed to mount /bin: %v", err)
	// 	return err
	// }

	// pivot_root
	if err := syscall.PivotRoot(".", oldRootMnt); err != nil {
		log.Printf("failed to pivot root: %v", err)
		return err
	}

	// Chdir to new root
	if err := syscall.Chdir("/"); err != nil {
		log.Printf("failed to chdir to /: %v", err)
		return err
	}

	// Mount proc
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		log.Printf("failed to mount /proc: %v", err)
		return err
	}

	// unmount the old root filesystem
	if err := syscall.Unmount(oldRootMnt, syscall.MNT_DETACH); err != nil {
		log.Printf("failed to unmount the old root filesystem: %v", err)
		return err
	}

	// Remove the old root filesystem
	if err := os.Remove(oldRootMnt); err != nil {
		log.Printf("failed to remove old root filesystem: %v", err)
		return err
	}

	return nil
}

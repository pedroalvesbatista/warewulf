//go:build linux
// +build linux

package child

import (
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	if os.Getpid() != 1 {
		wwlog.Printf(wwlog.ERROR, "PID is not 1: %d\n", os.Getpid())
		os.Exit(1)
	}

	containerName := args[0]

	if container.ValidSource(containerName) == false {
		wwlog.Printf(wwlog.ERROR, "Unknown Warewulf container: %s\n", containerName)
		os.Exit(1)
	}

	containerPath := container.RootFsDir(containerName)

	syscall.Mount("", "/", "", syscall.MS_PRIVATE, "")
	syscall.Mount("/dev", path.Join(containerPath, "/dev"), "", syscall.MS_BIND, "")

	for _, b := range binds {
		var source string
		var dest string

		bind := strings.Split(b, ":")
		source = bind[0]

		if len(bind) == 1 {
			dest = source
		} else {
			dest = bind[1]
		}

		err := syscall.Mount(source, path.Join(containerPath, dest), "", syscall.MS_BIND, "")
		if err != nil {
			fmt.Printf("BIND ERROR: %s\n", err)
			os.Exit(1)
		}
	}

	syscall.Chroot(containerPath)
	os.Chdir("/")

	syscall.Mount("rootfs", "rootfs", "", syscall.MS_BIND, "")
	syscall.Mount("/proc", "/proc", "proc", 0, "")

	ps1string := fmt.Sprintf("[%s] Warewulf> ", containerName)
	os.Setenv("PS1", ps1string)
	os.Setenv("HISTFILE", "/dev/null")

	err := syscall.Exec(args[1], args[1:], os.Environ())
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	return nil
}

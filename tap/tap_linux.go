// +build linux

package tap

import (
    "syscall"
)

type tapData struct {
    fd int
    
}

func newTap(name string) (td *tapData, err error)  {
    syscall.IFNAMESIZ
}


func createTap(name string) {
    
}

func ioctl(fd, cmd, ptr uintptr) (error) {
    _, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
    if e != nil {
        return e
    }
    return nil
}

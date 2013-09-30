package main

// #include <stdlib.h>
// #include <pwd.h>
import "C"
import "unsafe"

type Passwd struct {
	Uid   uint32
	Gid   uint32
	Dir   string
	Shell string
}

func Getpwnam(name string) *Passwd {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cpw := C.getpwnam(cname)
	return &Passwd{
		Uid:   uint32(cpw.pw_uid),
		Gid:   uint32(cpw.pw_gid),
		Dir:   C.GoString(cpw.pw_dir),
		Shell: C.GoString(cpw.pw_shell),
	}
}

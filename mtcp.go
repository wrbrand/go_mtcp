package mtcp

/* 
#cgo LDFLAGS: -lmtcp -L /usr/local/mtcp/lib -lpthread -lnuma -lrt
#cgo CFLAGS: -I /usr/local/mtcp/include
#define _GNU_SOURCE
#include <mtcp_api.h>

int zmtcp_getsockname(int cpuid, int fd, void *rsa, void *slen) {
  return mtcp_getsockname((mctx_t) &cpuid, fd, (struct sockaddr *) rsa, (socklen_t *) slen);
}

int zmtcp_write(int cpuid, int fd, void *p, unsigned len) {
  return mtcp_write((mctx_t) &cpuid, fd, (char *) p, len);
}

int zmtcp_connect(int cpuid, int fd, void *p, unsigned len) {
  return mtcp_connect((mctx_t) &cpuid,fd, (const struct sockaddr *) p, len);
}
*/
import "C"
import "unsafe"

var cpuid C.int

/*  To be called when setting affinity; does not actually update 
    underlying mtcp state, just caches a value to avoid syscalls */
func SetCPU(cpu int) {
  cpuid = C.int(cpu)
}

type Sockaddr interface {
  sockaddr() (ptr unsafe.Pointer, len _Socklen, err error)
}

type SockaddrInet4 struct {
  Port int
  Addr [4]byte
  raw RawSockaddrInet4
}

type SockaddrInet6 struct {
  Port int
  ZoneId uint32
  Addr [16]byte
  raw RawSockaddrInet6
}

func (sa *SockaddrInet4) sockaddr() (unsafe.Pointer, _Socklen, error) {
	if sa.Port < 0 || sa.Port > 0xFFFF {
		return nil, 0, EINVAL
	}
	sa.raw.Family = AF_INET
	p := (*[2]byte)(unsafe.Pointer(&sa.raw.Port))
	p[0] = byte(sa.Port >> 8)
	p[1] = byte(sa.Port)
	for i := 0; i < len(sa.Addr); i++ {
		sa.raw.Addr[i] = sa.Addr[i]
	}
	return unsafe.Pointer(&sa.raw), SizeofSockaddrInet4, nil
}

func (sa *SockaddrInet6) sockaddr() (unsafe.Pointer, _Socklen, error) {
	if sa.Port < 0 || sa.Port > 0xFFFF {
		return nil, 0, EINVAL
	}
	sa.raw.Family = AF_INET6
	p := (*[2]byte)(unsafe.Pointer(&sa.raw.Port))
	p[0] = byte(sa.Port >> 8)
	p[1] = byte(sa.Port)
	sa.raw.Scope_id = sa.ZoneId
	for i := 0; i < len(sa.Addr); i++ {
		sa.raw.Addr[i] = sa.Addr[i]
	}
	return unsafe.Pointer(&sa.raw), SizeofSockaddrInet6, nil
}

func Getsockname(fd int, rsa unsafe.Pointer) (err error) {
  var slen _Socklen = SizeofSockaddrAny
  e := C.zmtcp_getsockname(C.int(cpuid), C.int(fd), rsa, unsafe.Pointer(&slen))
  if e != 0 {
    panic("Error calling mtcp_getsockname")
  }
  return nil
}

func Write(fd int, p []byte, l int) (n int, err error) {
  e := C.zmtcp_write(C.int(cpuid), C.int(fd), unsafe.Pointer(&p), C.uint(l))
  if e < 0 {
    panic("Error calling mtcp_write")
  }
  return int(e), nil
}

func Connect(fd int, p unsafe.Pointer, l _Socklen) (err error) {
  e := C.zmtcp_connect(C.int(cpuid), C.int(fd), p, C.uint(l))
  if e < 0 {
    panic("Error calling mtcp_connect")
  }
  return nil
}

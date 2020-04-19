package mtcp

/* 
#cgo LDFLAGS: -lmtcp -L /usr/local/mtcp/lib -lpthread -lnuma -lrt
#cgo CFLAGS: -I /usr/local/mtcp/include
#define _GNU_SOURCE
#include <mtcp_api.h>

int zmtcp_getsockname(int cpuid, int fd, void *rsa, void *slen) {
  return mtcp_getsockname((mctx_t) &cpuid, fd, (struct sockaddr *) rsa, (socklen_t *) slen);
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

func Getsockname(fd int) (sa Sockaddr, err error) {
  var rsa RawSockaddrAny
  var slen _Socklen = SizeofSockaddrAny
  e := C.zmtcp_getsockname(C.int(cpuid), C.int(fd), unsafe.Pointer(&rsa), unsafe.Pointer(&slen))
  if e != 0 {
    panic("Error calling mtcp_getsockname")
  }

  // Clipped from anyToSockaddr
  switch rsa.Addr.Family {
  case AF_INET: // AF_INET
    pp := (*RawSockaddrInet4)(unsafe.Pointer(&rsa))
    sa := new(SockaddrInet4)
    p := (*[2]byte)(unsafe.Pointer(&pp.Port))
    sa.Port = int(p[0])<< 8 + int(p[1])
    for i := 0; i < len(sa.Addr); i++ {
      sa.Addr[i] = pp.Addr[i]
    }
    return sa, nil
  case AF_INET6: // AF_INET6
    pp := (*RawSockaddrInet6)(unsafe.Pointer(&rsa))
    sa := new(SockaddrInet6)
    p := (*[2]byte)(unsafe.Pointer(&pp.Port))
    sa.Port = int(p[0])<<8 + int(p[1])
    sa.ZoneId = pp.Scope_id
    for i := 0; i < len(sa.Addr); i++ {
      sa.Addr[i] = pp.Addr[i]
    }
    return sa, nil
  }
  return nil, EAFNOSUPPORT
}

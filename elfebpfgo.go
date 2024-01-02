//go:build org
package elfebpfgo

/*
#cgo LDFLAGS: -lelf -lz
#include "elfebpfgo.h"
*/
import "C"

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"
)

type BPFModule struct {
	obj *C.struct_bpf_object
	elf *elf.File
}

type BPFProg struct {
	name   string
	prog   *C.struct_bpf_program
	module *BPFModule
}
type BPFLink struct {
	link     *C.struct_bpf_link
	prog     *BPFProg
	linkType int
}

type BPFMap struct {
	name   string
	bpfMap *C.struct_bpf_map
	fd     C.int
	module *BPFModule
}

func CreateBPFModule(bpfObjPath string) (*BPFModule, error) {
	f, err := elf.Open(bpfObjPath)
	if err != nil {
		return nil, err
	}

	bpfFile := C.CString(bpfObjPath)
	defer C.free(unsafe.Pointer(bpfFile))

	obj, errno := C.bpf_object__open(bpfFile)
	if obj == nil {
		return nil, fmt.Errorf("failed to open BPF object at path %s: %w",bpfObjPath , errno)
	}

	return &BPFModule{
		obj: obj,
		elf: f,
	}, nil

}

func (m *BPFModule) LoadBpfObj() error {
	ret := C.bpf_object__load(m.obj)
	if ret != 0 {
		return fmt.Errorf("failed to load BPF object: %w", syscall.Errno(-ret))
	}
	defer m.elf.Close()

	return nil
}

func (m *BPFModule) FindProgByName(prog string) (*BPFProg, error) {
	cs := C.CString(prog)
	program, errno := C.bpf_object__find_program_by_name(m.obj, cs)
	C.free(unsafe.Pointer(cs))
	if program == nil {
		return nil, fmt.Errorf("failed to find BPF program %s: %w", prog, errno)
	}

	return &BPFProg{
		name:   prog,
		prog:   program,
		module: m,
	}, nil

}

func (p *BPFProg) AttachGeneric() (*BPFLink, error) {
	link, errno := C.bpf_program__attach(p.prog)
	if link == nil {
		return nil, fmt.Errorf("failed to attach program: %w", errno)
	}
	bpfLink := &BPFLink{
		link:     link,
		prog:     p,
		linkType: 1,
	}
	return bpfLink, nil
}

func (m *BPFModule) GetMap(mapName string) (*BPFMap, error) {
	cs := C.CString(mapName)
	bpfMap, errno := C.bpf_object__find_map_by_name(m.obj, cs)
	C.free(unsafe.Pointer(cs))
	if bpfMap == nil {
		return nil, fmt.Errorf("failed to find BPF map %s: %w", mapName, errno)
	}

	return &BPFMap{
		bpfMap: bpfMap,
		name:   mapName,
		fd:     C.bpf_map__fd(bpfMap),
		module: m,
	}, nil
}


func (bm *BPFMap) KeySize() int {
	return int(C.bpf_map__key_size(bm.bpfMap))
}



func (bm *BPFMap) Next(prevPtr unsafe.Pointer) bool{
	//prevPtr := unsafe.Pointer(nil)


	next := make([]byte, bm.KeySize())
	nextPtr := unsafe.Pointer(&next[0])

	_, err := C.bpf_map_get_next_key(bm.fd, prevPtr, nextPtr)

	if errno, ok := err.(syscall.Errno); ok && errno == C.ENOENT {
		return false
	}
  

   
  prevPtr = nextPtr    

  fmt.Println(binary.LittleEndian.Uint32(next))

	return true
}

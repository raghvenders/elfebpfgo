//go:build org

package main

import "C"

import (
	"log"
	"os"
	"time"
	"unsafe"

	"fmt"

	"github.com/raghvenders/elfebpfgo"
)

func main() {

	bpfModule, err := elfebpfgo.CreateBPFModule("main.bpf.o")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	//defer bpfModule.Close()

	fmt.Println("----- Program Loaded")

	err = bpfModule.LoadBpfObj()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	fmt.Println("Object loaded")
	prog, err := bpfModule.FindProgByName("sys_enter_kill")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	fmt.Println("FUnc cal-----")

	 l, err := prog.AttachGeneric()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

  _ = l
	

	fmt.Println("Attached:::::")

	killMap, err := bpfModule.GetMap("kill_map")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	//fd := killMap.FileDescriptor()

	//fmt.Println("------", fd)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
 
  prevPtr := unsafe.Pointer(nil)
	for range ticker.C {
			for killMap.Next(prevPtr) {
			log.Println("-------")

		}
	}

	/*
			eventsChannel := make(chan []byte)
			ringBuf, err := bpfModule.InitRingBuf("kill_map", eventsChannel)
			if err != nil {
				fmt.Println("couldn't init ringbuffer")
				os.Exit(-1)
			}
			ringBuf.Start()

			i := 0
		thisloop:
			for {
				b := <-eventsChannel
				result := binary.LittleEndian.Uint32(b)
				log.Println("-------", result)
				i++

				if i == 3 {
					break thisloop
				}

			}
	*/

}

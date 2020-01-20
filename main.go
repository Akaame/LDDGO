package main

import (
	"os"
	"log"
	"debug/elf"
	"fmt"
	"strings"
)

var searchPath []string
var soPath map[string]string

func main() {
	if len(os.Args) < 2 {
		panic("Invalid usage")
	}

	soPath = make(map[string]string)

	fd, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Populate ELF fields
	bin, err := elf.NewFile(fd)
	if err != nil {
		log.Fatal(err)
	}

	// Check ELF Class
	if bin.Class == elf.ELFCLASS32 {
		searchPath = append(searchPath,"/lib/")
	} else
	if bin.Class == elf.ELFCLASS64 {
		searchPath = append(searchPath,"/lib64/")
	} else {
		log.Fatal("Unknown ELF Class")
	}

	// env parse LD_LIBRARY_PATH
	env := os.Getenv("LD_LIBRARY_PATH")
	paths := strings.Split(env, ":")
	searchPath = append(searchPath, paths...)
	// SO files are searched through LD_LIBRARY_PATH and lib/lib64

	// Get list of needed shared libraries
	dynSym, err := bin.DynString(elf.DT_NEEDED)
	if err != nil {
		log.Fatal(err)
	}
	// Recurse
	recurseDynStrings(dynSym)

	for k, v := range soPath {
		fmt.Println(k, ": ", v)
	}
}

func recurseDynStrings(dynSym []string) {
	for _, el := range dynSym {
		// fmt.Println(el)
		// check file path here for library if it doesnot exists panic
		var fd *os.File
		for _, entry := range searchPath {
			path := entry+el
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				fd, err = os.OpenFile(path, os.O_RDONLY, 0644)
				if err != nil {
					log.Fatal(err)
				} else {
					soPath[el] = path
				}
			} else {
				// Nothing
			}		
		}

		bint, err := elf.NewFile(fd)
		if err != nil {
			log.Fatal(err)
		}
		
		bDynSym, err := bint.DynString(elf.DT_NEEDED)
		if err != nil {
			log.Fatal(err)
		}
		
		recurseDynStrings(bDynSym)
	}	
}

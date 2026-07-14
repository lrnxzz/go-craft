package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/lrnxzz/go-craft/viewer"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	screenshot := flag.String("screenshot", "", "render a single frame to this PNG and exit")
	flag.Parse()

	view, err := viewer.New(*screenshot == "")
	if err != nil {
		log.Fatal(err)
	}

	if *screenshot != "" {
		if err := view.Screenshot(*screenshot); err != nil {
			log.Fatal(err)
		}
		log.Printf("wrote %s", *screenshot)

		return
	}

	view.Run()
}

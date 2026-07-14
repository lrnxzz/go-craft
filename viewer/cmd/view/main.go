package main

import (
	"context"
	"flag"
	"log"
	"runtime"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/viewer"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	screenshot := flag.String("screenshot", "", "render a single frame to this PNG and exit")
	username := flag.String("username", "gocraft_view", "bot username")
	flag.Parse()

	address := flag.Arg(0)
	if address == "" {
		log.Fatal("usage: view [flags] <host[:port]>")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot, err := agent.Join(ctx, address, *username)
	if err != nil {
		log.Fatal(err)
	}

	spawn := make(chan gocraft.Vec3d, 1)
	bot.OnSpawn(func() {
		spawn <- bot.Player().Position
	})

	go bot.Run(ctx)

	focus := <-spawn
	time.Sleep(3 * time.Second)

	view, err := viewer.New(bot.World(), focus, *screenshot == "")
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

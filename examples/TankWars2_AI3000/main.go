package main

import (
	"TankWars2_AI/ai"
	"TankWars2_AI/gui"
	"TankWars2_AI/remote"
	"os"
	"time"
)

func main() {

	// config
	headless := false
	host := "127.0.0.1"
	port := "1234"

	// params
	if len(os.Args) >= 2 {
		host = os.Args[1]
		println(host)
	}
	if len(os.Args) >= 3 {
		headless = false
		println(headless)
	}

	// new client
	client, err := remote.NewClient(host, port)
	if err != nil {
		println(err.Error())
		os.Exit(11)
	}

	// load world from server
	world, _ := client.Status()

	// enable AI
	go ai.RunAI(client)

	// run gui (blocking)
	if !headless {
		// run gui
		if err := gui.RunGame("Tank Wars AI 3000", world, client); err != nil {
			panic(err)
		}
	} else {
		// hide gui
		for {
			time.Sleep(10 * time.Second)
		}
	}
}

package main

import (
	"flag"
	"fmt"
	"github.com/SchnorcherSepp/TankWars2/ai"
	"github.com/SchnorcherSepp/TankWars2/core"
	"github.com/SchnorcherSepp/TankWars2/gui"
	"github.com/SchnorcherSepp/TankWars2/gui/resources"
	"github.com/SchnorcherSepp/TankWars2/maps"
	"github.com/SchnorcherSepp/TankWars2/remote"
	"os"
	"time"
)

const VERSION = "2.2"

func main() {
	parseMode() // cli
}

//--------------------------------------------------------------------------------------------------------------------//

func parseMode() {

	// print program name and version
	println("TankWars " + VERSION)
	println()

	// help text for mode
	help := "Choose mode: local, server, client, editor"

	// check args
	if len(os.Args) < 2 {
		println(help)
		os.Exit(3)
	}

	// cut mode from args (flag parse don't work without)
	mode := os.Args[1]
	os.Args[1] = os.Args[0]
	os.Args = os.Args[1:]

	// select mode
	switch mode {
	case "local":
		parseLocal()
	case "server":
		parseServer()
	case "client":
		parseClient()
	case "editor":
		parseEditor()
	default:
		println(help)
		os.Exit(4)
	}
}

func parseLocal() {
	var mapFile string
	var mute bool

	// parse
	flag.StringVar(&mapFile, "map", "", "Path to map file")
	flag.BoolVar(&mute, "mute", false, "Mute sound")
	flag.Parse()

	// enforce map
	if mapFile == "" {
		flag.Usage()
		os.Exit(5)
	}

	// run program
	runLocal(mapFile, mute)
}

func parseServer() {
	var mapFile string
	var host string
	var port string
	var headless bool
	var mute bool

	// parse
	flag.StringVar(&mapFile, "map", "", "Path to map file")
	flag.StringVar(&host, "host", "", "Server host")
	flag.StringVar(&port, "port", "", "Server port")
	flag.BoolVar(&headless, "headless", false, "Run in headless mode")
	flag.BoolVar(&mute, "mute", false, "Mute sound")
	flag.Parse()

	// enforce map, host and port
	if mapFile == "" || host == "" || port == "" {
		flag.Usage()
		os.Exit(6)
	}

	// run program
	runServer(mapFile, host, port, headless, mute)
}

func parseClient() {
	var host string
	var port string
	var basicAI bool
	var headless bool

	// parse
	flag.StringVar(&host, "host", "", "Server host")
	flag.StringVar(&port, "port", "", "Server port")
	flag.BoolVar(&basicAI, "ai", false, "Use basic AI")
	flag.BoolVar(&headless, "headless", false, "Run in headless mode")
	flag.Parse()

	// enforce host and port
	if host == "" || port == "" {
		flag.Usage()
		os.Exit(7)
	}

	// run program
	runClient(host, port, basicAI, headless)
}

func parseEditor() {
	var mapFile string
	var newWidth int
	var newHeight int

	// parse
	flag.StringVar(&mapFile, "map", "", "Path to map file")
	flag.IntVar(&newWidth, "width", 15, "New width")
	flag.IntVar(&newHeight, "height", 8, "New height")
	flag.Parse()

	// enforce map
	if mapFile == "" {
		flag.Usage()
		os.Exit(8)
	}

	// run program
	runEditor(mapFile, newWidth, newHeight)
}

//--------------------------------------------------------------------------------------------------------------------//

func runLocal(mapFile string, mute bool) {
	title := fmt.Sprintf("Tank Wars %s (%s)", VERSION, "Local")

	// load map
	world, err := maps.Loader(mapFile)
	if err != nil {
		println("err: invalid map:", err.Error())
		os.Exit(9)
	}

	// run gui (blocking)
	if err := gui.RunGame(title, world, nil, mute); err != nil {
		panic(err)
	}
}

func runServer(mapFile, host, port string, headless, mute bool) {
	title := fmt.Sprintf("Tank Wars %s (%s)", VERSION, "Server")

	// load map
	world, err := maps.Loader(mapFile)
	if err != nil {
		println("err: invalid map:", err.Error())
		os.Exit(10)
	}

	// run server
	playerCount := world.PlayerCount()
	go remote.RunServer(host, port, world, playerCount)

	// run gui/server (blocking)
	if !headless {
		// GUI
		if err := gui.RunGame(title, world, nil, mute); err != nil {
			panic(err)
		}
	} else {
		// headless
		resources.MuteSound = true // play no sound without GUI
		for {
			world.Update()
			time.Sleep(time.Second / core.GameSpeed)
		}
	}
}

func runClient(host, port string, basicAI, headless bool) {
	title := fmt.Sprintf("Tank Wars %s (%s)", VERSION, "Client")

	// new client
	client, err := remote.NewClient(host, port)
	if err != nil {
		println(err.Error())
		os.Exit(11)
	}

	// load world from server
	world := client.Status()

	// enable AI
	if basicAI {
		go ai.RunAI(client)
	}

	// run gui (blocking)
	if !headless {
		const mute = true
		if err := gui.RunGame(title, world, client, mute); err != nil {
			panic(err)
		}
	} else {
		for {
			time.Sleep(1 * time.Second)
		}
	}

}

func runEditor(mapFile string, newWidth, newHeight int) {
	gui.RunEditor(mapFile, core.NewWorld(newWidth, newHeight))
}

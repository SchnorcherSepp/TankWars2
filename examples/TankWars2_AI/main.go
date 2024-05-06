package main

import (
	"TankWars2_AI/ai2000"
	"TankWars2_AI/gui"
	"TankWars2_AI/remote"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {

	headless := false
	host := "0.0.0.0"
	port := "1234"

	runServerAndTestAi := false
	twMap := "1v1_borderdispute_15x08.json" //  1v1_borderdispute_15x08.json  or  1v1_riverisland_21x13.json
	disable := false

	//######################################################################

	if len(os.Args) >= 2 {
		if os.Args[1] == "true" || os.Args[1] == "True" || os.Args[1] == "TRUE" || os.Args[1] == "1" {
			headless = true
		} else {
			headless = false
		}
	}
	if len(os.Args) >= 3 {
		host = os.Args[2]
	}
	if len(os.Args) >= 4 {
		port = os.Args[3]
	}
	if len(os.Args) >= 5 {
		if os.Args[4] == "true" || os.Args[4] == "True" || os.Args[4] == "TRUE" || os.Args[4] == "1" {
			runServerAndTestAi = true
		} else {
			runServerAndTestAi = false
		}
	}
	if len(os.Args) >= 6 {
		twMap = os.Args[5]
	}
	if len(os.Args) >= 7 {
		if os.Args[6] == "true" || os.Args[6] == "True" || os.Args[6] == "TRUE" || os.Args[6] == "1" {
			disable = true
		} else {
			disable = false
		}
	}

	fmt.Printf("args: headless=%v host=%s port=%s runServerAndTestAi=%v twMap=%s disable=%v\n", headless, host, port, runServerAndTestAi, twMap, disable)

	//-------------------------------------------------------------------------

	if runServerAndTestAi {
		// start server
		go runServer(twMap, host, port)
		time.Sleep(500 * time.Millisecond)

		// start test ai
		go func() {
			time.Sleep(700 * time.Millisecond)
			runBasicAiClient(disable, host, port)
		}()
	}

	runAi2000Client(disable, headless, host, port)
	if headless {
		for {
			time.Sleep(10 * time.Second)
		}
	}
}

//--------------------------------------------------------------------------------------------------------------------//

func runServer(mapFile, host, port string) {
	twExe, err := filepath.Abs(`Server\TankWars2.exe`)
	if err != nil {
		panic(err)
	}
	twMap, err := filepath.Abs("Server/maps/" + mapFile)
	if err != nil {
		panic(err)
	}
	command := exec.Command(twExe, "server", "-host", host, "-port", port, "-map", twMap, "-mute")
	command.Stderr = os.Stdout
	command.Stdout = os.Stdout
	err = command.Run()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func runBasicAiClient(disable bool, host, port string) {
	twExe, err := filepath.Abs(`Server\TankWars2.exe`)
	if err != nil {
		panic(err)
	}

	command := exec.Command(twExe, "client", "-host", host, "-port", port, "-ai", "-headless")
	if disable {
		command = exec.Command(twExe, "client", "-host", host, "-port", port, "-headless")
	}

	command.Stderr = os.Stdout
	command.Stdout = os.Stdout
	err = command.Run()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func runAi2000Client(disable, headless bool, host, port string) {

	// new client
	client, err := remote.NewClient(host, port)
	if err != nil {
		println(err.Error())
		os.Exit(11)
	}

	// load world from server
	world := client.Status()

	// enable AI
	go ai2000.RunAI(client, disable)

	// run gui (blocking)
	const mute = true
	if !headless {
		if err := gui.RunGame("Tank Wars AI 2000", world, client, mute); err != nil {
			panic(err)
		}
	}
}

package remote

/*
  This file provides a Server struct and related methods for handle client connections to this game server.
*/

import (
	"TankWars2/core"
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strconv"
	"strings"
)

// RunServer runs a server (BLOCKING!).
// The server receives commands from the clients and implements them in "World".
// The first connecting client controls player 1.
// The second connecting client controls player 2, and so on.
func RunServer(host, port string, world *core.World, maxPlayer int) {
	world.Freeze = true // wait for all player

	// Listen for incoming connections.
	l, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("RunServer: %v\n", err)
	}

	// Close the listener when the application closes.
	defer func(l net.Listener) {
		_ = l.Close()
	}(l)

	// start server
	fmt.Println("START SERVER [" + host + ":" + port + "]")
	for player := uint8(1); player < 50; player++ {

		// wait for incoming connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}

		// Handle connections in a new goroutine.
		go handleRequest(conn, world, player)
		fmt.Printf("player %d from %v\n", player, conn.RemoteAddr())

		// start game with all player
		if int(player) == maxPlayer {
			world.Freeze = false // START GAME
			fmt.Printf("START GAME\n")
		}
	}
}

// handleRequest handles incoming requests from a client.
func handleRequest(conn net.Conn, w *core.World, player uint8) {

	// prepare line reader
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	// close at end
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	// loop
	for {
		// read one line (ended with \n or \r\n)
		line, err := tp.ReadLine()
		if err != nil {
			break // EXIT
		}

		// trim line and split args
		args := strings.Split(strings.TrimSpace(line), " ")

		// extract com
		var com string
		if len(args) > 0 {
			com = args[0]
		}

		// CHECK COMMANDS
		switch com {
		case "PLAYER":
			comResponse(conn, strconv.Itoa(int(player)))
		case "STATUS":
			world := core.Censorship(w, player)
			comResponse(conn, world.Json())
		case core.FIRE:
			x1, y1, x2, y2 := saveNums(args)
			comResponseErr(conn, w.Fire(w.Tile(x1, y1), w.Tile(x2, y2), player))
		case core.MOVE:
			x1, y1, x2, y2 := saveNums(args)
			_, err = w.Move(w.Tile(x1, y1), w.Tile(x2, y2), player)
			comResponseErr(conn, err)
		default:
			comResponse(conn, "err: invalid command")
		}
	}

	// exit
	fmt.Printf("player %d has left\n", player)
}

//--------  Helper  --------------------------------------------------------------------------------------------------//

// comResponse is a helper function and sends messages back to the clients.
func comResponse(conn net.Conn, s string) {
	_, err := conn.Write([]byte(fmt.Sprintf("%s\r\n", s)))
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
}

// comResponseErr is a helper function and sends messages back to the clients with error handling.
func comResponseErr(conn net.Conn, err error) {
	if err != nil {
		comResponse(conn, err.Error())
	} else {
		comResponse(conn, "OK")
	}
}

// saveArgs is a helper function and returns 4 string arguments from the client commands.
func saveArgs(args []string) (a1, a2, a3, a4 string) {
	sArgs := make([]string, 5)
	copy(sArgs, args)
	return sArgs[1], sArgs[2], sArgs[3], sArgs[4]
}

// saveNums is a helper function and returns 4 integer arguments from the client commands.
func saveNums(args []string) (n1, n2, n3, n4 int) {
	a1, a2, a3, a4 := saveArgs(args)
	n1, _ = strconv.Atoi(a1)
	n2, _ = strconv.Atoi(a2)
	n3, _ = strconv.Atoi(a3)
	n4, _ = strconv.Atoi(a4)
	return
}

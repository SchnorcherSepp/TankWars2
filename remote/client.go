package remote

/*
  This file provides a Client struct and related methods for establishing a remote connection to the game server.
*/

import (
	"TankWars2/core"
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client represents a remote connection to the game server, allowing communication and interaction with the game world.
type Client struct {
	conn *net.TCPConn      // TCP connection to the game server
	tp   *textproto.Reader // Text protocol reader for the connection
	mux  *sync.Mutex       // Mutex for thread-safe operations

	world *core.World // Current game world status
}

// NewClient creates a new Client instance and establishes a connection to the game server at the provided host and port.
// It initializes the TCP connection and polls the world status every 100ms.
func NewClient(host, port string) (*Client, error) {

	// Resolve TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		return nil, err
	}

	// Establish TCP connection
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	// Create a new Client instance
	c := &Client{
		conn: conn,
		tp:   textproto.NewReader(bufio.NewReader(conn)),
		mux:  new(sync.Mutex),
	}

	// Start a goroutine to continuously update the game world
	go func(c *Client) {
		errCount := 0
		for {
			// update world
			if nil != c.updateWorld() {
				errCount++
			}
			// check error count
			if errCount >= 3 {
				println("CLIENT: exit update loop")
				break
			}
			// sleep
			time.Sleep(100 * time.Millisecond)
		}
	}(c)

	// return
	return c, nil
}

// Player returns the player's ID associated with this client session (see core.PLAYERS).
func (c *Client) Player() uint8 {
	c.mux.Lock()
	defer c.mux.Unlock()

	resp := c.command("PLAYER")

	value, err := strconv.ParseUint(resp, 10, 8)
	if err != nil {
		fmt.Println("err:", err)
		return 0
	}

	return uint8(value)
}

// Status returns the current game world status as new World object.
// This status is the censored version with the information visible to this player (see core.Censorship).
func (c *Client) Status() *core.World {

	// wait for world (max 1 sec)
	for n := 0; n < 10; n++ {
		if c.world != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// return world
	return c.world
}

// Fire sends a 'Fire' command to the game server to initiate an attack from one tile to another.
// (see Fire methode from core.World)
func (c *Client) Fire(fromX, fromY, toX, toY int) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	resp := c.command(fmt.Sprintf("%s %d %d %d %d", core.FIRE, fromX, fromY, toX, toY))
	if resp == "OK" {
		return nil // success
	} else {
		return fmt.Errorf("err: %s", resp)
	}
}

// Move sends a 'Move' command to the game server to move a unit from one tile to another.
// (see Move methode from core.World)
func (c *Client) Move(fromX, fromY, toX, toY int) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	resp := c.command(fmt.Sprintf("%s %d %d %d %d", core.MOVE, fromX, fromY, toX, toY))
	if resp == "OK" {
		return nil // success
	} else {
		return fmt.Errorf("err: %s", resp)
	}
}

//---------------- HELPER --------------------------------------------------------------------------------------------//

// command send the cmd to the server and return the response
func (c *Client) command(cmd string) string {
	if c == nil || c.conn == nil || c.tp == nil {
		return "err: TcpClient connection closed."
	}

	// remove protocol break
	cmd = strings.ReplaceAll(cmd, "\n", "")
	cmd = strings.ReplaceAll(cmd, "\r", "")
	cmd = strings.ReplaceAll(cmd, "  ", " ")

	// send command
	_, err := c.conn.Write([]byte(fmt.Sprintf("%s\r\n", cmd)))
	if err != nil {
		return fmt.Sprintf("err: TcpClient write: %v", err)
	}

	// read response
	resp, err := c.tp.ReadLine()
	if err != nil {
		return fmt.Sprintf("err: TcpClient read: %v", err)
	}

	// return server response
	return resp
}

// updateWorld retrieves the current game world status from the server and override the local world instance.
func (c *Client) updateWorld() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	// request status
	wJSON := c.command("STATUS")
	retWorld := core.World{}

	// parse JSON
	if err := json.Unmarshal([]byte(wJSON), &retWorld); err != nil {
		println("err:", wJSON)
		return err
	}

	// set new world
	c.world = &retWorld
	return nil
}

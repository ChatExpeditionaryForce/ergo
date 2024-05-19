package irc

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const PORT = "127.0.0.1:22843"

type Grumble struct {
	connections map[string]net.Conn
	server      *Server
}

func (instance *Grumble) grumbleMessage(action string, extra ...string) {
	if instance == nil {
		return
	}
	str := fmt.Sprintf("%s %s\n", action, strings.Join(extra, " "))
	for _, conn := range instance.connections {
		conn.Write([]byte(str))
	}
}

func (instance *Grumble) kickFromGrumble(channel string, user string) {
	instance.grumbleMessage("KICK", channel, user)
}

func grumbleConnection(server *Server) *Grumble {
	// create a tcp listener on the given port
	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println("Unable to open Grumble listener:", err)
		os.Exit(1)
	}
	fmt.Printf("Grumble listener on %s active\n", PORT)
	instance := Grumble{connections: make(map[string]net.Conn), server: server}
	go grumbleListener(listener, &instance)

	return &instance
}

func grumbleListener(listener net.Listener, instance *Grumble) {
	// listen for new connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("failed to accept grumble connection, err:", err)
			continue
		}
		instance.connections[conn.RemoteAddr().String()] = conn

		// pass an accepted connection to a handler goroutine
		go handleConnection(conn, instance)
	}
}

func handleConnection(conn net.Conn, instance *Grumble) {
	defer delete(instance.connections, conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)
	println("Connection with Grumble established")
	for {
		// read client request data
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err != io.EOF {
				fmt.Println("failed to read data, err:", err)
			}
			return
		}
		convertedLine := string(bytes[:len(bytes)-1])
		line := strings.Split(strings.Trim(convertedLine, " \n"), " ")
		fmt.Printf("grumble: %+q\n", line)
		switch line[0] {
		case "PART":
			if len(line) == 1 || len(line[1]) == 0 {
				println("skipping malformed line")
				continue
			}
			channel := instance.server.channels.Get(line[1])
			for _, member := range channel.Members() {
				for _, session := range member.Sessions() {
					session.Send(nil, member.server.name, "VOICEPART", line[1], line[2])
				}
			}
			break
		case "VOICESTATE":
			channel := instance.server.channels.Get(line[1])
			for _, member := range channel.Members() {
				for _, session := range member.Sessions() {
					session.Send(nil, member.server.name, "VOICESTATE", line[1], line[2], line[3], line[4])
				}
			}
		default:
			println("Unknown grumble message: ", line[0])
		}

	}
}

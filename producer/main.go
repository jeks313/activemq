package main

import (
	"fmt"
	"os"

	stomp "github.com/go-stomp/stomp"
)

//Connect to ActiveMQ and produce messages
func main() {
	conn, err := stomp.Dial("tcp", "localhost:61613")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Producer(conn, 1000)
	err = conn.Disconnect()
	if err != nil {
		fmt.Println(err)
	}
}

func Producer(conn *stomp.Conn, count int) {
	var n int
	for {
		err := conn.Send(
			"/topic/VirtualTopic.Test",
			"text/plain",
			[]byte(fmt.Sprintf("Test message #%d", n)))
		n++
		if err != nil {
			fmt.Println(err)
			return
		}
		if n >= count {
			return
		}
	}
}

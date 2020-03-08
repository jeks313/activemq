package main

import (
	"fmt"

	stomp "github.com/go-stomp/stomp"
)

// Connect to ActiveMQ and listen for messages to archive
func main() {
	conn, err := stomp.Dial("tcp", "localhost:61613")
	if err != nil {
		fmt.Println(err)
	}

	sub, err := conn.Subscribe("/queue/Consumer.Test.VirtualTopic.Test", stomp.AckAuto)
	if err != nil {
		fmt.Println(err)
	}

	for {
		msg := <-sub.C
		if msg == nil {
			break
		}
		fmt.Println(string(msg.Body))
	}

	err = sub.Unsubscribe()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Disconnect()
}

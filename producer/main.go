package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/bxcodec/faker"
	stomp "github.com/go-stomp/stomp"
	"github.com/google/uuid"
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

func randomUUIDs(num int) []string {
	var uuids []string
	for i := 0; i < num; i++ {
		id := uuid.New()
		log.Printf("created uuid: %s", id.String())
		uuids = append(uuids, id.String())
	}
	return uuids
}

// just some random fakey struct to put some JSON data in the topic
type fakey struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Time        time.Time `json:"time"`
	SoftwareIDs []int     `json:"software_id"`
	Model       string    `json:"model"`
}

const numUUIDs = 10

func Producer(conn *stomp.Conn, count int) {
	var n int
	uuids := randomUUIDs(numUUIDs)
	var d fakey
	var err error
	for {
		err = faker.FakeData(&d)
		if err != nil {
			log.Fatal("failed to fake my data")
		}
		d.UUID = uuids[rand.Intn(numUUIDs)]
		b, err := json.Marshal(&d)
		if err != nil {
			log.Fatal("failed to convert fake doc to JSON")
		}
		err = conn.Send("/topic/VirtualTopic.Test", "text/plain", b)
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

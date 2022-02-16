package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	tagsChannel := make(chan string)
	quitChannel := make(chan os.Signal, 1)
	// Create an abstraction of the Reader, DeviceConnection string is empty if you want the library to autodetect your reader
	rfidReader, err := NewTagReader("")
	if err != nil {
		log.Println("Could not create NFC tag reader: ", err)

		if err := ResetDevice(); err != nil {
			log.Println("Could not reset device: ", err)
		}

		return
	}

	// Listen for an RFID/NFC tag in another goroutine and check every 1 second
	frequencyTicker := time.NewTicker(1 * time.Second)
	go rfidReader.ListenForTags(tagsChannel, frequencyTicker)

	for {
		// fmt.Printf("%s: Waiting for a tag \n", time.Now().String())
		select {
		case tagId := <-tagsChannel:
			fmt.Println("Tag ID: ", tagId)
			continue
		case <-quitChannel:
			frequencyTicker.Stop()
			if err := rfidReader.Close(); err != nil {
				fmt.Println("Could not close reader ", err)
			}
			break
		default:
			time.Sleep(time.Millisecond * 300)
			continue
		}
	}
}

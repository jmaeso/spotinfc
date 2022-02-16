package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	nfc "github.com/clausecker/nfc/v2"
	"github.com/warthog618/gpiod"
)

const resetPin = 19

//Listen for all the modulations specified
var modulations = []nfc.Modulation{
	{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106},
	{Type: nfc.ISO14443b, BaudRate: nfc.Nbr106},
	{Type: nfc.Felica, BaudRate: nfc.Nbr212},
	{Type: nfc.Felica, BaudRate: nfc.Nbr424},
	{Type: nfc.Jewel, BaudRate: nfc.Nbr106},
	{Type: nfc.ISO14443biClass, BaudRate: nfc.Nbr106},
}

type TagReader struct {
	device *nfc.Device
}

func NewTagReader(deviceConnection string) (*TagReader, error) {
	dev, err := nfc.Open(deviceConnection)
	if err != nil {
		return nil, fmt.Errorf("Cannot communicate with the device: %w", err)
	}

	if err := dev.InitiatorInit(); err != nil {
		return nil, fmt.Errorf("Failed to initialize: %w", err)
	}

	return &TagReader{
		device: &dev,
	}, nil
}

// Reset Implements the hardware reset by pressing the ResetPin(19) low and then releasing.
func (r TagReader) ResetDevice() error {
	log.Println("Resetting the reader..")

	// refer to gpiod docs
	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		return fmt.Errorf("Could not open GPIO device: %w ", err)
	}

	pin, err := c.RequestLine(resetPin, gpiod.AsOutput(0))
	if err != nil {
		return fmt.Errorf("Could not prepare GPIO line: %w ", err)
	}

	if err := pin.SetValue(1); err != nil {
		return fmt.Errorf("Could not set reset signal: %w ", err)
	}

	time.Sleep(time.Millisecond * 400) // margin

	if err := pin.SetValue(0); err != nil {
		return fmt.Errorf("Could not clear reset signal: %w ", err)
	}

	time.Sleep(time.Millisecond * 400) // margin

	if err := pin.SetValue(1); err != nil {
		return fmt.Errorf("Could not set reset signal: %w ", err)
	}

	time.Sleep(time.Millisecond * 100) // margin

	return nil
}

func (r *TagReader) Close() error {
	if err := r.device.Close(); err != nil {
		return fmt.Errorf("Could not close device: %w ", err)
	}

	return nil
}

func (r *TagReader) ListenForTags(tagsChannel chan string, frequencyTicker *time.Ticker) {
	for range frequencyTicker.C {
		// Poll for 300ms
		tagCount, target, err := r.device.InitiatorPollTarget(modulations, 1, 300*time.Millisecond)
		if err != nil {
			fmt.Println("Error polling the reader", err)
			continue
		}

		// Check if any tag was detected
		if tagCount > 0 {
			var UID string

			fmt.Printf(target.String())
			// Transform the target to a specific tag Type and send the UID to the channel
			switch target.Modulation() {
			case nfc.Modulation{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106}:
				var card = target.(*nfc.ISO14443aTarget)
				var UIDLen = card.UIDLen
				var ID = card.UID
				UID = hex.EncodeToString(ID[:UIDLen])
				break
			case nfc.Modulation{Type: nfc.ISO14443b, BaudRate: nfc.Nbr106}:
				var card = target.(*nfc.ISO14443bTarget)
				var UIDLen = len(card.ApplicationData)
				var ID = card.ApplicationData
				UID = hex.EncodeToString(ID[:UIDLen])
				break
			case nfc.Modulation{Type: nfc.Felica, BaudRate: nfc.Nbr212}:
				var card = target.(*nfc.FelicaTarget)
				var UIDLen = card.Len
				var ID = card.ID
				UID = hex.EncodeToString(ID[:UIDLen])
				break
			case nfc.Modulation{Type: nfc.Felica, BaudRate: nfc.Nbr424}:
				var card = target.(*nfc.FelicaTarget)
				var UIDLen = card.Len
				var ID = card.ID
				UID = hex.EncodeToString(ID[:UIDLen])
				break
			case nfc.Modulation{Type: nfc.Jewel, BaudRate: nfc.Nbr106}:
				var card = target.(*nfc.JewelTarget)
				var ID = card.ID
				var UIDLen = len(ID)
				UID = hex.EncodeToString(ID[:UIDLen])
				break
			case nfc.Modulation{Type: nfc.ISO14443biClass, BaudRate: nfc.Nbr106}:
				var card = target.(*nfc.ISO14443biClassTarget)
				var ID = card.UID
				var UIDLen = len(ID)
				UID = hex.EncodeToString(ID[:UIDLen])
				break
			}

			// Send the UID of the tag to main goroutine
			tagsChannel <- UID
		}
	}
}

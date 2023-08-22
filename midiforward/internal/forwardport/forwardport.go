package forwardport

import (
	// "gitlab.com/gomidi/midi/v2"
	"fmt"
	"sync"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

var connectLock = sync.Mutex{}

type ForwardPort struct {
	name string
	port drivers.Out
}

func New(name string) (ForwardPort, error) {
	var o = ForwardPort{name: name}
	err := o.connectOutput()
	return o, err
}

func (o *ForwardPort) connectOutput() error {
	locked := connectLock.TryLock()
	defer connectLock.Unlock()

	if !locked {
		fmt.Println("Output port is locked, skipping port creation...")
	} else if o.port == nil {
		newPort, err := drivers.OutByName(o.name)
		if err != nil {
			return err
		}
		o.port = newPort
	}
	return nil
}

// TODO: what does int32 param contain? 'timestampms'?
// https://github.com/gomidi/midi/blob/eb01aef2d7aa5ecb65343052708db05997af3315/examples/logger/main.go#L20
func (o ForwardPort) ForwardMessage(msg midi.Message, timestampms int32) {
	if o.port != nil {
		fmt.Printf("Forwarding message: %s\n", msg)
		err := o.port.Send([]byte(string(msg)))
		if err != nil {
			fmt.Printf("Error forwarding message: %s\n", err)
		}
	} else {
		err := o.connectOutput()
		if err != nil {
			fmt.Printf("Error connecting output: %s\n", err)
		}
	}
}

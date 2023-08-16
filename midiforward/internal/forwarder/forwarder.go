package forwarder

import (
	"fmt"
	"midiforward/internal/forwardport"
	"midiforward/internal/utils"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

var connectedInPorts midi.InPorts

func containsPort(port drivers.In, ports midi.InPorts) bool {
	for _, p := range ports {
		if p.String() == port.String() {
			return true
		}
	}
	return false
}

func connectInPort(in drivers.In, fp *forwardport.ForwardPort) {
	_, err := midi.ListenTo(
		in,
		fp.ForwardMessage,
		midi.UseSysEx(),
	)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	connectedInPorts = append(connectedInPorts, in)
	fmt.Println("Connected new input port: ", in)
}

func removeFromList(port drivers.In, ports midi.InPorts) midi.InPorts {
	for i, p := range ports {
		if p.String() == port.String() {
			return append(ports[:i], ports[i+1:]...)
		}
	}
	return ports
}

func connectPorts(
	forwardPort *forwardport.ForwardPort,
	ingorePorts map[string]struct{},
) {
	ports := midi.GetInPorts()
	// remove any ports that are no longer connected
	for _, port := range connectedInPorts {
		if !containsPort(port, ports) {
			fmt.Println("Removing disconnected port: ", port.String())
			connectedInPorts = removeFromList(port, connectedInPorts)
		}
	}
	// connect any new ports
	for _, port := range ports {
		_, isExcluded := ingorePorts[port.String()]
		if !containsPort(port, connectedInPorts) && !isExcluded {
			connectInPort(port, forwardPort)
		}
	}
}

func StartForwarding(outputPortName string, ignorePorts map[string]struct{}) error {
	forwardPort, err := forwardport.New(outputPortName)
	if err != nil {
		fmt.Println("Can't connect output port: ", outputPortName)
		return err
	}
	utils.LogPorts()
	fmt.Println("Listening...")
	for {
		connectPorts(&forwardPort, ignorePorts)
		time.Sleep(5 * time.Second)
	}
}

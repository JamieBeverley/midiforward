package utils

import (
	"bufio"
	"fmt"
	"os"

	"gitlab.com/gomidi/midi/v2"
)

func ReadOutPort() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter midi output port: ")
	return reader.ReadString('\n')
}

func LogPorts() {
	outPorts := midi.GetOutPorts()
	fmt.Println("MIDI Ports:")
	fmt.Printf("Outputs: \n%s", outPorts)
	inPorts := midi.GetInPorts()
	fmt.Printf("Inputs: \n%s\n", inPorts)
}

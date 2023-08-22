package main

import (
	"fmt"
	"midiforward/internal/forwarder"
	"midiforward/internal/forwardport"
	"midiforward/internal/server"
	"midiforward/internal/utils"
	"os"
	"strings"
	"sync"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func main() {
	defer midi.CloseDriver()

	flags := []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "out",
			Aliases: []string{"o"},
			Usage:   "midi output port name",
		}),
		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name:    "ignore",
			Aliases: []string{"i"},
			Usage:   "midi input ports to ignore",
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:    "list",
			Aliases: []string{"ls", "log"},
			Usage:   "log midi ports",
		}),
		altsrc.NewInt64Flag(&cli.Int64Flag{
			Name:  "serverPort",
			Usage: "port for udp server",
			Value: 8000,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "serverAddress",
			Usage: "port for udp server",
			Value: "0.0.0.0",
		}),
		&cli.StringFlag{
			Name:  "config",
			Usage: "path to .json file with settings (alternative to CLI arguments)",
		},
	}

	app := &cli.App{
		Name: "midiforward",
		Usage: `Forward midi messages to another port. Settings can be defined in a 
		json file indicated by --config or overriden as CLI arguments.`,
		Action: func(c *cli.Context) error {
			if c.IsSet("list") {
				utils.LogPorts()
				return nil
			}

			outPortName := c.String("out")
			if outPortName == "" {
				outPortName, _ = utils.ReadOutPort()
			}

			ignore := c.StringSlice("ignore")
			ignorePorts := make(map[string]struct{})
			for _, port := range ignore {
				ignorePorts[port] = struct{}{}
			}
			fmt.Println("\nSETTINGS:")
			fmt.Printf("Output port: %s\n", outPortName)
			fmt.Printf("Ignoring ports: %s\n\n", strings.Join(ignore, ", "))
			forwardPort, err := forwardport.New(outPortName)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				return err
			}

			udpServer := server.New(
				c.String("serverAddress"),
				c.Int("serverPort"),
				&forwardPort,
			)

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := forwarder.StartForwarding(&forwardPort, ignorePorts)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err)
				}
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := udpServer.Listen()
				if err != nil {
					fmt.Printf("ERROR: %s\n", err)
				}
			}()
			wg.Wait()
			return nil
		},
		Before: altsrc.InitInputSourceWithContext(flags, altsrc.NewJSONSourceFromFlagFunc("config")),
		Flags:  flags,
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
}

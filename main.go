package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	vm "github.com/josiahsams/virsh-client/cmd/vm"
)

func main() {
    app := &cli.App{
		Name:  "zosvm",
		Usage: "Utility for managing a libvirt z/OS VM",
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create a libvirt z/OS VM",
                Action: func(c *cli.Context) error {
							return vm.HandleCreateVM(c)
						},
				Flags: []cli.Flag{
                    &cli.StringFlag{
                        Name:    "name",
                        Usage:   "Virtual Machine Name",
                        Aliases: []string{"n"},
                        Required: true,
                    },
                    &cli.StringFlag{
                        Name:    "osImgSrc",
                        Usage:   "OS image src path",
                        Aliases: []string{"p"},
                        Required: true,
                    },
                    &cli.StringFlag{
                        Name:    "cloudInitSrc",
                        Usage:   "Cloud Init Image src path",
                        Aliases: []string{"ci"},
                        Required: true,
                    },
                    &cli.StringFlag{
                        Name:    "mode",
                        Usage:   "CPU mode",
                        Aliases: []string{"m"},
                        Value: "host-passthrough",
                    },
                    &cli.UintFlag{
                        Name:    "cpu",
                        Usage:   "Number of VCPUs",
                        Aliases: []string{"c"},
                        Value: 2,
                    },
                    &cli.UintFlag{
                        Name:    "memory",
                        Usage:   "Memory size to allocate",
                        Aliases: []string{"m"},
                        Value: 4194304,
                    },
                },
            },
        },
    }

    err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

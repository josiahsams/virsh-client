package vm

import (
	"fmt"
	"log"

	vm "github.com/josiahsams/virsh-client/internal/pkg/vm"
	libvirt "github.com/libvirt/libvirt-go"
	"github.com/urfave/cli/v2"
)

// HandleCreateVM ..
func HandleCreateVM(c *cli.Context) error {

	conn, err := libvirt.NewConnect("qemu:///system")
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer conn.Close()

    if err != nil {
        panic(err)
    }

    var flags libvirt.DomainCreateFlags
    // flags = libvirt.DOMAIN_START_PAUSED
    flags = libvirt.DOMAIN_NONE

	vmName := c.String("name")
	vpcu := c.Uint("cpu")
	memory := c.Uint("memory")
	mode := c.String("mode")
	osImgSrc := c.String("osImgSrc")
	cloudInitSrc := c.String("cloudInitSrc")

    newVM := vm.New(vmName, memory, vpcu, mode, osImgSrc, cloudInitSrc)
    xml, err := newVM.CreateXML()

    domain, err := conn.DomainCreateXML(xml, flags)
	if err != nil {
        panic(err)
    }

    domainName, err := domain.GetName()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Domain created successfully : %s !!\n", domainName)
	return nil
}

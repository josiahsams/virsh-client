package vm

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	cloudinit "github.com/josiahsams/virsh-client/internal/pkg/cloudinit"
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
    userdata := c.String("userdata")

    if cloudInitSrc != "" {
        _, err = os.Stat(cloudInitSrc)
        if os.IsNotExist(err) {
            fmt.Printf("CloudInit image will be created and add it to the VM.\n")
            ci := cloudinit.New(cloudInitSrc, userdata)
            script := "export RUNZ_COMMIT='2cc9801+';" + 
                    "export UID=1001;" +
                    "export GID=1001;" +
                    "nohup proxy -id xyz &"

            ci.AddStartScripts("runz", script)
            err := ci.PrepareImg(false)
            if err != nil {
                panic(err)
            }
        } else {
            fmt.Printf("Skip creating CloudInit image as its already found. \n")
        }
    }

    // Create a new OS image keep the baseOS image as a backing store
    newOSImgSrc := osImgSrc + "-" + vmName
    _, err = exec.Command("qemu-img create -f qcow2 -F qcow2 -b ",
                "" + osImgSrc + " " + newOSImgSrc).CombinedOutput()
	if err != nil {
		panic(err)
	} 

    fmt.Println("Created a new clone out of the baseOS image: ", newOSImgSrc)
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

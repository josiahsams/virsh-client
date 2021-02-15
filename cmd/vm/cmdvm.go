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
    zvolumes := c.String("zvolumes")
    share := c.Bool("share")

    if cloudInitSrc != "" {
        _, err = os.Stat(cloudInitSrc)
        if os.IsNotExist(err) {
            fmt.Printf("CloudInit image will be created and add it to the VM.\n")
            ci := cloudinit.New(cloudInitSrc, userdata)
            script := "#!/bin/bash \n" +
                    "sleep 30 \n"+
                    "ulimit -l unlimited \n"+
                    "chown runz:runz /dev/net/tun \n" +
                    "chown runz:runz /dev/kvm \n" + 
                    "chown -R runz:runz /volumes \n" + 
                    "chmod 0666 /dev/kvm \n" +
                    "export RUNZ_COMMIT='2cc9801+' \n" + 
                    "export UID=1001 \n" +
                    "export GID=1001 \n" +
                    "export AUTOIPL=1 \n" +
                    "export PROCESSORS=2 \n" +
                    "export SSH_PUBLICKEY= \n" +
                    "export TCP_PORTS=1-65535 \n" +
                    "export MEMORY=8192 \n" +
                    "export X509_CERTIFICATE=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURhekNDQWxPZ0F3SUJBZ0lVSGR6dFdNamlmQnBjRERXSFpvajkvNlhtS3Jrd0RRWUpLb1pJaHZjTkFRRUwKQlFBd1JURUxNQWtHQTFVRUJoTUNRVlV4RXpBUkJnTlZCQWdNQ2xOdmJXVXRVM1JoZEdVeElUQWZCZ05WQkFvTQpHRWx1ZEdWeWJtVjBJRmRwWkdkcGRITWdVSFI1SUV4MFpEQWVGdzB5TURBM01Ea3hNVEU0TlRoYUZ3MHlNVEEzCk1Ea3hNVEU0TlRoYU1FVXhDekFKQmdOVkJBWVRBa0ZWTVJNd0VRWURWUVFJREFwVGIyMWxMVk4wWVhSbE1TRXcKSHdZRFZRUUtEQmhKYm5SbGNtNWxkQ0JYYVdSbmFYUnpJRkIwZVNCTWRHUXdnZ0VpTUEwR0NTcUdTSWIzRFFFQgpBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRRElTbmhIdzF5anJVM0lETVJTRDZBMGRmOTAvVVlOTzE2ZWtycURPWnIxCjJ4aTlYN3RaNmZWeTZmNDhqOGh6VkZJR3ZnR3RoRE1Kck9xbzJHK1ZsSXo0WXVoV29BaTgxRitFLzJlcDlVaTQKNk1wWnlFOHVmNEoyWDlhZEVCUHBmU1ZiWjVqV1dSQUFFMExHMTRKQ3NuT01DdkJwam04eHRkVFIrS3pUVVdVMgpMeHlmQnJMMmIwSkR4YldMczY3ZmJCcUs0U1luZ29kT3l4ckkwMU8xeFJubHNKY1o1bCt1WFBUM2o2L0RKaitEClpUMVpCdkFqdk83QklQN2NTaStRancwWGVzTXFlVUtwMHpPNjlXb205OFVBSDJsdmxOdWhLWXNlR0Y4MUhrbSsKUWUzZ25QcE43N0VwT0lDMkF6RDdaRmFGK0dNNE5aV01jQ2Y5UUU4akVmeGpBZ01CQUFHalV6QlJNQjBHQTFVZApEZ1FXQkJUcFgxdjQ5V284S3RhOXB0eStnQTZ4cDNrQnRUQWZCZ05WSFNNRUdEQVdnQlRwWDF2NDlXbzhLdGE5CnB0eStnQTZ4cDNrQnRUQVBCZ05WSFJNQkFmOEVCVEFEQVFIL01BMEdDU3FHU0liM0RRRUJDd1VBQTRJQkFRQ0wKWjBmeVF1ZUpaN1NqTmp4anRrVFNrQTFyQUJJNXVzdnRSbWR2cTBINEVLRVFpWDFPaG5FbzZjZEZ4SEhUVlpCWApPai8zRVpWSDRrTVIrZnlmNUo5RnQycUlRelhSem1mNFlxYzBJSWxScGZLRkphUkw5YWpkMWNWTU1ER1JKdGpNCkxCSU9Yb2llZjJIdnh6ZDU2SVdpWFlnWEZuelZZMDdwdW9oNERwOEZDN1FQc2F1ZFhnbTdvYlZqSERhc1YzNDAKNXV1QmwyZ3pNVDJvbzQ1WkpxS3ZuNjU0cGt4UUNOZzlSRDluZThHVDBQbUdMMUFYTlVyU1NESi9TTTNjUlFzNgpGVDd0MzRqQVEvcHRnUzNISzkvL05CWUZHVm8xb2E3eDI3MitrQWV5cXhCQkdwSDBMQUxkeDdqNm0zWmMycWtHCkhnTHpkeU85cUxzM3dGRkNQWXd3Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K \n" +
                    "nohup /proxy -id xyz &"

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
    newOSImgSrc := osImgSrc
    if share {
        newOSImgSrc := newOSImgSrc + "-" + vmName
        _, err = exec.Command("qemu-img", "create", "-f",  "qcow2", "-F", "qcow2", "-b",
                    osImgSrc , newOSImgSrc).CombinedOutput()
        if err != nil {
            panic(err)
        }
    }

    fmt.Println("Created a new clone out of the baseOS image: ", newOSImgSrc)
    newVM := vm.New(vmName, memory, vpcu, mode, newOSImgSrc, cloudInitSrc, zvolumes)
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

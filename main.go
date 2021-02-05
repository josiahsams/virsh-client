package main

import (
	"fmt"
	"log"
    "runtime"

	libvirt "github.com/libvirt/libvirt-go"
    libvirtxml "github.com/libvirt/libvirt-go-xml"
)

func main() {
    fmt.Println("GOARCH : "+ runtime.GOARCH)
    var platform string
    var s390x bool = false

    if runtime.GOARCH == "amd64" {
        platform = "x86_64"
    } else if runtime.GOARCH == "s390x" {
        s390x = true
        platform = "s390x"
    } else {
        log.Fatalf("Unsupported platform")
    }

    conn, err := libvirt.NewConnect("qemu:///system")
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer conn.Close()
    var port uint = 0

    domcfg := &libvirtxml.Domain{}

    if s390x {
        domcfg.Type = "kvm"
    } else {
        domcfg.Type = "qemu"
    }
    domcfg.Name  = "ub18-1"

    domcfg.Memory = &libvirtxml.DomainMemory{
          Unit: "KiB",
          Value: 2097152,
       }
    domcfg.VCPU = &libvirtxml.DomainVCPU{ Value: 2 }

    if s390x {
         domcfg.CPU = &libvirtxml.DomainCPU{
            Mode: "host-passthrough",
        }
    } else {
        domcfg.CPU = &libvirtxml.DomainCPU{
            Mode: "custom",
            Model: &libvirtxml.DomainCPUModel{
                Fallback: "forbid",
                Value: "EPYC",
            },
        }
    }
    domcfg.Devices =  &libvirtxml.DomainDeviceList{
          Emulator: "/usr/bin/qemu-system-" + platform,
          Disks: []libvirtxml.DomainDisk{
             {
                Device: "disk",
                Source:  &libvirtxml.DomainDiskSource{
                    File: &libvirtxml.DomainDiskSourceFile{
                        File: "/tmp/ub18.img",
                    }},
                Target:  &libvirtxml.DomainDiskTarget{
                    Dev: "vda",
                    Bus: "virtio",
                },
                Driver:  &libvirtxml.DomainDiskDriver{
                    Name: "qemu",
                    Type: "qcow2",
                },
             },
             {
                Device: "cdrom",
                Source:  &libvirtxml.DomainDiskSource{
                    File: &libvirtxml.DomainDiskSourceFile{
                        File: "/tmp/cidata.iso",
                    }},
                Target:  &libvirtxml.DomainDiskTarget{
                    Dev: "sda",
                    Bus: "scsi",
                },
                Driver:  &libvirtxml.DomainDiskDriver{
                    Name: "qemu",
                    Type: "raw",
                },
             },
          },
          Interfaces : []libvirtxml.DomainInterface{
             {
                Source:  &libvirtxml.DomainInterfaceSource{
                    Bridge: &libvirtxml.DomainInterfaceSourceBridge {
                        Bridge: "virbr0",
                    },
                },
             },
          },
          Serials: []libvirtxml.DomainSerial {
             {
                 Target:  &libvirtxml.DomainSerialTarget{ Port: &port},
             },
          },
          Consoles: []libvirtxml.DomainConsole {
             {
                 Target:  &libvirtxml.DomainConsoleTarget{
                     Type: "serial",
                     Port: &port,
                 },
             },
          },
       }

    domcfg.OS = &libvirtxml.DomainOS{
           Type: &libvirtxml.DomainOSType {Arch: platform, Type: "hvm"},
          BootDevices: []libvirtxml.DomainBootDevice {
             {
                Dev: "hd",
             },
          },
       }

    domcfg.Clock = &libvirtxml.DomainClock  {
          Offset : "utc",
       }

    domcfg.OnPoweroff = "destroy"
    domcfg.OnReboot = "restart"
    domcfg.OnCrash = "destroy"

    xml, err := domcfg.Marshal()
    if err != nil {
        panic(err)
    }

    var flags libvirt.DomainCreateFlags
    // flags = libvirt.DOMAIN_START_PAUSED
    flags = libvirt.DOMAIN_NONE

    domain, err := conn.DomainCreateXML(xml, flags)
    if err != nil {
        panic(err)
    }

    domainName, err := domain.GetName()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Domain created successfully : %s !!\n", domainName)

    doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
    if err != nil {
        log.Fatalf("failed to list  domains: %v", err)
    }

    fmt.Printf("%d running domains:\n", len(doms))
    for _, dom := range doms {
        name, err := dom.GetName()
        if err == nil {
            fmt.Printf("  %s\n", name)
        }
        dom.Free()
    }
}

package vm

import (
	"log"
	"runtime"

	lvxml "github.com/libvirt/libvirt-go-xml"
)

// Instance ..
type Instance struct {
	Name string
	memoryInKB uint
	vcpu uint
	mode string
	diskSource string
	cloudInitSource string
    zvolumes string
}

// New ...
func New(name string, memory uint, vcpu uint, mode string, diskSrc string, cloudInitSrc string, zvolumes string ) *Instance {
	return &Instance{ 
		Name : name, 
		memoryInKB: memory, 
		vcpu : vcpu, 
		mode: mode,
		diskSource: diskSrc, 
		cloudInitSource: cloudInitSrc,
        zvolumes:  zvolumes,
	}
}

func createFileDisk(driveName string, driverType string, srcFile string) (disk *lvxml.DomainDisk) {
	disk = &lvxml.DomainDisk{
		Device: "disk",
		Driver: &lvxml.DomainDiskDriver{
			Name:        "qemu",
			Type:        driverType,
		},
		Source: &lvxml.DomainDiskSource{
			File: &lvxml.DomainDiskSourceFile{File: srcFile},
		},
		Target: &lvxml.DomainDiskTarget{
			Bus: "virtio",
			Dev: driveName,
		},
	}
	return
}


// CreateXML ..
func (inst *Instance) CreateXML() (xml string, err error ) {

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

    var port uint = 0

    domcfg := &lvxml.Domain{}

    if s390x {
        domcfg.Type = "kvm"
    } else {
        domcfg.Type = "qemu"
    }
    domcfg.Name  = inst.Name

    domcfg.Memory = &lvxml.DomainMemory{
          Unit: "KiB",
          Value: inst.memoryInKB,
       }
    domcfg.VCPU = &lvxml.DomainVCPU{ Value: inst.vcpu }

    if s390x {
         domcfg.CPU = &lvxml.DomainCPU{
            Mode: inst.mode,
        }
    } else {
        domcfg.CPU = &lvxml.DomainCPU{
            Mode: "custom",
            Model: &lvxml.DomainCPUModel{
                Fallback: "forbid",
                Value: "EPYC",
            },
        }
    }

	domainDisks := make([]lvxml.DomainDisk, 0, 2)

	osDisk := createFileDisk("vda", "qcow2", inst.diskSource)
	domainDisks = append(domainDisks, *osDisk)

    if inst.cloudInitSource != "" {
        cloudInitDisk := createFileDisk("vdb", "raw", inst.cloudInitSource)
        domainDisks = append(domainDisks, *cloudInitDisk)
    }

    if inst.zvolumes != "" {
        zosVolumesDisk := createFileDisk("vdc", "raw", inst.zvolumes)
	    domainDisks = append(domainDisks, *zosVolumesDisk)
    }

    domcfg.Devices =  &lvxml.DomainDeviceList{
          Emulator: "/usr/bin/qemu-system-" + platform,
          Disks: domainDisks,
          Interfaces : []lvxml.DomainInterface{
             {
                Source:  &lvxml.DomainInterfaceSource{
                    Bridge: &lvxml.DomainInterfaceSourceBridge {
                        Bridge: "virbr0",
                    },
                },
             },
          },
          Serials: []lvxml.DomainSerial {
             {
                 Target:  &lvxml.DomainSerialTarget{ Port: &port},
             },
          },
          Consoles: []lvxml.DomainConsole {
             {
                 Target:  &lvxml.DomainConsoleTarget{
                     Type: "serial",
                     Port: &port,
                 },
             },
          },
       }

    domcfg.OS = &lvxml.DomainOS{
           Type: &lvxml.DomainOSType {Arch: platform, Type: "hvm"},
          BootDevices: []lvxml.DomainBootDevice {
             {
                Dev: "hd",
             },
          },
       }

    domcfg.Clock = &lvxml.DomainClock  {
          Offset : "utc",
       }

    domcfg.OnPoweroff = "destroy"
    domcfg.OnReboot = "restart"
    domcfg.OnCrash = "destroy"

    xml, err = domcfg.Marshal()
    if err != nil {
        return
    }
	return

}
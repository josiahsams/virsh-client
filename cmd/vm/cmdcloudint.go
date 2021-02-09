package vm

import (
	"fmt"

	cloudinit "github.com/josiahsams/virsh-client/internal/pkg/cloudinit"
	"github.com/urfave/cli/v2"
)

// HandleCreateCloudInitImg ..
func HandleCreateCloudInitImg(c *cli.Context) error {

	imgpath := c.String("imgpath")
	userdata := c.String("userdata")
	retainFlag := c.Bool("retain")

	ci := cloudinit.New(imgpath, userdata)

	script := "export RUNZ_COMMIT='2cc9801+';" + 
 			  "export UID=1001;" +
			  "export GID=1001;" +
			  "nohup proxy -id xyz &"

	ci.AddStartScripts("runz", script)
	err := ci.PrepareImg(retainFlag)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Cloud-Init image created successfully : %s !!\n", imgpath)
	return nil
}
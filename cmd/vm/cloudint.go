package vm

import (
	ci "github.com/josiahsams/virsh-client/internal/pkg/cloudinit"
	"github.com/urfave/cli/v2"
)

// HandleCreateCloudInitImg ..
func HandleCreateCloudInitImg(c *cli.Context) error {

	imgpath := c.String("imgpath")
	userdata := c.String("userdata")

	err := ci.PrepareImg(imgpath, userdata)
	if err != nil {
		panic(err)
	}

	return nil
}
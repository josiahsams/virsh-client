package cloudint

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// cloudinit ..
type cloudinit struct {
	basePath string
	userData string
}

// PrepareImg ..
func PrepareImg(basePath, userData string, retainFlag bool) (err error) {

	ci := &cloudinit{basePath, userData}

	// create directory if it doesn't exist
	err = ci.createDir()
	if err != nil {
		return
	}
	err = ci.buildMetaData()
	if err != nil {
		return
	}
	err = ci.buildUserData()
	if err != nil {
		return
	}
	err = ci.buildVendordata()
	if err != nil {
		return
	}

	err = ci.createImage()
	if err != nil {
		return
	}

	// don't remove the generated files if retain flag is set
	if ! retainFlag {
		os.Remove(basePath+".metadata")
		os.Remove(basePath+".userData")
		os.Remove(basePath+".vendordata")
	}
	return
}

func (ci *cloudinit) createDir() error {
	// create the path; if it exists, there is no error
	if err := os.MkdirAll(filepath.Dir(ci.basePath), 0777); err != nil {
		return err
	}
	return nil
}

func wrapError(err error, msg string) error {
	return fmt.Errorf("%s %w", msg, err) 
}

func (ci *cloudinit) buildMetaData() (err error) {

	meta := metadataInfoCI{
		InstanceID:    "sample-instance-id",
		LocalHostname: "sample1",
	}
	metaDataYAML, err := yaml.Marshal(meta)
	if err != nil {
		return wrapError(err, "could not marshal metadata to yaml")
	}

	err = ci.updateInitDataFile(".metadata", metaDataYAML)
	return 
}

func (ci *cloudinit) buildUserData() (err error) {
	var userData []byte
	if (ci.userData != "") {
		userData, err = ioutil.ReadFile(ci.userData)
	}
	err = ci.updateInitDataFile(".userData", userData)
	return 
}

func (ci *cloudinit) buildVendordata() (err error) {
	var vendorData []byte

	// build the vendordata file as a multi-part MIME file containing the vendordata defaults
	vendorData, err = ci.assembleVendordata()
	if err != nil {
		return
	}
	return ci.updateInitDataFile(".vendordata", vendorData)
}

func (ci *cloudinit) updateInitDataFile(fileName string, newData []byte) (err error) {
	initDataFilePath := ci.basePath + fileName
	err = ioutil.WriteFile(initDataFilePath, newData, 0777)
	if err != nil {
		return wrapError(err, "could not write "+path.Base(initDataFilePath)+" to disk")
	}
	return
}

func (ci *cloudinit) createImage() (err error) {
	imgPath := ci.basePath
	_, err = exec.Command("cloud-localds", "--disk-format=raw", 
				"--vendor-data="+ci.basePath+".vendordata", 
				imgPath, ci.basePath+".userData", 
				ci.basePath+".metadata").CombinedOutput()
	if err != nil {
		return wrapError(err, "could not create cloud-init disk")
	}
	return nil
}

func (ci *cloudinit) assembleVendordata () (vendorData []byte, err error) {
	// create multipart buffer and writer to write data to
	body, writer, err := ci.createBufferAndWriter()
	if err != nil {
		return nil, err
	}

	powerState := powerStateCI{
		Delay: "now",
		Mode: "reboot",
		Message:  "reboot VM after kernel patches",
		Timeout: 0,
		Condition:  "dpkg --configure -a 2>&1 | grep -i done",
	}
	
	out, err := yaml.Marshal(powerState)
	if err != nil {
		return nil, wrapError(err, "could not marshal metadata to yaml")
	}
	configString := "#cloud-config\n" + string(out)

	err = ci.writeDataSection(writer, textCloudConfig, configString)
	if err != nil {
		return
	}

	// add the final boundary
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	vendorData = body.Bytes()
	return
}

func (ci *cloudinit) createBufferAndWriter() (*bytes.Buffer, *multipart.Writer, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	req, err := http.NewRequest("POST", "", body)
	if err != nil {
		return nil, nil, wrapError(err, "could not make request for multipart data construction")
	}
	// hardcode boundary so we can tell if the data has changed
	writer.SetBoundary(LinuxMimeBoundary)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	err = req.Header.Write(body)
	if err != nil {
		return nil, nil, wrapError(err, "could not write multipart data header")
	}
	_, err = body.WriteString("MIME-Version: 1.0\n\n")
	if err != nil {
		return nil, nil, wrapError(err, "could not write multipart data header")
	}
	return body, writer, nil
}

func (ci *cloudinit) writeDataSection(writer *multipart.Writer, contentType, content string) (err error) {
	mh := make(textproto.MIMEHeader)
	mh.Set("Content-Type", contentType)
	partWriter, err := writer.CreatePart(mh)
	if err != nil {
		return wrapError(err, "could not create data section")
	}
	_, err = io.Copy(partWriter, bytes.NewBufferString(content))
	if err != nil {
		return wrapError(err, "could not create data section")
	}
	return nil
}

// LinuxMimeBoundary ..
const LinuxMimeBoundary = "3efa30189c9e0e8ebc24a4decbbf4c2be7b26120c1cdd7cb7bc2ecb0c07c"
const textCloudConfig = "text/cloud-config"
const textShellScript = "text/x-shellscript"
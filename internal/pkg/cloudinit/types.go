package cloudint


type powerStateCI struct {
	Delay           string   	`yaml:"delay"`
	Mode        	string     	`yaml:"mode"`
	Message         string    	`yaml:"message"`
	Timeout 	    int 		`yaml:"timeout"`
	Condition       string   	`yaml:"condition"`
}

type metadataInfoCI struct {
	InstanceID    string `yaml:"instance-id"`
	LocalHostname string `yaml:"local-hostname"`
}
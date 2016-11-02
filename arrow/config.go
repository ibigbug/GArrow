package arrow

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// Config struct
type Config struct {
	ServerAddress string `yaml:"server,omitempty"`
	LocalAddress  string `yaml:"local,omitempty"`
	Password      string `yaml:"password,omitempty"`
}

// NewConfig factory
func NewConfig(p string) (c *Config) {
	c = &Config{}
	fd, err := os.Open(p)
	defer fd.Close()
	if os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "No such file: "+p)
		os.Exit(1)
	}
	checkError(err)

	bytes, err := ioutil.ReadAll(fd)
	checkError(err)

	err = yaml.Unmarshal(bytes, c)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid config: ", err.Error())
		os.Exit(1)
	}
	return
}

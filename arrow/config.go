package arrow

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	ServerAddress string `yaml:"server,omitempty"`
	LocalAddress  string `yaml:"local,omitempty"`
	Password      string `yaml:"password,omitempty"`
}

func NewConfig(p string) (c *Config) {
	c = &Config{}
	fd, err := os.Open(p)
	defer fd.Close()
	if os.IsNotExist(err) {
		fmt.Println("No such file: " + p)
		os.Exit(1)
	}
	checkError(err)

	bytes, err := ioutil.ReadAll(fd)
	checkError(err)

	err = yaml.Unmarshal(bytes, c)
	if err != nil {
		fmt.Println("Invalid config: ", err.Error())
		os.Exit(1)
	}
	return
}

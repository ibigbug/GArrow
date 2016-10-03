package connpool

import (
	"fmt"
	"os"
)

func debug(formatter string, a ...interface{}) {
	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, formatter, a...)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/djthorpe/goapp/pkg/location"
)

func main() {
	l := location.NewLocation("World")
	fmt.Println("Hello,", l)
	os.Exit(-1)
}

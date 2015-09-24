package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/cbarrick/evo/example/queens"
)

func main() {
	var example string
	if len(os.Args) == 1 {
		fmt.Println("usage: go run example/run.go <example> <dimension>")
		os.Exit(0)
	}
	example = os.Args[1]

	var dim int
	if len(os.Args) > 2 {
		var err error
		dim, err = strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("ERROR: Invalid dimension")
			os.Exit(1)
		}
	}

	switch example {
	case "queens":
		queens.Main(dim)

	default:
		fmt.Printf("ERROR: Unknown example: '%v'\n", example)
		os.Exit(1)
	}
}

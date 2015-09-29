package main

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/cbarrick/evo/example/queens"
	"github.com/cbarrick/evo/example/tsp"
)

func main() {
	var example string
	if len(os.Args) == 1 {
		cmd := path.Base(os.Args[0])
		fmt.Printf(
			"usage: %v <example> [<arg>...]\n" +
			"\n" +
			"examples:\n" +
			" queens  n-queens, takes n as arg, default n=256\n" +
			" tsp     travelling salesman, takes no args\n", cmd)
		os.Exit(0)
	}
	example = os.Args[1]

	switch example {
	case "queens":
		var dim int
		var err error
		if len(os.Args) > 2 {
			dim, err = strconv.Atoi(os.Args[2])
			if err != nil {
				panic(err.Error())
			}
		} else {
			dim = -1
		}
		queens.Main(dim)
	case "tsp":
		tsp.Main()

	default:
		fmt.Printf("ERROR: Unknown example: '%v'\n", example)
		os.Exit(1)
	}
}

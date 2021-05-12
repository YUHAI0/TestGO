package main

import (
	"fmt"
	"os"
)

func main() {

	f, _ := os.Create("1mrz")

	count := 10000
	c := 0

	for {
		f.WriteString(fmt.Sprintf("%-10d\t###############################\n", c))
		c += 1
		if c > count {
			break
		}
	}

	f.Close()
}

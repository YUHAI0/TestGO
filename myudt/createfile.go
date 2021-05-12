package main

import (
	"fmt"
	"os"
)

func main() {

	f, _ := os.Create("1mrz")

	count := 100000
	for {
		f.WriteString(fmt.Sprintf("%-10d\t###############################\n", count))
		count -= 1
		if count <= 0 {
			break
		}
	}

	f.Close()
}

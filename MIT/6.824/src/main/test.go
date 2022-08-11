package main

import (
	"fmt"
	"os"
)

func main()  {
	tmpDir, _ := os.CreateTemp("", "tmp")
	fmt.Println(tmpDir)
}

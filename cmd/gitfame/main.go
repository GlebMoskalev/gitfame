package main

import (
	"fmt"
	"github.com/GlebMoskalev/gitfame/internal/git"
)

func main() {
	files, err := git.GetFilesRepository(
		"",
		"",
		"",
		"",
	)
	fmt.Println(files, err)
}

package main

import "github.com/GlebMoskalev/gitfame/internal/git"

func main() {
	git.CalculateStats("/Users/glebmoskalev/Downloads/blame",
		"HEAD",
		"",
		"",
		"",
		"",
		false)
}

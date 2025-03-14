package main

import "github.com/GlebMoskalev/gitfame/internal/git"

func main() {
	git.CalculateStats("/Users/glebmoskalev/Downloads/blame",
		"HEAD",
		"",
		"",
		"",
		"yaml,c++,markdown,gopher",
		"lines",
		"json-lines",
		false,
	)
}

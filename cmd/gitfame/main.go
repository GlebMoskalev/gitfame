package main

import "github.com/GlebMoskalev/gitfame/internal/git"

func main() {
	git.CalculateStats("/Users/glebmoskalev/Учеба/FlowerApp", "HEAD", "", "", "", "c#", false)
	//files, err := git.GetFilesRepository(
	//	".",
	//	"HEAD",
	//	"",
	//	"",
	//)
	//fmt.Println(files, err)
}

package main

import (
	"fmt"
	"malmodel"
	"malspider"
	"os"
)

func main() {
	pgSettings := os.Args[1]
	i := 1
	for {
		fmt.Printf("start %v downloading loop\n", i)
		i++
		malspider.UpdateTitlesFromTop(false, 1, pgSettings)
		malspider.StartSpider(false, 1, pgSettings)
		malmodel.GroupAllAnimeModels(pgSettings)
	}
}

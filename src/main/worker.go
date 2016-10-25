package main

import (
	"malspider"
	"os"
	"strconv"
)

func main() {
	workers, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	postgres := os.Args[2]
	malspider.StartSpider(false, workers, postgres)
}

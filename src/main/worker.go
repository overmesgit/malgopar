package main

import (
	"malspider"
	"os"
	"strconv"
)

func main() {
	from, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	to, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	workers, err := strconv.Atoi(os.Args[3])
	if err != nil {
		panic(err)
	}
	postgres := os.Args[4]
	malspider.StartSpider(from, to, false, workers, postgres)
}

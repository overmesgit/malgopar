package main

import (
	"malspider"
)

func main() {
	malspider.StartSpider(1, 100, false, 1, "host=127.0.0.1 port=5432 user=user dbname=user sslmode=disable password=user")
}

package malspider

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"io/ioutil"
	"mal/model"
	"mal/parser"
	"net/http"
	"strconv"
	"sync"
)

const (
	MAL_ANIME_URL = "https://myanimelist.net/anime/"
)

func StartSpider(start, end int, manga bool, workers int) {
	queue := make(chan int, 100)
	result := make(chan malparser.Anime, 100)

	go startFillWorker(queue, start, end)

	var wgParser sync.WaitGroup
	for i := 0; i < workers; i++ {
		wgParser.Add(1)
		go startDownloadWorker(&wgParser, queue, result, manga)
	}

	var wgSaver sync.WaitGroup
	wgSaver.Add(1)
	go startSaver(&wgSaver, result, manga)

	wgParser.Wait()
	close(result)
	wgSaver.Wait()
}

func startFillWorker(queue chan int, start, end int) {
	for i := start; i <= end; i++ {
		queue <- i
	}
	close(queue)
}

func startDownloadWorker(wg *sync.WaitGroup, queue chan int, result chan malparser.Anime, manga bool) {
	defer wg.Done()

	for i := range queue {
		url := MAL_ANIME_URL + strconv.Itoa(i)
		fmt.Printf("download %v\n", url)
		dat, err := http.Get(url)
		if err != nil || dat.StatusCode != http.StatusOK {
			fmt.Printf("can't load url %v, error %v, status %v\n", url, err, dat.StatusCode)
			continue
		}
		body, err := ioutil.ReadAll(dat.Body)
		dat.Body.Close()
		if err != nil {
			fmt.Printf("can't read body %v, error %v\n", url, err)
			continue
		}
		anime, err := malparser.ParseAnimePage(body)
		if err != nil {
			fmt.Printf("can't parse url %v, error %v\n", url, err)
			continue
		}
		anime.Id = i
		result <- anime
	}
}

func startSaver(wg *sync.WaitGroup, result chan malparser.Anime, manga bool) {
	defer wg.Done()

	db, err := gorm.Open("postgres", "host=127.0.0.1 port=5432 user=user dbname=user sslmode=disable password=user")
	if err != nil {
		fmt.Printf("can't connect to db %v\n", err)
		return
	}
	defer db.Close()
	db.AutoMigrate(&malmodel.AnimeModel{})

	for anime := range result {
		fmt.Printf("save data to db %v", anime.Id)

		animeModel := malmodel.GetAnimeModelFromParsedAnime(anime)

		var count int
		db.First(&malmodel.AnimeModel{}, anime.Id).Count(&count)
		if count > 0 {
			db.Save(animeModel)
		} else {
			db.Create(animeModel)
		}
	}
}

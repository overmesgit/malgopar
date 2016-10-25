package malspider

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"malmodel"
	"malparser"
	"sync"
)

func UpdateTitlesFromTop(manga bool, workers int, pgSettings string) {
	queue := make(chan int, 1000)

	for i := 0; i < 200; i++ {
		queue <- i
	}

	var wgParser sync.WaitGroup
	for i := 0; i < workers; i++ {
		wgParser.Add(1)
		go UpdateDateFromAnimeTopWorker(&wgParser, queue, manga, pgSettings)
	}
	wgParser.Wait()
}

func UpdateDateFromAnimeTopWorker(wg *sync.WaitGroup, queue chan int, manga bool, pgSettings string) {
	defer wg.Done()

	db, err := gorm.Open("postgres", pgSettings)
	if err != nil {
		fmt.Printf("error: connect to db %v\n", err)
		return
	}
	defer db.Close()
	db.AutoMigrate(&malmodel.AnimeModel{})

	for i := range queue {
		topPageUrl := fmt.Sprintf(MALTopAnimePage, i*100)

		fmt.Printf("download %v\n", topPageUrl)
		dat, _, err := getUrlData(topPageUrl)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		animeList, err := malparser.ParseAnimeTopPage(dat)
		if len(animeList) == 0 {
			fmt.Println("stop parsing: empty top")
			break
		}

		for _, anime := range animeList {
			fmt.Printf("save data to db: animeId %v\n", anime.Id)

			animeModel := &malmodel.AnimeModel{Id: anime.Id, Title: anime.Title, Score: anime.Score}
			if animeModel.Exist(db) {
				db.Model(animeModel).Update(animeModel)
			} else {
				animeModel = malmodel.GetAnimeModelFromParsedAnime(anime)
				db.Create(animeModel)
			}
		}

	}
}

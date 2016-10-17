package malspider

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"io/ioutil"
	"malmodel"
	"malparser"
	"net/http"
	"strconv"
	"sync"
)

const (
	MALMainUrl            = "http://myanimelist.net"
	MalAnimeUrl           = MALMainUrl + "/anime/"
	MALCharactersUrl      = MALMainUrl + "/anime/%v/rand/characters"
	MALCharacterDetailUrl = MALMainUrl + "/character/%v/rand/pictures"
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

func getUrlData(url string) ([]byte, error) {
	var body []byte
	dat, err := http.Get(url)
	if err != nil || dat.StatusCode != http.StatusOK {
		return body, errors.New(fmt.Sprintf("error: load url %v, error %v, status %v\n", url, err, dat.StatusCode))
	}
	body, err = ioutil.ReadAll(dat.Body)
	dat.Body.Close()
	if err != nil {
		return body, errors.New(fmt.Sprintf("error: read body %v, error %v\n", url, err))
	}
	return body, nil
}

func startDownloadWorker(wg *sync.WaitGroup, queue chan int, result chan malparser.Anime, manga bool) {
	defer wg.Done()

	parsedCharacters := map[int]bool{}

	for i := range queue {
		url := MalAnimeUrl + strconv.Itoa(i)

		fmt.Printf("download %v\n", url)
		body, err := getUrlData(url)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		anime, err := malparser.ParseAnimePage(body)
		anime.Id = i
		if err != nil {
			fmt.Printf("error: parse url %v, error %v\n", url, err)
			continue
		}

		url = fmt.Sprintf(MALCharactersUrl, i)
		body, err = getUrlData(url)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		characters, err := malparser.ParseAnimeCharacters(body)
		if err != nil {
			fmt.Printf("error: parse url %v, error %v\n", url, err)
			continue
		}
		anime.Characters = characters

		for charIndex := range anime.Characters {
			//Update characters only one time
			if _, ok := parsedCharacters[anime.Characters[charIndex].Id]; !ok {
				url = fmt.Sprintf(MALCharacterDetailUrl, anime.Characters[charIndex].Id)
				body, err = getUrlData(url)
				if err != nil {
					fmt.Printf(err.Error())
					continue
				}
				favorites, images, err := malparser.ParseCharacterPage(body)
				if err != nil {
					fmt.Printf("error: parse character %v", anime.Characters[charIndex].Id)
					continue
				}
				anime.Characters[charIndex].Favorites = favorites
				anime.Characters[charIndex].Images = images
				parsedCharacters[anime.Characters[charIndex].Id] = true
			}

		}
		result <- anime
	}
}

func startSaver(wg *sync.WaitGroup, result chan malparser.Anime, manga bool) {
	defer wg.Done()

	db, err := gorm.Open("postgres", "host=127.0.0.1 port=5432 user=user dbname=user sslmode=disable password=user")
	if err != nil {
		fmt.Printf("error: connect to db %v\n", err)
		return
	}
	defer db.Close()
	db.AutoMigrate(&malmodel.AnimeModel{})
	db.AutoMigrate(&malmodel.CharacterModel{})

	for anime := range result {
		fmt.Printf("save data to db: animeId %v\n", anime.Id)

		animeModel := malmodel.GetAnimeModelFromParsedAnime(anime)
		err := animeModel.SaveModel(db)
		if err != nil {
			fmt.Printf("error: save data %v\n", err)
		}
		err = malmodel.SaveCharacters(anime.Characters, db)
		if err != nil {
			fmt.Printf("error: save char data %v\n", err)
		}

	}
}

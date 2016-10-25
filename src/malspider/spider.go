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
	"time"
)

const (
	MALMainUrl            = "https://myanimelist.net"
	MalAnimeUrl           = MALMainUrl + "/anime/"
	MALCharactersUrl      = MALMainUrl + "/anime/%v/rand/characters"
	MALCharacterDetailUrl = MALMainUrl + "/character/%v/rand/pictures"
	MALTopAnimePage       = MALMainUrl + "/topanime.php?limit=%v"
)

func StartSpider(manga bool, workers int, pgSettings string) {
	queue := make(chan int, 5)
	result := make(chan malparser.Anime, 100)

	go startFillWorker(queue, pgSettings)

	var wgParser sync.WaitGroup
	for i := 0; i < workers; i++ {
		wgParser.Add(1)
		go startDownloadWorker(&wgParser, queue, result, manga)
	}

	var wgSaver sync.WaitGroup
	wgSaver.Add(1)
	go startSaver(&wgSaver, result, manga, pgSettings)

	wgParser.Wait()
	close(result)
	wgSaver.Wait()
}

func startFillWorker(queue chan int, pgSettings string) {
	db, err := gorm.Open("postgres", pgSettings)
	if err != nil {
		fmt.Printf("error: connect to db %v\n", err)
		return
	}
	db.AutoMigrate(&malmodel.AnimeModel{})

	var ids []int
	db.Table("anime_models").Pluck("id", &ids)
	db.Close()

	for _, id := range ids {
		queue <- id
	}
	close(queue)
}

func getUrlData(url string) ([]byte, int, error) {
	var body []byte
	var err error
	var dat *http.Response
	var retry int
	for retry = 0; retry < 5; retry++ {
		dat, err = http.Get(url)
		if err != nil || dat.StatusCode != http.StatusTooManyRequests {
			break
		}
		dat.Body.Close()
		time.Sleep(time.Second * time.Duration(retry))
	}
	if err != nil || dat.StatusCode != http.StatusOK {
		dat.Body.Close()
		return body, dat.StatusCode, errors.New(fmt.Sprintf("error: load url %v, error %v, status %v", url, err, dat.StatusCode))
	}
	body, err = ioutil.ReadAll(dat.Body)
	dat.Body.Close()
	if err != nil {
		return body, dat.StatusCode, errors.New(fmt.Sprintf("error: read body %v, error %v", url, err))
	}
	return body, dat.StatusCode, nil
}

func startDownloadWorker(wg *sync.WaitGroup, queue chan int, result chan malparser.Anime, manga bool) {
	defer wg.Done()

	parsedCharacters := map[int]bool{}

	for i := range queue {
		titleUrl := MalAnimeUrl + strconv.Itoa(i)

		fmt.Printf("download %v\n", titleUrl)
		dat, _, err := getUrlData(titleUrl)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		anime, err := malparser.ParseAnimePage(i, dat)
		if err != nil {
			fmt.Printf("error: parse url %v, error %v\n", titleUrl, err)
			continue
		}
		charUrl := fmt.Sprintf(MALCharactersUrl, i)
		dat, _, err = getUrlData(charUrl)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		characters, err := malparser.ParseAnimeCharacters(dat)
		if err != nil {
			fmt.Printf("error: parse url %v, error %v\n", charUrl, err)
			continue
		}
		anime.Characters = characters

		for charIndex := range anime.Characters {
			//Update characters only one time
			if _, ok := parsedCharacters[anime.Characters[charIndex].Id]; !ok {
				favorites, images, err := updateCharacterDetail(&anime.Characters[charIndex])
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

func updateCharacterDetail(char *malparser.Character) (int, []string, error) {
	var imgRes []string
	charDetailUrl := fmt.Sprintf(MALCharacterDetailUrl, char.Id)
	body, _, err := getUrlData(charDetailUrl)
	if err != nil {
		return 0, imgRes, err
	}
	return malparser.ParseCharacterPage(body)

}

func startSaver(wg *sync.WaitGroup, result chan malparser.Anime, manga bool, pgSettings string) {
	defer wg.Done()

	db, err := gorm.Open("postgres", pgSettings)
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

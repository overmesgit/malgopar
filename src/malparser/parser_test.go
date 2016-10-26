package malparser

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TestTopAnimePage(t *testing.T) {
	dat, err := ioutil.ReadFile("topanime_test.html")
	check(err)
	animeList, err := ParseAnimeTopPage(dat)
	if err != nil {
		t.Error("Parser error:\n", err.Error())
	}
	if len(animeList) != 50 {
		t.Error("not all anime found", len(animeList))
	}

	for _, anime := range animeList {
		if anime.Id == 0 || anime.Title == "" || anime.Score == 0 {
			t.Error("wrong anime data", anime)
		}
	}

	firstTitle := animeList[0]
	if firstTitle.Title != "Kimi no Na wa." || firstTitle.Id != 32281 || firstTitle.Score != 9.39 {
		t.Error("wrong title", firstTitle)
	}
	fmt.Println(animeList)
}

func TestAnimeParser(t *testing.T) {
	dat, err := ioutil.ReadFile("anime_test.html")
	check(err)
	anime, err := ParseAnimePage(1, dat)
	if err != nil {
		t.Error("Parser error:\n", err.Error())
	}
	if anime.Score != 9.39 {
		t.Error("anime.Score != 9.39,", anime.Score)
	}
	if anime.ScoreCount != 42754 {
		t.Error("anime.ScoredBy != 42754,", anime.ScoreCount)
	}
	if len(anime.Related) != 2 {
		t.Error("wrong related,", anime.Related)
	}
	if anime.Title != "Kimi no Na wa." {
		t.Error("title != Kimi no Na wa.", anime.Title)
	}
	if anime.English != "Your Name." {
		t.Error("english != Your Name.", anime.English)
	}
	if anime.Status != FINISHED_AIRING_STATUS {
		t.Error("status != finished airing", anime.Status)
	}
	adaptation := false
	other := false
	for _, rel := range anime.Related {
		if rel.TitleId == 99314 && rel.Type == ADAPTATION_RELATION && rel.TitleType == MANGA_TYPE {
			adaptation = true
		}
		if rel.TitleId == 33902 && rel.Type == OTHER_RELATION && rel.TitleType == ANIME_TYPE {
			other = true
		}
	}
	if !adaptation || !other {
		t.Error("wrong related,", anime.Related)
	}
}

func TestStrangeAnimeParser(t *testing.T) {
	dat, err := ioutil.ReadFile("anime_test_strange.html")
	check(err)
	anime, err := ParseAnimePage(1, dat)
	if err != nil {
		t.Error("Parser error:\n", err.Error())
	}
	if anime.Score != 6.70 {
		t.Error("anime.Score != 6.70,", anime.Score)
	}
	if anime.ScoreCount != 43 {
		t.Error("anime.ScoredBy != 43,", anime.ScoreCount)
	}
}

package malparser

import (
	"io/ioutil"
	"testing"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TestAnimeParser(t *testing.T) {
	dat, err := ioutil.ReadFile("anime_test.html")
	check(err)
	anime, err := ParseAnimePage(dat)
	if err != nil {
		t.Error("Parser error:\n", err.Error())
	}
	if anime.Score != 9.41 {
		t.Error("anime.Score != 9.41,", anime.Score)
	}
	if anime.ScoreCount != 15616 {
		t.Error("anime.ScoredBy != 15616,", anime.ScoreCount)
	}
}

func TestMangaParser(t *testing.T) {
	if 1 != 1 {
		t.Error("1 != 2")
	}
}

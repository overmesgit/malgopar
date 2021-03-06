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

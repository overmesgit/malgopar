package malmodel

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"mal/parser"
)

type AnimeModel struct {
	gorm.Model
	Score      float64
	ScoreCount int
}

func NewAnimeModel() {

}

func GetAnimeModelFromParsedAnime(anime malparser.Anime) *AnimeModel {
	return &AnimeModel{Model: gorm.Model{ID: uint(anime.Id)}, Score: anime.Score, ScoreCount: anime.ScoreCount}
}

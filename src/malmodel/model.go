package malmodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"malparser"
	"strconv"
)

type StoredChar struct {
	Id   int
	Main bool
}

type AnimeModel struct {
	Id          int     `gorm:"primary_key"`
	Score       float64 `gorm:"index"`
	ScoreCount  int     `gorm:"index"`
	Title       string
	English     string
	Group       int    `gorm:"index"`
	RelatedJSON string `sql:"type:jsonb"`
	CharsJSON   string `sql:"type:jsonb"`
}

type CharacterModel struct {
	Id         int `gorm:"primary_key"`
	Name       string
	Favorites  int
	ImagesJSON string `sql:"type:jsonb"`
}

func (c CharacterModel) GetImages() []string {
	images := make([]string, 0)
	json.Unmarshal([]byte(c.ImagesJSON), &images)
	return images
}

func (a AnimeModel) GetStoredChars() []StoredChar {
	chars := make([]StoredChar, 0)
	json.Unmarshal([]byte(a.CharsJSON), &chars)
	return chars
}

func (a AnimeModel) GetRelatedCharacters(db *gorm.DB) ([]CharacterModel, error) {
	storedChars := a.GetStoredChars()
	titleCharactersIds := make([]string, len(storedChars))
	for i, char := range storedChars {
		titleCharactersIds[i] = strconv.Itoa(char.Id)
	}
	var characters []CharacterModel
	query := db.Where("id in (?)", titleCharactersIds).Find(&characters)
	errs := query.GetErrors()
	if len(errs) > 0 {
		return characters, errors.New(fmt.Sprint(errs))
	}
	return characters, nil
}

func (m *AnimeModel) SaveModel(db *gorm.DB) error {
	var count int
	db.Find(&AnimeModel{Id: m.Id}).Count(&count)

	var query *gorm.DB
	if count > 0 {
		query = db.Save(m)
	} else {
		query = db.Create(m)
	}
	errs := query.GetErrors()
	if len(errs) > 0 {
		return errors.New(fmt.Sprint(errs))
	}
	return nil
}

func GetAnimeModelFromParsedAnime(anime malparser.Anime) *AnimeModel {
	relatedJson, _ := json.Marshal(anime.Related)

	animeChars := make([]StoredChar, len(anime.Characters))
	for _, char := range anime.Characters {
		animeChars = append(animeChars, StoredChar{Id: char.Id, Main: char.Main})
	}
	charsJson, _ := json.Marshal(animeChars)

	model := AnimeModel{Id: anime.Id, Score: anime.Score, ScoreCount: anime.ScoreCount, Title: anime.Title,
		English: anime.English, RelatedJSON: string(relatedJson), CharsJSON: string(charsJson)}

	return &model
}

func SaveCharacters(characters malparser.CharacterSlice, db *gorm.DB) error {
	for _, char := range characters {
		var count int
		db.Find(&CharacterModel{Id: char.Id}).Count(&count)
		imagesJson, _ := json.Marshal(char.Images)
		m := CharacterModel{Id: char.Id, Name: char.Name, Favorites: char.Favorites,
			ImagesJSON: string(imagesJson)}
		if count == 0 || len(char.Images) > 0 {
			var query *gorm.DB
			if count > 0 {
				query = db.Save(&m)
			} else {
				query = db.Create(&m)
			}
			errs := query.GetErrors()
			if len(errs) > 0 {
				return errors.New(fmt.Sprint(errs))
			}
		}
	}
	return nil
}

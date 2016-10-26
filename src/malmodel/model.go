package malmodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"malparser"
)

type StoredCharSlice []StoredChar
type StoredChar struct {
	Id   int
	Main bool
}

type AnimeModel struct {
	Id          int                  `gorm:"primary_key"`
	Score       float64              `gorm:"index"`
	ScoreCount  int                  `gorm:"index"`
	Status      malparser.StatusType `gorm:"index"`
	Title       string
	English     string
	GroupId     int    `gorm:"index"`
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

func (a AnimeModel) GetStoredChars() StoredCharSlice {
	chars := make(StoredCharSlice, 0)
	json.Unmarshal([]byte(a.CharsJSON), &chars)
	return chars
}

func (chars StoredCharSlice) GetIds() []int {
	var titleCharacters []int
	for _, char := range chars {
		titleCharacters = append(titleCharacters, char.Id)
	}
	return titleCharacters
}

func (chars StoredCharSlice) GetMainCharsMap() map[int]bool {
	res := map[int]bool{}
	for _, char := range chars {
		if char.Main {
			res[char.Id] = true
		}
	}
	return res
}

func (a *AnimeModel) GetRelatedTitles() malparser.RelationSlice {
	relatedTitles := make(malparser.RelationSlice, 0)
	json.Unmarshal([]byte(a.RelatedJSON), &relatedTitles)
	return relatedTitles
}

func (a *AnimeModel) GetRelatedCharacters(db *gorm.DB) ([]CharacterModel, map[int]bool, error) {
	storedChars := a.GetStoredChars()
	mainMap := storedChars.GetMainCharsMap()
	var characters []CharacterModel
	query := db.Where("id in (?)", storedChars.GetIds()).Find(&characters)
	if errs := query.GetErrors(); len(errs) > 0 {
		return characters, mainMap, errors.New(fmt.Sprint(errs))
	}
	return characters, mainMap, nil
}

func (m *AnimeModel) SaveModel(db *gorm.DB) error {
	var count int
	saveModel := AnimeModel{Id: m.Id}
	db.First(&saveModel).Count(&count)

	var query *gorm.DB
	if count > 0 {
		m.GroupId = saveModel.GroupId
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

func (m *AnimeModel) Exist(db *gorm.DB) bool {
	var count int
	saveModel := AnimeModel{Id: m.Id}
	db.First(&saveModel).Count(&count)
	return count > 0
}

func GetAnimeModelFromParsedAnime(anime malparser.Anime) *AnimeModel {
	if anime.Related == nil {
		anime.Related = make(malparser.RelationSlice, 0)
	}
	relatedJson, _ := json.Marshal(anime.Related)

	animeChars := make([]StoredChar, 0)
	for _, char := range anime.Characters {
		animeChars = append(animeChars, StoredChar{Id: char.Id, Main: char.Main})
	}
	charsJson, _ := json.Marshal(animeChars)

	model := AnimeModel{Id: anime.Id, Score: anime.Score, ScoreCount: anime.ScoreCount, Status: anime.Status, Title: anime.Title,
		English: anime.English, RelatedJSON: string(relatedJson), CharsJSON: string(charsJson), GroupId: anime.Id}

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

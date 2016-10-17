package malparser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"strings"
	"time"
)

const (
	ANIME_TYPE = 0
	MANGA_TYPE = 1

	ADAPTATION_RELATION          = 0
	SIDE_STORY_RELATION          = 1
	SEQUEL_RELATION              = 2
	ALTERNATIVE_VERSION_RELATION = 3
	SPIN_OFF_RELATION            = 4
	ALTERNATIVE_SETTING_RELATION = 5
	CHARACTER_RELATION           = 6
	OTHER_RELATION               = 7
	PREQUEL_RELATION             = 8
	PARENT_STORY_RELATION        = 9
	FULL_STORY_RELATION          = 10
	SUMMARY_RELATION             = 11
)

type Relation struct {
	TitleId   int
	Type      int
	TitleType int
}

type Anime struct {
	Id         int
	Title      string
	English    string
	Japanese   string
	Type       string
	Episodes   int
	Status     int
	AiredFrom  time.Time
	AiredTo    time.Time
	Producers  []string
	Genres     []string
	Duration   int
	Rating     string
	Score      float64
	ScoreCount int
	Ranked     int
	Popularity int
	Members    int
	Favorites  int
	Related    []Relation
	Characters CharacterSlice
}

type ParserError struct {
	errors []error
}

func NewParserError() *ParserError {
	return &ParserError{make([]error, 0)}
}

func (p *ParserError) Add(err error) {
	p.errors = append(p.errors, err)
}

func (p *ParserError) GetError() error {
	if len(p.errors) > 0 {
		strErrors := make([]string, len(p.errors))
		for i, err := range p.errors {
			strErrors[i] = err.Error()
		}
		return errors.New(strings.Join(strErrors, "\n"))
	}
	return nil
}

func ParseAnimePage(pageHTML []byte) (Anime, error) {
	res := Anime{}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	if err != nil {
		return res, err
	}

	parserError := NewParserError()

	res.Score = GetScore(doc, parserError)
	res.ScoreCount = GetScoreCount(doc, parserError)
	res.Related = GetRelated(doc, parserError)
	res.Title = GetTitle(doc, parserError)
	res.English = GetEnglish(doc, parserError)

	return res, parserError.GetError()
}

func GetTitle(doc *goquery.Document, parserError *ParserError) string {
	return doc.Find("h1 span").Text()
}

func GetEnglish(doc *goquery.Document, parserError *ParserError) string {
	rowText := doc.Find(`span:contains("English:")`).Parent().Text()
	return strings.Trim(strings.Replace(rowText, "English:", "", -1), " \n")
}

func GetScore(doc *goquery.Document, parserError *ParserError) float64 {
	scoreText := doc.Find(`[itemprop="ratingValue"]`).Text()
	res, err := strconv.ParseFloat(scoreText, 64)
	if err != nil {
		parserError.Add(errors.New(fmt.Sprintf("GetScore error: %v", err.Error())))
	}
	return res
}

func GetScoreCount(doc *goquery.Document, parserError *ParserError) int {
	scoreText := doc.Find(`[itemprop="ratingCount"]`).Text()
	scoreText = strings.Replace(scoreText, ",", "", 100)
	res, err := strconv.Atoi(scoreText)
	if err != nil {
		parserError.Add(errors.New(fmt.Sprintf("GetScoreCount error: %v", err.Error())))
	}
	return int(res)
}

var RelationMap = map[string]int{
	"adaptation":          ADAPTATION_RELATION,
	"side story":          SIDE_STORY_RELATION,
	"sequel":              SEQUEL_RELATION,
	"alternative version": ALTERNATIVE_VERSION_RELATION,
	"spin-off":            SPIN_OFF_RELATION,
	"alternative setting": ALTERNATIVE_SETTING_RELATION,
	"character":           CHARACTER_RELATION,
	"other":               OTHER_RELATION,
	"prequel":             PREQUEL_RELATION,
	"parent story":        PARENT_STORY_RELATION,
	"full story":          FULL_STORY_RELATION,
	"summary":             SUMMARY_RELATION,
}

var IdTypeMap = map[string]int{
	"anime": ANIME_TYPE,
	"manga": MANGA_TYPE,
}

func GetRelated(doc *goquery.Document, parserError *ParserError) []Relation {
	relations := make([]Relation, 0)
	doc.Find(".anime_detail_related_anime tr").Each(func(i int, tr *goquery.Selection) {
		relation := tr.Find("td").First().Text()
		relation = strings.Replace(relation, ":", "", -1)
		relation = strings.ToLower(relation)
		relationType, ok := RelationMap[relation]
		if !ok {
			parserError.Add(errors.New(fmt.Sprintf("GetRelated error: not found %v", relation)))
			return
		}
		tr.Find("a").Each(func(j int, link *goquery.Selection) {
			href, _ := link.Attr("href")
			splitUrl := strings.Split(href, "/")
			idType := IdTypeMap[splitUrl[1]]
			id, err := strconv.Atoi(splitUrl[2])
			if err != nil {
				parserError.Add(errors.New(fmt.Sprintf("GetRelated error: %v", err.Error())))
				return
			}
			relations = append(relations, Relation{TitleId: id, TitleType: idType, Type: relationType})
		})
	})
	return relations
}

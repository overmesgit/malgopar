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

type Relation struct {
	Id   int
	Name int
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
}

type ParserError struct {
	errors []error
}

func NewParserError() *ParserError {
	return &ParserError{make([]error, 0)}
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

	return res, parserError.GetError()
}

func GetScore(doc *goquery.Document, parserError *ParserError) float64 {
	scoreText := doc.Find(`[itemprop="ratingValue"]`).Text()
	res, err := strconv.ParseFloat(scoreText, 64)
	if err != nil {
		parserError.errors = append(parserError.errors, errors.New(fmt.Sprintf("GetScore error: %v", err.Error())))
	}
	return res
}

func GetScoreCount(doc *goquery.Document, parserError *ParserError) int {
	scoreText := doc.Find(`[itemprop="ratingCount"]`).Text()
	scoreText = strings.Replace(scoreText, ",", "", 100)
	res, err := strconv.Atoi(scoreText)
	if err != nil {
		parserError.errors = append(parserError.errors, errors.New(fmt.Sprintf("GetScoreCount error: %v", err.Error())))
	}
	return int(res)
}

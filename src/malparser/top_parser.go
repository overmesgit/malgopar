package malparser

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"strings"
)

func ParseAnimeTopPage(pageHTML []byte) ([]Anime, error) {
	res := make([]Anime, 0)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	if err != nil {
		return res, err
	}
	doc.Find(".ranking-list").Each(func(i int, s *goquery.Selection) {
		titleLink := s.Find(".detail a").First()
		title := titleLink.Text()

		attr, _ := titleLink.Attr("href")
		urlSplit := strings.Split(attr, "/")
		titleId, err := strconv.Atoi(urlSplit[len(urlSplit)-2])
		if err != nil {
			fmt.Println("error: can't parse anime id", err)
		}

		score, err := strconv.ParseFloat(strings.Trim(s.Find(".score").Text(), " \n"), 64)
		if err != nil {
			fmt.Println("error: can't parse anime score", err)
		}
		res = append(res, Anime{Id: titleId, Title: title, Score: score})

	})

	return res, nil
}

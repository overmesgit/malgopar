package malparser

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strconv"
	"strings"
)

type Character struct {
	Id        int
	Name      string
	Main      bool
	Favorites int
	Images    []string
}

type CharacterSlice []Character

func (c CharacterSlice) GetIds() []int {
	result := make([]int, len(c))
	for i, v := range c {
		result[i] = v.Id
	}
	return result
}

func (c CharacterSlice) Len() int {
	return len(c)
}

func (c CharacterSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c CharacterSlice) Less(i, j int) bool {
	return c[i].Favorites < c[j].Favorites
}

func ParseAnimeCharacters(pageHTML []byte) (CharacterSlice, error) {
	result := make(CharacterSlice, 0)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	if err != nil {
		return result, err
	}

	doc.Find(".js-scrollfix-bottom-rel tr").Each(func(i int, s *goquery.Selection) {
		fileUrl, charName, charType, charUrl := "", "", "", ""
		s.Find(`td > div .picSurround > a[href^="/character"]`).Each(func(i int, s *goquery.Selection) {
			fileUrl = s.Find("img").AttrOr("data-src", "")
		})
		if fileUrl != "" {
			charInfo := s.Find(`td > a[href^="/character"]`)
			charName = charInfo.Text()
			charUrl = charInfo.AttrOr("href", "")
			charType = s.Find(`td > div > small`).Text()
		}

		if fileUrl != "" && !strings.Contains(fileUrl, "questionmark") {
			charId, err := strconv.Atoi(strings.Split(charUrl, "/")[2])
			if err != nil {
				fmt.Printf("error: parse char id %v\n", err.Error())
			}
			character := Character{Id: charId, Name: charName, Favorites: 0, Images: make([]string, 0)}
			if charType == "Main" {
				character.Main = true
			}
			result = append(result, character)
		}
	})

	return result, nil
}

func ParseCharacterPage(pageHTML []byte) (int, []string, error) {
	images := make([]string, 0)
	var favorites int

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))

	if err != nil {
		return favorites, images, err
	}

	var favorRegexp = regexp.MustCompile(`Member Favorites: ([\d]*)`)
	favorFullString := favorRegexp.FindString(doc.Text())
	favorString := strings.TrimSpace(strings.Split(favorFullString, ":")[1])
	favorites, err = strconv.Atoi(favorString)
	if err != nil {
		return favorites, images, err
	}

	mainImage := doc.Find(`[style="text-align: center;"] img`).AttrOr("src", "")
	if mainImage != "" {
		images = append(images, mainImage)
	}

	doc.Find(`[width="225"][align="center"] img`).Each(func(i int, s *goquery.Selection) {
		imageUrl := s.AttrOr("src", "")
		if imageUrl != "" && imageUrl != mainImage {
			images = append(images, imageUrl)
		}
	})
	return favorites, images, nil
}

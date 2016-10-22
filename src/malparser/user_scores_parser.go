package malparser

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type UserScoresParserError struct {
	Msg  string
	Name string
	Id   int
}

func (e *UserScoresParserError) Error() string {
	return fmt.Sprintf("id: %v name: %v msg: %v", e.Name, e.Id, e.Msg)
}

type Title struct {
	Status      uint   `xml:"my_status"`
	Score       uint   `xml:"my_score"`
	LastUpdate  int    `xml:"my_last_updated"`
	MyStartDate string `xml:"my_start_date"`
	MyLastDate  string `xml:"my_finish_date"`
}

type TitleFace interface {
	GetScore() uint
}

type AnimeTitle struct {
	Title
	Id uint `xml:"series_animedb_id"`
}

func (t AnimeTitle) GetScore() uint {
	return t.Score
}

type MangaTitle struct {
	Title
	Id uint `xml:"series_mangadb_id"`
}

func (t MangaTitle) GetScore() uint {
	return t.Score
}

type TitleLastUpdater interface {
	SetLastUpdate()
}

func GetMALDate(date string) int {
	layout := "2006-01-02"
	date = strings.Replace(date, "-00", "-01", -1)
	myLast, err := time.Parse(layout, date)
	//int32 unix time
	if myLast.Year() < 1950 || myLast.Year() > 2030 {
		return 0
	}
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return int(myLast.Unix())
}

func (t *Title) SetLastUpdate() {
	if t.LastUpdate == 0 {
		if !strings.Contains(t.MyLastDate, "0000") {
			t.LastUpdate = GetMALDate(t.MyLastDate)
		} else {
			if !strings.Contains(t.MyStartDate, "0000") {
				t.LastUpdate = GetMALDate(t.MyStartDate)
			}
		}
	}
}

type MyInfo struct {
	UserId int `xml:"user_id"`
}

type Result struct {
	XMLName xml.Name `xml:"myanimelist"`
	MyInfo  MyInfo   `xml:"myinfo"`
}

type MalApiError struct {
	XMLName xml.Name `xml:"myanimelist"`
	Error   string   `xml:"error"`
}

type AnimeResult struct {
	Result
	TitleList []AnimeTitle `xml:"anime"`
}

type MangaResult struct {
	Result
	TitleList []MangaTitle `xml:"manga"`
}

type UserList struct {
	UserId    int
	UserName  string
	AnimeList []AnimeTitle
	MangaList []MangaTitle
}

type UserListParser interface {
	ParseAnime(io.Reader)
	ParseManga(io.Reader)
	ScoresCount() int
	ToArrayFormat()
}

func TitlesAvg(scoresSlice interface{}) float32 {
	sum := 0
	count := 0
	scores := reflect.ValueOf(scoresSlice)
	for i := 0; i < scores.Len(); i++ {
		s := scores.Index(i).Interface().(TitleFace)
		if s.GetScore() > 0 {
			sum += int(s.GetScore())
			count++
		}
	}
	return float32(sum) / float32(count)
}

func (l *UserList) AnimeAvg() float32 {
	return TitlesAvg(l.AnimeList)
}

func (l *UserList) MangaAvg() float32 {
	return TitlesAvg(l.MangaList)
}

func (l *UserList) ToArrayFormat() map[string][3]uint {
	result := map[string][3]uint{}
	for i := range l.AnimeList {
		current := l.AnimeList[i]
		scoreArray := [3]uint{current.Score, current.Status, uint(current.LastUpdate)}
		result[strconv.Itoa(int(current.Id))] = scoreArray
	}
	for i := range l.MangaList {
		current := l.MangaList[i]
		scoreArray := [3]uint{current.Score, current.Status, uint(current.LastUpdate)}
		result[strconv.Itoa(-int(current.Id))] = scoreArray
	}
	return result
}

func (l *UserList) ScoresCount() int {
	return len(l.AnimeList) + len(l.MangaList)
}

func (l *UserList) ParseTitles(data []byte, type_ string) error {
	var err error

	switch {
	case type_ == "anime":
		v := AnimeResult{}
		err = xml.Unmarshal(data, &v)
		if err != nil {
			return err
		}

		for i := range v.TitleList {
			v.TitleList[i].SetLastUpdate()
		}
		l.AnimeList = v.TitleList
		l.UserId = v.MyInfo.UserId
	case type_ == "manga":
		v := MangaResult{}
		err = xml.Unmarshal(data, &v)
		if err != nil {
			fmt.Printf("error: %v", err)
			return err
		}

		for i := range v.TitleList {
			v.TitleList[i].SetLastUpdate()
		}
		l.MangaList = v.TitleList
		l.UserId = v.MyInfo.UserId
	case true:
		return errors.New("Wrong type")
	}
	return nil

}

const (
	mainUrl    = "http://myanimelist.net"
	profileUrl = "/comments.php?id=%v"
	apiUrl     = "/malappinfo.php?u=%v&status=all&type=%s"
)

func GetUserNameById(userId int, retry int) (string, error) {
	fullUrl := mainUrl + fmt.Sprintf(profileUrl, userId)
	var err error
	var resp *http.Response
	for i := 0; i <= retry; i++ {
		resp, err = http.Get(fullUrl)
		if err == nil {
			break
		}
		fmt.Println(err)

	}
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("User name not found %v", userId))
	}
	body := make([]byte, 250)
	resp.Body.Read(body)

	userNameRegexp := regexp.MustCompile(`<title>([\w\W]*)&#039;s Comments[\w\W]*<\/title>`)
	userNameMatch := userNameRegexp.FindStringSubmatch(string(body))
	if len(userNameMatch) == 0 {
		return "", errors.New(fmt.Sprintf("User name not found %v", userId))
	}
	return strings.TrimSpace(userNameMatch[1]), nil
}

func GetUserScoresById(userId int, retry int) (UserList, error) {
	userName, err := GetUserNameById(userId, 3)
	if err != nil {
		return UserList{UserId: userId}, err
	}
	userList, err := GetUserScoresByName(userName, 3)
	userList.UserId = userId
	return userList, err
}

func getUserApiPage(url string) ([]byte, error) {
	var body []byte
	resp, err := http.Get(url)
	if err != nil {
		return body, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if resp.StatusCode == 429 {
			fmt.Println("Get user page error 429")
			time.Sleep(3 * time.Second)
		}
		return body, errors.New(fmt.Sprintf("User page error %v %v", url, resp.StatusCode))
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	v := MalApiError{}
	err = xml.Unmarshal(body, &v)
	if v.Error != "" {
		return body, errors.New(v.Error)
	}
	return body, nil

}

func GetUserScoresByName(userName string, retry int) (UserList, error) {
	userList := UserList{UserName: userName}
	for _, content := range [2]string{"anime", "manga"} {
		url := mainUrl + fmt.Sprintf(apiUrl, userName, content)

		var err error
		var body []byte
		for try := 1; try == 1 || (try <= retry && err != nil); try++ {
			body, err = getUserApiPage(url)
		}
		if err != nil {
			return userList, err
		}
		err = userList.ParseTitles(body, content)
		if err != nil {
			return userList, err
		}
	}
	return userList, nil
}

package malmodel

import (
	"encoding/json"
	"malparser"
	"sort"
	"testing"
)

func (r rel) json() string {
	json, _ := json.Marshal(r)
	return string(json)
}

var adap = malparser.ADAPTATION_RELATION
var side = malparser.SIDE_STORY_RELATION
var seq = malparser.SEQUEL_RELATION
var alt = malparser.ALTERNATIVE_VERSION_RELATION
var spin = malparser.SPIN_OFF_RELATION
var set = malparser.ALTERNATIVE_SETTING_RELATION
var char = malparser.CHARACTER_RELATION
var other = malparser.OTHER_RELATION
var preq = malparser.PREQUEL_RELATION
var par = malparser.PARENT_STORY_RELATION
var full = malparser.FULL_STORY_RELATION
var sum = malparser.SUMMARY_RELATION

var ani = malparser.ANIME_TYPE
var man = malparser.MANGA_TYPE

func ItemsEqual(l, r []int) bool {
	if len(r) != len(l) {
		return false
	}
	sort.Ints(r)
	sort.Ints(l)
	for i, v := range r {
		if l[i] != v {
			return false
		}
	}
	return true
}

type rel []malparser.Relation

func TestSimpleGroup(t *testing.T) {
	grouper := NewTitleGrouper()

	anime1 := AnimeModel{Id: 1, RelatedJSON: rel{}.json(), GroupId: 1}
	anime2 := AnimeModel{Id: 2, RelatedJSON: rel{{1, par, ani}}.json()}

	grouper.GroupModels([]AnimeModel{anime1, anime2})
	result := grouper.TitleGroups
	if result[anime1.Id] != anime1.Id || result[anime2.Id] != anime1.Id {
		t.Errorf("wrong group %v", result)
	}

	anime3 := AnimeModel{Id: 3, RelatedJSON: rel{{4, par, ani}}.json()}
	anime4 := AnimeModel{Id: 4, RelatedJSON: rel{}.json()}
	grouper.GroupModels([]AnimeModel{anime3, anime4})
	if result[anime3.Id] != anime3.Id || result[anime4.Id] != anime3.Id {
		t.Errorf("wrong group %v", result)
	}

	anime6 := AnimeModel{Id: 6}
	grouper.GroupModels([]AnimeModel{anime6})

	anime5 := AnimeModel{Id: 5, RelatedJSON: rel{{2, par, ani}, {3, par, ani}}.json()}
	grouper.GroupModels([]AnimeModel{anime5})
	if result[anime5.Id] != anime1.Id || result[anime3.Id] != anime1.Id {
		t.Errorf("wrong group %v", result)
	}

	changedGroups := grouper.GetChangedGroups()
	if len(changedGroups) != 1 || !ItemsEqual(changedGroups[anime1.Id], []int{anime2.Id, anime3.Id, anime4.Id, anime5.Id}) {
		t.Errorf("wrong group %v", changedGroups)
	}

}

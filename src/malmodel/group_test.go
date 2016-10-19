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
	anime1 := AnimeModel{Id: 1, RelatedJSON: rel{}.json()}
	anime2 := AnimeModel{Id: 2, RelatedJSON: rel{{1, par, ani}}.json()}

	grouper := NewTitleGrouper()
	grouper.GroupModels([]AnimeModel{anime1, anime2})
	result := grouper.GetChangedGroups()
	if ids, ok := result[2]; !ok || !ItemsEqual(ids, []int{anime2.Id}) {
		t.Errorf("wrong group %v", ids)
	}

}

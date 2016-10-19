package malmodel

import (
	"encoding/json"
	"testing"
)

func relToJson(rel []Relation) string {
	json, _ := json.Marshal(rel)
	return string(json)
}

func TestSimpleGroup(t *testing.T) {
	anime1 := malmodel.AnimeModel{Id: 1, RelatedJSON: relToJson([]Relation{})}
	anime2 := malmodel.AnimeModel{Id: 2, RelatedJSON: relToJson([]Relation{{1, PARENT_STORY_RELATION, ANIME_TYPE}})}
}

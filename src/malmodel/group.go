package malmodel

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"malparser"
)

type TitleGrouper struct {
	TitleGroups    map[int]int
	PreviousGroups map[int]int
}

func NewTitleGrouper() *TitleGrouper {
	return &TitleGrouper{make(map[int]int), make(map[int]int)}
}

func GroupAllAnimeModels(pgSettings string) {
	db, err := gorm.Open("postgres", pgSettings)
	if err != nil {
		fmt.Printf("error: connect to db %v\n", err)
		return
	}
	defer db.Close()
	db.AutoMigrate(&AnimeModel{})

	grouper := NewTitleGrouper()
	var models []AnimeModel
	page := 0
	for {
		pageLimit := 100
		db.Order("id").Limit(pageLimit).Offset(page * pageLimit).Find(&models)
		if len(models) > 0 {
			page++
			grouper.GroupModels(models)
		} else {
			break
		}
	}
	changedGroups := grouper.GetChangedGroups()
	for group, modelIds := range changedGroups {
		fmt.Printf("Change group %v for %v\n", group, modelIds)
		db.Table("anime_models").Where("id in (?)", modelIds).UpdateColumn("group_id", group)
	}
}

func (g *TitleGrouper) getRoot(src int) int {
	currentParent := g.TitleGroups[src]
	currentSrc := src
	for currentParent != currentSrc {
		currentSrc = currentParent
		currentParent = g.TitleGroups[currentParent]
	}
	return currentParent
}

func (g *TitleGrouper) addNode(src int) {
	if _, ok := g.TitleGroups[src]; !ok {
		g.TitleGroups[src] = src
	}
}

func (g *TitleGrouper) link(srcRoot, dstRoot int) {
	if srcRoot < dstRoot {
		g.TitleGroups[dstRoot] = srcRoot
	} else {
		g.TitleGroups[srcRoot] = dstRoot
	}
}

func (g *TitleGrouper) addRelation(src, dst int) {
	g.addNode(src)
	g.addNode(dst)

	srcRoot := g.getRoot(src)
	dstRoot := g.getRoot(dst)
	if srcRoot != dstRoot {
		g.link(srcRoot, dstRoot)
	}
}

func (g *TitleGrouper) GroupModels(models []AnimeModel) {
	for modelIndex := range models {
		g.PreviousGroups[models[modelIndex].Id] = models[modelIndex].GroupId
		modelsRelations := models[modelIndex].GetRelatedTitles()
		g.addNode(models[modelIndex].Id)
		for relationIndex := range modelsRelations {
			tType := modelsRelations[relationIndex].Type
			if tType != malparser.ADAPTATION_RELATION && tType != malparser.OTHER_RELATION && tType != malparser.CHARACTER_RELATION {
				if modelsRelations[relationIndex].TitleType != malparser.MANGA_TYPE {
					g.addRelation(models[modelIndex].Id, modelsRelations[relationIndex].TitleId)
				}
			}
		}
	}

}

func (g *TitleGrouper) GetChangedGroups() map[int][]int {
	result := make(map[int][]int)
	for titleId, previousGroup := range g.PreviousGroups {
		currentGroup := g.getRoot(titleId)
		if currentGroup != previousGroup {
			if _, ok := result[currentGroup]; !ok {
				result[currentGroup] = make([]int, 0)
			}
			result[currentGroup] = append(result[currentGroup], titleId)
		}
	}

	return result
}

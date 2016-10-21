package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"malmodel"
	"os"
)

func main() {
	db, err := gorm.Open("postgres", os.Args[1])
	if err != nil {
		fmt.Printf("error: connect to db %v\n", err)
		return
	}
	defer db.Close()
	db.AutoMigrate(&malmodel.AnimeModel{})

	grouper := malmodel.NewTitleGrouper()
	var models []malmodel.AnimeModel
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

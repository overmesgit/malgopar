package malmodel

type TitleGrouper struct {
	TitleGroups map[int]int
	PreviousGroups map[int]int
}

func NewTitleGrouper() *TitleGrouper {
	return &TitleGrouper{}
}

func (g *TitleGrouper) GroupModels([]AnimeModel) {

}

func (g *TitleGrouper) GetChangedGroups() map[int][]int {
	return make(map[int][]int)
}

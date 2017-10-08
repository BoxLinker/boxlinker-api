package boxlinker

type PageConfig struct {
	CurrentPage int
	PageCount   int
	TotalCount  int
}

func (pc PageConfig) Limit() int {
	return pc.PageCount
}
func (pc PageConfig) Offset() int {
	return pc.PageCount * (pc.CurrentPage - 1)
}

func (pc PageConfig) PaginationJSON() map[string]int {
	m := map[string]int{}
	m["current_page"] = pc.CurrentPage
	m["page_count"] = pc.PageCount
	m["total_count"] = pc.TotalCount
	return m
}

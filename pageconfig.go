package boxlinker


type PageConfig struct {
	CurrentPage int
	PageCount int
}


func (pc PageConfig) Limit() int {
	return pc.PageCount
}
func (pc PageConfig) Offset() int {
	return pc.PageCount*(pc.CurrentPage - 1)
}


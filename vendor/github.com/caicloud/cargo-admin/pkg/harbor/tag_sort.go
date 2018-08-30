package harbor

type TagsSortByDateDes []*HarborTag

func (r TagsSortByDateDes) Len() int {
	return len(r)
}

func (r TagsSortByDateDes) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r TagsSortByDateDes) Less(i, j int) bool {
	return r[i].Created.After(r[j].Created)
}

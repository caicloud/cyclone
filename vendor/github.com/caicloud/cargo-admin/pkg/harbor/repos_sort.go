package harbor

type ReposSortByNameAsc []*HarborRepo

func (r ReposSortByNameAsc) Len() int {
	return len(r)
}

func (r ReposSortByNameAsc) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ReposSortByNameAsc) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

type ReposSortByNameDes []*HarborRepo

func (r ReposSortByNameDes) Len() int {
	return len(r)
}

func (r ReposSortByNameDes) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ReposSortByNameDes) Less(i, j int) bool {
	return r[i].Name > r[j].Name
}

type ReposSortByDateAsc []*HarborRepo

func (r ReposSortByDateAsc) Len() int {
	return len(r)
}

func (r ReposSortByDateAsc) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ReposSortByDateAsc) Less(i, j int) bool {
	return r[i].CreationTime.Before(r[j].CreationTime)
}

type ReposSortByDateDes []*HarborRepo

func (r ReposSortByDateDes) Len() int {
	return len(r)
}

func (r ReposSortByDateDes) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ReposSortByDateDes) Less(i, j int) bool {
	return r[i].CreationTime.After(r[j].CreationTime)
}

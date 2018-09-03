package form

type Pagination struct {
	Start int `source:"query,start,default=0"`
	Limit int `source:"query,limit,default=99999"`
}

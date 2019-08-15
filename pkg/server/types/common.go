package types

import "reflect"

// QueryParams describes pagination of request, start and limit defined.
type QueryParams struct {
	Start  uint64 `source:"query,start,default=0"`
	Limit  uint64 `source:"query,limit,default=99999"`
	Filter string `source:"query,filter"`
	// Sort will sorts the results by metadata.creationTimestamp
	Sort bool `source:"query,sort"`
	// Ascending will sort the results by ascending order, otherwise by descending order
	Ascending bool `source:"query,ascending"`
	// Detail determines whether to return details of list items.
	Detail bool `source:"query,detail"`
}

// ListMeta describes the structure of list metadata
type ListMeta struct {
	Total       int `json:"total"`
	ItemsLength int `json:"itemsLength"`
}

// ListResponse describes a list
type ListResponse struct {
	Metadata ListMeta    `json:"metadata"`
	Items    interface{} `json:"items"`
}

// NewListResponse create a ListResponse of a list.
func NewListResponse(total int, items interface{}) *ListResponse {
	value := reflect.ValueOf(items)
	typ := value.Type()
	kind := typ.Kind()
	if kind != reflect.Array && kind != reflect.Slice {
		panic("items must be an array or slice.")
	}
	return &ListResponse{
		ListMeta{
			Total:       total,
			ItemsLength: value.Len(),
		},
		items,
	}
}

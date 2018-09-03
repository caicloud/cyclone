/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package types

import "reflect"

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

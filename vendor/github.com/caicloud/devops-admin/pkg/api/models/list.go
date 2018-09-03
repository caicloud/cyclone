/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package models

import "reflect"

// ListMeta describes the structure of list metadata
type ListMeta struct {
	Total int `json:"total"`
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
		items = []interface{}{items}
	}
	return &ListResponse{
		ListMeta{
			Total: total,
		},
		items,
	}
}

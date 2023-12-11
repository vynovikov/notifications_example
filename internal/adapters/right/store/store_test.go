package store

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type storeSouite struct {
	suite.Suite
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(storeSouite))
}

func (s *storeSouite) TestDoFilter() {
	tt := []struct {
		name      string
		initData  []map[string]interface{}
		filters   []map[string]interface{}
		wantData  []map[string]interface{}
		wantError error
	}{
		{
			name: "dayTime",
			initData: []map[string]interface{}{
				{
					"name":       "azaza",
					"created_at": "2022-10-03T12:43:46.000000Z",
				},
				{
					"name":       "bzbzb",
					"created_at": "2022-10-04T12:43:46.000000Z",
				},
				{
					"name":       "czczc",
					"created_at": "2022-10-05T12:43:46.000000Z",
				},
			},
			filters: []map[string]interface{}{
				{
					"field": "created_at",
					"type":  "daytime",
					"value": map[string]interface{}{
						"from": "2022-10-02",
						"to":   "2022-10-04",
					},
				},
			},
			wantData: []map[string]interface{}{
				{
					"name":       "azaza",
					"created_at": "2022-10-03T12:43:46.000000Z",
				},
				{
					"name":       "bzbzb",
					"created_at": "2022-10-04T12:43:46.000000Z",
				},
			},
		},

		{
			name: "list",
			initData: []map[string]interface{}{
				{
					"name":       "azaza",
					"created_at": "2022-10-03T12:43:46.000000Z",
				},
				{
					"name":       "bzbzb",
					"created_at": "2022-10-04T12:43:46.000000Z",
				},
				{
					"name":       "czczc",
					"created_at": "2022-10-05T12:43:46.000000Z",
				},
			},
			filters: []map[string]interface{}{
				{
					"field": "name",
					"type":  "list",
					"value": []interface{}{
						"azaza", "bzbzb",
					},
				},
			},
			wantData: []map[string]interface{}{
				{
					"name":       "azaza",
					"created_at": "2022-10-03T12:43:46.000000Z",
				},
				{
					"name":       "bzbzb",
					"created_at": "2022-10-04T12:43:46.000000Z",
				},
			},
		},
		{
			name: "mixed",
			initData: []map[string]interface{}{
				{
					"name":       "azaza",
					"created_at": "2022-10-03T12:43:46.000000Z",
				},
				{
					"name":       "bzbzb",
					"created_at": "2022-10-04T12:43:46.000000Z",
				},
				{
					"name":       "czczc",
					"created_at": "2022-10-05T12:43:46.000000Z",
				},
			},
			filters: []map[string]interface{}{
				{
					"field": "created_at",
					"type":  "daytime",
					"value": map[string]interface{}{
						"from": "2022-10-02",
						"to":   "2022-10-04",
					},
				},
				{
					"field": "name",
					"type":  "list",
					"value": []interface{}{
						"azaza", "bzbzb",
					},
				},
			},
			wantData: []map[string]interface{}{
				{
					"name":       "azaza",
					"created_at": "2022-10-03T12:43:46.000000Z",
				},
				{
					"name":       "bzbzb",
					"created_at": "2022-10-04T12:43:46.000000Z",
				},
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			got, err := doFilter(v.initData, v.filters)
			if v.wantError != nil {
				s.Equal(v.wantError, err)
			}
			s.Equal(v.wantData, got)
		})
	}

}

func (s *storeSouite) TestDoSearch() {
	tt := []struct {
		name         string
		initData     []map[string]interface{}
		searchString string
		wantData     []map[string]interface{}
		wantError    error
	}{
		{
			name: "name only",
			initData: []map[string]interface{}{
				{
					"name":        "azaza",
					"created_at":  "2022-10-03T12:43:46.000000Z",
					"description": "desc_azaza",
				},
				{
					"name":        "bzbzb",
					"created_at":  "2022-10-04T12:43:46.000000Z",
					"description": "desc_bzbzb",
				},
				{
					"name":        "czczc",
					"created_at":  "2022-10-05T12:43:46.000000Z",
					"description": "desc_czczc",
				},
			},
			searchString: "aza",
			wantData: []map[string]interface{}{
				{
					"name":        "azaza",
					"created_at":  "2022-10-03T12:43:46.000000Z",
					"description": "desc_azaza",
				},
			},
		},

		{
			name: "desc only",
			initData: []map[string]interface{}{
				{
					"name":        "azaza",
					"created_at":  "2022-10-03T12:43:46.000000Z",
					"description": "desc_azaza",
				},
				{
					"name":        "bzbzb",
					"created_at":  "2022-10-04T12:43:46.000000Z",
					"description": "desc_bzbzb",
				},
				{
					"name":        "czczc",
					"created_at":  "2022-10-05T12:43:46.000000Z",
					"description": "desc_czczc",
				},
			},
			searchString: "desc_a",
			wantData: []map[string]interface{}{
				{
					"name":        "azaza",
					"created_at":  "2022-10-03T12:43:46.000000Z",
					"description": "desc_azaza",
				},
			},
		},

		{
			name: "name and desc",
			initData: []map[string]interface{}{
				{
					"name":        "azaza",
					"created_at":  "2022-10-03T12:43:46.000000Z",
					"description": "desc_azaza",
				},
				{
					"name":        "bzbzb",
					"created_at":  "2022-10-04T12:43:46.000000Z",
					"description": "az_desc_bzbzb",
				},
				{
					"name":        "czczc",
					"created_at":  "2022-10-05T12:43:46.000000Z",
					"description": "desc_czczc",
				},
			},
			searchString: "az",
			wantData: []map[string]interface{}{
				{
					"name":        "azaza",
					"created_at":  "2022-10-03T12:43:46.000000Z",
					"description": "desc_azaza",
				},
				{
					"name":        "bzbzb",
					"created_at":  "2022-10-04T12:43:46.000000Z",
					"description": "az_desc_bzbzb",
				},
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			got, err := doSearch(v.initData, v.searchString)
			if v.wantError != nil {
				s.Equal(v.wantError, err)
			}
			s.Equal(v.wantData, got)
		})
	}
}
func (s *storeSouite) TestDoSort() {
	tt := []struct {
		name      string
		initData  []map[string]interface{}
		by        string
		order     string
		wantData  []map[string]interface{}
		wantError error
	}{
		{
			name: "asc order",
			initData: []map[string]interface{}{
				{
					"name":        "azaza",
					"created_at":  "2022-10-03T12:43:46.000000Z",
					"description": "desc_azaza",
				},
				{
					"name":        "bzbzb",
					"created_at":  "2022-10-05T12:43:46.000000Z",
					"description": "az_desc_bzbzb",
				},
				{
					"name":        "czczc",
					"created_at":  "2022-10-04T12:43:46.000000Z",
					"description": "desc_czczc",
				},
			},
			by:    "created_at",
			order: "asc",
			wantData: []map[string]interface{}{
				{
					"name":        "azaza",
					"created_at":  "2022-10-03T12:43:46.000000Z",
					"description": "desc_azaza",
				},
				{
					"name":        "czczc",
					"created_at":  "2022-10-04T12:43:46.000000Z",
					"description": "desc_czczc",
				},
				{
					"name":        "bzbzb",
					"created_at":  "2022-10-05T12:43:46.000000Z",
					"description": "az_desc_bzbzb",
				},
			},
		},

		{
			name: "desc order",
			initData: []map[string]interface{}{
				{
					"name":        "azaza",
					"created_at":  "2022-10-03T12:43:46.000000Z",
					"description": "desc_azaza",
				},
				{
					"name":        "bzbzb",
					"created_at":  "2022-10-05T12:43:46.000000Z",
					"description": "az_desc_bzbzb",
				},
				{
					"name":        "czczc",
					"created_at":  "2022-10-04T12:43:46.000000Z",
					"description": "desc_czczc",
				},
			},
			by:    "created_at",
			order: "desc",
			wantData: []map[string]interface{}{
				{
					"name":        "bzbzb",
					"created_at":  "2022-10-05T12:43:46.000000Z",
					"description": "az_desc_bzbzb",
				},
				{
					"name":        "czczc",
					"created_at":  "2022-10-04T12:43:46.000000Z",
					"description": "desc_czczc",
				},
				{
					"name":        "azaza",
					"created_at":  "2022-10-03T12:43:46.000000Z",
					"description": "desc_azaza",
				},
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			got, err := doSort(v.initData, v.by, v.order)
			if v.wantError != nil {
				s.Equal(v.wantError, err)
			}
			s.Equal(v.wantData, got)
		})
	}
}

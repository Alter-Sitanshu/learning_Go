package database

import (
	"net/http"
	"strconv"
	"strings"
)

type FilteringQuery struct {
	Search string   `json:"search" validate:"max=100"`
	Tags   []string `json:"tags"`
	Limit  int      `json:"limit" validate:"gte=1, lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
}

func (fq *FilteringQuery) Parse(r *http.Request) error {
	query := r.URL.Query()
	limit := query.Get("limit")
	offset := query.Get("offset")
	search := query.Get("search")
	tags := query.Get("tags")
	sort := query.Get("sort")
	var err error
	if limit != "" {
		fq.Limit, err = strconv.Atoi(limit)
		if err != nil {
			return err
		}
	}

	if offset != "" {
		fq.Offset, err = strconv.Atoi(offset)
		if err != nil {
			return err
		}
	}

	if search != "" {
		fq.Search = search
	}

	if sort != "" {
		fq.Sort = sort
	}

	if tags != "" {
		fq.Tags = strings.Split(tags, ",")
	}

	return nil

}

package mapper

import (
	"strconv"

	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func ToGroupResponse(g db.GroupsModuleGroup) response.GroupResponse {
	var defPrice *float64
	if g.DefaultPrice.Valid {
		f, _ := strconv.ParseFloat(g.DefaultPrice.String, 64)
		defPrice = &f
	}
	var note *string
	if g.Note.Valid {
		note = &g.Note.String
	}
	return response.GroupResponse{
		ID:           g.ID,
		Name:         g.Name,
		DefaultPrice: defPrice,
		Note:         note,
		CreatedAt:    g.CreatedAt,
	}
}

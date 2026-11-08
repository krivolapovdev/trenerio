package mapper

import (
	"strconv"

	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
)

func ToGroupModel(g db.GroupsModuleGroup) domain.Group {
	var defPrice *float64
	if g.DefaultPrice.Valid {
		f, _ := strconv.ParseFloat(g.DefaultPrice.String, 64)
		defPrice = &f
	}
	var note *string
	if g.Note.Valid {
		note = &g.Note.String
	}
	return domain.Group{
		ID:           g.ID,
		Name:         g.Name,
		DefaultPrice: defPrice,
		Note:         note,
		CreatedAt:    g.CreatedAt,
	}
}

func ToGroupResponse(g db.GroupsModuleGroup) domain.Group {
	return ToGroupModel(g)
}

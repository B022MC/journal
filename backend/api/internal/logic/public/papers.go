package public

import (
	"journal/api/internal/types"
	"journal/rpc/paper/paper"
)

func toPaperItems(items []*paper.PaperItem) []types.PaperItem {
	result := make([]types.PaperItem, 0, len(items))
	for _, item := range items {
		result = append(result, toPaperType(item))
	}
	return result
}

func toPaperType(item *paper.PaperItem) types.PaperItem {
	return types.PaperItem{
		Id:               item.Id,
		Title:            item.Title,
		TitleEn:          item.TitleEn,
		Abstract:         item.Abstract,
		AbstractEn:       item.AbstractEn,
		Content:          item.Content,
		AuthorId:         item.AuthorId,
		AuthorName:       item.AuthorName,
		Discipline:       item.Discipline,
		Zone:             item.Zone,
		ShitScore:        item.ShitScore,
		AvgRating:        item.AvgRating,
		RatingCount:      item.RatingCount,
		ViewCount:        item.ViewCount,
		ControversyIndex: item.ControversyIndex,
		Doi:              item.Doi,
		Keywords:         item.Keywords,
		FilePath:         item.FilePath,
		CreatedAt:        item.CreatedAt,
		PromotedAt:       item.PromotedAt,
	}
}

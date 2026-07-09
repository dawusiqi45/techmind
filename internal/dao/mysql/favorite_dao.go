package mysql

import (
	"techmind/internal/model"

	"gorm.io/gorm/clause"
)

// CreateFavorite 添加收藏，已存在则忽略（INSERT IGNORE 等价）
func CreateFavorite(f *model.Favorite) error {
	return DB.Clauses(clause.OnConflict{DoNothing: true}).Create(f).Error
}

// DeleteFavorite 取消收藏
func DeleteFavorite(userID, articleID int64) error {
	return DB.Where("user_id = ? AND article_id = ?", userID, articleID).
		Delete(&model.Favorite{}).Error
}

// ExistsFavorite 判断是否已收藏
func ExistsFavorite(userID, articleID int64) (bool, error) {
	var count int64
	err := DB.Model(&model.Favorite{}).
		Where("user_id = ? AND article_id = ?", userID, articleID).
		Count(&count).Error
	return count > 0, err
}

func ListFavoritesByUserID(userID int64, page, pageSize int) ([]*model.ArticleListItem, int, error) {
	offset := (page - 1) * pageSize

	var total int64
	if err := DB.Model(&model.Favorite{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.ArticleListItem
	err := DB.Raw(`
		SELECT `+articleListSelect+`
		FROM article a
		JOIN user u ON u.id = a.author_id
		JOIN favorite f ON f.article_id = a.id
		WHERE f.user_id = ? AND a.status = 1
		ORDER BY f.created_at DESC
		LIMIT ? OFFSET ?`, userID, pageSize, offset).Scan(&list).Error
	return list, int(total), err
}

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

package mysql

import (
	"techmind/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ToggleFavorite 在文章行锁保护下切换收藏关系，并重新计算冗余计数。
func ToggleFavorite(userID, articleID int64) (bool, error) {
	favorited := false
	err := DB.Transaction(func(tx *gorm.DB) error {
		var article model.Article
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = 1", articleID).First(&article).Error; err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&model.Favorite{}).
			Where("user_id = ? AND article_id = ?", userID, articleID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			if err := tx.Where("user_id = ? AND article_id = ?", userID, articleID).
				Delete(&model.Favorite{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(&model.Favorite{UserID: userID, ArticleID: articleID}).Error; err != nil {
				return err
			}
			favorited = true
		}

		if err := tx.Model(&model.Favorite{}).Where("article_id = ?", articleID).Count(&count).Error; err != nil {
			return err
		}
		return tx.Model(&model.Article{}).Where("id = ?", articleID).
			UpdateColumn("favorite_count", count).Error
	})
	return favorited, err
}

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

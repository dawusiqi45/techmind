package mysql

import (
	"techmind/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ToggleUserLike 在文章行锁保护下切换点赞关系，并重新计算冗余计数。
func ToggleUserLike(userID, articleID int64) (bool, error) {
	liked := false
	err := DB.Transaction(func(tx *gorm.DB) error {
		var article model.Article
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = 1", articleID).First(&article).Error; err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&model.UserLike{}).
			Where("user_id = ? AND article_id = ?", userID, articleID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			if err := tx.Where("user_id = ? AND article_id = ?", userID, articleID).
				Delete(&model.UserLike{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(&model.UserLike{UserID: userID, ArticleID: articleID}).Error; err != nil {
				return err
			}
			liked = true
		}

		if err := tx.Model(&model.UserLike{}).Where("article_id = ?", articleID).Count(&count).Error; err != nil {
			return err
		}
		return tx.Model(&model.Article{}).Where("id = ?", articleID).
			UpdateColumn("like_count", count).Error
	})
	return liked, err
}

func CreateUserLike(userID, articleID int64) error {
	return DB.Create(&model.UserLike{UserID: userID, ArticleID: articleID}).Error
}

func DeleteUserLike(userID, articleID int64) error {
	return DB.Where("user_id = ? AND article_id = ?", userID, articleID).Delete(&model.UserLike{}).Error
}

func ExistsUserLike(userID, articleID int64) (bool, error) {
	var count int64
	err := DB.Model(&model.UserLike{}).Where("user_id = ? AND article_id = ?", userID, articleID).Count(&count).Error
	return count > 0, err
}

func ListUserLikesByUserID(userID int64, page, pageSize int) ([]*model.ArticleListItem, int, error) {
	offset := (page - 1) * pageSize

	var total int64
	if err := DB.Model(&model.UserLike{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.ArticleListItem
	err := DB.Raw(`
		SELECT `+articleListSelect+`
		FROM article a
		JOIN user u ON u.id = a.author_id
		JOIN user_like ul ON ul.article_id = a.id
		WHERE ul.user_id = ? AND a.status = 1
		ORDER BY ul.created_at DESC
		LIMIT ? OFFSET ?`, userID, pageSize, offset).Scan(&list).Error
	return list, int(total), err
}

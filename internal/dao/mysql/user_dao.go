package mysql

import (
	"errors"
	"fmt"

	"techmind/internal/model"

	"gorm.io/gorm"
)

// CreateUser 插入新用户
func CreateUser(u *model.User) error {
	return DB.Create(u).Error
}

// GetUserByUsername 按用户名查询，未找到返回 nil
func GetUserByUsername(username string) (*model.User, error) {
	var u model.User
	err := DB.Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// GetUserByID 按 ID 查询，未找到返回 nil
func GetUserByID(id int64) (*model.User, error) {
	var u model.User
	err := DB.First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// ExistsUsername 判断用户名是否已存在
func ExistsUsername(username string) (bool, error) {
	var count int64
	err := DB.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// UpdateUserAvatar 更新头像
func UpdateUserAvatar(userID int64, avatar string) error {
	return DB.Model(&model.User{}).Where("id = ?", userID).Update("avatar", avatar).Error
}

// GetUserStats 获取用户统计（文章数、获赞总数、收藏总数）
func GetUserStats(userID int64) (articleCount, likeCount, favoriteCount int, err error) {
	var total int64
	if err = DB.Model(&model.Article{}).
		Where("author_id = ? AND status = 1", userID).
		Count(&total).Error; err != nil {
		return
	}
	articleCount = int(total)

	var stats struct {
		LikeTotal     int64
		FavoriteTotal int64
	}
	err = DB.Model(&model.Article{}).
		Select("COALESCE(SUM(like_count),0) AS like_total, COALESCE(SUM(favorite_count),0) AS favorite_total").
		Where("author_id = ? AND status = 1", userID).
		Scan(&stats).Error
	if err != nil {
		return 0, 0, 0, fmt.Errorf("user stats: %w", err)
	}
	likeCount = int(stats.LikeTotal)
	favoriteCount = int(stats.FavoriteTotal)
	return
}

func UpdateUserProfile(userID int64, username, email string) error {
	updates := map[string]interface{}{}
	if username != "" {
		updates["username"] = username
	}
	if email != "" {
		updates["email"] = email
	}
	if len(updates) == 0 {
		return nil
	}
	return DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

func UpdateUserEmail(userID int64, email string) error {
	return DB.Model(&model.User{}).Where("id = ?", userID).Update("email", email).Error
}

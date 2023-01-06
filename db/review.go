package db

import (
	common "reviewmakerback/common"

	"gorm.io/gorm"
)

func GetReview(rid string, selectText string) (Review, *gorm.DB) {
	var review Review

	tx := Db.Select(selectText).Where("review_id = ?", rid).Find(&review)
	return review, tx
}

func ExistsReview(rid string) bool {
	var cnt int64

	_, tx := GetReview(rid, "review_id")

	tx.Count(&cnt)
	return cnt == 1
}

func CreateReviewId(userId string, tierId string) (string, error) {
	var id string
	var err error
	for i := 0; i < retryCreateCnt; i++ {
		// ランダムな文字列を生成して、IDにする
		id, err = common.MakeRandomChars(idSize, userId+tierId)
		if err != nil {
			return "", err
		}
		if !ExistsReview(id) {
			return id, err
		}
	}
	return "", err
}

func CreateReview(
	userId string,
	tierId string,
	reviewId string,
	name string,
	title string,
	// 画像の保存パス、"nochange"なら画像を保存しない
	path string,
	reviewFactors string,
	sections string,
) error {
	var tier Review
	if path == "nochange" {
		tier = Review{
			ReviewId:      reviewId,
			UserId:        userId,
			TierId:        tierId,
			Title:         title,
			Name:          name,
			IconUrl:       "",
			ReviewFactors: reviewFactors,
			Sections:      sections,
		}
	} else {
		tier = Review{
			ReviewId:      reviewId,
			UserId:        userId,
			TierId:        tierId,
			Title:         title,
			Name:          name,
			IconUrl:       path,
			ReviewFactors: reviewFactors,
			Sections:      sections,
		}
	}
	tx := Db.Create(&tier)
	return tx.Error
}

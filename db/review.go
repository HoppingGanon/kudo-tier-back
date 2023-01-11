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

// 検索条件に従ってレビュー配列を取得
func GetReviews(userId string, tierId string, word string, sortType string, page int, pageSize int, includeSection bool) ([]Review, error) {
	/**
	"updatedAtDesc",
	"updatedAtAsc",
	"createdAtDesc",
	"createdAtAsc",
	*/

	tx := Db.Where("user_id = ?", userId)

	if !includeSection {
		// セクションを含めないでselectする
		tx = tx.Select(ExcludeSelect(Review{}, "sections"))
	}

	if word != "" {
		// 検索文字列指定有
		tx = tx.Where(SearchWord([]string{"name", "sections"}, word))
	}

	if tierId != "" {
		// TierId指定有
		tx = tx.Where("tier_id = ?", tierId)
	}

	if sortType == "updatedAtDesc" {
		tx = tx.Order("updated_at desc")
	} else if sortType == "updatedAtAsc" {
		tx = tx.Order("updated_at asc")
	} else if sortType == "createdAtDesc" {
		tx = tx.Order("created_at desc")
	} else if sortType == "createdAtAsc" {
		tx = tx.Order("created_at asc")
	}

	var reviews []Review
	tx.Offset(pageSize * (page - 1)).Limit(pageSize).Find(&reviews)

	return reviews, nil
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

func UpdateReview(
	review Review,
	name string,
	title string,
	// 画像の保存パス、"nochange"なら変更しない
	path string,
	reviewFactors string,
	sections string,
) error {
	var tx *gorm.DB
	review.Name = name
	review.Title = title
	review.ReviewFactors = reviewFactors
	review.Sections = sections
	if path != "nochange" {
		review.IconUrl = path
	}
	tx = Db.Save(&review)
	return tx.Error
}

func DeleteReview(reviewId string) error {
	tx := Db.Select("review_id").Where("review_id = ?", reviewId).Delete(&Review{})
	return tx.Error
}

func DeleteReviews(tierId string) error {
	tx := Db.Select("tier_id").Where("tier_id = ?", tierId).Delete(&Review{})
	return tx.Error
}

func GetReviewCountInUser(userId string) int64 {
	var cnt int64
	Db.Select("review_id").Where("user_id = ?", userId).Find(&Review{}).Count(&cnt)
	return cnt
}

func GetReviewCountInTier(tierId string) int64 {
	var cnt int64
	Db.Select("review_id").Where("tier_id = ?", tierId).Find(&Review{}).Count(&cnt)
	return cnt
}

package db

import (
	common "reviewmakerback/common"

	"gorm.io/gorm"
)

func GetTier(tid string, selectText string) (Tier, *gorm.DB) {
	var tier Tier

	tx := Db.Select(selectText).Where("tier_id = ?", tid).Find(&tier)
	return tier, tx
}

func ExistsTier(tid string) bool {
	var cnt int64

	_, tx := GetTier(tid, "tier_id")

	tx.Count(&cnt)
	return cnt == 1
}

func CreateTierId(userId string) (string, error) {
	var id string
	var err error
	for i := 0; i < retryCreateCnt; i++ {
		// ランダムな文字列を生成して、IDにする
		id, err = common.MakeRandomChars(idSize, userId)
		if err != nil {
			return "", err
		}
		if !ExistsTier(id) {
			return id, err
		}
	}
	return "", err
}

func CreateTier(
	userId string,
	tierId string,
	name string,
	// 画像の保存パス、"nochange"なら画像を保存しない
	path string,
	parags string,
	pointType string,
	reviewFactorParams string,
) error {
	var tier Tier
	if path == "nochange" {
		tier = Tier{
			TierId:       tierId,
			UserId:       userId,
			Name:         name,
			ImageUrl:     "",
			Parags:       parags,
			PointType:    pointType,
			FactorParams: reviewFactorParams,
		}
	} else {
		tier = Tier{
			TierId:       tierId,
			UserId:       userId,
			Name:         name,
			ImageUrl:     path,
			Parags:       parags,
			PointType:    pointType,
			FactorParams: reviewFactorParams,
		}
	}
	tx := Db.Create(&tier)
	return tx.Error
}

func UpdateTier(
	tier Tier,
	userId string,
	tierId string,
	name string,
	// 画像の保存パス、"nochange"なら変更しない
	imageUrl string,
	parags string,
	pointType string,
	factorParams string,
) error {
	var tx *gorm.DB
	tier.TierId = tierId
	tier.Name = name
	tier.Parags = parags
	tier.PointType = pointType
	tier.FactorParams = factorParams
	// tier.UpdatedAt = time.Now()
	if imageUrl != "nochange" {
		tier.ImageUrl = imageUrl
	}
	tx = Db.Save(&tier)
	return tx.Error
}

func GetTiers(userId string, word string, sortType string, page int, pageSize int) ([]Tier, error) {
	/**
	"updatedAtDesc",
	"updatedAtAsc",
	"createdAtDesc",
	"createdAtAsc",
	*/
	tx := Db

	if word == "" {
		// 検索文字列指定無
		tx = tx.Where("user_id = ?", userId)
	} else {
		// 検索文字列指定有
		tx = tx.Where("user_id = ?", userId).Where(SearchWord([]string{"name", "parags"}, word))
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

	var tiers []Tier
	tx.Offset(pageSize * (page - 1)).Limit(pageSize).Find(&tiers)

	return tiers, nil
}

func DeleteTier(tierId string) error {
	tx := Db.Select("tier_id").Where("tier_id = ?", tierId).Delete(&Tier{})
	return tx.Error
}

func GetTierCountInUser(userId string) int64 {
	var cnt int64
	Db.Select("tier_id").Where("user_id = ?", userId).Find(&Tier{}).Count(&cnt)
	return cnt
}

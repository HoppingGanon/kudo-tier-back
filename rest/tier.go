package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	common "reviewmakerback/common"
	db "reviewmakerback/db"
	"strconv"

	"github.com/labstack/echo"
	"gorm.io/gorm"
)

// 一度に取得可能なTier/レビュー数
const postPageSize = 10

// レビューの最大登録数
const ReviewMaxInTier = 255

type TierValidation struct {
	// tier名の最大文字数最大
	nameLenMax int
	// 評価項目の合計数の上限
	paramsLenMax int
	// 評価項目名の文字数の上限
	paramNameLenMax int
	// tierの画像サイズの最大(KB)
	imgMaxBytes float64
	// tierの画像サイズの一辺最大
	imgMaxEdge int
	// 画像のアスペクト比
	imgAspectRate float32
}

// Tierに関するバリデーション
var tierValidation = TierValidation{
	nameLenMax:      100,
	paramsLenMax:    16,
	paramNameLenMax: 16,
	imgMaxBytes:     5000,
	imgMaxEdge:      1080,
	imgAspectRate:   10.0 / 3.0,
}

// Tierのバリデーション
func validTier(tierData TierEditingData) (bool, *ErrorResponse) {
	// バリデーションチェック
	// Name
	f, e := validText("Tier名", "vtir-001", tierData.Name, true, -1, tierValidation.nameLenMax, "", "")
	if !f {
		return f, e
	}

	// Paragsのチェック
	if tierData.Parags == nil {
		return false, MakeError("vtir-002", "説明文等がNULLです")
	} else if len(tierData.Parags) > sectionValidation.paragsLenMax {
		return false, MakeError("vtir-003", fmt.Sprintf("説明文等の合計数が最大の%d個を超えています", sectionValidation.paragsLenMax))
	}

	for _, v := range tierData.Parags {
		// タイプのチェック
		if !IsParagraphType(v.Type) {
			return false, MakeError("vtir-004", "説明文/リンクのタイプが異常です")
		} else {
			if v.Type == "text" {
				// 説明文
				f, e := validText("説明文", "vtir-005", v.Body, false, -1, sectionValidation.paragTextLenMax, "", "")
				if !f {
					return f, e
				}
			} else if v.Type == "twitterLink" {
				// Twitterリンク
				f, e := validText("Twitterリンク", "vtir-006", v.Body, true, -1, sectionValidation.paragLinkLenMax, `^https:\/\/twitter\.com\/.*`, "正しい文字列")
				if !f {
					return f, e
				}
			}
		}
	}

	// PointTypeのチェック
	if !IsPointType(tierData.PointType) {
		return false, MakeError("vtir-007", "ポイント表示方法が異常です")
	}

	// 画像が既定のサイズ以下であることを確認する
	if tierData.ImageBase64 != "nochange" {
		if len(tierData.ImageBase64) > int(tierValidation.imgMaxBytes*1024*8/6) {
			return false, MakeError("vtir-008", "画像のサイズが大きすぎます")
		}
	}

	// ReviewFactorParamsのチェック
	if tierData.ReviewFactorParams == nil {
		return false, MakeError("vtir-009", "評価項目がNULLです")
	} else {
		f := false
		for _, v := range tierData.ReviewFactorParams {
			f = f || v.IsPoint
		}
		if !f {
			return false, MakeError("vtir-010", "ポイントの評価項目が少なくとも一つ以上必要です")
		}
	}

	for _, v := range tierData.ReviewFactorParams {
		// 評価項目名の文字数チェック
		f, e := validText("評価項目名", "vtir-011", v.Name, true, -1, tierValidation.paramsLenMax, "", "")
		if !f {
			return f, e
		}
	}

	for _, v := range tierData.ReviewFactorParams {
		// 評価項目の重み範囲チェック
		if v.IsPoint {
			f, e := validInteger("評価項目名", "vtir-012", v.Weight, 0, 100)
			if !f {
				return f, e
			}
		}
	}

	return true, nil
}

func removeParamIndex(params []ReviewParamData) []ReviewParam {
	list := make([]ReviewParam, len(params))
	for i, v := range params {
		list[i] = ReviewParam{
			Name:    v.Name,
			IsPoint: v.IsPoint,
			Weight:  v.Weight,
		}
	}
	return list
}

func postReqTier(c echo.Context) error {
	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	// 最小投稿頻度のチェック
	if db.CheckLastPost(session) {
		return c.JSON(400, commonError.tooFrequently)
	}

	requestIp := net.ParseIP(c.RealIP()).String()

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}
	var tierData TierEditingData
	err = json.Unmarshal(b, &tierData)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	f, e := validTier(tierData)
	if !f {
		return c.JSON(400, e)
	}

	params, err := json.Marshal(removeParamIndex(tierData.ReviewFactorParams))
	if err != nil {
		return c.JSON(400, MakeError("ptir-001", "重みの登録に失敗しました"))
	}

	var params2 []ReviewParam
	err = json.Unmarshal([]byte(params), &params2)
	if err != nil {
		return c.JSON(400, MakeError("ptir-002", "重みの登録に失敗しました"))
	}

	params3, err := json.Marshal(params2)
	if err != nil {
		return c.JSON(400, MakeError("ptir-003", "重みの登録に失敗しました"))
	}

	parags, err := json.Marshal(tierData.Parags)
	if err != nil {
		return c.JSON(400, MakeError("ptir-004", "説明文の登録に失敗しました"))
	}

	tierId, err := db.CreateTierId(session.UserId)
	if err != nil {
		return c.JSON(400, MakeError("ptir-005", "TierIDが生成出来ませんでした しばらく時間を開けて実行してください"))
	}

	// 画像データの名前を生成
	path := ""
	var er *ErrorResponse
	if tierData.ImageIsChanged {
		code, err := common.MakeRandomChars(16, tierId)
		if err != nil {
			return c.JSON(400, MakeError("ptir-006", "Tierの画像保存に失敗しました しばらく時間を開けて実行してください"))
		}
		fname := "image_" + code + ".jpg"

		// 画像の保存
		path, er = savePicture(session.UserId, "tier", tierId, fname, "", tierData.ImageBase64, "ptir-007", tierValidation.imgMaxEdge, tierValidation.imgAspectRate, 80)
		if er != nil {
			return c.JSON(400, er)
		}
	}

	err = db.CreateTier(session.UserId, tierId, tierData.Name, path, string(parags), tierData.PointType, string(params3))
	// 投稿時間を記録
	db.UpdateLastPostAt(session)
	if err != nil {
		db.WriteErrorLog(session.UserId, requestIp, "ptir-008", "Tierの作成に失敗しました", err.Error())
		return c.JSON(400, MakeError("ptir-008", "Tierの作成に失敗しました"))
	}

	db.WriteOperationLog(session.UserId, requestIp, "ptir", tierId)
	return c.String(201, tierId)
}

func updateReqTier(c echo.Context) error {
	tid := c.Param("tid")

	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	requestIp := net.ParseIP(c.RealIP()).String()

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}
	var tierData TierEditingData
	err = json.Unmarshal(b, &tierData)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	// Tierのチェック
	var cnt int64
	tier, tx := db.GetTier(tid, "*")
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(400, MakeError("utir-001", "該当するTierがありません"))
	}

	// 編集ユーザーとTier所有ユーザーチェック
	if session.UserId != tier.UserId {
		return c.JSON(403, commonError.userNotEqual)
	}

	f, e := validTier(tierData)
	if !f {
		return c.JSON(400, e)
	}

	// 新しく保存する対象の評価項目を定義する
	newParamsLen := len(tierData.ReviewFactorParams)
	newParams := make([]ReviewParam, newParamsLen)
	for i, param := range tierData.ReviewFactorParams {
		newParams[i] = ReviewParam{
			Name:    param.Name,
			IsPoint: param.IsPoint,
			Weight:  param.Weight,
		}
	}

	newParamsStr, err := json.Marshal(newParams)
	if err != nil {
		return c.JSON(400, MakeError("utir-004", "評価項目の登録に失敗しました"))
	}

	parags, err := json.Marshal(tierData.Parags)
	if err != nil {
		return c.JSON(400, MakeError("utir-005", "説明文の登録に失敗しました"))
	}

	// 画像データの名前を生成
	path := ""
	var er *ErrorResponse
	if tierData.ImageIsChanged {
		code, err := common.MakeRandomChars(16, tid)
		if err != nil {
			return c.JSON(400, MakeError("utir-006", "TierIDが生成出来ませんでした しばらく時間を開けて実行してください"))
		}
		fname := "icon_" + code + ".jpg"

		path, er = savePicture(session.UserId, "tier", tid, fname, "", tierData.ImageBase64, "utir-007", tierValidation.imgMaxEdge, tierValidation.imgAspectRate, 80)
		if er != nil {
			return c.JSON(400, er)
		}
	}

	var reviews []db.Review
	var oldFactors []ReviewFactorData
	var newFactors []ReviewFactorData
	var newFactorsBin []byte
	var oldIndex int

	err = db.Db.Transaction(func(tx *gorm.DB) error {
		// 旧データを取得
		tx1 := tx.Select("review_id, review_factors").Where("tier_id = ?", tier.TierId).Find(&reviews)

		if tx1.Error != nil {
			return tx1.Error
		}

		for _, review := range reviews {
			// 旧データをJSON化
			err = json.Unmarshal([]byte(review.ReviewFactors), &oldFactors)
			if err != nil {
				return err
			}
			// 新しい評価要素を入れる配列
			newFactors = make([]ReviewFactorData, newParamsLen)
			for i := range newFactors {
				// 受け取ったデータから、旧配列のときにあった場所を読み取る
				oldIndex = tierData.ReviewFactorParams[i].Index
				if oldIndex < 0 {
					// 負数であれば、新規追加されたものとして初期化する
					newFactors[i] = ReviewFactorData{
						Info:  "",
						Point: 0,
					}
				} else if oldIndex < len(oldFactors) {
					// 0以上であれば、旧配列の位置から新配列の位置に移動する
					newFactors[i] = oldFactors[oldIndex]
				}
				newFactorsBin, err = json.Marshal(newFactors)
				if err != nil {
					return err
				}
				tx1 = tx.Model(&review).Update("review_factors", string(newFactorsBin))

				if tx1.Error != nil {
					return tx1.Error
				}
			}
		}

		// トランザクション内でTierを更新する
		err = db.UpdateTierTx(tx, tier, session.UserId, tid, tierData.Name, path, tierData.ImageIsChanged, string(parags), tierData.PointType, string(newParamsStr))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// 新しく保存した方の画像削除
		er = daletePicture("utir-007", path)
		if er != nil {
			db.WriteErrorLog(session.UserId, requestIp, er.Code, er.Message, err.Error())
		}
		db.WriteErrorLog(session.UserId, requestIp, "utir-008", "Tierの更新に失敗しました", err.Error())
		return c.JSON(400, MakeError("utir-003", "Tierに紐づくレビューの評価要素の登録に失敗しました"))
	}

	// 古いほうの画像削除
	er = daletePicture("utir-007", tier.ImageUrl)
	if er != nil {
		return c.JSON(400, er)
	}

	db.WriteOperationLog(session.UserId, requestIp, "utir", tid)
	return c.String(200, tid)
}

func getReqTier(c echo.Context) error {
	tid := c.Param("tid")

	var cnt int64

	tier, tx := db.GetTier(tid, "*")
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(404, MakeError("gtir-002", "Tierが存在しません"))
	}

	user, tx := db.GetUser(tier.UserId, "*")
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(404, MakeError("gtir-001", "ユーザーが存在しません"))
	}

	tierData, er := makeTierData(tid, user, tier, "gtir-003")
	if er != nil {
		return c.JSON(400, er)
	}

	reviews, err := db.GetReviews(user.UserId, tid, "", "updatedAtDesc", 1, ReviewMaxInTier, false)
	if err != nil {
		return c.JSON(404, MakeError("gtir-003", "Tierに紐づくレビューが取得できませんでした"))
	}

	reviewDataList := make([]ReviewData, len(reviews))
	for i, review := range reviews {
		reviewData, err := makeReviewData(review.ReviewId, user, review, tier.PointType, "")
		if err != nil {
			return c.JSON(404, MakeError("gtir-004", "Tierに紐づくレビューが取得できませんでした"))
		}
		reviewDataList[i] = reviewData
	}

	tierData.Reviews = reviewDataList

	return c.JSON(200, tierData)
}

func makeTierData(tid string, user db.User, tier db.Tier, code string) (TierData, *ErrorResponse) {
	imageUrl2 := ""
	if tier.ImageUrl != "" {
		imageUrl2 = tier.ImageUrl
	}

	var err error
	var parags []ParagData
	if tier.Parags == "" {
		parags = []ParagData{}
	} else {
		err = json.Unmarshal([]byte(tier.Parags), &parags)
		if err != nil {
			return TierData{}, MakeError(code+"-01", "説明文の取得に失敗しました")
		}
	}

	var params []ReviewParamData
	err = json.Unmarshal([]byte(tier.FactorParams), &params)
	if err != nil {
		return TierData{}, MakeError(code+"-02", "評価項目の取得に失敗しました")
	}

	for i := range params {
		params[i].Index = i
	}

	return TierData{
		TierId:             tid,
		UserName:           user.Name,
		UserId:             user.UserId,
		UserIconUrl:        user.IconUrl,
		Name:               tier.Name,
		ImageUrl:           imageUrl2,
		Parags:             parags,
		Reviews:            []ReviewData{},
		PointType:          tier.PointType,
		ReviewFactorParams: params,
		CreatedAt:          common.DateToString(tier.CreatedAt),
		UpdatedAt:          common.DateToString(tier.UpdatedAt),
	}, nil
}

func getReqTiers(c echo.Context) error {
	userId := c.QueryParam("userid")
	word := c.QueryParam("word")
	sortType := c.QueryParam("sorttype")
	page, err := strconv.Atoi(c.QueryParam("page"))

	if err != nil {
		return c.JSON(400, MakeError("gtrs-001", "ページ指定が異常です"))
	} else if page < 0 {
		return c.JSON(400, MakeError("gtrs-002", "ページ指定が異常です"))
	}

	if !IsTierSortType(sortType) {
		return c.JSON(400, MakeError("gtrs-003", "ソートタイプが異常です"))
	}

	var cnt int64
	user, tx := db.GetUser(userId, "*")
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(404, MakeError("gtrs-004", "指定されたユーザーは存在しません"))
	}

	var er *ErrorResponse
	tiers, err := db.GetTiers(userId, word, sortType, page, postPageSize)
	if err != nil {
		return c.JSON(400, MakeError("gtrs-005", "Tierが取得できません"))
	}

	tierDataList := make([]TierData, len(tiers))
	for i, tier := range tiers {
		tierDataList[i], er = makeTierData(tier.TierId, user, tier, "gtrs-006")
		if er != nil {
			c.JSON(400, *er)
		}
	}
	return c.JSON(200, tierDataList)
}

func deleteReqTier(c echo.Context) error {
	tid := c.Param("tid")

	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	requestIp := net.ParseIP(c.RealIP()).String()

	tier, tx := db.GetTier(tid, "tier_id, user_id")

	// 編集ユーザーとTier所有ユーザーチェック
	if session.UserId != tier.UserId {
		return c.JSON(403, commonError.userNotEqual)
	}

	var cnt int64
	tx.Count(&cnt)

	if cnt != 1 {
		return c.JSON(400, MakeError("dtir-001", "対象のTierがありません"))
	}

	err = db.DeleteTier(tid)

	if err != nil {
		db.WriteErrorLog(session.UserId, requestIp, "dtir-002", "Tierの削除に失敗しました", err.Error())
		return c.JSON(400, MakeError("dtir-002", "Tierの削除に失敗しました"))
	}

	err = db.DeleteReviews(tid)

	if err != nil {
		db.WriteErrorLog(session.UserId, requestIp, "dtir-003", "Tierに紐づくレビューの削除に失敗しました", err.Error())
		return c.JSON(400, MakeError("dtir-003", "Tierに紐づくレビューの削除に失敗しました"))
	}

	db.WriteOperationLog(session.UserId, requestIp, "dtir", tid)
	return c.NoContent(200)
}

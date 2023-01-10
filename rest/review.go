package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	common "reviewmakerback/common"
	db "reviewmakerback/db"
	"strconv"

	"github.com/labstack/echo"
)

type ReviewValidation struct {
	// tier名の最大文字数
	nameLenMax int
	// Tierタイトルの最大文字数
	titleLenMax int
	// セクションの最大数
	sectionLenMax int
	// 評価情報の文字数の上限
	factorInfoLenMax int
	// レビューアイコンサイズの最大(KB)
	iconMaxBytes float64
	// レビューアイコンサイズの一辺最大
	iconMaxEdge int
	// 画像のアスペクト比
	iconAspectRate float32
}

// レビューに関するバリデーション
var reviewValidation = ReviewValidation{
	nameLenMax:       50,
	titleLenMax:      100,
	sectionLenMax:    8,
	factorInfoLenMax: 16,
	iconMaxBytes:     5000,
	iconMaxEdge:      256,
	iconAspectRate:   1.0,
}

// Tierのバリデーション
func validReview(reviewData ReviewEditingData, factorParams []ReviewParamData, pointType string) (bool, *ErrorResponse) {
	// バリデーションチェック
	// Name
	f, e := validText("レビュー名", "vrev-001", reviewData.Name, true, -1, reviewValidation.nameLenMax, "", "")
	if !f {
		return f, e
	}

	// Title
	f, e = validText("レビュータイトル", "vrev-002", reviewData.Title, false, -1, reviewValidation.titleLenMax, "", "")
	if !f {
		return f, e
	}

	// 評点のチェック
	if reviewData.ReviewFactors == nil {
		return false, MakeError("vtir-002", "レビュー評点や情報がNULLです")
	} else if len(reviewData.ReviewFactors) != len(factorParams) {
		return false, MakeError("vtir-003", "Tierの評価項目数とレビューの評点・情報の数が一致しません")
	}
	for i, factor := range reviewData.ReviewFactors {
		if factorParams[i].IsPoint {
			if pointType != "unlimited" {
				f, e = ValidFloat("評価情報", "vrev-004", float64(factor.Point), 0, 100)
				if !f {
					return false, e
				}
			}
		} else {
			f, e = validText("評価情報", "vrev-004", factor.Info, false, -1, reviewValidation.factorInfoLenMax, "", "")
			if !f {
				return false, e
			}
		}
	}

	// セクションのチェック
	if reviewData.Sections == nil {
		return false, MakeError("vtir-005", "説明文等がNULLです")
	}
	if len(reviewData.Sections) > reviewValidation.sectionLenMax {
		return false, MakeError("vtir-005", "説明文等がNULLです")
	}
	for _, sec := range reviewData.Sections {
		f, e = validText("見出し", "vrev-006", sec.Title, false, -1, sectionValidation.sectionTitleLen, "", "")
		if !f {
			return f, e
		}
		f, e = validParagraphs(sec.Parags)
		if !f {
			return f, e
		}
	}

	// 画像が既定のサイズ以下であることを確認する
	if reviewData.IconBase64 != "nochange" {
		if len(reviewData.IconBase64) > int(reviewValidation.iconMaxBytes*1024*8/6) {
			return false, MakeError("vrev-007", "画像のサイズが大きすぎます")
		}
	}

	return true, nil
}

func postReqReview(c echo.Context) error {
	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	requestIp := net.ParseIP(c.RealIP()).String()

	// Bodyの読み取り
	b, _ := ioutil.ReadAll(c.Request().Body)
	var reviewData ReviewEditingData
	err = json.Unmarshal(b, &reviewData)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	// Tier検索
	tier, tx := db.GetTier(reviewData.TierId, "tier_id, factor_params, user_id")
	if tx.Error != nil {
		return c.JSON(400, MakeError("prev-001", "レビューに対応するTierが存在しません"))
	}
	var cnt int64
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(400, MakeError("prev-002", "レビューに対応するTierが存在しません"))
	}

	if db.GetReviewCountInTier(tier.TierId) > ReviewMaxInTier {
		return c.JSON(400, MakeError("prev-003", fmt.Sprintf("登録できるレビューはTier一つにつき%d個までです", ReviewMaxInTier)))
	}

	// 編集ユーザーとTier所有ユーザーチェック
	if session.UserId != tier.UserId {
		return c.JSON(403, commonError.userNotEqual)
	}

	var params []ReviewParamData
	err = json.Unmarshal([]byte(tier.FactorParams), &params)
	if err != nil {
		return c.JSON(400, MakeError("prev-004", "Tierの情報取得に失敗しました"))
	}

	f, e := validReview(reviewData, params, tier.PointType)
	if !f {
		return c.JSON(400, e)
	}

	reviewId, err := db.CreateReviewId(session.UserId, tier.TierId)
	if err != nil {
		return c.JSON(400, MakeError("prev-005", "レビューIDが生成出来ませんでした しばらく時間を開けて実行してください"))
	}

	factors, err := json.Marshal(reviewData.ReviewFactors)
	if err != nil {
		return c.JSON(400, MakeError("prev-006", ""))
	}

	sections, err := json.Marshal(reviewData.Sections)
	if err != nil {
		return c.JSON(400, MakeError("prev-007", ""))
	}

	// 画像データの名前を生成
	code, err := common.MakeRandomChars(16, reviewId)
	if err != nil {
		return c.JSON(400, MakeError("prev-008", "レビューアイコンの保存に失敗しました しばらく時間を開けて実行してください"))
	}
	fname := "icon_" + code + ".jpg"

	// 画像の保存
	path, er := savePicture(session.UserId, "review", reviewId, fname, "", reviewData.IconBase64, "prev-009", reviewValidation.iconMaxEdge, reviewValidation.iconAspectRate, 92)
	if er != nil {
		return c.JSON(400, er)
	}

	err = db.CreateReview(session.UserId, reviewData.TierId, reviewId, reviewData.Name, reviewData.Title, path, string(factors), string(sections))
	if err != nil {
		db.WriteErrorLog(session.UserId, requestIp, "prev-010", "Tierの作成に失敗しました", err.Error())
		return c.JSON(400, MakeError("prev-010", "Tierの作成に失敗しました"))
	}

	db.WriteOperationLog(session.UserId, requestIp, "create review("+reviewId+")")
	return c.String(201, reviewId)
}

func updateReqReview(c echo.Context) error {
	rid := c.Param("rid")

	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	requestIp := net.ParseIP(c.RealIP()).String()

	// Bodyの読み取り
	b, _ := ioutil.ReadAll(c.Request().Body)
	var reviewData ReviewEditingData
	err = json.Unmarshal(b, &reviewData)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	// 元レビュー検索
	orgReview, tx := db.GetReview(rid, "review_id, user_id")
	if tx.Error != nil {
		return c.JSON(400, MakeError("prev-000", "対応するレビューが存在しません"))
	}
	var cnt int64
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(400, MakeError("prev-000", "レビューに対応するTierが存在しません"))
	}

	// Tier検索
	tier, tx := db.GetTier(rid, "tier_id, factor_params")
	if tx.Error != nil {
		return c.JSON(400, MakeError("prev-000", "レビューに対応するTierが存在しません"))
	}
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(400, MakeError("prev-000", "レビューに対応するTierが存在しません"))
	}

	// 編集ユーザーとTier・レビュー所有ユーザーチェック
	if session.UserId != orgReview.UserId || session.UserId != tier.UserId {
		return c.JSON(403, commonError.userNotEqual)
	}

	var params []ReviewParamData
	err = json.Unmarshal([]byte(tier.FactorParams), &params)
	if err != nil {
		return c.JSON(400, MakeError("prev-001", "Tierの情報取得に失敗しました"))
	}

	f, e := validReview(reviewData, params, tier.PointType)
	if !f {
		return c.JSON(400, e)
	}

	factors, err := json.Marshal(reviewData.ReviewFactors)
	if err != nil {
		return c.JSON(400, MakeError("prev-003", ""))
	}

	sections, err := json.Marshal(reviewData.Sections)
	if err != nil {
		return c.JSON(400, MakeError("prev-004", ""))
	}

	// 画像データの名前を生成
	code, err := common.MakeRandomChars(16, orgReview.ReviewId)
	if err != nil {
		return c.JSON(400, MakeError("prev-005", "レビューアイコンの保存に失敗しました しばらく時間を開けて実行してください"))
	}
	fname := "icon_" + code + ".jpg"

	// 画像の保存
	path, er := savePicture(session.UserId, "review", orgReview.ReviewId, fname, orgReview.IconUrl, reviewData.IconBase64, "prev-006", reviewValidation.iconMaxEdge, reviewValidation.iconAspectRate, 92)
	if er != nil {
		return c.JSON(400, er)
	}

	err = db.UpdateReview(orgReview, reviewData.Name, reviewData.Title, path, string(factors), string(sections))
	if err != nil {
		db.WriteErrorLog(session.UserId, requestIp, "prev-007", "Tierの作成に失敗しました", err.Error())
		return c.JSON(400, MakeError("prev-007", "Tierの作成に失敗しました"))
	}

	db.WriteOperationLog(session.UserId, requestIp, "update review("+orgReview.ReviewId+")")
	return c.String(201, orgReview.ReviewId)
}

func makeReviewData(rid string, user db.User, review db.Review, pointType string, code string) (ReviewData, *ErrorResponse) {
	imageUrl := ""
	if review.IconUrl != "" {
		imageUrl = os.Getenv("AP_BASE_URL") + "/" + review.IconUrl
	}

	var sections []SectionData
	err := json.Unmarshal([]byte(review.Sections), &sections)
	if err != nil {
		return ReviewData{}, MakeError(code+"-01", "説明文の取得に失敗しました")
	}

	var factors []ReviewFactorData
	err = json.Unmarshal([]byte(review.ReviewFactors), &factors)
	if err != nil {
		return ReviewData{}, MakeError(code+"-02", "評価点・情報の取得に失敗しました")
	}

	return ReviewData{
		ReviewId:      rid,
		UserName:      user.Name,
		UserId:        user.UserId,
		UserIconUrl:   user.IconUrl,
		TierId:        review.TierId,
		Title:         review.Title,
		Name:          review.Name,
		IconUrl:       imageUrl,
		ReviewFactors: factors,
		PointType:     pointType,
		Sections:      sections,
		CreatedAt:     common.DateToString(review.CreatedAt),
		UpdatedAt:     common.DateToString(review.UpdatedAt),
	}, nil
}

func getReqReview(c echo.Context) error {
	rid := c.Param("rid")

	var cnt int64

	review, tx := db.GetReview(rid, "*")
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(404, MakeError("grev-002", "レビューが存在しません"))
	}

	user, tx := db.GetUser(review.UserId, "*")
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(404, MakeError("grev-001", "ユーザーが存在しません"))
	}

	tier, tx := db.GetTier(review.TierId, "point_type, factor_params")
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(404, MakeError("grev-003", "レビューに紐づいたTier情報の取得に失敗しました"))
	}

	var params []ReviewParamData
	err := json.Unmarshal([]byte(tier.FactorParams), &params)
	if err != nil {
		return c.JSON(404, MakeError("grev-004", "評価項目の取得に失敗しました"))
	}

	reviewData, er := makeReviewData(rid, user, review, tier.PointType, "grev-005")
	if er != nil {
		return c.JSON(400, er)
	}
	return c.JSON(200, ReviewDataWithParams{
		Review: reviewData,
		Params: params,
	})
}

func getReqReviewPairs(c echo.Context) error {
	userId := c.QueryParam("userid")
	word := c.QueryParam("word")
	sortType := c.QueryParam("sorttype")
	page, err := strconv.Atoi(c.QueryParam("page"))

	if err != nil {
		return c.JSON(400, MakeError("grvs-001", "ページ指定が異常です"))
	} else if page < 0 {
		return c.JSON(400, MakeError("grvs-002", "ページ指定が異常です"))
	}

	if !IsTierSortType(sortType) {
		return c.JSON(400, MakeError("grvs-003", "ソートタイプが異常です"))
	}

	var cnt int64
	user, tx := db.GetUser(userId, "*")
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(404, MakeError("grvs-004", "指定されたユーザーは存在しません"))
	}

	var er *ErrorResponse
	// TierIdは指定せず、ユーザーに紐づくレビューを取得
	reviews, err := db.GetReviews(userId, "", word, sortType, page, postPageSize, true)
	if err != nil {
		return c.JSON(400, MakeError("grvs-005", "Tierが取得できません"))
	}

	var reviewData ReviewData
	var pointType string
	var params []ReviewParam
	var parsedParams []ReviewParamData
	reviewPairList := make([]ReviewDataWithParams, len(reviews))

	for i, review := range reviews {
		// Tier取得
		tier, _ := db.GetTier(review.TierId, "point_type, factor_params")
		if tier.PointType == "" {
			pointType = "stars"
		} else {
			pointType = tier.PointType
		}

		err = json.Unmarshal([]byte(tier.FactorParams), &params)
		if err != nil {
			return c.JSON(400, MakeError("grvs-006", "評価項目が取得できません"))
		}
		parsedParams = make([]ReviewParamData, len(params))
		for j, v := range params {
			parsedParams[j] = ReviewParamData{
				Name:    v.Name,
				IsPoint: v.IsPoint,
				Weight:  v.Weight,
				Index:   j,
			}
		}

		reviewData, er = makeReviewData(review.ReviewId, user, review, pointType, "grvs-007")

		if er != nil {
			c.JSON(400, *er)
		}

		// レビューデータの作成
		reviewPairList[i] = ReviewDataWithParams{
			Review: reviewData,
			Params: parsedParams,
		}
	}
	return c.JSON(200, reviewPairList)
}

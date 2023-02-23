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
		return false, MakeError("vtir-003", "レビュー評点や情報がNULLです")
	} else if len(reviewData.ReviewFactors) != len(factorParams) {
		return false, MakeError("vtir-004", "Tierの評価項目数とレビューの評点・情報の数が一致しません")
	}
	for i, factor := range reviewData.ReviewFactors {
		if factorParams[i].IsPoint {
			if pointType != "unlimited" {
				f, e = ValidFloat("評価情報", "vrev-005", float64(factor.Point), 0, 100)
				if !f {
					return false, e
				}
			}
		} else {
			f, e = validText("評価情報", "vrev-006", factor.Info, false, -1, reviewValidation.factorInfoLenMax, "", "")
			if !f {
				return false, e
			}
		}
	}

	// セクションのチェック
	if reviewData.Sections == nil {
		return false, MakeError("vtir-007", "説明文等がNULLです")
	}
	if len(reviewData.Sections) > reviewValidation.sectionLenMax {
		return false, MakeError("vtir-008", "説明文等がNULLです")
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
			return false, MakeError("vrev-009", "画像のサイズが大きすぎます")
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

	// 画像データを保存
	path := ""
	var er *ErrorResponse
	if reviewData.IconIsChanged {
		// 画像の保存
		path, er = savePicture(session.UserId, "review", reviewId, "icon_", "", reviewData.IconBase64, "prev-007", reviewValidation.iconMaxEdge, reviewValidation.iconAspectRate, 92)
		if er != nil {
			return c.JSON(400, er)
		}
	}

	// セクションを加工、Parag内の画像を保存
	madeSections, imageMap, er := createSections(reviewData.Sections, sections2ImageList([]SectionData{}), session.UserId, "review", reviewId, "image_")
	if er != nil {
		deleteSectionImg(madeSections)
		return c.JSON(400, er)
	}

	// セクションをJSONテキスト化
	sections, err := json.Marshal(madeSections)
	if err != nil {
		// 新しく作成した途中の画像ファイルを削除
		deleteSectionImg(madeSections)
		return c.JSON(400, MakeError("prev-008", ""))
	}

	// 使用しなくなったファイルを強制削除(POSTならば存在しない)
	deleteImageMap(imageMap)

	err = db.CreateReview(session.UserId, reviewData.TierId, reviewId, reviewData.Name, reviewData.Title, path, string(factors), string(sections))
	// 投稿時間を記録
	db.UpdateLastPostAt(session)
	if err != nil {
		// 新しく作成した途中の画像ファイルを削除
		deleteSectionImg(madeSections)
		db.WriteErrorLog(session.UserId, requestIp, "prev-009", "レビューの更新に失敗しました", err.Error())
		return c.JSON(400, MakeError("prev-009", "レビューの更新に失敗しました"))
	}

	db.WriteOperationLog(session.UserId, requestIp, "prev", reviewId)
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
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}
	var reviewData ReviewEditingData
	err = json.Unmarshal(b, &reviewData)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	// 元レビュー検索
	orgReview, tx := db.GetReview(rid, "*")
	if tx.Error != nil {
		return c.JSON(400, MakeError("urev-001", "レビューが存在しません"))
	}
	var cnt int64
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(400, MakeError("urev-002", "レビューが存在しません"))
	}

	// Tier検索
	tier, tx := db.GetTier(orgReview.TierId, "tier_id, user_id, factor_params")
	if tx.Error != nil {
		return c.JSON(400, MakeError("urev-003", "レビューに対応するTierが存在しません"))
	}
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(400, MakeError("urev-004", "レビューに対応するTierが存在しません"))
	}

	// 編集ユーザーとTier・レビュー所有ユーザーチェック
	if session.UserId != orgReview.UserId || session.UserId != tier.UserId {
		return c.JSON(403, commonError.userNotEqual)
	}

	var params []ReviewParamData
	err = json.Unmarshal([]byte(tier.FactorParams), &params)
	if err != nil {
		return c.JSON(400, MakeError("urev-005", "Tierの情報取得に失敗しました"))
	}

	f, e := validReview(reviewData, params, tier.PointType)
	if !f {
		return c.JSON(400, e)
	}

	factors, err := json.Marshal(reviewData.ReviewFactors)
	if err != nil {
		return c.JSON(400, MakeError("urev-006", "評価項目の登録に失敗しました"))
	}

	path := ""
	var er *ErrorResponse
	if reviewData.IconIsChanged {
		// 画像の保存
		path, er = savePicture(session.UserId, "review", orgReview.ReviewId, "icon_", orgReview.IconUrl, reviewData.IconBase64, "urev-007", reviewValidation.iconMaxEdge, reviewValidation.iconAspectRate, 92)
		if er != nil {
			return c.JSON(400, er)
		}
	} else {
		path = orgReview.IconUrl
	}

	var orgSections []SectionData
	err = json.Unmarshal([]byte(orgReview.Sections), &orgSections)
	if err != nil {
		return c.JSON(400, MakeError("urev-008", "説明文等の登録に失敗しました"))
	}

	// セクションを加工、Parag内の画像を保存
	madeSections, imageMap, er := createSections(reviewData.Sections, sections2ImageList(orgSections), session.UserId, "review", orgReview.ReviewId, "image_")
	if er != nil {
		deleteSectionImg(madeSections)
		return c.JSON(400, er)
	}

	// セクションをJSONテキスト化
	sections, err := json.Marshal(madeSections)
	if err != nil {
		// 新しく作成した途中の画像ファイルを削除
		deleteSectionImg(madeSections)
		return c.JSON(400, MakeError("urev-009", "セクションの変換に失敗しました"))
	}

	// 使用しなくなったファイルを強制削除(POSTならば存在しない)
	deleteImageMap(imageMap)

	err = db.UpdateReview(orgReview, reviewData.Name, reviewData.Title, path, reviewData.IconIsChanged, string(factors), string(sections))
	if err != nil {
		// 新しく作成した途中の画像ファイルを削除
		deleteSectionImg(madeSections)
		db.WriteErrorLog(session.UserId, requestIp, "urev-010", "Tierの作成に失敗しました", err.Error())
		return c.JSON(400, MakeError("urev-010", "Tierの作成に失敗しました"))
	}

	// 古いほうの画像削除
	if reviewData.IconIsChanged {
		daleteFile("", orgReview.IconUrl)
	}

	db.WriteOperationLog(session.UserId, requestIp, "urev", orgReview.ReviewId)
	return c.String(201, orgReview.ReviewId)
}

func makeReviewData(rid string, user db.User, review db.Review, pointType string, code string) (ReviewData, *ErrorResponse) {
	imageUrl := ""
	if review.IconUrl != "" {
		imageUrl = review.IconUrl
	}

	var err error
	var sections []SectionData
	if review.Sections == "" {
		sections = []SectionData{}
	} else {
		err = json.Unmarshal([]byte(review.Sections), &sections)
		if err != nil {
			return ReviewData{}, MakeError(code+"-001", "説明文の取得に失敗しました")
		}
	}

	var factors []ReviewFactorData
	err = json.Unmarshal([]byte(review.ReviewFactors), &factors)
	if err != nil {
		return ReviewData{}, MakeError(code+"-002", "評価点・情報の取得に失敗しました")
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
		return c.JSON(404, MakeError("grev-001", "レビューが存在しません"))
	}

	user, tx := db.GetUser(review.UserId, "*")
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(404, MakeError("grev-002", "ユーザーが存在しません"))
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
			Review:      reviewData,
			Params:      parsedParams,
			PullingDown: tier.PullingDown,
			PullingUp:   tier.PullingUp,
		}
	}
	return c.JSON(200, reviewPairList)
}

func deleteReviewReq(c echo.Context) error {
	rid := c.Param("rid")

	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	requestIp := net.ParseIP(c.RealIP()).String()

	var cnt int64
	review, tx := db.GetReview(rid, "user_id")
	tx.Count(&cnt)

	if cnt != 1 {
		return c.JSON(404, MakeError("drev-001", "レビューが存在しません"))
	}

	if review.UserId != session.UserId {
		return c.JSON(403, commonError.userNotEqual)
	}

	err = db.DeleteReview(rid)
	deleteFolder(review.UserId, "review", rid, "drev-003", requestIp)

	if err != nil {
		db.WriteErrorLog(session.UserId, requestIp, "drev-002", "レビューの削除に失敗しました", err.Error())
		return c.JSON(400, MakeError("drev-002", "レビューの削除に失敗しました"))
	}

	db.WriteOperationLog(session.UserId, requestIp, "drev", rid)
	return c.NoContent(200)
}

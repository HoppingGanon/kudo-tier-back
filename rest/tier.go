package rest

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	common "reviewmakerback/common"
	db "reviewmakerback/db"

	"github.com/labstack/echo"
)

// Tierのバリデーション
func validTier(tierData TierEditingData) (bool, *ErrorResponse) {
	// バリデーションチェック
	// Name
	f, e := validText("Tier名", "vtir-001", tierData.Name, true, -1, tierValidation.tierNameLenMax, "", "")
	if !f {
		return f, e
	}

	// Paragsのチェック
	if tierData.Parags == nil {
		return false, MakeError("vtir-002", "説明文等がNULLです")
	}

	for _, v := range tierData.Parags {
		// タイプのチェック
		if !IsParagraphType(v.Type) {
			return false, MakeError("vtir-003-00", "説明文/リンクのタイプが異常です")
		} else {
			if v.Type == "text" {
				// 説明文
				f, e := validText("説明文", "vtir-004", v.Body, true, -1, tierValidation.paragTextLenMax, "", "")
				if !f {
					return f, e
				}
			} else if v.Type == "twitterLink" {
				// Twitterリンク
				f, e := validText("Twitterリンク", "vtir-005", v.Body, true, -1, tierValidation.paragLinkLenMax, "", "")
				if !f {
					return f, e
				}
			}
		}
	}

	// PointTypeのチェック
	if !IsPointType(tierData.PointType) {
		return false, MakeError("vtir-006", "ポイント表示方法が異常です")
	}

	// 画像が既定のサイズ以下であることを確認する
	if tierData.ImageBase64 != "nochange" {
		if len(tierData.ImageBase64) > tierValidation.tierImgMaxBytes*1024*8/6 {
			return false, MakeError("vtir-007", "画像のサイズが大きすぎます")
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
		f, e := validText("評価項目名", "vtir-008", v.Name, true, -1, tierValidation.paramsLenMax, "", "")
		if !f {
			return f, e
		}
	}

	return true, nil
}

func postReqTier(c echo.Context) error {
	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	requestIp := net.ParseIP(c.RealIP()).String()

	// Bodyの読み取り
	b, _ := ioutil.ReadAll(c.Request().Body)
	var tierData TierEditingData
	err = json.Unmarshal(b, &tierData)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	f, e := validTier(tierData)
	if !f {
		return c.JSON(400, e)
	}

	params, err := json.Marshal(tierData.ReviewFactorParams)
	if err != nil {
		return c.JSON(400, MakeError("ptir-001", "重みの登録に失敗しました"))
	}

	parags, err := json.Marshal(tierData.Parags)
	if err != nil {
		return c.JSON(400, MakeError("ptir-002", "説明文の登録に失敗しました"))
	}

	tierId, err := db.CreateTierId(session.UserId)
	if err != nil {
		return c.JSON(400, MakeError("utir-003", "TierIDが生成出来ませんでした しばらく時間を開けて実行してください"))
	}

	// 画像データの名前を生成
	code, err := common.MakeRandomChars(16, tierId)
	if err != nil {
		return c.JSON(400, MakeError("utir-007", "TierIDが生成出来ませんでした しばらく時間を開けて実行してください"))
	}
	fname := "icon_" + code + ".jpg"

	path, er := savePicture(session.UserId, "tier", tierId, fname, "", tierData.ImageBase64, "ptir-005")
	if err != nil {
		return c.JSON(400, er)
	}

	err = db.CreateTier(session.UserId, tierId, tierData.Name, path, string(parags), tierData.PointType, string(params))
	if err != nil {
		db.WriteErrorLog(session.UserId, requestIp, "ptir-006", "Tierの作成に失敗しました", err.Error())
		return c.JSON(400, MakeError("ptir-006", "Tierの作成に失敗しました"))
	}
	return c.String(201, tierId)
}

func updateReqTier(c echo.Context) error {
	// セッションの存在チェック
	session, err := db.CheckSession(c)
	if err != nil {
		return c.JSON(403, commonError.noSession)
	}

	requestIp := net.ParseIP(c.RealIP()).String()

	// Bodyの読み取り
	b, _ := ioutil.ReadAll(c.Request().Body)
	var tierData TierEditingData
	err = json.Unmarshal(b, &tierData)
	if err != nil {
		return c.JSON(400, commonError.unreadableBody)
	}

	// Tierのチェック
	var cnt int64
	tier, tx := db.GetTier(tierData.TierId, session.UserId)
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(400, MakeError("utir-000", "該当するTierがありません"))
	}

	f, e := validTier(tierData)
	if !f {
		return c.JSON(400, e)
	}

	params, err := json.Marshal(tierData.ReviewFactorParams)
	if err != nil {
		return c.JSON(400, MakeError("utir-001", "重みの登録に失敗しました"))
	}

	parags, err := json.Marshal(tierData.Parags)
	if err != nil {
		return c.JSON(400, MakeError("utir-002", "説明文の登録に失敗しました"))
	}

	// 画像データの名前を生成
	code, err := common.MakeRandomChars(16, tierData.TierId)
	if err != nil {
		return c.JSON(400, MakeError("utir-004", "TierIDが生成出来ませんでした しばらく時間を開けて実行してください"))
	}
	fname := "icon_" + code + ".jpg"

	path, er := savePicture(session.UserId, "tier", tierData.TierId, fname, tier.ImageUrl, tierData.ImageBase64, "utir-005")
	if er != nil {
		return c.JSON(400, er)
	}

	err = db.UpdateTier(tier, session.UserId, tierData.TierId, tierData.Name, path, string(parags), tierData.PointType, string(params))
	if err != nil {
		db.WriteErrorLog(session.UserId, requestIp, "utir-003", "Tierの作成に失敗しました", err.Error())
		return c.JSON(400, MakeError("utir-003", "Tierの作成に失敗しました"))
	}

	return c.String(201, tierData.TierId)
}

func getReqTier(c echo.Context) error {
	uid := c.Param("uid")
	tid := c.Param("tid")

	var cnt int64

	user, tx := db.GetUser(uid)
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(404, MakeError("gtir-002", "ユーザーが存在しません"))
	}

	tier, tx := db.GetTier(tid, uid)
	tx.Count(&cnt)
	if cnt != 1 {
		return c.JSON(404, MakeError("gtir-001", "Tierが存在しません"))
	}

	var parags []ParagData
	err := json.Unmarshal([]byte(tier.Parags), &parags)
	if err != nil {
		return c.JSON(400, MakeError("gtir-003", "説明文の取得に失敗しました"))
	}

	var params []ReviewParamData
	err = json.Unmarshal([]byte(tier.FactorParams), &params)
	if err != nil {
		return c.JSON(400, MakeError("gtir-004", "評価項目の取得に失敗しました"))
	}

	imageUrl := ""

	if tier.ImageUrl != "" {
		imageUrl = os.Getenv("AP_BASE_URL") + "/" + tier.ImageUrl
	}

	res := TierData{
		TierId:             tid,
		UserName:           user.Name,
		UserId:             user.UserId,
		UserIconUrl:        user.IconUrl,
		Name:               tier.Name,
		ImageUrl:           imageUrl,
		Parags:             parags,
		Reviews:            []ReviewData{},
		PointType:          tier.PointType,
		ReviewFactorParams: params,
		CreateAt:           common.DateToString(tier.CreatedAt),
		UpdatedAt:          common.DateToString(tier.UpdatedAt),
	}

	return c.JSON(200, res)
}

package rest

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

// 1つの発信元IPあたりの最大保持一時セッション数
const maxSessionPerIp = 16

// codeVeriferの文字数
const codeVeriferCnt = 64

type TierValidation struct {
	// tier名の最大文字数最大
	nameLenMax int
	// 評価項目の合計数の上限
	paramsLenMax int
	// 評価項目名の文字数の上限
	paramNameLenMax int
	// tierの画像サイズの最大(KB)
	imgMaxBytes int
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

type SectionValidation struct {
	// セクションタイトルの最大文字数
	sectionTitleLen int
	// 説明文の文字数の上限
	paragTextLenMax int
	// セクション中に存在できるパラグラフ最大数
	paragsLenMax int
	// リンクの文字数の長さの上限
	paragLinkLenMax int
}

var sectionValidation = SectionValidation{
	sectionTitleLen: 100,
	paragTextLenMax: 16,
	paragsLenMax:    400,
	paragLinkLenMax: 100,
}

// 一度に取得可能なTier数
const tierPageSize = 10

// リンクの文字数の長さの上限
const paragLinkLenMax = 100

// アスペクト比の振れ幅
const aspectRateAmp = 0.1

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
	iconMaxBytes int
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
	iconMaxBytes:     1000,
	iconMaxEdge:      1080,
	iconAspectRate:   1.0,
}

type CommonError struct {
	noSession      ErrorResponse
	unreadableBody ErrorResponse
}

var commonError = CommonError{
	noSession: ErrorResponse{
		Code:    "gen0-a-001-00",
		Message: "セッションがありません",
	},
	unreadableBody: ErrorResponse{
		Code:    "gen0-a-002-00",
		Message: "データを読み取ることができませんでした",
	},
}

func MakeError(code string, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}

// 文字列に対するバリデーション
// minまたはmaxが0未満であればチェックしない
// regを指定すると正規表現によるチェックを行い、エラーだった場合はregMessageを使用してエラーメッセージを作成する
func validText(title string, code string, text string, required bool, min int, max int, reg string, regMessage string) (bool, *ErrorResponse) {
	if required && len(text) == 0 {
		// 必須入力
		return false, MakeError(code+"-00", fmt.Sprintf("%sは必須入力です", title))
	} else if min > 0 && utf8.RuneCountInString(text) < min {
		// 最低文字数
		return false, MakeError(code+"-01", fmt.Sprintf("%sは最低でも%d文字の入力が必要です", title, min))
	} else if max > 0 && utf8.RuneCountInString(text) > max {
		// 最大文字数
		return false, MakeError(code+"-02", fmt.Sprintf("%sは%d文字以下で入力する必要があります", title, max))
	} else if reg != "" && regexp.MustCompile(reg).MatchString(text) {
		// 正規表現
		return false, MakeError(code+"-03", fmt.Sprintf("%sは%sで入力する必要があります", title, regMessage))
	}
	return true, nil
}

// 整数に対するバリデーション
func validInteger(title string, code string, val int, min int, max int) (bool, *ErrorResponse) {
	if val < min {
		return false, MakeError(code+"-00", fmt.Sprintf("%sは%d以上の整数を入力してください", title, min))
	} else if val < max {
		return false, MakeError(code+"-01", fmt.Sprintf("%sは%d以下の整数を入力してください", title, max))
	}
	return true, nil
}

func contains(s string, a []string) bool {
	for _, v := range a {
		if s == v {
			return true
		}
	}
	return false
}

func IsPointType(v string) bool {
	return contains(v, []string{
		"stars",
		"rank7",
		"rank14",
		"score",
		"point",
		"unlimited",
	})
}

func validParagraphs(parags []ParagData) (bool, *ErrorResponse) {
	for _, parag := range parags {
		if !IsParagraphType(parag.Type) {
			return false, MakeError("vpgs-01", "説明文/リンクのタイプが異常です")
		} else {
			if parag.Type == "text" {
				f, e := validText("説明文", "vpgs-02", parag.Body, false, -1, sectionValidation.paragTextLenMax, "", "")
				if !f {
					return false, e
				}
			} else if parag.Type == "twitterLink" {
				f, e := validText("Twitterリンク", "vpgs-03", parag.Body, false, -1, paragLinkLenMax, "", "")
				if !f {
					return false, e
				}
			} else if parag.Type == "imageLink" {

			}
		}
	}
	return true, nil
}

func IsParagraphType(v string) bool {
	return contains(v, []string{
		"text",
		"twitterLink",
		"imageLink",
	})
}

func IsTierSortType(v string) bool {
	return contains(v, []string{
		"updatedAtDesc",
		"updatedAtAsc",
		"createdAtDesc",
		"createdAtAsc",
	})
}

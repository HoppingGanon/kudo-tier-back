package rest

import (
	"fmt"
	"regexp"
	"reviewmakerback/common"
	"reviewmakerback/db"
	"unicode/utf8"
)

// 1つの発信元IPあたりの最大保持一時セッション数
const maxSessionPerIp = 16

// codeVeriferの文字数
const codeVeriferCnt = 64

type UserValidation struct {
	// ユーザー表示名の最大文字数
	nameLenMax int
	// プロフィールの最大文字数
	profileLenMax int
}

var userValidation = UserValidation{
	nameLenMax:    50,
	profileLenMax: 400,
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
	paragTextLenMax: 400,
	paragsLenMax:    16,
	paragLinkLenMax: 100,
}

// リンクの文字数の長さの上限
const paragLinkLenMax = 100

// アスペクト比の振れ幅
const aspectRateAmp = 0.1

type CommonError struct {
	noSession      ErrorResponse
	unreadableBody ErrorResponse
	userNotEqual   ErrorResponse
	tooFrequently  ErrorResponse
}

var commonError = CommonError{
	noSession: ErrorResponse{
		Code:    "gen0-001-00",
		Message: "セッションがありません",
	},
	unreadableBody: ErrorResponse{
		Code:    "gen0-002-00",
		Message: "データを読み取ることができませんでした",
	},
	userNotEqual: ErrorResponse{
		Code:    "gen0-003-00",
		Message: "このユーザーの編集権限はありません",
	},
	tooFrequently: ErrorResponse{
		Code:    "gen0-004-00",
		Message: fmt.Sprintf("投稿は%d秒以上あけて実行してください", db.PostSpanMin),
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
	} else if reg != "" && !regexp.MustCompile(reg).MatchString(text) {
		// 正規表現
		return false, MakeError(code+"-03", fmt.Sprintf("%sは%sで入力する必要があります", title, regMessage))
	}
	return true, nil
}

// 整数に対するバリデーション
func validInteger(title string, code string, val int, min int, max int) (bool, *ErrorResponse) {
	if val < min {
		return false, MakeError(code+"-00", fmt.Sprintf("%sは%d以上の整数を入力してください", title, min))
	} else if val > max {
		return false, MakeError(code+"-01", fmt.Sprintf("%sは%d以下の整数を入力してください", title, max))
	}
	return true, nil
}

// 浮動小数に対するバリデーション
func ValidFloat(title string, code string, val float64, min float64, max float64) (bool, *ErrorResponse) {
	if val < min {
		return false, MakeError(code+"-00", fmt.Sprintf("%sは%f以上の数値を入力してください", title, min))
	} else if val > max {
		return false, MakeError(code+"-01", fmt.Sprintf("%sは%f以下の数値を入力してください", title, max))
	}
	return true, nil
}

func IsPointType(v string) bool {
	return common.Contains(v, []string{
		"stars",
		"rank7",
		"rank14",
		"score",
		"point",
		"unlimited",
	})
}

func validParagraphs(parags []ParagData) (bool, *ErrorResponse) {
	if len(parags) > sectionValidation.paragsLenMax {
		return false, MakeError("vpgs-002", fmt.Sprintf("説明文/リンクは合計%d個以下にする必要があります", sectionValidation.paragsLenMax))
	}

	for _, parag := range parags {
		if !IsParagraphType(parag.Type) {
			return false, MakeError("vpgs-002", "説明文/リンクのタイプが異常です")
		} else {
			if parag.Type == "text" {
				f, e := validText("説明文", "vpgs-003", parag.Body, false, -1, sectionValidation.paragTextLenMax, "", "")
				if !f {
					return false, e
				}
			} else if parag.Type == "twitterLink" {
				f, e := validText("Twitterリンク", "vpgs-004", parag.Body, false, -1, paragLinkLenMax, "", "")
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
	return common.Contains(v, []string{
		"text",
		"twitterLink",
		"imageLink",
	})
}

func IsTierSortType(v string) bool {
	return common.Contains(v, []string{
		"updatedAtDesc",
		"updatedAtAsc",
		"createdAtDesc",
		"createdAtAsc",
	})
}

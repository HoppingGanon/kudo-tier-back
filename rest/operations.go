package rest

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	common "reviewmakerback/common"
	"reviewmakerback/db"

	"github.com/labstack/echo"
	"github.com/nfnt/resize"
)

const saveRetryCount = 3

// テスト用の関数
func getReqHello(c echo.Context) error {
	return c.String(http.StatusOK, "{\"Hello\": \"World\"}")
}

func getUserFile(c echo.Context) error {
	userId := c.Param("uid")
	data := c.Param("method")
	id := c.Param("id")
	fname := c.Param("fname")

	// 不正なファイル名へのアクセスを防ぐ
	if !common.TestRegexp(`^[a-zA-Z0-9]*$`, userId) {
		// ユーザーID
		return c.JSON(http.StatusBadRequest, MakeError("gen0-000", "不正なディレクトリが指定されました"))
	}
	if !common.Contains(data, []string{"tier", "review", "user"}) {
		// データ(機能種別)
		return c.JSON(http.StatusBadRequest, MakeError("gen0-001", "不正なディレクトリが指定されました"))
	}
	if !common.TestRegexp(`^[a-zA-Z0-9]*$`, id) {
		// ID
		return c.JSON(http.StatusBadRequest, MakeError("gen0-002", "不正なディレクトリが指定されました"))
	}
	if !common.TestRegexp(`^[a-zA-Z0-9/._]*$`, fname) {
		// ファイル名
		return c.JSON(http.StatusBadRequest, MakeError("gen0-003", "不正なファイルが指定されました"))
	}

	path := os.Getenv("AP_FILE_PATH") + "/" + userId + "/" + data + "/" + id + "/" + fname
	// アクセスされたファイルを返す
	return c.File(path)
}

func daleteFile(errorCode string, delpath string) *ErrorResponse {
	// ファイル削除
	if delpath != "" {
		_, err := os.Stat(delpath)
		if err == nil {
			// ファイルが存在した場合
			err = os.Remove(delpath)
			if err != nil {
				// エラーコードはsavePicと重複
				return MakeError(errorCode+"-10", "画像の削除に失敗しました")
			}
		}
	}
	return nil
}

func deleteFolder(userId string, data string, id string, errorCode string, ipAddress string) {
	err := os.RemoveAll((fmt.Sprintf("%s/%s/%s/%s", os.Getenv("AP_FILE_PATH"), userId, data, id)))
	if os.IsNotExist(err) {
		db.WriteErrorLog(userId, ipAddress, errorCode, "フォルダが削除できませんでした", fmt.Sprintf("'%s/%s/%s/%s' ", os.Getenv("AP_FILE_PATH"), userId, data, id)+err.Error())
	}
}

// 画像を上書き保存する
// delpath 省略可能
func savePicture(userId string, data string, id string, fname string, delpath string, imageBase64 string, errorCode string, imgMaxEdge int, aspectRate float32, quality int) (string, *ErrorResponse) {
	path := ""
	// Base64文字列をバイト列に変換する
	if imageBase64 == "" {
		// ファイル削除
		er := daleteFile(errorCode, delpath)
		if er != nil {
			return path, er
		}
	} else {
		byteAry, err := b64.StdEncoding.DecodeString(imageBase64)
		if err != nil {
			return path, MakeError(errorCode+"-01", "画像の登録に失敗しました")
		}

		// バイト列をReaderに変換
		r := bytes.NewReader(byteAry)
		img, _, err := image.Decode(r)
		if err != nil {
			return path, MakeError(errorCode+"-02", "画像の登録に失敗しました")
		}

		x := img.Bounds().Dx()
		y := img.Bounds().Dy()

		// (画像のアスペクト比 / 既定のアスペクト比) がプラスマイナスaspectRateAmpになってるか確認
		if ((float32(x)/float32(y))/aspectRate)-(1.0-aspectRateAmp) > aspectRateAmp*2 {
			return path, MakeError(errorCode+"-03", "画像のアスペクト比が異常です")
		}

		resizedImg := resize.Thumbnail(uint(imgMaxEdge), uint(imgMaxEdge), img, resize.NearestNeighbor)
		err = os.MkdirAll(fmt.Sprintf("%s/%s/%s/%s", os.Getenv("AP_FILE_PATH"), userId, data, id), os.ModePerm)
		if err != nil {
			return path, MakeError(errorCode+"-04", "画像の登録に失敗しました")
		}

	lo:
		for i := 0; i < saveRetryCount; i++ {
			code, err := common.MakeRandomChars(16, fmt.Sprintf("%s%s_%d", userId, id, i))
			if err != nil {
				return "", MakeError(errorCode+"-05", "画像の登録に失敗しました しばらく時間を空けてもう一度実行してください")
			}
			path = fmt.Sprintf("%s/%s/%s/%s/%s%s.jpg", os.Getenv("AP_FILE_PATH"), userId, data, id, fname, code)

			_, err = os.Stat(path)
			if os.IsNotExist(err) {
				break lo
			} else if i == saveRetryCount-1 {
				// リトライ上限に到達
				return "", MakeError(errorCode+"-06", "画像の登録に失敗しました しばらく時間を空けてもう一度実行してください")
			}
		}

		out, err := os.Create(path)
		if err != nil {
			if out != nil {
				out.Close()
			}
			return "", MakeError(errorCode+"-07", "画像の登録に失敗しました")
		}

		opts := &jpeg.Options{
			Quality: quality,
		}

		// ファイル削除
		er := daleteFile(errorCode, delpath)
		if er != nil {
			out.Close()
			return path, er
		}

		err = jpeg.Encode(out, resizedImg, opts)
		out.Close()

		if err != nil {
			return path, MakeError(errorCode+"-08", "画像の登録に失敗しました")
		}
	}
	return path, nil
}

// セクション配列から画像のパスをマップ化したものを取得
func sections2ImageList(sections []SectionData) map[string]bool {
	m := make(map[string]bool)
	for _, section := range sections {
		for _, parag := range section.Parags {
			if parag.Type == "imageLink" && parag.Body != "" {
				m[parag.Body] = false
			}
		}
	}
	return m
}

// parag配列から画像のパスをマップ化したものを取得
func parags2DelImageMap(oldParags []ParagData) map[string]bool {
	m := make(map[string]bool)
	for _, parag := range oldParags {
		if parag.Type == "imageLink" && parag.Body != "" {
			m[parag.Body] = false
		}
	}
	return m
}

// 編集データをセクション配列に変換
// oldImageMapはもともと存在していたparagsのなかに存在するファイルのパスのマップで、対応する値は全てfalseにしておく
// 返すmapは、もともと存在していたparagsのなかに存在するかつ削除せずに残しておくファイル
func createParags(parags []ParagEditingData, oldImageMap map[string]bool, userId string, data string, id string, fname string) ([]ParagData, map[string]bool, *ErrorResponse) {
	madeParags := make([]ParagData, len(parags))
	var exists bool
	for i, parag := range parags {
		if parag.Type == "imageLink" {
			// 画像ファイルの場合
			if !parag.IsChanged {
				// クライアント側で変更がない
				_, exists = oldImageMap[parag.Body]
				if exists {
					// クライアント側から送られてきたリンクが、もともと存在していたparagsのなかにも存在する
					// 残すためのフラグを立てておく
					oldImageMap[parag.Body] = true
					madeParags[i].Type = "imageLink"
					madeParags[i].Body = parag.Body
				} else {
					// 存在しない場合は異常なケース
					return madeParags, oldImageMap, MakeError("cpgs-02", "説明画像に存在しないファイルが指定されました")
				}
			} else {
				// クライアント側で変更あり
				path, er := savePicture(userId, data, id, fname, "", parag.Body, "cpgs-01", sectionValidation.paragImgMax, sectionValidation.paragImgAspect, 80)
				if er != nil {
					return madeParags, oldImageMap, er
				}
				madeParags[i].Type = "imageLink"
				madeParags[i].Body = path
			}
		} else {
			// 画像ファイル以外
			madeParags[i].Type = parag.Type
			madeParags[i].Body = parag.Body
		}
	}

	return madeParags, oldImageMap, nil
}

// 編集データをセクション配列に変換
// oldImageMapはもともと存在していたparagsのなかに存在するファイルのパスのマップで、対応する値は全てfalseにしておく
// 返すmapは、もともと存在していたparagsのなかに存在するかつ削除せずに残しておくファイル
func createSections(sections []SectionEditingData, oldImageMap map[string]bool, userId string, data string, id string, fname string) ([]SectionData, map[string]bool, *ErrorResponse) {
	madeSections := make([]SectionData, len(sections))
	var parags []ParagData
	var er *ErrorResponse = nil
	for i, section := range sections {
		parags, oldImageMap, er = createParags(section.Parags, oldImageMap, userId, data, id, fname)
		if er != nil {
			return madeSections, oldImageMap, er
		}
		madeSections[i] = SectionData{
			Title:  section.Title,
			Parags: parags,
		}
	}
	return madeSections, oldImageMap, nil
}

func deleteParagsImg(parags []ParagData) {
	for _, parag := range parags {
		daleteFile("", parag.Body)
	}
}

func deleteSectionImg(sections []SectionData) {
	for _, section := range sections {
		deleteParagsImg(section.Parags)
	}
}

func deleteImageMap(oldImageMap map[string]bool) {
	for path, f := range oldImageMap {
		if !f {
			daleteFile("", path)
		}
	}
}

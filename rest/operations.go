package rest

import (
	"bytes"
	b64 "encoding/base64"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	common "reviewmakerback/common"

	"github.com/labstack/echo"
	"github.com/nfnt/resize"
)

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
		return c.JSON(http.StatusBadRequest, MakeError("gen0-00", "不正なディレクトリが指定されました"))
	}
	if !contains(data, []string{"tier", "review", "user"}) {
		// データ(機能種別)
		return c.JSON(http.StatusBadRequest, MakeError("gen0-01", "不正なディレクトリが指定されました"))
	}
	if !common.TestRegexp(`^[a-zA-Z0-9]*$`, id) {
		// ID
		return c.JSON(http.StatusBadRequest, MakeError("gen0-02", "不正なディレクトリが指定されました"))
	}
	if !common.TestRegexp(`^[a-zA-Z0-9/._]*$`, fname) {
		// ファイル名
		return c.JSON(http.StatusBadRequest, MakeError("gen0-03", "不正なファイルが指定されました"))
	}

	path := os.Getenv("AP_FILE_PATH") + "/" + userId + "/" + data + "/" + id + "/" + fname
	// アクセスされたファイルを返す
	return c.File(path)
}

func savePicture(userId string, data string, id string, fname string, delpath string, imageBase64 string, errorCode string) (string, *ErrorResponse) {
	path := ""
	// Base64文字列をバイト列に変換する
	if imageBase64 == "nochange" {
		// 更新しない
		return "nochange", nil
	} else if imageBase64 == "" {
		// ファイル削除
		_, err := os.Stat(delpath)
		if err == nil {
			// ファイルが存在した場合
			err = os.Remove(delpath)
			if err != nil {
				return path, MakeError(errorCode+"-05", "画像の登録に失敗しました")
			}
		}
	} else {
		byteAry, err := b64.StdEncoding.DecodeString(imageBase64)
		if err != nil {
			return path, MakeError(errorCode+"-00", "画像の登録に失敗しました")
		}

		// バイト列をReaderに変換
		r := bytes.NewReader(byteAry)
		img, _, err := image.Decode(r)
		if err != nil {
			return path, MakeError(errorCode+"-01", "画像の登録に失敗しました")
		}

		resizedImg := resize.Thumbnail(uint(tierImgMaxEdge), uint(tierImgMaxEdge), img, resize.NearestNeighbor)
		err = os.MkdirAll(os.Getenv("AP_FILE_PATH")+"/"+userId+"/"+data+"/"+id, os.ModePerm)
		if err != nil {
			return path, MakeError(errorCode+"-02", "画像の登録に失敗しました")
		}

		path = os.Getenv("AP_FILE_PATH") + "/" + userId + "/" + data + "/" + id + "/" + fname

		out, err := os.Create(path)
		if err != nil {
			out.Close()
			return path, MakeError(errorCode+"-03", "画像の登録に失敗しました")
		}

		opts := &jpeg.Options{
			Quality: 92,
		}

		// ファイル削除
		_, err = os.Stat(delpath)
		if err == nil {
			// ファイルが存在した場合
			err = os.Remove(delpath)
			if err != nil {
				out.Close()
				return path, MakeError(errorCode+"-06", "画像の登録に失敗しました")
			}
		}

		err = jpeg.Encode(out, resizedImg, opts)
		out.Close()

		if err != nil {
			return path, MakeError(errorCode+"-04", "画像の登録に失敗しました")
		}
	}
	return path, nil
}

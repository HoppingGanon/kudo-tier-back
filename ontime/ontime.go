package ontime

import (
	"context"
	db "reviewmakerback/db"
	"time"
)

func Start() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go ArrangeSession(ctx)
	return ctx, cancel
}

func ArrangeSession(ctx context.Context) {
	// タイマーを設定する
	ticker := time.NewTicker(db.SessionDelSpan * time.Second)

	// 処理終了時、タイマーを終了する
	defer ticker.Stop()

	// 最初の一回を実行
	db.ArrangeSession()

	for {
		select {
		case <-ctx.Done():
			// キャンセルされた場合
			return
		case <-ticker.C:
			// タイマーが周回した際
			db.ArrangeSession()
		}
	}
}

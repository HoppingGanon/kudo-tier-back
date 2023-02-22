package rest

import (
	"os"

	"github.com/labstack/echo"
)

func Route(e *echo.Echo) {
	e.GET("/hello", getReqHello)
	e.GET("/auth/tempsession/:service/:version", getReqTempSession)
	e.POST("/auth/session/:service/:version", postReqSession)
	e.DELETE("/auth/session", delReqSession)
	e.GET("/auth/check-session", getReqCheckSession)
	e.POST("/user", postReqUser)
	e.DELETE("/user1/:uid", deleteUser1)
	e.DELETE("/user2/:uid", deleteUser2)
	e.GET("/user/:uid", getReqUserData)
	e.PATCH("/user/:uid", updateReqUser)
	e.POST("/tier", postReqTier)
	e.GET("/tier/:tid", getReqTier)
	e.GET("/"+os.Getenv("BACK_AP_FILE_PATH")+"/:uid/:method/:id/:fname", getUserFile)
	e.PATCH("/tier/:tid", updateReqTier)
	e.DELETE("/tier/:rid", deleteReqTier)
	e.GET("/tiers", getReqTiers)
	e.POST("/review", postReqReview)
	e.GET("/review/:rid", getReqReview)
	e.PATCH("/review/:rid", updateReqReview)
	e.DELETE("/review/:rid", deleteReviewReq)
	e.GET("/review-pairs", getReqReviewPairs)
	e.GET("/latest-post-lists/:uid", getReqLatestPostLists)
}

package rest

import (
	"os"

	"github.com/labstack/echo"
)

func Route(e *echo.Echo) {
	e.GET("/hello", getReqHello)
	e.GET("/auth/tempsession", getReqTempSession)
	e.GET("/auth/session", getReqSession)
	e.DELETE("/auth/session", delReqSession)
	e.POST("/user", postReqUser)
	e.GET("/user/:uid", getReqUserData)
	e.POST("/tier", postReqTier)
	e.GET("/tier/:tid", getReqTier)
	e.GET("/"+os.Getenv("AP_FILE_PATH")+"/:uid/:method/:id/:fname", getUserFile)
	e.PATCH("/tier/:tid", updateReqTier)
	e.DELETE("/tier/:rid", deleteReqTier)
	e.GET("/tiers", getReqTiers)
	e.POST("/review", postReqReview)
	e.GET("/review/:rid", getReqReview)
	e.PATCH("/review/:rid", updateReqReview)
	e.GET("/review-pairs", getReqReviewPairs)
	e.GET("/latest-post-lists/:uid", getReqLatestPostLists)
}

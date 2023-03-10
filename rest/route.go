package rest

import (
	"github.com/labstack/echo"
)

func Route(e *echo.Echo) {
	e.GET("/auth/tempsession/:service/:version", getReqTempSession)
	e.POST("/auth/session/:service/:version", postReqSession)
	e.PATCH("/auth/service/:service/:version", updateService)
	e.DELETE("/auth/service/:service", deleteService)
	e.DELETE("/auth/session", delReqSession)
	e.GET("/auth/check-session", getReqCheckSession)
	e.POST("/user", postReqUser)
	e.DELETE("/user/:uid/try", deleteUser1)
	e.DELETE("/user/:uid/commit", deleteUser2)
	e.GET("/user/:uid", getReqUserData)
	e.PATCH("/user/:uid", updateReqUser)
	e.POST("/tier", postReqTier)
	e.GET("/tier/:tid", getReqTier)
	e.GET("/userfile/:uid/:method/:id/:fname", getUserFile)
	e.PATCH("/tier/:tid", updateReqTier)
	e.DELETE("/tier/:rid", deleteReqTier)
	e.GET("/tiers", getReqTiers)
	e.POST("/review", postReqReview)
	e.GET("/review/:rid", getReqReview)
	e.PATCH("/review/:rid", updateReqReview)
	e.DELETE("/review/:rid", deleteReviewReq)
	e.GET("/review-pairs", getReqReviewPairs)
	e.GET("/latest-post-lists/:uid", getReqLatestPostLists)
	e.GET("/common/notifications", getNotifications)
	e.GET("/common/notifications-count", getNotificationsCount)
	e.PATCH("/common/notification-read/:nid", updateNotificationRead)
}

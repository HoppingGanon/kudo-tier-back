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
	e.GET("/user/:id", getReqUserData)
	e.POST("/tier", postReqTier)
	e.GET("/tier/:uid/:tid", getReqTier)
	e.GET("/"+os.Getenv("AP_FILE_PATH")+"/:uid/:method/:id/:fname", getUserFile)
	e.PATCH("/tier", updateReqTier)
}

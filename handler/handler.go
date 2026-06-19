package handler

import (
	"net/http"
	"redpacket/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

func CreateActivity(c *gin.Context) {
	var req service.CreateActivityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "invalid parameters: "+err.Error())
		return
	}

	activity, err := service.CreateActivity(&req)
	if err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	Success(c, activity)
}

func GrabRedPacket(c *gin.Context) {
	var req service.GrabRedPacketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "invalid parameters: "+err.Error())
		return
	}

	record, err := service.GrabRedPacket(&req)
	if err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	Success(c, record)
}

func GetActivity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Fail(c, http.StatusBadRequest, "invalid id")
		return
	}

	activity, err := service.GetActivity(uint(id))
	if err != nil {
		Fail(c, http.StatusNotFound, err.Error())
		return
	}
	Success(c, activity)
}

func ListActivities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	activities, total, err := service.ListActivities(page, pageSize)
	if err != nil {
		Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	Success(c, gin.H{
		"list":      activities,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetRecords(c *gin.Context) {
	idStr := c.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Fail(c, http.StatusBadRequest, "invalid id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	records, total, err := service.GetRecords(uint(activityID), page, pageSize)
	if err != nil {
		Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	Success(c, gin.H{
		"list":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetUserRedPackets(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		Fail(c, http.StatusBadRequest, "user_id is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	packets, total, err := service.GetUserRedPackets(userID, page, pageSize)
	if err != nil {
		Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	Success(c, gin.H{
		"list":      packets,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

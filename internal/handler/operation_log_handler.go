package handler

import (
	"strconv"
	"time"

	"go-shop/internal/entity"
	"go-shop/internal/repository"
	"go-shop/pkg/response"

	"github.com/gin-gonic/gin"
)

// OperationLogHandler 操作日志处理器
type OperationLogHandler struct {
	operationLogRepo repository.OperationLogRepository
}

// NewOperationLogHandler 创建操作日志处理器实例
func NewOperationLogHandler(operationLogRepo repository.OperationLogRepository) *OperationLogHandler {
	return &OperationLogHandler{
		operationLogRepo: operationLogRepo,
	}
}

// GetOperationLogs 获取操作日志列表
// GET /api/v2/admin/logs
// 需求：11.4 - 当管理员查看操作日志时，系统应显示操作时间、操作人、操作类型和操作详情
// 需求：11.5 - 当管理员查询日志时，系统应支持按时间范围、操作人、操作类型筛选
func (h *OperationLogHandler) GetOperationLogs(c *gin.Context) {
	// 绑定查询参数
	var req struct {
		UserID    *int64  `form:"user_id" binding:"omitempty,gt=0"`
		UserType  *string `form:"user_type" binding:"omitempty,oneof=user admin"`
		Module    *string `form:"module" binding:"omitempty"`
		Operation *string `form:"operation" binding:"omitempty"`
		StartTime *string `form:"start_time" binding:"omitempty"`
		EndTime   *string `form:"end_time" binding:"omitempty"`
		Page      int     `form:"page" binding:"omitempty,gt=0"`
		PageSize  int     `form:"page_size" binding:"omitempty,gt=0,lte=100"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize

	// 解析时间参数
	var startTime, endTime *time.Time
	if req.StartTime != nil {
		t, err := time.Parse("2006-01-02 15:04:05", *req.StartTime)
		if err != nil {
			// 尝试只解析日期
			t, err = time.Parse("2006-01-02", *req.StartTime)
			if err != nil {
				response.InvalidParam(c, "无效的开始时间格式，请使用 YYYY-MM-DD 或 YYYY-MM-DD HH:MM:SS")
				return
			}
		}
		startTime = &t
	}
	if req.EndTime != nil {
		t, err := time.Parse("2006-01-02 15:04:05", *req.EndTime)
		if err != nil {
			// 尝试只解析日期
			t, err = time.Parse("2006-01-02", *req.EndTime)
			if err != nil {
				response.InvalidParam(c, "无效的结束时间格式，请使用 YYYY-MM-DD 或 YYYY-MM-DD HH:MM:SS")
				return
			}
			// 如果只提供了日期，设置为当天的23:59:59
			t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
		}
		endTime = &t
	}

	// 转换用户类型
	var userType *entity.UserType
	if req.UserType != nil {
		ut := entity.UserType(*req.UserType)
		userType = &ut
	}

	// 获取日志列表
	logs, total, err := h.operationLogRepo.ListWithFilters(
		c.Request.Context(),
		req.UserID,
		userType,
		req.Module,
		req.Operation,
		startTime,
		endTime,
		offset,
		req.PageSize,
	)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	// 计算总页数
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	response.Success(c, gin.H{
		"items":       logs,
		"total":       total,
		"page":        req.Page,
		"page_size":   req.PageSize,
		"total_pages": totalPages,
	})
}

// GetOperationLogDetail 获取日志详情
// GET /api/v2/admin/logs/:id
// 需求：11.4 - 当管理员查看操作日志时，系统应显示操作时间、操作人、操作类型和操作详情
func (h *OperationLogHandler) GetOperationLogDetail(c *gin.Context) {
	// 获取日志ID
	logIDStr := c.Param("id")
	if logIDStr == "" {
		response.InvalidParam(c, "日志ID不能为空")
		return
	}

	logID, err := strconv.ParseInt(logIDStr, 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的日志ID")
		return
	}

	// 获取日志详情
	log, err := h.operationLogRepo.GetByID(c.Request.Context(), logID)
	if err != nil {
		if err.Error() == "operation log not found" {
			response.NotFound(c, "日志不存在")
			return
		}
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, log)
}

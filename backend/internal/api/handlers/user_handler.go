package handlers

import (
	"net/http"
	"shared-models/common"
	"strconv"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/services"
	"autocodeweb-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "注册请求"
// @Success 200 {object} common.Response{data=models.LoginResponse}
// @Failure 400 {object} common.ErrorResponse
// @Failure 409 {object} common.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      common.VALIDATION_ERROR,
			Message:   "请求参数错误: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	response, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "邮箱已存在" || err.Error() == "用户名已存在" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, common.Response{
			Code:      common.ERROR_CODE,
			Message:   err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "注册成功",
		Data:      response,
		Timestamp: utils.GetCurrentTime(),
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录并获取访问令牌
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "登录请求"
// @Success 200 {object} common.Response{data=models.LoginResponse}
// @Failure 400 {object} common.ErrorResponse
// @Failure 401 {object} common.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      common.VALIDATION_ERROR,
			Message:   "请求参数错误: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	response, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if err.Error() == "用户不存在或密码错误" {
			statusCode = http.StatusUnauthorized
		} else if err.Error() == "用户账户已被禁用" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, common.Response{
			Code:      common.ERROR_CODE,
			Message:   "登录失败, " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "登录成功",
		Data:      response,
		Timestamp: utils.GetCurrentTime(),
	})
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出并清除会话
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} common.Response
// @Failure 401 {object} common.ErrorResponse
// @Router /api/v1/users/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	// 调用登出服务（对于纯JWT实现，主要是客户端清除token）
	err := h.userService.Logout(c.Request.Context(), userID, "")
	if err != nil {
		c.JSON(http.StatusOK, common.Response{
			Code:      common.ERROR_CODE,
			Message:   "登出失败: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "登出成功",
		Timestamp: utils.GetCurrentTime(),
	})
}

// GetUserProfile 获取用户档案
// @Summary 获取用户档案
// @Description 获取当前用户的档案信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} common.Response{data=models.UserInfo}
// @Failure 401 {object} common.ErrorResponse
// @Failure 404 {object} common.ErrorResponse
// @Router /api/v1/users/profile [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	response, err := h.userService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "用户不存在" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, common.Response{
			Code:      common.ERROR_CODE,
			Message:   "获取用户档案失败, " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "获取用户档案成功",
		Data:      response,
		Timestamp: utils.GetCurrentTime(),
	})
}

// UpdateUserProfile 更新用户档案
// @Summary 更新用户档案
// @Description 更新当前用户的档案信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.UpdateProfileRequest true "更新档案请求"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 401 {object} common.ErrorResponse
// @Failure 409 {object} common.ErrorResponse
// @Router /api/v1/users/profile [put]
func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      common.VALIDATION_ERROR,
			Message:   "请求参数错误: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	err := h.userService.UpdateUserProfile(c.Request.Context(), userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "用户名已被使用" || err.Error() == "邮箱已被使用" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, common.Response{
			Code:      common.ERROR_CODE,
			Message:   "更新用户档案失败, " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "更新用户档案成功",
		Timestamp: utils.GetCurrentTime(),
	})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 401 {object} common.ErrorResponse
// @Router /api/v1/users/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      common.VALIDATION_ERROR,
			Message:   "请求参数错误: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	err := h.userService.ChangePassword(c.Request.Context(), userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "旧密码错误" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, common.Response{
			Code:      common.ERROR_CODE,
			Message:   "修改密码失败, " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "密码修改成功",
		Timestamp: utils.GetCurrentTime(),
	})
}

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表（需要管理员权限）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} common.Response{data=models.PaginationResponse}
// @Failure 401 {object} common.ErrorResponse
// @Failure 403 {object} common.ErrorResponse
// @Router /api/v1/users [get]
func (h *UserHandler) GetUserList(c *gin.Context) {
	// 从JWT中获取用户ID和角色（这里简化处理，实际应该从JWT中解析）
	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, common.Response{
			Code:      common.UNAUTHORIZED,
			Message:   "未授权",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 检查权限
	if userRole != "admin" {
		c.JSON(http.StatusForbidden, common.Response{
			Code:      common.FORBIDDEN,
			Message:   "权限不足",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.userService.GetUserList(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, common.Response{
			Code:      common.ERROR_CODE,
			Message:   "获取用户列表失败: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "获取成功",
		Data:      response,
		Timestamp: utils.GetCurrentTime(),
	})
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定用户（需要管理员权限）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param user_id path string true "用户ID"
// @Success 200 {object} common.Response
// @Failure 401 {object} common.ErrorResponse
// @Failure 403 {object} common.ErrorResponse
// @Router /api/v1/users/{user_id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 从中间件获取用户ID和角色
	currentUserID := c.GetString("user_id")
	userRole := c.GetString("user_role")

	// 检查权限
	if userRole != "admin" {
		c.JSON(http.StatusForbidden, common.Response{
			Code:      common.FORBIDDEN,
			Message:   "权限不足",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 获取要删除的用户ID
	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      common.VALIDATION_ERROR,
			Message:   "用户ID不能为空",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 不能删除自己
	if targetUserID == currentUserID {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      common.FORBIDDEN,
			Message:   "不能删除自己的账户",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	err := h.userService.DeleteUser(c.Request.Context(), targetUserID)
	if err != nil {
		c.JSON(http.StatusOK, common.Response{
			Code:      common.INTERNAL_ERROR,
			Message:   "删除用户失败: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "删除成功",
		Timestamp: utils.GetCurrentTime(),
	})
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param refresh_token query string true "刷新令牌"
// @Success 200 {object} common.Response{data=models.LoginResponse}
// @Failure 400 {object} common.ErrorResponse
// @Failure 401 {object} common.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	refreshToken := c.Query("refresh_token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      common.VALIDATION_ERROR,
			Message:   "刷新令牌不能为空",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	response, err := h.userService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if err.Error() == "无效的刷新令牌" {
			statusCode = http.StatusUnauthorized
		} else if err.Error() == "用户账户已被禁用" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, common.Response{
			Code:      common.ERROR_CODE,
			Message:   "令牌刷新失败, " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "令牌刷新成功",
		Data:      response,
		Timestamp: utils.GetCurrentTime(),
	})
}

package handlers

import (
	"net/http"
	"strconv"

	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/backend/internal/models"
	"github.com/lighthought/app-maker/backend/internal/services"

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
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "请求参数错误: "+err.Error()))
		return
	}

	response, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == common.MESSAGE_EMAIL_ALREADY_EXISTS || err.Error() == common.MESSAGE_USERNAME_ALREADY_EXISTS {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, utils.GetErrorResponse(common.ERROR_CODE, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("注册成功", response))
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
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "请求参数错误: "+err.Error()))
		return
	}

	response, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if err.Error() == common.MESSAGE_USER_OR_PASSWORD_ERROR {
			statusCode = http.StatusUnauthorized
		} else if err.Error() == common.MESSAGE_USER_DISABLED {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, utils.GetErrorResponse(common.ERROR_CODE, "登录失败, "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("登录成功", response))
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
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "登出失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("登出成功", nil))
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
		if err.Error() == common.MESSAGE_USER_NOT_FOUND {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, utils.GetErrorResponse(common.ERROR_CODE, "获取用户档案失败, "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取用户档案成功", response))
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
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "请求参数错误: "+err.Error()))
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	err := h.userService.UpdateUserProfile(c.Request.Context(), userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == common.MESSAGE_USERNAME_ALREADY_EXISTS || err.Error() == common.MESSAGE_EMAIL_ALREADY_EXISTS {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, utils.GetErrorResponse(common.ERROR_CODE, "更新用户档案失败, "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("更新用户档案成功", nil))
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
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "请求参数错误: "+err.Error()))
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	err := h.userService.ChangePassword(c.Request.Context(), userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == common.MESSAGE_OLD_PASSWORD_ERROR {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, utils.GetErrorResponse(common.ERROR_CODE, "修改密码失败, "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("密码修改成功", nil))
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
		c.JSON(http.StatusUnauthorized, utils.GetErrorResponse(common.UNAUTHORIZED, "未授权"))
		return
	}

	// 检查权限
	if userRole != common.UserRoleAdmin {
		c.JSON(http.StatusForbidden, utils.GetErrorResponse(common.FORBIDDEN, "权限不足"))
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.userService.GetUserList(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "获取用户列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取用户列表成功", response))
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
	if userRole != common.UserRoleAdmin {
		c.JSON(http.StatusForbidden, utils.GetErrorResponse(common.FORBIDDEN, "权限不足"))
		return
	}

	// 获取要删除的用户ID
	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "用户ID不能为空"))
		return
	}

	// 不能删除自己
	if targetUserID == currentUserID {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.FORBIDDEN, "不能删除自己的账户"))
		return
	}

	err := h.userService.DeleteUser(c.Request.Context(), targetUserID)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "删除用户失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("删除成功", nil))
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
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "刷新令牌不能为空"))
		return
	}

	response, err := h.userService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if err.Error() == common.MESSAGE_INVALID_REFRESH_TOKEN {
			statusCode = http.StatusUnauthorized
		} else if err.Error() == common.MESSAGE_USER_DISABLED {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, utils.GetErrorResponse(common.ERROR_CODE, "令牌刷新失败, "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("令牌刷新成功", response))
}

// GetUserSettings 获取用户设置
// @Summary 获取用户设置
// @Description 获取当前用户的开发设置（CLI工具、模型配置等）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} common.Response{data=models.UserSettingsResponse}
// @Failure 401 {object} common.ErrorResponse
// @Router /api/v1/users/settings [get]
func (h *UserHandler) GetUserSettings(c *gin.Context) {
	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	response, err := h.userService.GetUserSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GetErrorResponse(common.ERROR_CODE, "获取用户设置失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取用户设置成功", response))
}

// UpdateUserSettings 更新用户设置
// @Summary 更新用户设置
// @Description 更新当前用户的开发设置（CLI工具、模型配置等）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.UpdateUserSettingsRequest true "更新设置请求"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 401 {object} common.ErrorResponse
// @Router /api/v1/users/settings [put]
func (h *UserHandler) UpdateUserSettings(c *gin.Context) {
	var req models.UpdateUserSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "请求参数错误: "+err.Error()))
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	err := h.userService.UpdateUserSettings(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GetErrorResponse(common.ERROR_CODE, "更新用户设置失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("更新用户设置成功", nil))
}

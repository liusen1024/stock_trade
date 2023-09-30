package handler

import (
	"stock/api-gateway/dao"
	"stock/api-gateway/model"
	"stock/api-gateway/util"

	"github.com/gin-gonic/gin"
)

// AgentHandler 代理
type AgentHandler struct {
}

// NewAgentHandler 单例
func NewAgentHandler() *AgentHandler {
	return &AgentHandler{}
}

// Register 注册handler
func (h *AgentHandler) Register(e *gin.Engine) {
	e.GET("/cms/agent/list", JSONWrapper(h.Agent))
	e.GET("/cms/agent/get_modules", JSONWrapper(h.GetModules))
	e.POST("/cms/agent/create", JSONWrapper(h.Create))
	e.GET("/cms/agent/get_by_id", JSONWrapper(h.GetByID))
}

// GetByID 根据ID查询代理商
func (h *AgentHandler) GetByID(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		ID int64 `form:"id"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	role, err := dao.RoleDaoInstance().GetRoleByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	// 获取目录模块
	modules, err := dao.RoleModuleDaoInstance().GetModulesByRoleID(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	roleModules := make([]int64, 0)
	for _, it := range modules {
		roleModules = append(roleModules, it.ModuleID)
	}
	return map[string]interface{}{
		"id":        role.ID,
		"user_name": role.UserName,
		"password":  role.Password,
		"name":      role.UserName,
		"status":    role.Status,
		"module":    roleModules,
	}, nil
}

// Create 创建代理
func (h *AgentHandler) Create(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		ID       int64   `form:"id" json:"id"`
		UserName string  `form:"user_name" json:"user_name"`
		Password string  `form:"password" json:"password"`
		Name     string  `form:"name" json:"name"`
		Status   bool    `form:"status" json:"status"`
		Module   []int64 `form:"module" json:"module"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	role, err := dao.RoleDaoInstance().Create(ctx, &model.Role{
		ID:       req.ID,
		UserName: req.UserName,
		Password: req.Password,
		Status:   req.Status,
		IsAdmin:  req.UserName == "admin", // 超级管理员=admin
	})
	if err != nil {
		return nil, err
	}
	// 目录权限设置
	if err := dao.RoleModuleDaoInstance().Delete(ctx, role.ID); err != nil {
		return nil, err
	}
	modules := make([]*model.RoleModule, 0)
	for _, moduleID := range req.Module {
		modules = append(modules, &model.RoleModule{
			RoleID:   role.ID,
			Module:   model.RoleModuleMap[moduleID],
			ModuleID: moduleID,
		})
	}
	if err := dao.RoleModuleDaoInstance().Create(ctx, modules); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// GetModules 获取目录模块
func (h *AgentHandler) GetModules(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type Module struct {
		ID int64 `json:"id"` // 模块ID
	}
	list := make([]int64, 0)
	if IsAdmin(c) {
		for k, _ := range model.RoleModuleMap {
			list = append(list, k)
		}
	} else {
		role, err := dao.RoleDaoInstance().GetRoleByUserName(ctx, Username(c))
		if err != nil {
			return nil, err
		}
		modules, err := dao.RoleModuleDaoInstance().GetModulesByRoleID(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		for _, it := range modules {
			list = append(list, it.ModuleID)
		}
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// Agent 代理列表
func (h *AgentHandler) Agent(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type agent struct {
		ID       int64  `json:"id"`
		UserName string `json:"user_name"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Status   string `json:"status"`
	}
	agents := make([]*agent, 0)
	if len(Username(c)) == 0 {
		return map[string]interface{}{
			"list": agents,
		}, nil
	}

	roles, err := dao.RoleDaoInstance().GetRoles(ctx)
	if err != nil {
		return nil, err
	}
	for _, it := range roles {
		status := "激活"
		if !it.Status {
			status = "冻结"
		}
		// 非管理员情况,并且用户名不是当前代理账户 则跳过
		if !IsAdmin(c) && Username(c) != it.UserName {
			continue
		}
		agents = append(agents, &agent{
			ID:       it.ID,
			UserName: it.UserName,
			Password: it.Password,
			Name:     it.UserName,
			Status:   status,
		})
	}

	return map[string]interface{}{
		"list": agents,
	}, nil
}

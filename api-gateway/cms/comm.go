package handler

import (
	"bytes"
	"context"
	"fmt"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/util"
	"stock/common/log"
	"time"

	"gorm.io/gorm"

	"github.com/tealeg/xlsx"

	"github.com/gin-gonic/gin"
)

func AgentFilter(c *gin.Context) []int64 {
	ctx := util.RPCContext(c)
	userIDs := make([]int64, 0)
	type request struct {
		Agents []string `form:"agent[]" json:"agent[]"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return userIDs
	}

	// 管理员查询全部
	if IsAdmin(c) && len(req.Agents) == 0 {
		list, err := dao.UserDaoInstance().GetUsers(ctx)
		if err != nil {
			return userIDs
		}
		for _, it := range list {
			userIDs = append(userIDs, it.ID)
		}
		return userIDs
	}
	// 非管理员,智能查询自己名下的
	if !IsAdmin(c) {
		req.Agents = []string{Username(c)}
	}
	roles, err := dao.RoleDaoInstance().GetRolesByName(ctx, req.Agents)
	if err != nil {
		log.Errorf("GetRolesByName err:%+v", err)
		return userIDs
	}
	roleIDs := make([]int64, 0)
	for _, it := range roles {
		roleIDs = append(roleIDs, it.ID)
	}
	if len(roleIDs) == 0 {
		return userIDs
	}
	users, err := dao.UserDaoInstance().GetUserByRoleIDs(ctx, roleIDs)
	if err != nil {
		return userIDs
	}
	for _, it := range users {
		userIDs = append(userIDs, it.ID)
	}
	return userIDs
}

func ContractIDFilter(c *gin.Context, tx *gorm.DB) {
	ctx := context.Background()
	contractID, err := Int64(c, "contract_id")
	if err != nil {
		return
	}
	if contractID > 0 {
		contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
		if err != nil {
			tx.Where("uid = ?", 0)
			return
		}
		tx.Where("uid = ?", contract.UID)
	}
	return
}

func UserNameFilter(c *gin.Context, tx *gorm.DB) {
	ctx := context.Background()
	userName, err := String(c, "user_name")
	if err != nil {
		return
	}
	if len(userName) == 0 {
		return
	}
	user, err := dao.UserDaoInstance().GetUserByUserName(ctx, userName)
	if err != nil {
		tx.Where("uid = ?", 0)
		return
	}
	tx.Where("uid = ?", user.ID)
}

// RoleMap map[role.id][role.name]
func RoleMap(ctx context.Context) map[int64]string {
	roleMap := make(map[int64]string)
	roles, err := dao.RoleDaoInstance().GetRoles(ctx)
	if err != nil {
		return roleMap
	}
	for _, it := range roles {
		roleMap[it.ID] = it.UserName
	}
	return roleMap
}

// ContractMap map[contract.id]*model.Contract
func ContractMap(ctx context.Context) map[int64]*model.Contract {
	contractMap := make(map[int64]*model.Contract)
	contracts, err := dao.ContractDaoInstance().GetContracts(ctx)
	if err != nil {
		return contractMap
	}
	for _, it := range contracts {
		contractMap[it.ID] = it
	}
	return contractMap
}

func Online(ctx context.Context, uid int64) bool {
	return db.Get(ctx, fmt.Sprintf("user_online_%d", uid)).Val() == "1"
}

// UsersMap map[user.id]*model.User
func UsersMap(ctx context.Context) map[int64]*model.User {
	usersMap := make(map[int64]*model.User)
	users, err := dao.UserDaoInstance().GetUsers(ctx)
	if err != nil {
		log.Errorf("GetUsers err:%+v", err)
		return usersMap
	}
	for _, it := range users {
		usersMap[it.ID] = it
	}
	return usersMap
}

// Download 导出excel
func Download(c *gin.Context, titles []string, list []interface{}) {
	file := xlsx.NewFile()
	sheet, _ := file.AddSheet("Sheet1")
	// 插入表头
	titleRow := sheet.AddRow()
	for _, v := range titles {
		cell := titleRow.AddCell()
		cell.Value = v
	}
	// 插入内容
	for _, v := range list {
		row := sheet.AddRow()
		row.WriteStruct(v, -1)
	}
	var buf bytes.Buffer
	_ = file.Write(&buf)
	fileName := fmt.Sprintf("%s-%s.xlsx", time.Now().Format("2006-01-02"), Username(c))
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Writer.Header().Add("Content-Type", "application/octet-stream")
	file.Write(c.Writer)
}

package handler

import (
	"fmt"
	"sort"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/api-gateway/service"
	"stock/api-gateway/util"
	"stock/common/errgroup"
	"stock/common/log"
	"stock/common/timeconv"

	"github.com/gin-gonic/gin"
)

const (
	maxExpireTTL         = 3600 * 24 * 3
	randomStrSalt        = "hxcms_salt"
	userSessionLayoutKey = "cms_user_session_%s_v2"
)

// UserHandler 用户管理
type UserHandler struct {
}

// NewUserHandler 单例
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// Register 注册handler
func (h *UserHandler) Register(e *gin.Engine) {
	e.GET("/cms/user/list", JSONWrapper(h.UserList))
	e.GET("/cms/user/get_by_id", JSONWrapper(h.GetByID))
	e.POST("/cms/user/update", JSONWrapper(h.UpdateUser))
	e.POST("/cms/user/update_status", JSONWrapper(h.UpdateStatus))
	e.GET("/cms/user/recharge", JSONWrapper(h.Recharge))         // 充值
	e.POST("/cms/user/set_recharge", JSONWrapper(h.SetRecharge)) // 用户充值-确认
	e.GET("/cms/user/withdraw", JSONWrapper(h.Withdraw))         // 用户提现列表
	e.POST("/cms/user/set_withdraw", JSONWrapper(h.SetWithdraw)) // 用户提现-确认

}

// SetWithdraw 用户提现-确认
func (h *UserHandler) SetWithdraw(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		ID     int64 `form:"id" json:"id"`
		Status bool  `form:"status" json:"status"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	transfer, err := dao.TransferDaoInstance().GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, transfer.UID)
	if err != nil {
		return nil, err
	}

	if req.Status {
		// 转出成功
		transfer.Status = model.TransferStatusSuccess
		user.FreezeMoney -= transfer.Money
	} else {
		// 转出失败
		transfer.Status = model.TransferStatusFail
		user.Money += transfer.Money
		user.FreezeMoney -= transfer.Money
	}
	if err := dao.TransferDaoInstance().Create(ctx, transfer); err != nil {
		log.Errorf("Create err:%+v", err)
		return nil, err
	}

	wg := errgroup.GroupWithCount(2)
	wg.Go(func() error {
		if err := dao.UserDaoInstance().CreateUser(ctx, user); err != nil {
			return err
		}
		return nil
	})
	wg.Go(func() error {
		if req.Status {
			if err := service.SmsServiceInstance().SendSms(ctx, fmt.Sprintf("您转出资金到账:%0.2f,请打开App查看", transfer.Money), user.UserName); err != nil {
				log.Errorf("短信提醒失败:%+v", err)
			}
		}
		return nil
	})

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"result": true,
	}, nil
}

// Withdraw 用户提现列表
func (h *UserHandler) Withdraw(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		Agent     []string `form:"agent" json:"agent"`
		BeginDate int32    `form:"begin_date" json:"begin_date"`
		EndDate   int32    `form:"end_date" json:"end_date"`
		UserName  string   `form:"user_name" json:"user_name"`
		Limit     int      `form:"limit" json:"limit"`
		Offset    int      `json:"offset" json:"offset"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	tx := db.StockDB().WithContext(ctx).Table("transfer")
	tx.Where("uid in (?)", AgentFilter(c))
	tx.Where("status != ? ", model.TransferStatusPre) // 跳过预审
	tx.Where("type = ?", model.TransferTypeWithdraw)  // 提现
	UserNameFilter(c, tx)
	if req.BeginDate > 0 {
		tx.Where("order_time >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("order_time <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}

	var transfer []*model.Transfer
	if err := tx.Find(&transfer).Error; err != nil {
		return nil, err
	}
	// 排序:待审核、时间
	a := make([]*model.Transfer, 0)
	b := make([]*model.Transfer, 0)
	for _, it := range transfer {
		if it.Status == model.TransferStatusWaitExam {
			a = append(a, it) // 待审核
		} else {
			b = append(b, it) // 成功、失败
		}
	}
	sort.SliceStable(a, func(i, j int) bool {
		return timeconv.TimeToInt64(a[i].OrderTime) > timeconv.TimeToInt64(a[j].OrderTime)
	})
	sort.SliceStable(b, func(i, j int) bool {
		return timeconv.TimeToInt64(b[i].OrderTime) > timeconv.TimeToInt64(b[j].OrderTime)
	})

	transfer = transfer[0:0]
	transfer = append(transfer, a...)
	transfer = append(transfer, b...)

	usersMap := UsersMap(ctx)
	roleMap := RoleMap(ctx)
	list := make([]*model.WithdrawResp, 0)
	for _, it := range transfer {
		user, ok := usersMap[it.UID]
		if !ok {
			continue
		}
		list = append(list, &model.WithdrawResp{
			ID:       it.ID,
			UserName: user.UserName,
			Name:     user.Name,
			Agent:    roleMap[user.RoleID],
			Time:     it.OrderTime.Format("2006-01-02 15:04:05"),
			Money:    it.Money,
			BankName: it.Name,
			BankNo:   it.BankNo,
			Status:   it.Status, // 1待审核 2成功 3失败
		})
	}

	if IsDownload(c) {
		var res []interface{}
		for _, it := range list {
			res = append(res, it)
		}
		Download(c, []string{"流水号", "用户名称", "姓名", "代理机构", "时间", "金额", "收款人姓名", "银行卡号", "审核状态:1待审核2成功3失败"}, res)
	}

	count := len(list)
	start, end := SlicePage(c, count)
	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil
}

// SetRecharge 用户管理-用户充值-确认
func (h *UserHandler) SetRecharge(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		ID     int64 `form:"id" json:"id"`
		Status bool  `form:"status" json:"status"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	transfer, err := dao.TransferDaoInstance().GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, transfer.UID)
	if err != nil {
		return nil, err
	}

	if req.Status {
		transfer.Status = model.TransferStatusSuccess
		user.Money += transfer.Money
		if err := dao.UserDaoInstance().CreateUser(ctx, user); err != nil {
			return nil, err
		}
		if err := service.SmsServiceInstance().SendSms(ctx, fmt.Sprintf("您转入资金到账:%0.2f,请打开App查看", transfer.Money), user.UserName); err != nil {
			log.Errorf("短信提醒失败:%+v", err)
		}
	} else {
		transfer.Status = model.TransferStatusFail
	}
	if err := dao.TransferDaoInstance().Create(ctx, transfer); err != nil {
		log.Errorf("Create err:%+v", err)
		return nil, err
	}

	return map[string]interface{}{
		"result": true,
	}, nil
}

// Recharge 用户管理-用户充值
func (h *UserHandler) Recharge(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		Agent     []string `form:"agent" json:"agent"`
		BeginDate int32    `form:"begin_date" json:"begin_date"`
		EndDate   int32    `form:"end_date" json:"end_date"`
		UserName  string   `form:"user_name" json:"user_name"`
		Limit     int      `form:"limit" json:"limit"`
		Offset    int      `json:"offset" json:"offset"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	tx := db.StockDB().WithContext(ctx).Table("transfer")
	tx.Where("uid in (?)", AgentFilter(c))
	tx.Where("status != 0 ") // 跳过预审
	tx.Where("type = ?", model.TransferTypeRecharge)
	UserNameFilter(c, tx)
	if req.BeginDate > 0 {
		tx.Where("order_time >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("order_time <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}

	var transfer []*model.Transfer
	if err := tx.Find(&transfer).Error; err != nil {
		return nil, err
	}
	// 排序
	a := make([]*model.Transfer, 0)
	b := make([]*model.Transfer, 0)
	for _, it := range transfer {
		if it.Status == model.TransferStatusWaitExam {
			a = append(a, it) // 待审核
		} else {
			b = append(b, it) // 成功、失败
		}
	}
	sort.SliceStable(a, func(i, j int) bool {
		return timeconv.TimeToInt64(a[i].OrderTime) > timeconv.TimeToInt64(a[j].OrderTime)
	})
	sort.SliceStable(b, func(i, j int) bool {
		return timeconv.TimeToInt64(b[i].OrderTime) > timeconv.TimeToInt64(b[j].OrderTime)
	})
	transfer = transfer[0:0]
	transfer = append(transfer, a...)
	transfer = append(transfer, b...)

	list := make([]*model.RechargeResp, 0)
	usersMap := UsersMap(ctx)
	roleMap := RoleMap(ctx)
	for _, it := range transfer {
		user, ok := usersMap[it.UID]
		if !ok {
			continue
		}
		list = append(list, &model.RechargeResp{
			ID:       it.ID,
			UserName: user.UserName,
			Name:     user.Name,
			Agent:    roleMap[user.RoleID],
			Time:     it.OrderTime.Format("2006-01-02 15:04:05"),
			Money:    it.Money,
			OrderNo:  it.OrderNo,
			Channel:  it.Channel,
			Status:   it.Status,
		})
	}

	if IsDownload(c) {
		var res []interface{}
		for _, it := range list {
			res = append(res, it)
		}
		Download(c, []string{"流水号", "用户名称", "姓名", "代理机构", "时间", "金额", "订单流水号", "充值渠道", "审核状态:1待审核2成功3失败"}, res)
	}

	count := len(list)
	start, end := SlicePage(c, count)
	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil
}

// UpdateStatus 更新用户状态
func (h *UserHandler) UpdateStatus(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		ID     int64 `form:"id" json:"id"`
		Status bool  `form:"status" json:"status"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if !req.Status {
		user.Status = model.UserStatusActive
	} else {
		user.Status = model.UserStatusFrezze
	}
	if err := dao.UserDaoInstance().CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		ID          int64   `form:"id" json:"id"`
		Agent       string  `form:"agent" json:"agent"`
		Money       float64 `form:"money" json:"money"`
		FreezeMoney float64 `form:"freeze_money" json:"freeze_money"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if !IsAdmin(c) {
		return nil, serr.ErrBusiness("更新失败")
	}
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if len(req.Agent) > 0 {
		role, err := dao.RoleDaoInstance().GetRoleByUserName(ctx, req.Agent)
		if err != nil {
			return nil, err
		}
		user.RoleID = role.ID
	}
	if req.Money > 0 {
		user.Money = req.Money
	}
	if req.FreezeMoney > 0 {
		user.FreezeMoney = req.FreezeMoney
	}
	if err := dao.UserDaoInstance().CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// GetByID 根据ID查询用户
func (h *UserHandler) GetByID(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		ID int64 `form:"id" json:"id"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	status := "激活"
	if user.Status != model.UserStatusActive {
		status = "冻结"
	}
	roleMap := RoleMap(ctx)
	return map[string]interface{}{
		"id":           user.ID,
		"user_name":    user.UserName,
		"name":         user.Name,
		"password":     user.Password,
		"status":       status,
		"id_no":        user.ICCID,
		"agent":        roleMap[user.RoleID],
		"money":        user.Money,
		"freeze_money": user.FreezeMoney,
	}, nil
}

// UserList 用户列表
func (h *UserHandler) UserList(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		BeginDate int32 `form:"begin_date" json:"begin_date"`
		EndDate   int32 `form:"end_date" json:"end_date"`
		Limit     int   `form:"limit" json:"limit"`
		Offset    int   `json:"offset" json:"offset"`
	}
	var req request
	if err := c.ShouldBind(&req); err != nil {
		return nil, err
	}

	tx := db.StockDB().WithContext(ctx).Table("users")
	tx.Where("id in (?)", AgentFilter(c))
	if req.BeginDate > 0 {
		tx.Where("created_at >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("created_at <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}

	var users []*model.User
	if err := tx.Find(&users).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(users, func(i, j int) bool {
		return timeconv.TimeToInt64(users[i].CreateAt) > timeconv.TimeToInt64(users[j].CreateAt)
	})
	contracts, err := dao.ContractDaoInstance().GetContracts(ctx)
	if err != nil {
		return nil, err
	}
	// contractMap[uid][合约数量]
	contractMap := make(map[int64]int64)
	for _, it := range contracts {
		contractNum, ok := contractMap[it.UID]
		if !ok {
			contractMap[it.UID] = 0
			continue
		}
		if it.Status != model.ContractStatusEnable {
			continue
		}
		contractNum++
		contractMap[it.UID] = contractNum
	}
	roleMap := RoleMap(ctx)
	userList := make([]*model.UserListResp, 0)
	for _, it := range users {
		userList = append(userList, &model.UserListResp{
			ID:             it.ID,
			UserName:       it.UserName,
			Password:       it.Password,
			Name:           it.Name,
			Agent:          roleMap[it.RoleID],
			Money:          it.Money,
			FreezeMoney:    it.FreezeMoney,
			Broker:         true,
			Contract:       contractMap[it.ID],
			Authentication: len(it.ICCID) > 0,
			Online:         Online(ctx, it.ID),
			RegisterTime:   it.CreateAt.Format("2006-01-02 15:04:05"),
			Status:         it.Status == model.UserStatusActive,
		})
	}

	// 下载则下发文件
	if IsDownload(c) {
		var res []interface{}
		for _, it := range userList {
			res = append(res, it)
		}
		Download(c, []string{
			"ID", "user_name", "password",
			"name", "代理", "金额", "冻结资金", "券商", "合约",
			"实名认证", "在线状态", "注册时间", "状态",
		}, res)
	}

	count := len(userList)
	start, end := SlicePage(c, count)

	return map[string]interface{}{
		"list":  userList[start:end],
		"total": count,
	}, nil
}

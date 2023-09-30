package handler

import (
	"sort"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/quote"
	"stock/api-gateway/service"
	"stock/api-gateway/util"
	"stock/common/timeconv"

	"github.com/gin-gonic/gin"
)

// ContractHandler 合约
type ContractHandler struct {
}

// NewContractHandler 单例
func NewContractHandler() *ContractHandler {
	return &ContractHandler{}
}

// Register 注册handler
func (h *ContractHandler) Register(e *gin.Engine) {
	e.GET("/cms/contract/get_by_id", JSONWrapper(h.GetByID))           // 合约信息
	e.POST("/cms/contract/update", JSONWrapper(h.Update))              // 更新合约
	e.GET("/cms/contract/list", JSONWrapper(h.List))                   // 合约列表
	e.GET("/cms/contract/fund_detail", JSONWrapper(h.FundDetail))      // 合约管理-资金明细
	e.GET("/cms/contract/fund_detail/items", JSONWrapper(h.FundItems)) // 合约管理-资金明细-费项列表
}

// FundItems 合约管理-资金明细-费项列表
func (h *ContractHandler) FundItems(c *gin.Context) (interface{}, error) {

	list := make([]string, 0)
	for _, v := range model.ContractFeeTypeMap {
		list = append(list, v)
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// FundDetail 资金明细
func (h *ContractHandler) FundDetail(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		BeginDate int32  `form:"begin_date" json:"begin_date"`
		EndDate   int32  `form:"end_date" json:"end_date"`
		Item      string `form:"item" json:"item"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	tx := db.StockDB().WithContext(ctx).Table("contract_fee")
	tx.Where("uid in (?)", AgentFilter(c))
	UserNameFilter(c, tx)
	ContractIDFilter(c, tx)
	if len(req.Item) > 0 {
		tx.Where("type = ?", model.ContractFeeType(req.Item))
	}
	if req.BeginDate > 0 {
		tx.Where("order_time >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("order_time <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}

	var contractFee []*model.ContractFee
	if err := tx.Find(&contractFee).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(contractFee, func(i, j int) bool {
		return timeconv.TimeToInt64(contractFee[i].OrderTime) > timeconv.TimeToInt64(contractFee[j].OrderTime)
	})

	roleMap := RoleMap(ctx)
	userMap := UsersMap(ctx)
	contractMap := ContractMap(ctx)

	list := make([]*model.CmsContractFeeResp, 0)
	for _, it := range contractFee {
		user, ok := userMap[it.UID]
		if !ok {
			continue
		}
		contract, ok := contractMap[it.ContractID]
		if !ok {
			contract = &model.Contract{}
		}
		list = append(list, &model.CmsContractFeeResp{
			ID:           it.ContractID,                              // 合约ID
			ContractName: contract.FullName(),                        // 合约名称
			UserName:     user.UserName,                              // 用户名称
			Name:         user.Name,                                  // 用户姓名
			Agent:        roleMap[user.RoleID],                       // 代理机构
			Time:         it.OrderTime.Format("2006-01-02 15:04:05"), // 时间
			StockCode:    it.Code,                                    // 股票代码
			StockName:    it.Name,                                    // 股票名称
			Balance:      it.Money,                                   // 交易金额
			Item:         model.ContractFeeTypeMap[it.Type],          // 费项
			Detail:       it.Detail,                                  // 明细
		})
	}

	// 下载则下发文件
	if IsDownload(c) {
		var res []interface{}
		for _, it := range list {
			res = append(res, it)
		}
		Download(c, []string{
			"合约ID", "合约名称", "用户名称", "用户姓名", "代理机构", "时间", "股票代码", "股票名称", "交易金额",
			"费项", "明细",
		}, res)
	}

	count := len(list)
	start, end := SlicePage(c, count)

	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil
}

// List 合约列表
func (h *ContractHandler) List(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		BeginDate int32 `form:"begin_date" json:"begin_date"`
		EndDate   int32 `form:"end_date" json:"end_date"`
		Status    int64 `form:"status" json:"status"` // 状态 0:全部 1有效 2失效
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	tx := db.StockDB().WithContext(ctx).Table("contract")
	tx.Where("uid in (?)", AgentFilter(c))
	if req.BeginDate > 0 {
		tx.Where("order_time >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("order_time <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}
	if req.Status == 1 { // 有效合约
		tx.Where("status = ?", model.ContractStatusEnable)
	} else if req.Status == 2 { // 失效合约
		tx.Where("status = ?", model.ContractStatusDisabled)
	} else { // 全部状态:排除预申请的
		tx.Where("status != ?", model.ContractStatusApply)
	}

	UserNameFilter(c, tx)

	var contracts []*model.Contract
	if err := tx.Find(&contracts).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(contracts, func(i, j int) bool {
		return timeconv.TimeToInt64(contracts[i].OrderTime) > timeconv.TimeToInt64(contracts[j].OrderTime)
	})

	roleMap := RoleMap(ctx)
	userMap := UsersMap(ctx)

	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		return nil, err
	}

	// 持仓
	positions, err := dao.PositionDaoInstance().GetPositions(ctx)
	if err != nil {
		return nil, err
	}
	// 查询行情
	codes := make([]string, 0)
	for _, it := range positions {
		codes = append(codes, it.StockCode)
	}
	qt, err := quote.QtServiceInstance().GetQuoteByTencent(codes)
	if err != nil {
		return nil, err
	}

	// positionMap[contract.ID][]*model.Position
	positionMap := make(map[int64][]*model.Position)
	for _, it := range positions {
		stock, ok := qt[it.StockCode]
		if ok {
			it.CurPrice = stock.CurrentPrice
		}
		v, ok := positionMap[it.ContractID]
		if !ok {
			v = make([]*model.Position, 0)
		}
		v = append(v, it)
		positionMap[it.ContractID] = v
	}

	list := make([]*model.CmsContractResp, 0)
	for _, it := range contracts {
		user, ok := userMap[it.UID]
		if !ok {
			continue
		}
		contract := &model.CmsContractResp{
			ID:           it.ID,
			ContractName: it.FullName(),
			UserName:     user.UserName,
			Name:         user.Name,
			Agent:        roleMap[user.RoleID],
			Time:         it.OrderTime.Format("2006-01-02 15:04:05"),
		}
		if it.Status != model.ContractStatusEnable {
			contract.Status = "失效"
			list = append(list, contract)
			continue
		}

		contract.InitMoney = it.InitMoney
		contract.Money = it.Money
		contract.ValMoney = it.ValMoney
		contract.AppendMoney = it.AppendMoney
		contract.Warn = service.ContractServiceInstance().Warn(it, sys)
		contract.Close = service.ContractServiceInstance().Close(it, sys)
		contract.Status = "有效"

		// 总资产:合约可用资金+股票市值
		if _, ok := positionMap[it.ID]; ok {
			contract.Asset = contract.ValMoney + model.CalculatePositionMarketValue(positionMap[it.ID]) // 总资产
			contract.MarketValue = model.CalculatePositionMarketValue(positionMap[it.ID])               // 持仓市值
			contract.Profit = util.FloatRound(model.CalculatePositionProfit(positionMap[it.ID]), 2)
			if it.InitMoney*sys.ClosePct > it.Money+contract.Profit {
				contract.Risk = "触发平仓线"
			} else if it.InitMoney*sys.WarnPct > it.Money+contract.Profit {
				contract.Risk = "触发警戒线"
			} else {
				contract.Risk = "安全"
			}
		}

		list = append(list, contract)
	}

	// 下载则下发文件
	if IsDownload(c) {
		var res []interface{}
		for _, it := range list {
			res = append(res, it)
		}
		Download(c, []string{
			"ID", "合约名称", "用户名称", "用户姓名", "代理机构", "时间", "总资产", "持仓市值", "盈亏金额",
			"原始保证金", "现保证金", "可用资金", "追加保证金", "警戒线", "平仓线", "合约状态", "合约风控",
		}, res)
	}

	count := len(list)
	start, end := SlicePage(c, count)

	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil
}

// Update 更新
func (h *ContractHandler) Update(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		ID        int64   `form:"id" json:"id"`
		Money     float64 `form:"money" json:"money"`
		InitMoney float64 `form:"init_money" json:"init_money"`
		ValMoney  float64 `form:"val_money" json:"val_money"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	contract.Money = req.Money
	contract.InitMoney = req.InitMoney
	contract.ValMoney = req.ValMoney
	if err := dao.ContractDaoInstance().UpdateContract(ctx, contract); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// GetByID 代理列表
func (h *ContractHandler) GetByID(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	id, err := Int64(c, "id")
	if err != nil {
		return nil, err
	}
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, contract.UID)
	if err != nil {
		return nil, err
	}
	status := "有效"
	if contract.Status != model.ContractStatusEnable {
		status = "失效"
	}
	return map[string]interface{}{
		"id":            contract.ID,
		"user_name":     user.UserName,
		"name":          user.Name,
		"contract_name": contract.FullName(),
		"lever":         contract.Lever,
		"status":        status,
		"time":          contract.OrderTime.Format("2006-01-02 15:04:05"),
		"money":         contract.Money,
		"val_money":     contract.ValMoney,
		"init_money":    contract.InitMoney,
	}, nil
}

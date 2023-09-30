package handler

import (
	"context"
	"fmt"
	"sort"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/quote"
	"stock/api-gateway/service"
	"stock/api-gateway/util"
	"stock/common/timeconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// TradeHandle 管理
type TradeHandle struct {
}

// NewTradeHandler 单例
func NewTradeHandler() *TradeHandle {
	return &TradeHandle{}
}

// Register 注册handler
func (h *TradeHandle) Register(e *gin.Engine) {
	e.GET("/cms/trade/buy", JSONWrapper(h.Buy))                                              // 股票交易-买入交易
	e.GET("/cms/trade/sell", JSONWrapper(h.Sell))                                            // 股票交易-卖出交易
	e.GET("/cms/trade/position", JSONWrapper(h.Position))                                    // 股票交易-持仓
	e.GET("/cms/trade/position/dividend", JSONWrapper(h.Dividend))                           // 股票交易-分红配送
	e.GET("/cms/trade/position/get_sell_stock_by_id", JSONWrapper(h.GetPositionByEntrustID)) // 股票交易-持仓-平仓查询
	e.POST("/cms/trade/position/sell_stock", JSONWrapper(h.SellStock))                       // 股票交易-持仓-平仓
	e.GET("/cms/trade/detail", JSONWrapper(h.TradeDetail))                                   // 股票交易-明细
	e.GET("/cms/trade/entrust", JSONWrapper(h.TradeEntrust))                                 // 股票交易-委托
}

// TradeEntrust 股票交易-委托
func (h *TradeHandle) TradeEntrust(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		BeginDate int32 `form:"begin_date" json:"begin_date"`
		EndDate   int32 `form:"end_date" json:"end_date"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	tx := db.StockDB().WithContext(ctx).Table("entrust")
	tx.Where("uid in (?)", AgentFilter(c))
	ContractIDFilter(c, tx)
	UserNameFilter(c, tx)

	if req.BeginDate > 0 {
		tx.Where("order_time >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("order_time <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}

	var entrusts []*model.Entrust
	if err := tx.Find(&entrusts).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(entrusts, func(i, j int) bool {
		return timeconv.TimeToInt64(entrusts[i].OrderTime) > timeconv.TimeToInt64(entrusts[j].OrderTime)
	})

	roleMap := RoleMap(ctx)
	userMap := UsersMap(ctx)
	contractMap := ContractMap(ctx)

	entrustIDs := make([]int64, 0)
	for _, it := range entrusts {
		entrustIDs = append(entrustIDs, it.ID)
	}
	// 获取券商委托表ID map[entrust.id] []*model.BrokerEntrust
	brokerEntrustMap := getBrokerEntrustByEntrustIDs(ctx, entrustIDs)

	// 券商
	brokerIDs := make([]int64, 0)
	for _, v := range brokerEntrustMap {
		for _, it := range v {
			brokerIDs = append(brokerIDs, it.BrokerID)
		}
	}
	brokerMap := getBrokerByIDs(ctx, brokerIDs)

	list := make([]*model.TradeEntrustResp, 0)
	for _, it := range entrusts {
		user, ok := userMap[it.UID]
		if !ok {
			continue
		}
		contract, ok := contractMap[it.ContractID]
		if !ok {
			contract = &model.Contract{}
		}
		// 券商委托表
		brokerEntrusts, ok := brokerEntrustMap[it.ID]
		if !ok {
			brokerEntrusts = make([]*model.BrokerEntrust, 0)
		}

		// 券商
		brokerAccount := "" // 券商委托账号:平安证券-XSD123
		brokerOrderNo := "" // 券商委托编号: 平安证券-SXF101
		for _, it := range brokerEntrusts {
			broker, ok := brokerMap[it.BrokerID]
			if !ok {
				continue
			}
			brokerAccount += fmt.Sprintf("%s-%s;", broker.BrokerName, broker.FundAccount)
			brokerOrderNo += fmt.Sprintf("%s-%s;", broker.BrokerName, it.BrokerEntrustNo)
		}
		// 委托失败原因填写
		if it.IsBrokerEntrust && len(it.Remark) > 0 {
			brokerAccount, brokerOrderNo = it.Remark, it.Remark
		}

		typ := "限价"
		if it.EntrustProp == model.EntrustPropTypeMarketPrice {
			typ = "市价"
		}
		if it.EntrustBS == model.EntrustBsTypeBuy {
			typ += "买入"
		} else {
			typ += "卖出"
		}

		list = append(list, &model.TradeEntrustResp{
			ID:            it.ID,
			UserName:      user.UserName,
			Name:          user.Name,
			Agent:         roleMap[user.RoleID],
			Time:          it.OrderTime.Format("2006-01-02 15:04:05"),
			ContractID:    it.ContractID,
			ContractName:  contract.FullName(),
			StockCode:     it.StockCode,
			StockName:     it.StockName,
			Price:         it.Price,
			Amount:        it.Amount,
			Type:          typ,
			Status:        model.EntrustStatusMap[it.Status],
			Broker:        it.IsBrokerEntrust,
			BrokerAccount: brokerAccount,
			BrokerOrderNo: brokerOrderNo,
			Remark:        it.Remark,
		})
	}

	// 下载则下发文件
	if IsDownload(c) {
		var res []interface{}
		for _, it := range list {
			res = append(res, it)
		}
		Download(c, []string{
			"ID", "用户名称", "姓名", "代理机构", "时间", "合约ID", "合约名称", "股票代码", "股票名称",
			"委托价格", "委托数量", "交易类型", "状态", "对接券商", "对接券商账号", "券商委托序号", "备注",
		}, res)
	}

	count := len(list)
	start, end := SlicePage(c, count)
	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil
}

// TradeDetail 股票交易-明细
func (h *TradeHandle) TradeDetail(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	entrustID, err := Int64(c, "id")
	if err != nil {
		return nil, err
	}
	entrust, err := dao.EntrustDaoInstance().GetEntrustByID(ctx, entrustID)
	if err != nil {
		return nil, err
	}
	// 委托为卖出时候,查找最早的一笔买入记录委托
	if entrust.EntrustBS == model.EntrustBsTypeSell {
		e, err := dao.EntrustDaoInstance().GetFirstBuyEntrustByPositionID(ctx, entrust)
		if err != nil {
			return nil, err
		}
		entrust = e
	}

	user, err := dao.UserDaoInstance().GetUserByUID(ctx, entrust.UID)
	if err != nil {
		return nil, err
	}
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, entrust.ContractID)
	if err != nil {
		return nil, err
	}
	roleMap := RoleMap(ctx)

	// 券商账号
	brokerAccount := ""
	if entrust.IsBrokerEntrust {
		brokerEntrusts, err := dao.BrokerEntrustDaoInstance().GetByEntrustID(ctx, entrust.ID)
		if err != nil {
			return nil, err
		}
		brokerIDs := make([]int64, 0)
		for _, brokerEntrust := range brokerEntrusts {
			brokerIDs = append(brokerIDs, brokerEntrust.BrokerID)
		}
		brokers, err := dao.BrokerDaoInstance().GetBrokersByIDs(ctx, brokerIDs)
		if err != nil {
			return nil, err
		}
		brokerMap := make(map[int64]*model.Broker)
		for _, it := range brokers {
			brokerMap[it.ID] = it
		}
		accountList := make([]string, 0)
		for _, brokerEntrust := range brokerEntrusts {
			broker, ok := brokerMap[brokerEntrust.ID]
			if !ok {
				continue
			}
			brokerAccount = fmt.Sprintf("%s-%s", broker.BrokerName, broker.FundAccount)
		}
		brokerAccount = strings.Join(accountList, ";")
	}

	// 买入列表
	buys, err := dao.BuyDaoInstance().GetByPositionID(ctx, entrust.PositionID)
	if err != nil {
		return nil, err
	}
	buyItems := make([]*model.BuyDetailItem, 0)
	for _, it := range buys {
		typ := "市价买入"
		if it.EntrustProp == model.EntrustPropTypeLimitPrice {
			typ = "限价买入"
		}
		buyItems = append(buyItems, &model.BuyDetailItem{
			ID:        it.ID,
			OrderTime: it.OrderTime.Format("2006-01-02 15:04:05"),
			StockCode: it.StockCode,
			StockName: it.StockName,
			Price:     it.Price,
			Amount:    it.Amount,
			Balance:   it.Balance,
			Type:      typ,
			Fee:       it.Fee,
		})
	}

	// 卖出列表
	sells, err := dao.SellDaoInstance().GetByPositionID(ctx, entrust.PositionID)
	if err != nil {
		return nil, err
	}
	sellItems := make([]*model.SellDetailItem, 0)
	for _, it := range sells {
		typ := "市价卖出"
		if it.EntrustProp == model.EntrustPropTypeLimitPrice {
			typ = "限价卖出"
		}
		mode := "主动卖出"
		if it.Mode == 2 {
			mode = "系统卖出"
		}
		sellItems = append(sellItems, &model.SellDetailItem{
			ID:            it.ID,
			OrderTime:     it.OrderTime.Format("2006-01-02 15:04:05"),
			StockCode:     it.StockCode,
			StockName:     it.StockName,
			PositionPrice: it.PositionPrice,
			DealPrice:     it.Price,
			Amount:        it.Amount,
			Balance:       it.Balance,
			Profit:        it.Profit,
			Fee:           it.Fee,
			Type:          typ,
			Mode:          mode,
			Reason:        it.Reason,
		})
	}
	return &model.TradeDetailResp{
		ID:             entrustID,
		UserName:       user.UserName,
		Name:           user.Name,
		Agent:          roleMap[user.RoleID],
		Time:           entrust.OrderTime.Format("2006-01-02 15:04:05"),
		ContractID:     entrust.ContractID,
		ContractName:   contract.FullName(),
		Broker:         entrust.IsBrokerEntrust,
		BrokerAccount:  brokerAccount,
		BuyDetailItem:  buyItems,
		SellDetailItem: sellItems,
	}, nil
}

// SellStock 股票交易-持仓-平仓
func (h *TradeHandle) SellStock(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		EntrustID int64   `form:"id" json:"id"`
		Price     float64 `form:"price" json:"price"`
		Amount    int64   `form:"amount" json:"amount"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	position, err := dao.PositionDaoInstance().GetPositionByEntrustID(ctx, req.EntrustID)
	if err != nil {
		return nil, err
	}
	if err := service.TradeServiceInstance().Sell(ctx, &model.EntrustPackage{
		UID:         position.UID,                    // 用户UID
		ContractID:  position.ContractID,             // 合约ID
		Code:        position.StockCode,              // 股票代码
		Price:       req.Price,                       // 股票价格
		Amount:      req.Amount,                      // 股票数量
		EntrustProp: model.EntrustPropTypeLimitPrice, // 委托类型:1限价 2市价
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// GetPositionByEntrustID 股票交易-持仓-平仓查询
func (h *TradeHandle) GetPositionByEntrustID(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		ID int64 `form:"id" json:"id"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	var position *model.Position
	tx := db.StockDB().WithContext(ctx).Table("position")
	if err := tx.Where("entrust_id = ?", req.ID).Take(&position).Error; err != nil {
		return nil, err
	}

	user, err := dao.UserDaoInstance().GetUserByUID(ctx, position.UID)
	if err != nil {
		return nil, err
	}

	qt, err := quote.QtServiceInstance().GetQuoteByTencent([]string{position.StockCode})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":             position.ID,
		"user_name":      user.UserName,
		"stock_code":     position.StockCode,
		"stock_name":     position.StockName,
		"position_price": position.Price,
		"current_price":  qt[position.StockCode].CurrentPrice,
		"amount":         position.Amount,
		"freeze_amount":  position.FreezeAmount,
	}, nil
}

// Dividend 股票交易-分红配送
func (h *TradeHandle) Dividend(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		BeginDate int32 `form:"begin_date" json:"begin_date"`
		EndDate   int32 `form:"end_date" json:"end_date"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	tx := db.StockDB().WithContext(ctx).Table("dividend")
	tx.Where("uid in (?)", AgentFilter(c))
	ContractIDFilter(c, tx)
	UserNameFilter(c, tx)
	if req.BeginDate > 0 {
		tx.Where("order_time >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("order_time <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}

	var dividend []*model.Dividend
	if err := tx.Find(&dividend).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(dividend, func(i, j int) bool {
		return timeconv.TimeToInt64(dividend[i].OrderTime) > timeconv.TimeToInt64(dividend[j].OrderTime)
	})

	roleMap := RoleMap(ctx)
	userMap := UsersMap(ctx)
	contractMap := ContractMap(ctx)

	list := make([]*model.TradeDividendResp, 0)
	for _, it := range dividend {
		user, ok := userMap[it.UID]
		if !ok {
			continue
		}
		contract, ok := contractMap[it.ContractID]
		if !ok {
			contract = &model.Contract{}
		}
		dividendAmount := it.PlanExplain
		dividendMoney := it.PlanExplain

		list = append(list, &model.TradeDividendResp{
			ID:             it.ID,
			UserName:       user.UserName,
			Name:           user.Name,
			Agent:          roleMap[user.RoleID],
			Time:           it.OrderTime.Format("2006-01-02 15:04:05"),
			ContractID:     it.ContractID,
			ContractName:   contract.FullName(),
			StockCode:      it.StockCode,
			StockName:      it.StockName,
			PositionPrice:  it.PositionPrice,
			PositionAmount: it.PositionAmount, // 持仓股数
			DividendAmount: dividendAmount,    // 转股比例
			DividendMoney:  dividendMoney,     // 现金分红比例
			IsBuyBack:      it.IsBuyBack,      // 是否零股回购
			BuyBackAmount:  it.BuyBackAmount,  // 零股回购数量
			BuyBackPrice:   it.BuyBackPrice,   // 零股回购价格
		})
	}

	// 下载则下发文件
	if IsDownload(c) {
		var res []interface{}
		for _, it := range list {
			res = append(res, it)
		}
		Download(c, []string{
			"ID", "用户名称", "姓名", "代理机构", "时间", "合约ID", "合约名称", "股票代码", "股票名称",
			"持仓价格", "持仓数量", "送股比例", "现金分红比例", "是否零股回购", "零股回购数量", "零股回购价格",
		}, res)
	}

	count := len(list)
	start, end := SlicePage(c, count)

	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil
}

// Position 股票交易-持仓
func (h *TradeHandle) Position(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		BeginDate int32 `form:"begin_date" json:"begin_date"`
		EndDate   int32 `form:"end_date" json:"end_date"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	tx := db.StockDB().WithContext(ctx).Table("position")
	tx.Where("uid in (?)", AgentFilter(c))
	ContractIDFilter(c, tx)
	UserNameFilter(c, tx)
	if req.BeginDate > 0 {
		tx.Where("order_time >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("order_time <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}

	var position []*model.Position
	if err := tx.Find(&position).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(position, func(i, j int) bool {
		return timeconv.TimeToInt64(position[i].OrderTime) > timeconv.TimeToInt64(position[j].OrderTime)
	})

	roleMap := RoleMap(ctx)
	userMap := UsersMap(ctx)
	contractMap := ContractMap(ctx)

	entrustIDs := make([]int64, 0)
	for _, it := range position {
		entrustIDs = append(entrustIDs, it.EntrustID)
	}
	entrustMap := getEntrustsByIDs(ctx, entrustIDs)

	// 获取股票行情现价
	codes := make([]string, 0)
	for _, it := range position {
		codes = append(codes, it.StockCode)
	}
	qt, err := quote.QtServiceInstance().GetQuoteByTencent(codes)
	if err != nil {
		return nil, err
	}

	list := make([]*model.TradePositionResp, 0)
	for _, it := range position {
		user, ok := userMap[it.UID]
		if !ok {
			continue
		}
		contract, ok := contractMap[it.ContractID]
		if !ok {
			contract = &model.Contract{}
		}

		entrust, ok := entrustMap[it.EntrustID]
		if !ok {
			entrust = &model.Entrust{}
		}

		list = append(list, &model.TradePositionResp{
			ID:            it.EntrustID,
			UserName:      user.UserName,
			Name:          user.Name,
			Agent:         roleMap[user.RoleID],
			Time:          it.OrderTime.Format("2006-01-02 15:04:05"),
			ContractID:    it.ContractID,
			ContractName:  contract.FullName(),
			StockCode:     it.StockCode,
			StockName:     it.StockName,
			PositionPrice: it.Price,
			CurrentPrice:  qt[it.StockCode].CurrentPrice,
			Amount:        it.Amount,
			FreezeAmount:  it.FreezeAmount,
			Profit:        util.FloatRound((qt[it.StockCode].CurrentPrice-it.Price)*float64(it.Amount), 4),
			ProfitPct:     fmt.Sprintf("%0.2f%%", 100*util.FloatRound((qt[it.StockCode].CurrentPrice-it.Price)/it.Price, 4)),
			Broker:        entrust.IsBrokerEntrust,
		})
	}

	// 下载则下发文件
	if IsDownload(c) {
		var res []interface{}
		for _, it := range list {
			res = append(res, it)
		}
		Download(c, []string{
			"ID", "用户名称", "姓名", "代理机构", "时间", "合约ID", "合约名称", "股票代码", "股票名称",
			"持仓价格", "当前价格", "持仓数量", "冻结数量", "盈亏金额", "盈亏比率", "对接券商",
		}, res)
	}

	count := len(list)
	start, end := SlicePage(c, count)
	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil
}

// Sell 股票交易-卖出交易
func (h *TradeHandle) Sell(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		BeginDate int32 `form:"begin_date" json:"begin_date"`
		EndDate   int32 `form:"end_date" json:"end_date"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	tx := db.StockDB().WithContext(ctx).Table("sell")
	tx.Where("uid in (?)", AgentFilter(c))
	ContractIDFilter(c, tx)
	UserNameFilter(c, tx)
	if req.BeginDate > 0 {
		tx.Where("order_time >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("order_time <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}

	var sell []*model.Sell
	if err := tx.Find(&sell).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(sell, func(i, j int) bool {
		return timeconv.TimeToInt64(sell[i].OrderTime) > timeconv.TimeToInt64(sell[j].OrderTime)
	})

	roleMap := RoleMap(ctx)
	userMap := UsersMap(ctx)
	contractMap := ContractMap(ctx)

	entrustIDs := make([]int64, 0)
	for _, it := range sell {
		entrustIDs = append(entrustIDs, it.EntrustID)
	}
	entrustMap := getEntrustsByIDs(ctx, entrustIDs)

	// 获取券商委托表ID map[entrust.id] []*model.BrokerEntrust
	brokerEntrustMap := getBrokerEntrustByEntrustIDs(ctx, entrustIDs)

	// 券商
	brokerIDs := make([]int64, 0)
	for _, v := range brokerEntrustMap {
		for _, it := range v {
			brokerIDs = append(brokerIDs, it.BrokerID)
		}
	}
	brokerMap := getBrokerByIDs(ctx, brokerIDs)

	list := make([]*model.TradeBuyResp, 0)
	for _, it := range sell {
		user, ok := userMap[it.UID]
		if !ok {
			continue
		}
		contract, ok := contractMap[it.ContractID]
		if !ok {
			contract = &model.Contract{}
		}
		typ := "限价"
		if it.EntrustProp == model.EntrustPropTypeMarketPrice {
			typ = "市价"
		}
		entrust, ok := entrustMap[it.EntrustID]
		if !ok {
			entrust = &model.Entrust{}
		}

		// 券商委托表
		brokerEntrusts, ok := brokerEntrustMap[entrust.ID]
		if !ok {
			brokerEntrusts = make([]*model.BrokerEntrust, 0)
		}

		// 券商
		brokerAccount := "" // 券商委托账号:平安证券-XSD123
		brokerOrderNo := "" // 券商委托编号: 平安证券-SXF101
		for _, it := range brokerEntrusts {
			broker, ok := brokerMap[it.BrokerID]
			if !ok {
				continue
			}
			brokerAccount += fmt.Sprintf("%s-%s;", broker.BrokerName, broker.FundAccount)
			brokerOrderNo += fmt.Sprintf("%s-%s;", broker.BrokerName, it.BrokerEntrustNo)
		}

		list = append(list, &model.TradeBuyResp{
			ID:            it.EntrustID,
			UserName:      user.UserName,
			Name:          user.Name,
			Agent:         roleMap[user.RoleID],
			Time:          it.OrderTime.Format("2006-01-02 15:04:05"),
			ContractID:    it.ContractID,
			ContractName:  contract.FullName(),
			StockCode:     it.StockCode,
			StockName:     it.StockName,
			Price:         it.Price,
			Amount:        it.Amount,
			Balance:       it.Balance,
			Profit:        it.Profit,
			ProfitPct:     "",
			Type:          typ,
			Fee:           it.Fee,
			Broker:        entrust.IsBrokerEntrust,
			BrokerAccount: brokerAccount,
			BrokerOrderNo: brokerOrderNo,
		})
	}

	// 下载则下发文件
	if IsDownload(c) {
		var res []interface{}
		for _, it := range list {
			res = append(res, it)
		}
		Download(c, []string{
			"ID", "用户名称", "姓名", "代理机构", "时间", "合约ID", "合约名称", "股票代码", "股票名称",
			"成交价格", "成交数量", "成交金额", "交易类型", "手续费", "对接券商", "对接券商账号", "券商委托序号",
		}, res)
	}

	count := len(list)
	start, end := SlicePage(c, count)

	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil
}

// Buy 股票交易-买入交易
func (h *TradeHandle) Buy(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		BeginDate int32 `form:"begin_date" json:"begin_date"`
		EndDate   int32 `form:"end_date" json:"end_date"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	tx := db.StockDB().WithContext(ctx).Table("buy")
	tx.Where("uid in (?)", AgentFilter(c))
	ContractIDFilter(c, tx)
	UserNameFilter(c, tx)
	if req.BeginDate > 0 {
		tx.Where("order_time >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("order_time <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}

	var buy []*model.Buy
	if err := tx.Find(&buy).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(buy, func(i, j int) bool {
		return timeconv.TimeToInt64(buy[i].OrderTime) > timeconv.TimeToInt64(buy[j].OrderTime)
	})

	roleMap := RoleMap(ctx)
	userMap := UsersMap(ctx)
	contractMap := ContractMap(ctx)

	entrustIDs := make([]int64, 0)
	for _, it := range buy {
		entrustIDs = append(entrustIDs, it.EntrustID)
	}
	entrustMap := getEntrustsByIDs(ctx, entrustIDs)

	// 获取券商委托表ID map[entrust.id] []*model.BrokerEntrust
	brokerEntrustMap := getBrokerEntrustByEntrustIDs(ctx, entrustIDs)

	// 券商
	brokerIDs := make([]int64, 0)
	for _, v := range brokerEntrustMap {
		for _, it := range v {
			brokerIDs = append(brokerIDs, it.BrokerID)
		}
	}
	brokerMap := getBrokerByIDs(ctx, brokerIDs)

	list := make([]*model.TradeBuyResp, 0)
	for _, it := range buy {
		user, ok := userMap[it.UID]
		if !ok {
			continue
		}
		contract, ok := contractMap[it.ContractID]
		if !ok {
			contract = &model.Contract{}
		}
		typ := "限价"
		if it.EntrustProp == model.EntrustPropTypeMarketPrice {
			typ = "市价"
		}
		entrust, ok := entrustMap[it.EntrustID]
		if !ok {
			entrust = &model.Entrust{}
		}

		// 券商委托表
		brokerEntrusts, ok := brokerEntrustMap[entrust.ID]
		if !ok {
			brokerEntrusts = make([]*model.BrokerEntrust, 0)
		}

		// 券商
		brokerAccount := "" // 券商委托账号:平安证券-XSD123
		brokerOrderNo := "" // 券商委托编号: 平安证券-SXF101
		for _, it := range brokerEntrusts {
			broker, ok := brokerMap[it.BrokerID]
			if !ok {
				continue
			}
			brokerAccount += fmt.Sprintf("%s-%s;", broker.BrokerName, broker.FundAccount)
			brokerOrderNo += fmt.Sprintf("%s-%s;", broker.BrokerName, it.BrokerEntrustNo)
		}

		list = append(list, &model.TradeBuyResp{
			ID:            it.EntrustID,
			UserName:      user.UserName,
			Name:          user.Name,
			Agent:         roleMap[user.RoleID],
			Time:          it.OrderTime.Format("2006-01-02 15:04:05"),
			ContractID:    it.ContractID,
			ContractName:  contract.FullName(),
			StockCode:     it.StockCode,
			StockName:     it.StockName,
			Price:         it.Price,
			Amount:        it.Amount,
			Balance:       it.Balance,
			Type:          typ,
			Fee:           it.Fee,
			Broker:        entrust.IsBrokerEntrust,
			BrokerAccount: brokerAccount,
			BrokerOrderNo: brokerOrderNo,
		})
	}

	// 下载则下发文件
	if IsDownload(c) {
		var res []interface{}
		for _, it := range list {
			res = append(res, it)
		}
		Download(c, []string{
			"ID", "用户名称", "姓名", "代理机构", "时间", "合约ID", "合约名称", "股票代码", "股票名称",
			"成交价格", "成交数量", "成交金额", "交易类型", "手续费", "对接券商", "对接券商账号", "券商委托序号",
		}, res)
	}

	count := len(list)
	start, end := SlicePage(c, count)
	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil
}

// getEntrustsByIDs map[entrust.id]*model.entrust
func getEntrustsByIDs(ctx context.Context, entrustIDs []int64) map[int64]*model.Entrust {
	result := make(map[int64]*model.Entrust)
	list, err := dao.EntrustDaoInstance().GetEntrustsByIDs(ctx, entrustIDs)
	if err != nil {
		return result
	}
	for _, it := range list {
		result[it.ID] = it
	}
	return result
}

// getBrokerEntrustByEntrustIDs map[entrust.id] []*model.BrokerEntrust
func getBrokerEntrustByEntrustIDs(ctx context.Context, ids []int64) map[int64][]*model.BrokerEntrust {
	result := make(map[int64][]*model.BrokerEntrust)
	list, err := dao.BrokerEntrustDaoInstance().GetByEntrustIDs(ctx, ids)
	if err != nil {
		return nil
	}
	for _, it := range list {
		v, ok := result[it.EntrustID]
		if !ok {
			v = make([]*model.BrokerEntrust, 0)
		}
		v = append(v, it)
		result[it.EntrustID] = v
	}
	return result
}

func getBrokerByIDs(ctx context.Context, ids []int64) map[int64]*model.Broker {
	result := make(map[int64]*model.Broker)
	list, err := dao.BrokerDaoInstance().GetBrokersByIDs(ctx, ids)
	if err != nil {
		return result
	}
	for _, it := range list {
		result[it.ID] = it
	}
	return result
}

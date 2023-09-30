package handler

import (
	"sort"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/service"
	"stock/api-gateway/util"
	"stock/common/log"
	"stock/common/timeconv"
	"time"

	"github.com/gin-gonic/gin"
)

// BrokerHandler 管理
type BrokerHandler struct {
}

// NewBrokerHandler 单例
func NewBrokerHandler() *BrokerHandler {
	return &BrokerHandler{}
}

// Register handler
func (h *BrokerHandler) Register(e *gin.Engine) {
	// 券商列表
	e.GET("/cms/broker/list", JSONWrapper(h.BrokerList))
	e.POST("/cms/broker/create", JSONWrapper(h.Create))
	e.GET("/cms/broker/position/get_by_id", JSONWrapper(h.GetByID))
	e.GET("/cms/broker/entrust", JSONWrapper(h.EntrustList))
	e.GET("/cms/broker/position", JSONWrapper(h.BrokerPosition))
	e.POST("/cms/broker/status", JSONWrapper(h.UpdateStatus))

}

// UpdateStatus 券商管理-更新状态
func (h *BrokerHandler) UpdateStatus(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		ID int64 `form:"id" json:"id"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	broker, err := dao.BrokerDaoInstance().GetBroker(ctx, req.ID)
	if err != nil {
		log.Errorf("GetBroker err:%+v", err)
		return nil, err
	}
	if broker.Status == model.BrokerStatusEnable {
		broker.Status = model.BrokerStatusDisabled
	} else {
		broker.Status = model.BrokerStatusEnable
	}
	log.Infof("broker:%+v", broker)
	if err := dao.BrokerDaoInstance().Create(ctx, broker); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// BrokerPosition 券商委托记录
func (h *BrokerHandler) BrokerPosition(c *gin.Context) (interface{}, error) {
	type request struct {
		BrokerID int64 `form:"id" json:"id"`
		Limit    int32 `form:"limit" json:"limit"`
		Offset   int32 `json:"offset" json:"offset"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	list := make([]*model.CmsBrokerPositionResp, 0)
	for _, broker := range service.BrokerServiceInstance().GetBrokers() {
		if req.BrokerID > 0 && broker.ID != req.BrokerID {
			continue
		}
		for _, position := range broker.BrokerPosition {
			list = append(list, &model.CmsBrokerPositionResp{
				BrokerID:      broker.ID,              // 券商编号
				BrokerName:    broker.BrokerName,      // 券商名称
				StockCode:     position.StockCode,     // 股票代码
				StockName:     position.StockName,     // 股票名称
				PositionPrice: position.PositionPrice, // 持仓价格
				CurrentPrice:  position.CurrentPrice,  // 当前价格
				Amount:        position.Amount,        // 持仓数量
				FreezeAmount:  position.FreezeAmount,  // 冻结股数
			})
		}
	}
	count := len(list)
	start, end := SlicePage(c, count)
	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil

}

// EntrustList 券商管理-委托记录
func (h *BrokerHandler) EntrustList(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		BrokerID  int64 `form:"id" json:"id"`
		BeginDate int32 `form:"begin_date" json:"begin_date"`
		EndDate   int32 `form:"end_date" json:"end_date"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	tx := db.StockDB().WithContext(ctx).Table("broker_entrust")
	if req.BrokerID > 0 {
		tx.Where("broker_id = ?", req.BrokerID)
	}
	if req.BeginDate > 0 {
		tx.Where("order_time >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("order_time <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}

	var brokerEntrust []*model.BrokerEntrust
	if err := tx.Find(&brokerEntrust).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(brokerEntrust, func(i, j int) bool {
		return timeconv.TimeToInt64(brokerEntrust[i].OrderTime) > timeconv.TimeToInt64(brokerEntrust[j].OrderTime)
	})

	// 券商
	brokerMap := make(map[int64]*model.Broker)
	brokers, err := dao.BrokerDaoInstance().GetBrokers(ctx)
	if err != nil {
		return nil, err
	}
	for _, it := range brokers {
		brokerMap[it.ID] = it
	}

	userMap := UsersMap(ctx)
	roleMap := RoleMap(ctx)

	list := make([]*model.CmsBrokerEntrustResp, 0)
	for _, it := range brokerEntrust {
		user, ok := userMap[it.UID]
		if !ok {
			continue
		}
		broker, ok := brokerMap[it.BrokerID]
		if !ok {
			broker = &model.Broker{}
		}
		entrustBs := "买入"
		if it.EntrustBs == model.EntrustBsTypeSell {
			entrustBs = "卖出"
		}
		prop := "限价"
		if it.EntrustProp == model.EntrustPropTypeMarketPrice {
			prop = "市价"
		}
		list = append(list, &model.CmsBrokerEntrustResp{
			ID:         it.BrokerID,                                // 券商编号
			BrokerName: broker.BrokerName,                          // 券商名称
			UserName:   user.UserName,                              // 用户账户
			Name:       user.Name,                                  // 用户姓名
			Agent:      roleMap[user.RoleID],                       // 代理机构
			Time:       it.OrderTime.Format("2006-01-02 15:04:05"), // 时间
			StockCode:  it.StockCode,                               // 股票代码
			StockName:  it.StockName,                               // 股票名称
			Price:      it.EntrustPrice,                            // 委托价格
			Amount:     it.EntrustAmount,                           // 委托数量
			DealAmount: it.DealAmount,                              // 成交数量
			Status:     model.EntrustStatusMap[it.Status],          // 状态
			Type:       entrustBs,                                  // 交易类型
			Prop:       prop,                                       // 委托类型
			EntrustNo:  it.BrokerEntrustNo,                         // 券商委托编号
		})
	}

	// 下载则下发文件
	if IsDownload(c) {
		var res []interface{}
		for _, it := range list {
			res = append(res, it)
		}
		Download(c, []string{
			"券商编号", "券商名称", "用户名称", "用户姓名", "代理机构", "委托时间", "股票代码", "股票名称", "委托价格", "委托数量",
			"成交数量", "状态", "交易类型", "委托类型", "券商委托编号",
		}, res)
	}

	count := len(list)
	start, end := SlicePage(c, count)

	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil
}

// GetByID 根据ID查询券商
func (h *BrokerHandler) GetByID(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	id, err := Int64(c, "id")
	if err != nil {
		return nil, err
	}
	broker, err := dao.BrokerDaoInstance().GetBroker(ctx, id)
	if err != nil {
		return nil, err
	}
	return &model.CmsBrokerResp{
		ID:              broker.ID,
		Priority:        broker.Priority,
		IP:              broker.IP,
		Port:            broker.Port,
		Name:            broker.BrokerName,
		Version:         broker.Version,
		BranchNo:        broker.BranchNo,
		Account:         broker.FundAccount,
		Password:        broker.TradePassword,
		CommPassword:    broker.TxPassword,
		SHHolderAccount: broker.SHHolderAccount,
		SZHolderAccount: broker.SZHolderAccount,
	}, nil
}

// Create 更新券商
func (h *BrokerHandler) Create(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	var req model.CmsBrokerResp
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	broker := &model.Broker{
		ID:              req.ID,
		IP:              req.IP,
		Port:            req.Port,
		Version:         "2.20", // 通达信默认版本
		BranchNo:        req.BranchNo,
		FundAccount:     req.Account,
		TradeAccount:    req.Account,
		TradePassword:   req.Password,
		TxPassword:      req.CommPassword,
		SHHolderAccount: req.SHHolderAccount,
		SZHolderAccount: req.SZHolderAccount,
		Priority:        req.Priority, // 顺序,数字越大,优先级越高
		Status:          1,            // 状态:1激活 2冻结
		BrokerName:      req.Name,     // 券商名称
		CreateTime:      time.Now(),   // 时间
	}
	if err := service.BrokerServiceInstance().Create(ctx, broker); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// BrokerList 券商列表
func (h *BrokerHandler) BrokerList(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	list := make([]*model.CmsBrokerResp, 0)
	brokers, err := dao.BrokerDaoInstance().GetBrokers(ctx)
	if err != nil {
		return nil, err
	}
	onlineBrokerMap := make(map[int64]*model.Broker)
	for _, it := range service.BrokerServiceInstance().GetBrokers() {
		if it.ClientID == 0 {
			continue
		}
		onlineBrokerMap[it.ID] = it
	}
	for _, it := range brokers {
		broker := &model.CmsBrokerResp{
			ID:              it.ID,
			Name:            it.BrokerName,
			IP:              it.IP,
			Port:            it.Port,
			Account:         it.FundAccount,
			Password:        it.TradePassword,
			CommPassword:    it.TxPassword,
			Priority:        it.Priority,
			BranchNo:        it.BranchNo,
			SHHolderAccount: it.SHHolderAccount,
			SZHolderAccount: it.SZHolderAccount,
			Status:          it.Status == 1,
		}
		if onlineBroker, ok := onlineBrokerMap[it.ID]; ok {
			broker.ValMoney = onlineBroker.ValMoney
			broker.Asset = onlineBroker.Asset
			broker.Version = onlineBroker.Version
		}

		list = append(list, broker)
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

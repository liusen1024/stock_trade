package service

import (
	"context"
	"errors"
	"sort"
	"stock/api-gateway/dao"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"
	"sync"
	"time"
)

// BrokerService 券商服务
type BrokerService struct {
	brokerMap map[int64]*model.Broker
}

var (
	brokerService *BrokerService
	brokerOnce    sync.Once
	mutex         sync.Mutex
	brokerMutex   sync.Mutex
)

// BrokerServiceInstance 实例
func BrokerServiceInstance() *BrokerService {
	brokerOnce.Do(func() {
		brokerService = &BrokerService{
			brokerMap: make(map[int64]*model.Broker),
		}
		ctx := context.Background()
		// 等待券商通道连接:查询数据库配置的券商,未连接的自动连接
		if err := brokerService.clientConn(ctx); err != nil {
			log.Errorf("clientConn err:%+v", err)
		}

		go func() {
			// 从数据库读取券商信息,调用客户端读取券商数据&更新券商信息到brokers
			for range time.Tick(2 * time.Minute) {
				mutex.Lock()
				if err := brokerService.clientConn(ctx); err != nil {
					log.Errorf("clientConn err:%+v", err)
					// 发送短信提醒
				}
				mutex.Unlock()
			}
		}()
		go func() {
			// 通过通道查询是否成交
			for range time.Tick(1 * time.Second) {
				mutex.Lock()
				if err := brokerService.query(ctx); err != nil {
					log.Errorf("query err:%+v", err)
				}
				mutex.Unlock()
			}
		}()
	})
	return brokerService
}

func (s *BrokerService) GetBrokers() []*model.Broker {
	brokers := make([]*model.Broker, 0)
	brokerMutex.Lock()
	for _, v := range s.brokerMap {
		brokers = append(brokers, v)
	}
	brokerMutex.Unlock()
	sort.SliceStable(brokers, func(i, j int) bool {
		return brokers[i].Priority > brokers[j].Priority
	})
	return brokers
}

// query 查询订单是否成交
func (s *BrokerService) query(ctx context.Context) error {
	brokers := s.GetBrokers()
	if len(brokers) == 0 {
		return nil
	}
	brokerTdxEntrustMap := make(map[int64][]*model.TDXTodayEntrust, 0)
	for _, broker := range brokers {
		if broker.ClientID == 0 {
			continue
		}
		// 1. 查询资金
		if err := s.queryFund(broker); err != nil {
			log.Errorf("queryFund err:%+v", err)
			s.disConnect(broker)
			return err
		}

		// 2. 查询持仓
		if err := s.queryPosition(broker); err != nil {
			log.Errorf("资金账号:%+v 查询持仓失败:%+v", broker.FundAccount, err)
			s.disConnect(broker)
			return err
		}

		// 3.查询成交
		tdxEntrusts, err := TDXServiceInstance().QueryTodayEntrust(broker)
		if err != nil {
			log.Errorf("资金账号:%+v 查询今日委托失败:%+v", broker.FundAccount, err)
			s.disConnect(broker)
			return err
		}
		if len(tdxEntrusts) != 0 {
			brokerTdxEntrustMap[broker.ID] = tdxEntrusts
		}
	}

	// 检查券商是否成交|撤单
	if err := TradeServiceInstance().BrokerTrade(ctx, brokerTdxEntrustMap); err != nil {
		log.Errorf("券商成交失败,BrokerTrade err:%+v", err)
		return err
	}
	return nil
}

// queryFund 查询资金
func (s *BrokerService) queryFund(broker *model.Broker) error {
	fund, err := TDXServiceInstance().QueryFund(broker)
	if err != nil {
		log.Errorf("QueryFund err:%+v", err)
		return err
	}
	broker.ValMoney = fund.ValMoney
	broker.Asset = fund.Asset
	return nil
}

// queryPosition 查询持仓
func (s *BrokerService) queryPosition(broker *model.Broker) error {
	positions, err := TDXServiceInstance().QueryPosition(broker)
	if err != nil {
		log.Errorf("资金账号:%+v 查询持仓失败:%+v", broker.FundAccount, err)
		return err
	}
	brokerPosition := make([]*model.BrokerPosition, 0)
	for _, it := range positions {
		brokerPosition = append(brokerPosition, &model.BrokerPosition{
			StockCode:     it.StockCode,     // 股票代码
			StockName:     it.StockName,     // 股票名称
			Amount:        it.Amount,        // 总数量
			FreezeAmount:  it.FreezeAmount,  // 冻结数量
			PositionPrice: it.PositionPrice, // 持仓价格
			CurrentPrice:  it.CurrentPrice,  // 当前价格
		})
	}
	broker.BrokerPosition = brokerPosition
	return nil
}

// disConnect 处理断开连接的券商，优化:连续2次查询失败则断开连接
func (s *BrokerService) disConnect(broker *model.Broker) {
	broker.IoTimes++
	brokerMutex.Lock()
	if broker.IoTimes > 3 {
		delete(s.brokerMap, broker.ID)
	} else {
		s.brokerMap[broker.ID] = broker
	}
	brokerMutex.Unlock()
}

// clientConn 客户端连接
func (s *BrokerService) clientConn(ctx context.Context) error {
	list, err := dao.BrokerDaoInstance().GetBrokers(ctx)
	if err != nil {
		log.Errorf("GetBrokers err:%+v", err)
		return err
	}
	if len(list) == 0 {
		return nil
	}

	conn := make(map[int64]*model.Broker)
	for _, broker := range list {
		// 过滤掉不生效的券商
		if broker.Status != model.BrokerStatusEnable {
			continue
		}
		if _, ok := s.brokerMap[broker.ID]; !ok {
			conn[broker.ID] = broker
		}
	}
	if len(conn) == 0 {
		return nil
	}

	for _, broker := range conn {
		clientID, err := TDXServiceInstance().Login(broker)
		if err != nil {
			log.Errorf("连接券商:%+v 失败:%+v", broker, err)
			continue
		}
		broker.ClientID = clientID

		brokerMutex.Lock()
		s.brokerMap[broker.ID] = broker
		brokerMutex.Unlock()
		log.Infof("券商:%+v 资金账号:%+v 连接成功!client_id:%+v", broker.BrokerName, broker.FundAccount, broker.ClientID)
	}

	return nil
}

// getMatchBroker 查询匹配的券商
func (s *BrokerService) getMatchBroker(e *model.Entrust) ([]*model.BrokerEntrust, error) {
	// 买入:找到匹配的券商
	buy := func() ([]*model.BrokerEntrust, error) {
		for _, broker := range s.GetBrokers() {
			// 加上手续费,是为了找出更合适的
			if broker.ValMoney > e.Balance+e.Fee {
				return []*model.BrokerEntrust{
					{
						UID:             e.UID,                           // 用户ID
						ContractID:      e.ContractID,                    // 合约编号
						BrokerID:        broker.ID,                       // 券商ID
						EntrustID:       e.ID,                            // 委托表ID
						OrderTime:       e.OrderTime,                     // 订单时间
						StockCode:       e.StockCode,                     // 股票代码
						StockName:       e.StockName,                     // 股票名称
						EntrustAmount:   e.Amount,                        // 委托总股数
						EntrustPrice:    e.Price,                         // 委托价格
						EntrustBalance:  e.Balance,                       // 委托总金额
						DealAmount:      0,                               // 成交数量
						DealPrice:       0,                               // 成交价格
						DealBalance:     0,                               // 成交总金额
						Status:          model.EntrustStatusTypeReported, // 订单状态:已申报
						EntrustBs:       e.EntrustBS,                     // 交易类型:1买入 2卖出
						EntrustProp:     e.EntrustProp,                   // 委托类型:1限价 2市价
						Fee:             e.Fee,                           // 券商交易总手续费
						BrokerEntrustNo: "",                              // 券商委托编号
						Broker:          broker,                          // 券商
					},
				}, nil
			}
		}

		// 多母账户分笔订单买入
		brokerEntrusts := make([]*model.BrokerEntrust, 0)
		var entrustedAmount int64
		for _, broker := range s.GetBrokers() {
			amount := (int64((broker.ValMoney-e.Fee)/e.Price) / 100) * 100
			if amount < 100 {
				continue
			}
			if amount >= e.Amount-entrustedAmount {
				amount = e.Amount - entrustedAmount
			}
			entrustedAmount += amount
			brokerEntrusts = append(brokerEntrusts, &model.BrokerEntrust{
				UID:             e.UID,                           // 用户ID
				ContractID:      e.ContractID,                    // 合约编号
				BrokerID:        broker.ID,                       // 券商ID
				EntrustID:       e.ID,                            // 委托表ID
				OrderTime:       e.OrderTime,                     // 订单时间
				StockCode:       e.StockCode,                     // 股票代码
				StockName:       e.StockName,                     // 股票名称
				EntrustAmount:   amount,                          // 委托总股数
				EntrustPrice:    e.Price,                         // 委托价格
				EntrustBalance:  e.Price * float64(amount),       // 委托总金额
				DealAmount:      0,                               // 成交数量
				DealPrice:       0,                               // 成交价格
				DealBalance:     0,                               // 成交总金额
				Status:          model.EntrustStatusTypeReported, // 订单状态:已申报
				EntrustBs:       e.EntrustBS,                     // 交易类型:1买入 2卖出
				EntrustProp:     e.EntrustProp,                   // 委托类型:1限价 2市价
				Fee:             0,                               // 券商交易总手续费(分笔买入券商手续费不准，填写0)
				BrokerEntrustNo: "",                              // 券商委托编号
				Broker:          broker,                          // 券商
			})
			if entrustedAmount == e.Amount {
				return brokerEntrusts, nil
			}
		}
		return nil, errors.New("无效券商通道或母账户资金不足")
	}

	// 卖出:找到持仓该股票的券商
	sell := func() ([]*model.BrokerEntrust, error) {
		var entrustedAmount int64
		brokerEntrusts := make([]*model.BrokerEntrust, 0)
		oddAmount := e.Amount % 100 // 零股
		for _, broker := range s.GetBrokers() {
			for _, it := range broker.BrokerPosition {
				if it.StockCode != e.StockCode {
					continue
				}
				// 零股卖出处理
				valAmount := it.Amount - it.FreezeAmount
				if valAmount <= 0 {
					continue
				}
				amount := int64((valAmount)/100) * 100
				if oddAmount > 0 && valAmount%100 == oddAmount {
					amount += oddAmount
				}

				if amount >= e.Amount-entrustedAmount {
					amount = e.Amount - entrustedAmount
				}
				entrustedAmount += amount
				brokerEntrusts = append(brokerEntrusts, &model.BrokerEntrust{
					UID:             e.UID,                           // 用户ID
					ContractID:      e.ContractID,                    // 合约编号
					BrokerID:        broker.ID,                       // 券商ID
					EntrustID:       e.ID,                            // 委托表ID
					OrderTime:       e.OrderTime,                     // 订单时间
					StockCode:       e.StockCode,                     // 股票代码
					StockName:       e.StockName,                     // 股票名称
					EntrustAmount:   amount,                          // 委托总股数
					EntrustPrice:    e.Price,                         // 委托价格
					EntrustBalance:  e.Price * float64(amount),       // 委托总金额
					DealAmount:      0,                               // 成交数量
					DealPrice:       0,                               // 成交价格
					DealBalance:     0,                               // 成交总金额
					Status:          model.EntrustStatusTypeReported, // 订单状态：已申报
					EntrustBs:       e.EntrustBS,                     // 交易类型:1买入 2卖出
					EntrustProp:     e.EntrustProp,                   // 委托类型:1限价 2市价
					Fee:             e.Fee,                           // 券商交易总手续费
					BrokerEntrustNo: "",                              // 券商委托编号
					Broker:          broker,                          // 券商客户ID
				})
				if entrustedAmount == e.Amount {
					return brokerEntrusts, nil
				}
			}
		}
		if entrustedAmount < e.Amount {
			log.Infof("股票代码:%+v 母账户可委托卖出数量:%+v", e.StockCode, entrustedAmount)
			return nil, errors.New("母账户持有该股票可卖股份不足")
		}
		return brokerEntrusts, nil
	}

	switch e.EntrustBS {
	case model.EntrustBsTypeBuy:
		return buy()
	case model.EntrustBsTypeSell:
		return sell()
	}
	return nil, serr.ErrBusiness("无效券商通道")
}

// entrust 调用券商通道申报订单
func (s *BrokerService) entrust(entrust *model.Entrust) ([]*model.BrokerEntrust, error) {
	brokers := s.GetBrokers()
	if len(brokers) == 0 {
		return nil, errors.New("未连接券商通道")
	}

	// 买入,卖出寻找合适的券商
	brokerEntrusts, err := s.getMatchBroker(entrust)
	if err != nil {
		log.Errorf("getMatchBroker err:%+v", err)
		return nil, err
	}

	// 逐笔委托
	for _, brokerEntrust := range brokerEntrusts {
		if err := TDXServiceInstance().Entrust(brokerEntrust); err != nil {
			log.Errorf("通达信委托失败:%+v", err)
			return nil, err
		}
	}

	return brokerEntrusts, nil
}

// Withdraw 撤单委托申请
func (s *BrokerService) Withdraw(entrust *model.BrokerEntrust, broker *model.Broker, entrustNo string) error {
	mutex.Lock()
	defer mutex.Unlock()
	return TDXServiceInstance().CancelOrder(entrust, broker, entrustNo)
}

// Entrust 券商委托申报
func (s *BrokerService) Entrust(entrust *model.Entrust) error {
	ctx := context.Background()
	if !entrust.IsBrokerEntrust {
		return nil
	}

	mutex.Lock()
	defer mutex.Unlock()
	// 券商通道正常,则进行委托撤单
	brokerEntrusts, err := s.entrust(entrust)
	if err != nil {
		// 券商委托失败,则废单处理
		log.Errorf("订单委托失败:%+v,err:%+v", entrust, err)
		return s.cancelEntrust(ctx, entrust, err.Error())
	}

	// 更新委托表
	entrust.Status = model.EntrustStatusTypeReported // 委托状态:已申报,未成交
	if err := dao.EntrustDaoInstance().Update(ctx, entrust); err != nil {
		log.Errorf("订单申报填写委托表失败 err:%+v", err)
		return err
	}

	// 创建券商委托表
	return dao.BrokerEntrustDaoInstance().MCreate(ctx, brokerEntrusts)
}

// cancelEntrust 券商委托失败，委托作废
func (s *BrokerService) cancelEntrust(ctx context.Context, entrust *model.Entrust, cancelReason string) error {
	entrust.Status = model.EntrustStatusTypeCancel
	entrust.Remark = cancelReason
	if err := dao.EntrustDaoInstance().Update(ctx, entrust); err != nil {
		return err
	}

	// 废单卖出更新冻结
	if entrust.EntrustBS == model.EntrustBsTypeSell {
		position, err := dao.PositionDaoInstance().GetPositionByID(ctx, entrust.PositionID)
		if err != nil {
			log.Errorf("卖出废单,查询持仓失败;GetPositionByID err:%+v", err)
			return err
		}
		position.FreezeAmount -= entrust.Amount
		if err := dao.PositionDaoInstance().Update(ctx, position); err != nil {
			log.Errorf("更新持仓失败:%+v", err)
			return err
		}
	}

	// 更新可用资金
	if err := ContractServiceInstance().UpdateValMoneyByID(ctx, entrust.ContractID); err != nil {
		return err
	}
	return nil
}

// Create 创建券商
func (s *BrokerService) Create(ctx context.Context, broker *model.Broker) error {
	if err := dao.BrokerDaoInstance().Create(ctx, broker); err != nil {
		return err
	}
	s.disConnect(broker)
	return nil
}

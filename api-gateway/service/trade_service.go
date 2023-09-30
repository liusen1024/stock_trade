package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/quote"
	"stock/api-gateway/serr"
	"stock/api-gateway/util"
	"stock/common/errgroup"
	"stock/common/log"
	"stock/common/timeconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// TradeService 内容服务
type TradeService struct {
}

var (
	tradeService *TradeService
	tradeOnce    sync.Once
)

// TradeServiceInstance TradeServiceInstance实例
func TradeServiceInstance() *TradeService {
	tradeOnce.Do(func() {
		tradeService = &TradeService{}
		ctx := context.Background()

		// 非券商委托状态下,检测交易自动成交
		go func() {
			for range time.Tick(2 * time.Second) {
				if err := tradeService.autoTrade(ctx); err != nil {
					log.Errorf("autoTrade err:%+v", err)
					continue
				}
			}
		}()

	})
	return tradeService
}

// genEntrust 生成成交委托
func (s *TradeService) genEntrust(ctx context.Context, entrust *model.Entrust, brokerEntrusts []*model.BrokerEntrust) *model.Entrust {
	var dealAmount int64
	dealPrice := entrust.Price
	for _, brokerEntrust := range brokerEntrusts {
		dealAmount += brokerEntrust.DealAmount
		// 成交价格小于0.01的则不作为最终成交价格
		if brokerEntrust.DealPrice < 0.01 {
			continue
		}
		// 价格纠正,取成交价格最有利
		if entrust.EntrustBS == model.EntrustBsTypeBuy && dealPrice < brokerEntrust.DealPrice {
			dealPrice = brokerEntrust.DealPrice
		}
		if entrust.EntrustBS == model.EntrustBsTypeSell && dealPrice > brokerEntrust.DealPrice {
			dealPrice = brokerEntrust.DealPrice
		}
	}

	if dealAmount == entrust.Amount {
		// 全部成交
		entrust.DealAmount = dealAmount
		entrust.Price = dealPrice
		entrust.Status = model.EntrustStatusTypeDeal
		entrust.Balance = float64(entrust.DealAmount) * entrust.Price
		if fee, err := s.GetTradeFee(ctx, entrust.Price, entrust.DealAmount, entrust.EntrustBS); err == nil {
			entrust.Fee = fee
		}
	} else if dealAmount > 0 {
		// 部撤
		entrust.DealAmount = dealAmount
		entrust.Price = dealPrice
		entrust.Status = model.EntrustStatusTypePartDealPartWithdraw
		entrust.Balance = float64(entrust.DealAmount) * entrust.Price
		if fee, err := s.GetTradeFee(ctx, entrust.Price, entrust.DealAmount, entrust.EntrustBS); err == nil {
			entrust.Fee = fee
		}
	} else {
		// 废单:有一个母账户订单是废单,则该笔委托则是废单委托
		entrust.Status = model.EntrustStatusTypeWithdraw

		for _, it := range brokerEntrusts {
			if it.Status == model.EntrustStatusTypeCancel {
				entrust.Status = model.EntrustStatusTypeCancel
			}
		}
	}
	entrust.BrokerEntrust = brokerEntrusts
	return entrust
}

// process 处理状态
func (s *TradeService) isDeal(entrust *model.Entrust, brokerEntrustsList []*model.BrokerEntrust, brokerTdxEntrustMap map[int64][]*model.TDXTodayEntrust) ([]*model.BrokerEntrust, bool) {
	if len(brokerEntrustsList) == 0 || len(brokerTdxEntrustMap) == 0 {
		return make([]*model.BrokerEntrust, 0), false
	}
	// 找到券商委托记录
	brokerEntrusts := make([]*model.BrokerEntrust, 0)
	size := 0
	for _, brokerEntrust := range brokerEntrustsList {
		if entrust.ID != brokerEntrust.EntrustID {
			continue
		}
		size++
		tdxEntrusts, ok := brokerTdxEntrustMap[brokerEntrust.BrokerID]
		if !ok {
			continue
		}
		for _, tdx := range tdxEntrusts {
			if tdx.EntrustNo != brokerEntrust.BrokerEntrustNo {
				continue
			}
			switch tdx.Status {
			case "已成":
				brokerEntrust.Status = model.EntrustStatusTypeDeal
				brokerEntrust.DealAmount = tdx.DealAmount
				brokerEntrust.DealPrice = tdx.DealPrice
				brokerEntrust.DealBalance = float64(brokerEntrust.DealAmount) * brokerEntrust.DealPrice
			case "部撤":
				brokerEntrust.Status = model.EntrustStatusTypePartDealPartWithdraw
				brokerEntrust.DealAmount = tdx.DealAmount
				brokerEntrust.DealPrice = tdx.DealPrice
				brokerEntrust.DealBalance = float64(brokerEntrust.DealAmount) * brokerEntrust.DealPrice
			case "已撤":
				brokerEntrust.Status = model.EntrustStatusTypeWithdraw
			case "废单":
				brokerEntrust.Status = model.EntrustStatusTypeCancel
			default:
				// 其他状态则表示非终态
				return make([]*model.BrokerEntrust, 0), false
			}
			brokerEntrusts = append(brokerEntrusts, brokerEntrust)
		}
	}

	if size != len(brokerEntrusts) {
		return make([]*model.BrokerEntrust, 0), false
	}
	return brokerEntrusts, true
}

// BrokerTrade 券商成交 brokerTdxEntrustMap[broker.id][]*model.TDXTodayEntrust
func (s *TradeService) BrokerTrade(ctx context.Context, brokerTdxEntrustMap map[int64][]*model.TDXTodayEntrust) error {
	if len(brokerTdxEntrustMap) == 0 {
		return nil
	}
	entrusts, err := dao.EntrustDaoInstance().GetTodayEntrusts(ctx)
	if err != nil {
		return err
	}
	brokerEntrustsList, err := dao.BrokerEntrustDaoInstance().GetTodayEntrusts(ctx)
	if err != nil {
		return err
	}
	for _, entrust := range entrusts {
		// 非券商委托不处理
		if !entrust.IsBrokerEntrust {
			continue
		}
		// 终态不处理
		if entrust.IsFinallyState() {
			continue
		}
		brokerEntrusts, ok := s.isDeal(entrust, brokerEntrustsList, brokerTdxEntrustMap)
		if !ok {
			continue
		}
		e := s.genEntrust(ctx, entrust, brokerEntrusts)
		if len(e.BrokerEntrust) == 0 {
			continue
		}
		log.Infof("委托订单[entrust]:%+v", e)
		for _, brokerEntrust := range e.BrokerEntrust {
			log.Infof("券商委托订单[broker_entrust]:%+v", brokerEntrust)
		}
		switch e.Status {
		case model.EntrustStatusTypeDeal:
			{
				// 已成
				if err := s.brokerEntrustDeal(ctx, e); err != nil {
					log.Errorf("brokerEntrustDeal err:%+v", err)
					return err
				}
				log.Infof("已成:%+v", e)
			}
		case model.EntrustStatusTypePartDealPartWithdraw:
			{
				// 部撤
				if err := s.brokerEntrustDeal(ctx, e); err != nil {
					log.Errorf("brokerEntrustDeal err:%+v", err)
					return err
				}
				log.Infof("部撤:%+v", e)
			}
		case model.EntrustStatusTypeWithdraw:
			{
				// 已撤
				if err := s.brokerEntrustWithdraw(ctx, e); err != nil {
					log.Errorf("brokerEntrustWithdraw err:%+v", err)
					return err
				}
				log.Infof("撤单:%+v", e)
			}
		case model.EntrustStatusTypeCancel:
			{
				// 废单
				if err := s.brokerCancelOrder(ctx, e); err != nil {
					log.Errorf("brokerCancelOrder err:%+v", err)
					return err
				}
				log.Infof("废单:%+v", e)
			}
		}
	}
	return nil
}

// brokerEntrustDeal 券商委托成交
func (s *TradeService) brokerEntrustDeal(ctx context.Context, entrust *model.Entrust) error {
	// 幂等:防止重复提交订单
	key := fmt.Sprintf("broker_entrust_deal_entrust_id_%+v", entrust.ID)
	if db.RedisClient().Exists(ctx, key).Val() == 1 {
		log.Errorf("订单异常:%+v 正在处理该笔订单!", entrust)
		return nil
	}
	db.RedisClient().SetNX(ctx, key, "1", 1*time.Minute)
	defer db.RedisClient().Del(ctx, key)

	// 买入成交
	if entrust.EntrustBS == model.EntrustBsTypeBuy {
		if err := BuyServiceInstance().CreateOrder(ctx, entrust); err != nil {
			log.Errorf("买入成交订单处理失败:%+v", err)
			return err
		}
	}
	// 卖出成交
	if entrust.EntrustBS == model.EntrustBsTypeSell {
		if err := SellServiceInstance().CreateOrder(ctx, entrust); err != nil {
			log.Errorf("卖出成交订单处理失败:%+v", err)
		}
	}

	return nil
}

// brokerEntrustWithdraw 券商委托撤单
func (s *TradeService) brokerEntrustWithdraw(ctx context.Context, entrust *model.Entrust) error {
	// 幂等:防止重复提交订单
	key := fmt.Sprintf("broker_entrust_withdraw_entrust_id_%+v", entrust.ID)
	if db.RedisClient().Exists(ctx, key).Val() == 1 {
		log.Errorf("订单异常:%+v 正在处理该笔订单!", entrust)
		return nil
	}
	db.RedisClient().SetNX(ctx, key, "1", 1*time.Minute)
	defer db.RedisClient().Del(ctx, key)

	tx := db.StockDB().WithContext(ctx).Begin()
	defer tx.Rollback()

	// 券商委托表设置撤单状态
	if entrust.IsBrokerEntrust && len(entrust.BrokerEntrust) > 0 {
		if err := dao.BrokerEntrustDaoInstance().MCreateWithTx(tx, entrust.BrokerEntrust); err != nil {
			log.Errorf("券商委托表更新失败:%+v", err)
			return err
		}
		log.Infof("撤单业务:委托编号:%+v [broker_entrust]更新券商委托表:%+v", entrust.ID, entrust.BrokerEntrust)
	}

	if err := dao.EntrustDaoInstance().UpdateStatusWithTx(tx, entrust); err != nil {
		log.Errorf("Update err:%+v", err)
		return err
	}
	log.Infof("撤单业务:委托编号:%+v [entrust]更新委托表:%+v", entrust.ID, entrust)

	// 卖出撤单,解冻股票
	if entrust.EntrustBS == model.EntrustBsTypeSell {
		if err := dao.PositionDaoInstance().UnFreezeAmountWithTx(tx, entrust.ContractID, entrust.StockCode, entrust.Amount); err != nil {
			log.Errorf("解冻股票失败:%+v", err)
			return serr.ErrBusiness("撤单失败")
		}
		log.Infof("撤单业务:委托编号:%+v [position]解除冻结股票,解除冻结数量:%+v", entrust.ID, entrust.Amount)
	}

	if err := tx.Commit().Error; err != nil {
		log.Errorf("撤单业务:委托编号:%+v 提交失败:%+v", entrust.ID, err)
		return err
	}
	log.Infof("撤单业务:委托编号:%+v 提交成功!", entrust.ID)

	if err := ContractServiceInstance().UpdateValMoneyByID(ctx, entrust.ContractID); err != nil {
		log.Errorf("UpdateValMoneyByID err:%+v", err)
	}
	return nil
}

// brokerCancelOrder 券商废单
func (s *TradeService) brokerCancelOrder(ctx context.Context, entrust *model.Entrust) error {
	// 幂等:防止重复提交订单
	key := fmt.Sprintf("broker_entrust_cancel_order_entrust_id_%+v", entrust.ID)
	if db.RedisClient().Exists(ctx, key).Val() == 1 {
		log.Errorf("订单异常:%+v 正在处理该笔订单!", entrust)
		return nil
	}
	db.RedisClient().SetNX(ctx, key, "1", 1*time.Minute)
	defer db.RedisClient().Del(ctx, key)

	tx := db.StockDB().WithContext(ctx).Begin()
	defer tx.Rollback()
	// 更新委托状态
	if err := dao.EntrustDaoInstance().UpdateWithTx(tx, entrust); err != nil {
		log.Errorf("UpdateWithTx err:%+v", err)
		return err
	}
	// 设置券商委托表
	if err := dao.BrokerEntrustDaoInstance().MCreateWithTx(tx, entrust.BrokerEntrust); err != nil {
		log.Errorf("Update err:%+v", err)
		return err
	}

	// 卖出废单,解除冻结股数
	if entrust.EntrustBS == model.EntrustBsTypeSell {
		if err := dao.PositionDaoInstance().UnFreezeAmountWithTx(tx, entrust.ContractID, entrust.StockCode, entrust.Amount); err != nil {
			log.Errorf("解冻股票失败:%+v", err)
			return serr.ErrBusiness("撤单失败")
		}
		log.Infof("撤单业务:委托编号:%+v [position]解除冻结股票,解除冻结数量:%+v", entrust.ID, entrust.Amount)
	}

	if err := tx.Commit().Error; err != nil {
		log.Errorf("撤单业务:委托编号:%+v 提交失败:%+v", entrust.ID, err)
		return err
	}

	// 更新可用资金
	if err := ContractServiceInstance().UpdateValMoneyByID(ctx, entrust.ContractID); err != nil {
		log.Errorf("UpdateValMoneyByID err:%+v", err)
	}
	return nil
}

// autoTrade 自动成交
func (s *TradeService) autoTrade(ctx context.Context) error {
	// 是否交易时间
	if !CalendarServiceInstance().IsTradeTime(ctx) {
		return nil
	}
	// 查询今日委托记录
	entrusts, err := dao.EntrustDaoInstance().GetTodayEntrusts(ctx)
	if err != nil {
		return err
	}
	if len(entrusts) == 0 {
		return nil
	}
	codes := make([]string, 0)
	for _, it := range entrusts {
		if it.Status == model.EntrustStatusTypeUnDeal && !it.IsBrokerEntrust {
			codes = append(codes, it.StockCode)
		}
	}
	if len(codes) == 0 {
		return nil
	}
	qts, err := quote.QtServiceInstance().GetQuoteByTencent(codes)
	if err != nil {
		return err
	}

	for _, entrust := range entrusts {
		// 券商委托,部分成交状态则不进行自动成交
		if entrust.Status != model.EntrustStatusTypeUnDeal || entrust.IsBrokerEntrust {
			continue
		}
		qt, ok := qts[entrust.StockCode]
		if !ok {
			continue
		}
		// 买入成交
		if entrust.EntrustBS == model.EntrustBsTypeBuy && qt.CurrentPrice <= entrust.Price {
			entrust.DealAmount = entrust.Amount // 全部成交
			entrust.Status = model.EntrustStatusTypeDeal
			if err := BuyServiceInstance().CreateOrder(ctx, entrust); err != nil {
				log.Errorf("买入成交订单处理失败:%+v", err)
				return err
			}
		}

		// 卖出成交
		if entrust.EntrustBS == model.EntrustBsTypeSell && qt.CurrentPrice >= entrust.Price {
			entrust.DealAmount = entrust.Amount // 全部成交
			entrust.Status = model.EntrustStatusTypeDeal
			if err := SellServiceInstance().CreateOrder(ctx, entrust); err != nil {
				log.Errorf("卖出成交订单处理失败:%+v", err)
			}
		}
	}
	return nil
}

func (s *TradeService) InitTrade(ctx context.Context, uid, contractID int64) (*model.InitTrade, error) {
	// 查询证券账户是否存在
	// 不存在则查找已选择的账户是否存在
	// 已经选择的账户为空，则查询该用户所有可用账户,
	// 所有账户都为空，则返回account=401 告诉用户去申请新账户
	if contractID == 0 {
		// 查询用户最近使用的contractID
		contract, err := s.getValidContract(ctx, uid)
		if err != nil {
			return nil, err
		}
		contractID = contract.ID
	}

	// 查询持仓
	positions := make([]*model.Position, 0)
	wg := errgroup.GroupWithCount(4)
	wg.Go(func() error {
		ret, err := PositionServiceInstance().GetPositionByContractID(ctx, contractID)
		if err != nil {
			log.Error("GetPositionByContractID error:%+v", err)
			return err
		}
		positions = ret
		return nil
	})
	contract := &model.Contract{}
	wg.Go(func() error {
		ret, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
		if err != nil {
			return err
		}
		contract = ret
		return nil
	})

	// 上个交易日的持仓
	lastDayPosition := make([]*model.Position, 0)
	wg.Go(func() error {
		ret, err := dao.HisPositionDaoInstance().GetYesterdayPositionByContractID(ctx, contractID)
		if err != nil {
			return err
		}
		lastDayPosition = ret
		return nil
	})
	// 今日盈亏
	var todayProfit float64
	wg.Go(func() error {
		ret, err := ContractServiceInstance().GetTodayProfitByContractID(ctx, contractID)
		if err != nil {
			log.Errorf("GetTodayProfitByContractID err:%+v", err)
			return err
		}
		todayProfit = ret
		return nil
	})
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	if contract.UID != uid {
		return nil, serr.New(serr.ErrCodeContractNoFound, "请申请合约")
	}
	// 昨日总资产:昨日股票市值(通过his_position来查询出来) + 昨日可用资金
	lastDayMarketValue := 0.00
	if len(lastDayPosition) != 0 {
		codes := make([]string, 0, len(lastDayPosition))
		for _, it := range lastDayPosition {
			codes = append(codes, it.StockCode)
		}
		qts, err := quote.QtServiceInstance().GetQuoteByTencent(codes)
		if err != nil {
			return nil, err
		}
		for _, it := range lastDayPosition {
			lastDayMarketValue += qts[it.StockCode].ClosePrice * float64(it.Amount)
		}
	}

	lmv := lastDayMarketValue + contract.ValMoney
	if lmv < 0.001 {
		for _, it := range positions {
			lmv += it.Balance
		}
	}

	return &model.InitTrade{
		ContractName:   contract.FullName(),
		ContractID:     contractID,
		TotalAssets:    util.FloatRound(contract.ValMoney+model.CalculatePositionMarketValue(positions), 2), // 总资产 = 可用资金 + 股票市值
		ValMoney:       contract.ValMoney,
		Margin:         contract.Money,
		TotalProfit:    util.FloatRound(model.CalculatePositionProfit(positions), 2), // 持仓盈亏
		TodayProfit:    util.FloatRound(todayProfit, 2),
		TodayProfitPct: util.FloatRound(todayProfit/(lmv), 4), // 今日盈亏率 = 今日盈亏总和 / 昨日总资产(持仓市值+可用资金)
		Positions:      s.positionList(positions),
	}, nil
}

// getValidContract 查询用户有效的合约ID
func (s *TradeService) getValidContract(ctx context.Context, uid int64) (*model.Contract, error) {
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if user.CurrentContractID != 0 {
		contract, err := dao.ContractDaoInstance().GetContractByID(ctx, user.CurrentContractID)
		if err == nil && contract.Status == model.ContractStatusEnable {
			return contract, nil
		}
	}

	// 任意一个contractID
	contract, err := dao.ContractDaoInstance().GetEnableContractByUID(ctx, uid)
	if err != nil {
		return nil, serr.New(serr.ErrCodeContractNoFound, "暂无有效合约")
	}

	// 更新用户当前合约id
	if err := dao.UserDaoInstance().UpdateCurrentContractID(ctx, uid, contract.ID); err != nil {
		log.Errorf("更新用户当前合约ID失败:%+v", err)
	}
	return contract, nil
}

// positionList 持仓列表
func (s *TradeService) positionList(position []*model.Position) []*model.PositionItem {
	result := make([]*model.PositionItem, 0)
	if len(position) == 0 {
		return result
	}
	sort.SliceStable(position, func(i, j int) bool {
		return timeconv.TimeToInt64(position[i].OrderTime) > timeconv.TimeToInt64(position[j].OrderTime)
	})
	for _, it := range position {
		result = append(result, &model.PositionItem{
			PositionID:  it.ID,                                                         // 持仓编号
			StockCode:   it.StockCode,                                                  // 股票代码
			StockName:   it.StockName,                                                  // 股票名称
			Amount:      it.Amount,                                                     // 持仓数量
			DealPrice:   util.FloatRound(it.Price, 2),                                  // 成本价
			Profit:      util.FloatRound((it.CurPrice-it.Price)*float64(it.Amount), 2), // 持仓盈亏
			MarketValue: util.FloatRound(it.CurPrice*float64(it.Amount), 2),            // 市值
			ValAmount:   it.Amount - it.FreezeAmount,                                   // 可用股数
			NowPrice:    util.FloatRound(it.CurPrice, 2),                               // 现价
			ProfitPct:   util.FloatRound((it.CurPrice-it.Price)/it.Price, 4),           // 盈亏比率
		})
	}
	return result
}

// PositionDetail 持仓明细
func (s *TradeService) PositionDetail(ctx context.Context, positionID int64) (*model.PositionDetail, error) {
	wg := errgroup.GroupWithCount(4)
	// 查询买入
	buy := make([]*model.Buy, 0)
	wg.Go(func() error {
		ret, err := dao.BuyDaoInstance().GetBuyByPositionIDs(ctx, []int64{positionID})
		if err != nil {
			log.Errorf("查询GetBuyByPositionIDs错误:%+v", err)
			return err
		}
		if len(ret) == 0 {
			return serr.ErrBusiness("查询持仓明细失败,买入记录为空")
		}
		buy = ret
		return nil
	})
	// 卖出
	sell := make([]*model.Sell, 0)
	wg.Go(func() error {
		ret, err := dao.SellDaoInstance().GetByPositionIDs(ctx, []int64{positionID})
		if err != nil {
			log.Errorf("查询GetByPositionIDs err:%+v", err)
			return err
		}
		sell = ret
		return nil
	})

	// 查询分红
	dividend := make([]*model.Dividend, 0)
	wg.Go(func() error {
		ret, err := dao.DividendDaoInstance().GetDividendByPositionIDs(ctx, positionID)
		if err != nil {
			log.Errorf("查询GetDividendByPositionIDs:%+v", err)
			return err
		}
		dividend = ret
		return nil
	})
	// 查询持仓
	position := &model.Position{}
	wg.Go(func() error {
		ret, err := dao.PositionDaoInstance().GetPositionByID(ctx, positionID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil
			}
			log.Errorf("查询GetPositionByID失败:%+v", err)
			return err
		}
		position = ret
		return nil
	})
	if err := wg.Wait(); err != nil {
		log.Errorf("持仓明细错误:%+v", err)
		return nil, serr.ErrBusiness("查询持仓明细失败")
	}

	var investMoney, retrieveMoney, totalFee float64 // 投资资金 回收资金 交易手续费
	// 投资资金 = 买入交易金额 + 买(卖)手续费
	// 回收资金 = 卖出金额 + 派息金额 + 系统回收股票资金
	// 交易手续费 = 买入手续费 + 卖出手续费
	for _, it := range buy {
		investMoney += it.Balance
		totalFee += it.Fee
	}
	for _, it := range sell {
		retrieveMoney += it.Balance
		totalFee += it.Fee
	}
	// 回收资金 = 回收资金 + 派息金额
	for _, it := range dividend {
		retrieveMoney += it.DividendMoney
	}

	// 总盈亏 总盈亏比率 持仓股票市值
	var totalProfit, totalProfitPct, marketValue float64

	// 有持仓情况
	if position.ID != 0 {
		qts, err := quote.QtServiceInstance().GetQuoteByTencent([]string{position.StockCode})
		if err != nil {
			return nil, serr.ErrBusiness("查询持仓明细失败")
		}
		// 市值
		marketValue = qts[position.StockCode].CurrentPrice * float64(position.Amount)
		// 总盈亏 = 回收资金 + 当前市值 - 投入资金
		totalProfit = retrieveMoney + marketValue - investMoney
		// 总盈亏比率 = 总盈亏金额 / 投资资金
		totalProfitPct = totalProfit / investMoney

	} else {
		// 清仓情况
		totalProfit = retrieveMoney - investMoney
		totalProfitPct = totalProfit / investMoney
	}

	return &model.PositionDetail{
		TotalProfit:        util.FloatRound(totalProfit, 2),           // 总盈亏:有持仓情况下,计算持仓收益+卖出收益;无持仓情况下,总投入资金-总回收资金
		TotalProfitPct:     util.FloatRound(totalProfitPct, 4),        // 总盈亏比率
		BeginDate:          buy[0].OrderTime.Format("2006-01-02"),     // 开始日期
		EndDate:            "",                                        // 结束日期
		MarketValue:        util.FloatRound(marketValue, 2),           // 市值
		InvestMoney:        investMoney,                               // 投入资金
		RetrieveMoney:      retrieveMoney,                             // 回收资金
		TotalFee:           util.FloatRound(totalFee, 2),              // 交易手续费
		PositionDetailItem: s.positionDetailList(buy, sell, dividend), // 持仓明细item
	}, nil
}

// generatePositionDetailList 生成持仓明细列表
func (s *TradeService) positionDetailList(buyRecord []*model.Buy, sellRecord []*model.Sell, dividend []*model.Dividend) []*model.PositionDetailItem {
	type item struct {
		Date   time.Time `json:"date"`   // 日期
		Type   string    `json:"type"`   // 买入
		Amount int64     `json:"amount"` // 数量
		Price  float64   `json:"price"`  // 价格
		Money  float64   `json:"money"`  // 交易金额
		Fee    float64   `json:"fee"`    // 费用
	}
	items := make([]*item, 0)
	for _, it := range buyRecord {
		items = append(items, &item{Date: it.OrderTime, Type: "买入", Amount: it.Amount, Price: it.Price, Money: it.Balance, Fee: it.Fee})
	}
	for _, it := range sellRecord {
		items = append(items, &item{Date: it.OrderTime, Type: "卖出", Amount: it.Amount, Price: it.Price, Money: it.Balance, Fee: it.Fee})
	}
	for _, it := range dividend {
		items = append(items, &item{Date: it.OrderTime, Type: "分红派息", Amount: it.DividendAmount, Price: 0, Money: it.DividendMoney, Fee: 0})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return timeconv.TimeToInt64(items[i].Date) > timeconv.TimeToInt64(items[j].Date)
	})
	result := make([]*model.PositionDetailItem, 0, len(items))
	for _, it := range items {
		result = append(result, &model.PositionDetailItem{
			Date:   it.Date.Format("01月02日"),
			Type:   it.Type,
			Amount: it.Amount,
			Price:  it.Price,
			Money:  it.Money,
			Fee:    it.Fee,
		})
	}
	return result
}

// TodayDeal 今日成交
func (s *TradeService) TodayDeal(ctx context.Context, contractID int64) ([]*model.TradeDeal, error) {
	entrusts, err := dao.EntrustDaoInstance().GetTodayEntrust(ctx, contractID)
	if err != nil {
		return nil, err
	}
	result := make([]*model.TradeDeal, 0)
	sort.SliceStable(entrusts, func(i, j int) bool {
		return timeconv.TimeToInt64(entrusts[i].OrderTime) > timeconv.TimeToInt64(entrusts[j].OrderTime)
	})
	for _, it := range entrusts {
		if it.Status != model.EntrustStatusTypeDeal && it.Status != model.EntrustStatusTypePartDealPartWithdraw {
			continue
		}
		result = append(result, &model.TradeDeal{
			EntrustID: it.ID,
			StockCode: it.StockCode,                    // 股票代码
			StockName: it.StockName,                    // 股票名称
			Time:      it.OrderTime.Format("15:04:05"), // 时间
			Type:      it.EntrustBS,                    // 类型
			Price:     it.Price,                        // 价格
			Amount:    it.DealAmount,                   // 数量
			Balance:   it.Balance,                      // 成交金额
		})
	}
	return result, nil
}

// HistoryDeal 查询历史成交
func (s *TradeService) HistoryDeal(ctx context.Context, contractID int64) ([]*model.TradeDeal, error) {
	entrusts, err := dao.EntrustDaoInstance().GetEntrustByContractID(ctx, contractID)
	if err != nil {
		return nil, serr.ErrBusiness("查询记录失败")
	}
	result := make([]*model.TradeDeal, 0)
	sort.SliceStable(entrusts, func(i, j int) bool {
		return timeconv.TimeToInt64(entrusts[i].OrderTime) > timeconv.TimeToInt64(entrusts[j].OrderTime)
	})
	for _, it := range entrusts {
		if it.Status != model.EntrustStatusTypeDeal && it.Status != model.EntrustStatusTypePartDealPartWithdraw {
			continue
		}
		result = append(result, &model.TradeDeal{
			EntrustID: it.ID,
			StockCode: it.StockCode,                      // 股票代码
			StockName: it.StockName,                      // 股票名称
			Time:      it.OrderTime.Format("2006-01-02"), // 时间
			Type:      it.EntrustBS,                      // 类型
			Price:     it.Price,                          // 价格
			Amount:    it.DealAmount,                     // 数量
			Balance:   it.Balance,                        // 成交金额
		})
	}
	return result, nil
}

// ContractFee 查询合约费用单
func (s *TradeService) ContractFee(ctx context.Context, contractID int64) ([]*model.Fee, error) {
	contractFee, err := dao.ContractFeeDaoInstance().GetContractFeeByID(ctx, contractID)
	if err != nil {
		log.Errorf("GetContractFeeByID失败:%+v", err)
		return nil, serr.ErrBusiness("查询费用失败")
	}
	sort.SliceStable(contractFee, func(i, j int) bool {
		return timeconv.TimeToInt64(contractFee[i].OrderTime) > timeconv.TimeToInt64(contractFee[j].OrderTime)
	})
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		return nil, err
	}
	result := make([]*model.Fee, 0)
	for _, it := range contractFee {
		// 只下发买入、卖出手续费、合约利息
		if it.Type != model.ContractFeeTypeBuy && it.Type != model.ContractFeeTypeSell && it.Type != model.ContractFeeTypeInterest {
			continue
		}
		name := it.Name
		if it.Type == model.ContractFeeTypeInterest {
			name = contract.FullName()
		}
		result = append(result, &model.Fee{
			Name:   name,                              // 名称(股票名称 & 合约)
			Date:   it.OrderTime.Format("2006-01-02"), // 日期
			Amount: it.Amount,                         // 数量
			Fee:    it.Money,                          // 费用
			Type:   model.ContractFeeTypeMap[it.Type], // 业务类型
		})
	}
	return result, nil
}

// TradeDetail 交易-成交明细
func (s *TradeService) TradeDetail(ctx context.Context, entrustID int64) (*model.TradeDetail, error) {
	entrust, err := dao.EntrustDaoInstance().GetEntrustByID(ctx, entrustID)
	if err != nil {
		log.Errorf("GetEntrustByID失败:%+v", err)
		return nil, serr.ErrBusiness("查询失败")
	}

	entrustType := "买入"
	if entrust.EntrustBS == model.EntrustBsTypeSell {
		entrustType = "卖出"
	}
	return &model.TradeDetail{
		StockCode:  entrust.StockCode,
		StockName:  entrust.StockName,
		Price:      entrust.Price,
		Amount:     entrust.DealAmount,
		Balance:    util.FloatRound(entrust.Price*float64(entrust.DealAmount), 2),
		Fee:        entrust.Fee,
		Status:     model.EntrustStatusMap[entrust.Status], // 1未成交 2成交 3已撤单
		Date:       entrust.OrderTime.Format("2006-01-02"), // 交易日期
		Time:       entrust.OrderTime.Format("15:04:05"),   // 交易时间
		EntrustID:  entrustID,                              // 交易序号
		Type:       entrustType,                            // 交易类型
		ContractID: entrust.ContractID,                     // 合约账户
	}, nil
}

// Position 持仓接口
func (s *TradeService) Position(ctx context.Context, uid, contractID int64) (*model.PositionResp, error) {
	// 合约id不存在则去查找有效合约id
	if contractID == 0 {
		contract, err := s.getValidContract(ctx, uid)
		if err != nil {
			return nil, serr.ErrBusiness("无效合约")
		}
		contractID = contract.ID
	}

	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		return nil, err
	}

	// 查询持仓
	position, err := PositionServiceInstance().GetPositionByContractID(ctx, contract.ID)
	if err != nil {
		log.Error("GetPositionByContractID error:%+v", err)
		return nil, err
	}

	return &model.PositionResp{
		ContractID:   contractID,
		ContractName: contract.FullName(),
		ValMoney:     contract.ValMoney,
		Positions:    s.positionList(position),
	}, nil
}

// InitBuy 买入界面初始化
func (s *TradeService) InitBuy(ctx context.Context, uid, contractID int64, code string) (*model.InitBuy, error) {
	if contractID == 0 {
		contract, err := s.getValidContract(ctx, uid)
		if err != nil {
			return nil, err
		}
		contractID = contract.ID
	}
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		return nil, err
	}
	qt, err := quote.QtServiceInstance().GetQuoteByTencent([]string{code})
	if err != nil {
		return nil, serr.ErrBusiness("证券代码不存在")
	}
	// 最大可买数量
	maxBuyAmount, err := s.maxBuyAmount(ctx, contract, qt[code])
	if err != nil {
		return nil, err
	}
	return &model.InitBuy{
		ContractID:   contractID,
		ContractName: contract.FullName(),
		MaxBuyAmount: maxBuyAmount,                  // 最大可买数量
		PanKou:       model.ConvertPanKou(qt[code]), // 盘口信息
	}, nil
}

// maxBuyAmount 最大可买数量
func (s *TradeService) maxBuyAmount(ctx context.Context, contract *model.Contract, stock *model.TencentQuote) (int64, error) {
	if stock.CurrentPrice < 0.01 {
		return 0, nil
	}
	maxBuyAmount := (int64(contract.ValMoney/stock.CurrentPrice) / 100) * 100
	if maxBuyAmount <= 0 {
		return 0, nil
	}
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		log.Errorf("获取系统参数失败:%+v", err)
		return 0, err
	}

	// 检查ST股票是否允许交易
	if strings.Contains(stock.Name, "ST") {
		if !sys.IsSupportSTStock {
			return 0, nil
		}
		if util.IsZero(sys.STLimitPct) && math.Abs(stock.ChgPercent) >= sys.STLimitPct {
			return 0, nil
		}
	}

	// 检查涨跌幅限制、板块是否允许交易
	switch util.StockBord(stock.Code) {
	case util.StockTypeNormal: // 普通股票
		// 检查普通股票涨跌幅
		if !util.IsZero(sys.LimitPct) && math.Abs(stock.ChgPercent) >= sys.LimitPct {
			return 0, nil
		}
	case util.StockTypeCYBBORD: // 创业板
		if !sys.IsSupportCYBBoard {
			return 0, nil
		}
		if !util.IsZero(sys.CYBLimitPct) && math.Abs(stock.ChgPercent) >= sys.CYBLimitPct {
			return 0, nil
		}
	case util.StockTypeKCBBORD: // 科创板
		if !sys.IsSupportKCBBoard {
			return 0, nil
		}
		if !util.IsZero(sys.KCBLimitPct) && math.Abs(stock.ChgPercent) >= sys.KCBLimitPct {
			return 0, nil
		}
	case util.StockTypeBJ: // 北交所
		if !sys.IsSupportBJBoard {
			return 0, nil
		}
	}

	wg := errgroup.GroupWithCount(4)
	var stockData *model.StockData
	wg.Go(func() error {
		// 股票代码是否存在&是否允许交易
		ret, err := StockDataServiceInstance().GetStockDataByCode(ctx, stock.Code)
		if err != nil {
			log.Errorf("GetStockDataByCode err:%+v", err)
			return err
		}
		stockData = ret
		return nil
	})

	isWarn := false
	wg.Go(func() error {
		// 低于警戒线是否允许开仓
		level, err := ContractServiceInstance().GetContractRiskLevel(ctx, contract)
		if err != nil {
			log.Errorf("IsWarnContract err:%+v", err)
			return err
		}
		if level != model.ContractRiskLevelHealth {
			isWarn = true // 触发警戒线
		}
		return nil
	})
	var positions []*model.Position
	wg.Go(func() error {
		ret, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contract.ID)
		if err != nil {
			log.Errorf("GetPositionByContractID err:%+v", err)
			return err
		}
		positions = ret
		return nil
	})
	var entrusts []*model.Entrust
	wg.Go(func() error {
		list, err := dao.EntrustDaoInstance().GetTodayEntrust(ctx, contract.ID)
		if err != nil {
			return err
		}
		entrusts = list
		return nil
	})
	if err := wg.Wait(); err != nil {
		return 0, err
	}
	// 股票代码是否存在&是否允许交易
	if stockData.Status == model.StockDataStatusDisable {
		return 0, nil
	}
	// 低于警戒线是否允许开仓
	if !sys.LowWarnCanBuy && isWarn {
		return 0, nil
	}
	// 单只股票最大持仓比例:合约总资金(init_money*lever + money ) / 当前股价 = 最大交易股数
	// 可交易最大股数 = 最大交易股数*sys.SinglePositionPct - 已持仓股数
	maxAmount := int64((contract.InitMoney*float64(contract.Lever) + contract.Money) / stock.CurrentPrice)
	if sys.SinglePositionPct > 0 && sys.SinglePositionPct < 1 {
		maxAmount = int64(float64(maxAmount) * sys.SinglePositionPct)
		// 减去持仓股票的数量
		for _, it := range positions {
			if stock.Code == it.StockCode {
				maxAmount -= it.Amount
			}
		}
		// 减去买入委托未成交的股票数量
		for _, it := range entrusts {
			if it.StockCode == stock.Code && it.EntrustBS == model.EntrustBsTypeBuy && it.Status == model.EntrustStatusTypeUnDeal {
				maxAmount -= it.Amount
			}
		}
		maxAmount = (maxAmount / 100) * 100
		if maxAmount < maxBuyAmount {
			maxBuyAmount = maxAmount
		}
	}

	if maxBuyAmount < 0 {
		maxBuyAmount = 0
	}
	return maxBuyAmount, nil
}

// HoldZeroShare 是否持有0股,true返回是,false返回否
func (s *TradeService) HoldZeroShare(ctx context.Context, contractID int64) bool {
	positions, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contractID)
	if err != nil {
		return false
	}
	for _, it := range positions {
		if it.Amount%100 != 0 { // 零股取摩100!=0
			return true
		}
	}
	return false
}

// Buy 交易:买入
func (s *TradeService) Buy(ctx context.Context, p *model.EntrustPackage) error {
	// 检查是否委托时间
	if !CalendarServiceInstance().IsEntrustTime(ctx) {
		return serr.ErrBusiness("委托失败:非交易时间")
	}

	// 幂等:防止重复提交订单
	key := fmt.Sprintf("buy_lock_contract_id_%+v", p.ContractID)
	if db.RedisClient().Exists(ctx, key).Val() == 1 {
		log.Errorf("重复提交订单:%+v", p)
		return nil
	}
	db.RedisClient().SetNX(ctx, key, "1", 1*time.Minute)
	defer db.RedisClient().Del(ctx, key)

	eg := errgroup.GroupWithCount(3)
	var contract *model.Contract
	eg.Go(func() error {
		c, err := dao.ContractDaoInstance().GetContractByID(ctx, p.ContractID)
		if err != nil {
			return serr.ErrBusiness("委托失败:合约不存在")
		}
		contract = c
		return nil
	})
	var sys *model.SysParam
	eg.Go(func() error {
		ret, err := dao.SysDaoInstance().GetSysParam(ctx)
		if err != nil {
			return serr.ErrBusiness("委托失败")
		}
		sys = ret
		return nil
	})
	var stock *model.TencentQuote
	eg.Go(func() error {
		ret, err := quote.QtServiceInstance().GetQuoteByTencent([]string{p.Code})
		if err != nil {
			return err
		}
		q, ok := ret[p.Code]
		if !ok {
			return serr.New(serr.ErrCodeBusinessFail, "委托失败")
		}
		stock = q
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	// 检查委托参数是否合法
	if contract.Status != model.ContractStatusEnable {
		return serr.ErrBusiness("委托失败:无效合约")
	}
	// 可用资金是否充足
	if contract.ValMoney < float64(p.Amount)*p.Price {
		return serr.ErrBusiness("委托失败:可用资金不足")
	}
	// 检查ST股票是否允许交易
	if strings.Contains(stock.Name, "ST") {
		if !sys.IsSupportSTStock {
			return serr.New(serr.ErrCodeBusinessFail, "不允许买入ST股[风控]")
		}
		if util.IsZero(sys.STLimitPct) && math.Abs(stock.ChgPercent) >= sys.STLimitPct {
			return serr.New(serr.ErrCodeBusinessFail, "委托失败[风控],ST股涨跌幅度过大")
		}
	}

	// 检查涨跌幅限制、板块是否允许交易
	switch util.StockBord(stock.Code) {
	case util.StockTypeNormal: // 普通股票
		// 检查普通股票涨跌幅
		if !util.IsZero(sys.LimitPct) && math.Abs(stock.ChgPercent) >= sys.LimitPct {
			return serr.New(serr.ErrCodeBusinessFail, "委托失败[风控],该股涨跌幅度过大")
		}
	case util.StockTypeCYBBORD: // 创业板
		if !sys.IsSupportCYBBoard {
			return serr.New(serr.ErrCodeBusinessFail, "委托失败[风控],创业板不允许交易")
		}
		if !util.IsZero(sys.CYBLimitPct) && math.Abs(stock.ChgPercent) >= sys.CYBLimitPct {
			return serr.New(serr.ErrCodeBusinessFail, "委托失败[风控],该股涨跌幅度过大")
		}
	case util.StockTypeKCBBORD: // 科创板
		// 科创板最小买入数量为200股
		if p.Amount < 200 {
			return serr.New(serr.ErrCodeBusinessFail, "科创板最小买入200股")
		}
		if !sys.IsSupportKCBBoard {
			return serr.New(serr.ErrCodeBusinessFail, "委托失败[风控],科创板不允许交易")
		}
		if !util.IsZero(sys.KCBLimitPct) && math.Abs(stock.ChgPercent) >= sys.KCBLimitPct {
			return serr.New(serr.ErrCodeBusinessFail, "委托失败[风控],该股涨跌幅度过大")
		}
	case util.StockTypeBJ: // 北交所
		if !sys.IsSupportBJBoard {
			return serr.New(serr.ErrCodeBusinessFail, "委托失败[风控],北交所股票不允许交易")
		}
	}

	wg := errgroup.GroupWithCount(3)
	wg.Go(func() error {
		// 单只股票最大持仓生效:检查允许买入最大股数
		if sys.SinglePositionPct > 0 && sys.SinglePositionPct < 1 {
			positions, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contract.ID)
			if err != nil {
				log.Errorf("GetPositionByContractID err:%+v", err)
				return serr.ErrBusiness("委托失败")
			}
			entrusts, err := dao.EntrustDaoInstance().GetTodayEntrust(ctx, contract.ID)
			if err != nil {
				log.Errorf("GetTodayEntrust err:%+v", err)
				return serr.ErrBusiness("委托失败")
			}
			// 单只股票最大持仓比例:合约总资金(init_money*lever + money ) / 当前股价 = 最大交易股数
			// 可交易最大股数 = 最大交易股数*sys.SinglePositionPct - 已持仓股数
			maxAmount := int64((contract.InitMoney*float64(contract.Lever) + contract.Money) / stock.CurrentPrice)
			maxAmount = int64(float64(maxAmount) * sys.SinglePositionPct)
			for _, it := range positions {
				if stock.Code == it.StockCode {
					maxAmount -= it.Amount
				}
			}
			// 减去买入委托未成交的股票数量
			for _, it := range entrusts {
				if it.StockCode == stock.Code && it.EntrustBS == model.EntrustBsTypeBuy && it.Status == model.EntrustStatusTypeUnDeal {
					maxAmount -= it.Amount
				}
			}
			maxAmount = (maxAmount / 100) * 100
			if p.Amount > maxAmount {
				log.Infof("委托失败:委托股数已达风控上限")
				return serr.ErrBusiness("委托失败:委托股数已达风控上限")
			}
		}
		return nil
	})
	wg.Go(func() error {
		// 股票代码是否存在&是否允许交易
		if err := StockDataServiceInstance().IsTrade(ctx, p.Code); err != nil {
			return serr.ErrBusiness("委托失败,[风控提示]该股票不可交易")
		}
		return nil
	})
	wg.Go(func() error {
		// 低于警戒线是否允许开仓
		level, err := ContractServiceInstance().GetContractRiskLevel(ctx, contract)
		if err != nil {
			log.Errorf("IsWarnContract err:%+v", err)
			return serr.ErrBusiness("委托失败")
		}
		if !sys.LowWarnCanBuy && level != model.ContractRiskLevelHealth {
			return serr.New(serr.ErrCodeBusinessFail, "委托失败[风控]:已触发警戒线")
		}
		return nil
	})
	if err := wg.Wait(); err != nil {
		return err
	}

	// 价格检查
	if p.EntrustProp == model.EntrustPropTypeLimitPrice {
		// 限价委托,如果委托价格大于市价则以市价为准
		if p.Price > stock.CurrentPrice {
			p.Price = stock.CurrentPrice
		}
		// 限价委托,如果委托价格小于跌停价,则报错
		if p.Price < stock.LimitDownPrice {
			return serr.ErrBusiness("委托失败:委托价格不能低于跌停价")
		}
	}

	// 市价委托,设置当前价格=市价
	if p.EntrustProp == model.EntrustPropTypeMarketPrice {
		p.Price = stock.CurrentPrice
	}

	// 获取交易手续费
	fee, err := s.GetTradeFee(ctx, p.Price, p.Amount, model.EntrustBsTypeBuy)
	if err != nil {
		return err
	}

	entrust := &model.Entrust{
		UID:         p.UID,                         // 用户ID
		ContractID:  p.ContractID,                  // 合约编号
		OrderTime:   time.Now(),                    // 订单时间
		StockCode:   p.Code,                        // 股票代码
		StockName:   stock.Name,                    // 股票名称
		Amount:      p.Amount,                      // 数量(股)
		Price:       p.Price,                       // 价格
		Balance:     float64(p.Amount) * p.Price,   // 委托金额
		Status:      model.EntrustStatusTypeUnDeal, // 委托状态:1未成交 2成交 3已撤单
		EntrustBS:   model.EntrustBsTypeBuy,        // 交易类型:1买入 2卖出
		EntrustProp: p.EntrustProp,                 // 委托类型:1限价 2市价
		Fee:         fee,                           // 总交易费用
	}

	// 券商委托
	if sys.IsSupportBroker {
		entrust.IsBrokerEntrust = true
	}

	// 创建委托表
	e, err := dao.EntrustDaoInstance().Create(ctx, entrust)
	if err != nil {
		log.Errorf("买入创建委托表失败:%+v", err)
		return serr.ErrBusiness("委托失败")
	}

	// 更新可用资金
	if err := ContractServiceInstance().UpdateValMoneyByID(ctx, entrust.ContractID); err != nil {
		log.Errorf("更新可用资金失败:%+v", err)
	}

	// 券商委托,则发送
	if entrust.IsBrokerEntrust {
		go func() {
			if err := BrokerServiceInstance().Entrust(e); err != nil {
				log.Errorf("委托交易失败:%+v", err)
				return
			}
		}()
	}

	return nil
}

// InitSell 卖出初始化
func (s *TradeService) InitSell(ctx context.Context, contractID int64, code string) (interface{}, error) {
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		return nil, err
	}
	qt, err := quote.QtServiceInstance().GetQuoteByTencent([]string{code})
	if err != nil {
		return nil, err
	}
	list, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contractID)
	if err != nil {
		return nil, err
	}
	var maxSellAmount int64
	for _, it := range list {
		if it.StockCode == code {
			maxSellAmount = it.Amount - it.FreezeAmount
		}
	}

	return &model.InitSell{
		ContractID:    contractID,
		ContractName:  contract.FullName(),
		MaxSellAmount: maxSellAmount,                 // 最大可卖数量
		PanKou:        model.ConvertPanKou(qt[code]), // 盘口信息
	}, nil
}

// GetTradeFee 交易手续费
func (s *TradeService) GetTradeFee(ctx context.Context, price float64, amount int64, entrustBs int64) (float64, error) {
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		return 0, err
	}
	var fee float64
	if entrustBs == model.EntrustBsTypeBuy {
		fee = price * float64(amount) * sys.BuyFee
	} else {
		fee = price * float64(amount) * sys.SellFee
	}
	if !util.IsZero(sys.MiniChargeFee) && sys.MiniChargeFee > fee {
		return sys.MiniChargeFee, nil
	}
	return fee, nil
}

// Sell 卖出
func (s *TradeService) Sell(ctx context.Context, p *model.EntrustPackage) error {
	if !CalendarServiceInstance().IsEntrustTime(ctx) {
		return serr.ErrBusiness("委托失败,非交易时间")
	}
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, p.ContractID)
	if err != nil {
		return serr.ErrBusiness("委托失败:合约不存在")
	}
	if contract.Status != model.ContractStatusEnable {
		return serr.ErrBusiness("委托失败:无效合约")
	}

	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		log.Errorf("GetSysParam err:%+v", err)
		return serr.ErrBusiness("委托失败")
	}

	// 检查可用股数是否满足卖出数量(卖出股数是否大于amount)
	position := &model.Position{}
	positions, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, p.ContractID)
	if err != nil {
		return err
	}
	for _, it := range positions {
		if it.StockCode == p.Code {
			position = it
			break
		}
	}
	if position.ID == 0 {
		return serr.ErrBusiness("委托失败:可卖股数为0")
	}
	if (p.Amount > position.Amount-position.FreezeAmount) || p.Amount > position.Amount {
		return serr.New(serr.ErrCodeBusinessFail, "委托失败,可卖股数不足")
	}

	// 行情
	qts, err := quote.QtServiceInstance().GetQuoteByTencent([]string{p.Code})
	if err != nil {
		return err
	}
	qt, ok := qts[p.Code]
	if !ok {
		return serr.ErrBusiness("委托交易失败")
	}

	// 限价委托,如果卖出价格小于市价则以市价为准
	if p.EntrustProp == model.EntrustPropTypeLimitPrice {
		if p.Price < qt.CurrentPrice {
			p.Price = qt.CurrentPrice
		}
		if p.Price > qt.LimitUpPrice {
			return serr.ErrBusiness("委托失败:委托价格高于涨停价")
		}
	}

	// 市价委托,设置当前价格=市价
	if p.EntrustProp == model.EntrustPropTypeMarketPrice {
		p.Price = qt.CurrentPrice
	}

	fee, err := s.GetTradeFee(ctx, p.Price, p.Amount, model.EntrustBsTypeSell)
	if err != nil {
		return serr.ErrBusiness("委托失败")
	}

	entrust := &model.Entrust{
		UID:             p.UID,                         // 用户ID
		ContractID:      p.ContractID,                  // 合约编号
		OrderTime:       time.Now(),                    // 订单时间
		StockCode:       p.Code,                        // 股票代码
		StockName:       qt.Name,                       // 股票名称
		Amount:          p.Amount,                      // 数量(股)
		Price:           p.Price,                       // 价格
		Balance:         p.Price * float64(p.Amount),   // 委托金额
		Status:          model.EntrustStatusTypeUnDeal, // 委托状态:1未成交 2成交 3已撤单 4部分成交 5等待撤单
		EntrustBS:       model.EntrustBsTypeSell,       // 交易类型:1买入 2卖出
		EntrustProp:     p.EntrustProp,                 // 委托类型:1限价 2市价
		PositionID:      position.ID,                   // 持仓表id(卖出时需填写)
		Fee:             fee,                           // 总交易费用
		IsBrokerEntrust: sys.IsSupportBroker,           // 是否券商委托
		Mode:            p.Mode,                        // 类型:0 主动卖出 1系统平仓
	}

	log.Infof("[业务]:委托卖出,持仓:%+v", position)
	// 创建委托表
	tx := db.StockDB().WithContext(ctx).Begin()
	defer tx.Rollback()
	e, err := dao.EntrustDaoInstance().CreateWithTx(tx, entrust)
	if err != nil {
		log.Errorf("卖出创建委托表失败:%+v", err)
		return serr.ErrBusiness("委托失败")
	}
	log.Infof("[业务]:卖出,创建委托成功:%+v", e)

	// 冻结持仓股数
	if err := dao.PositionDaoInstance().FreezeAmountWithTx(tx, entrust.ContractID, entrust.StockCode, entrust.Amount); err != nil {
		log.Errorf("冻结持仓股数失败:%+v", err)
		return serr.ErrBusiness("委托失败")
	}
	log.Infof("[业务]:卖出,冻结持仓股数:%+v成功!", entrust.Amount)

	if err := tx.Commit().Error; err != nil {
		log.Errorf("卖出冻结股票错误,提交事务失败:%+v", err)
		return serr.ErrBusiness("委托失败")
	}

	pos, err := dao.PositionDaoInstance().GetContractPositionByCode(ctx, entrust.ContractID, entrust.StockCode)
	if err != nil {
		log.Errorf("查询卖出股票错误:%+v", err)
	}
	log.Infof("[业务]:委托卖出,提交事务之后的持仓:%+v", pos)

	// 提交交易,发送到交易所
	go func() {
		if err := BrokerServiceInstance().Entrust(e); err != nil {
			log.Errorf("委托交易失败:%+v", err)
			return
		}
	}()

	return nil
}

// InitWithdraw 撤单页初始化
func (s *TradeService) InitWithdraw(ctx context.Context, contractID int64) (interface{}, error) {
	result := make([]*model.Withdraw, 0)
	list, err := dao.EntrustDaoInstance().GetTodayEntrust(ctx, contractID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return result, nil
	}
	sort.SliceStable(list, func(i, j int) bool {
		return timeconv.TimeToInt64(list[i].OrderTime) > timeconv.TimeToInt64(list[j].OrderTime)
	})
	var a, b []*model.Withdraw
	for _, it := range list {
		w := &model.Withdraw{
			EntrustID:     it.ID,                           // 委托id
			StockCode:     it.StockCode,                    // 股票代码
			StockName:     it.StockName,                    // 股票名称
			Time:          it.OrderTime.Format("15:04:05"), // 时间
			Type:          it.EntrustBS,                    // 类型:1买入 2卖出
			EntrustPrice:  it.Price,                        // 委托价格
			EntrustAmount: it.Amount,                       // 委托数量
			DealAmount:    it.DealAmount,                   // 成交数量
			DealStatus:    it.IsFinallyState(),             // 可撤单:true,不可撤单false
			StatusDesc:    model.EntrustStatusMap[it.Status],
		}
		if it.DealAmount > 0 {
			w.DealPrice = it.Price
		}
		if it.IsFinallyState() {
			// 终态:不可撤单
			a = append(a, w)
		} else {
			// 非终态:可撤单
			b = append(b, w)
		}
	}
	// 可撤单的优先在前面
	result = append(result, b...)
	result = append(result, a...)
	return result, nil
}

// Withdraw 撤单
func (s *TradeService) Withdraw(ctx context.Context, entrustID int64) error {
	entrust, err := dao.EntrustDaoInstance().GetEntrustByID(ctx, entrustID)
	if err != nil {
		return serr.ErrBusiness("委托订单不存在")
	}
	if entrust.IsFinallyState() {
		return serr.ErrBusiness("已撤单")
	} else if entrust.Status == model.EntrustStatusTypeWithdrawing {
		return serr.ErrBusiness("已申报,等待撤单中")
	}
	// 检查合约是否允许撤单
	if ContractServiceInstance().GetWithdrawStatus(ctx, entrust.ContractID) == model.ContractWithdrawStatusDisable {
		return serr.ErrBusiness("合约冻结,撤单失败")
	}

	if entrust.IsBrokerEntrust {
		return s.brokerWithdraw(ctx, entrust)
	}

	return s.WithdrawEntrust(ctx, entrust)

}

// WithdrawEntrust 撤单
func (s *TradeService) WithdrawEntrust(ctx context.Context, entrust *model.Entrust) error {
	entrust.Status = model.EntrustStatusTypeWithdraw
	if entrust.DealAmount == entrust.Amount {
		return serr.ErrBusiness("已成交:撤单失败")
	}
	// 检查买入和委托是否相同数量,修正买入数量

	if entrust.EntrustBS == model.EntrustBsTypeSell {
		if err := dao.PositionDaoInstance().UnFreezeAmount(ctx, entrust.ContractID, entrust.StockCode, entrust.Amount); err != nil {
			log.Errorf("解冻股票失败:%+v", err)
			return serr.ErrBusiness("撤单失败")
		}
	}

	// 券商委托表设置撤单状态
	if entrust.IsBrokerEntrust && len(entrust.BrokerEntrust) > 0 {
		if err := dao.BrokerEntrustDaoInstance().MCreate(ctx, entrust.BrokerEntrust); err != nil {
			log.Errorf("券商委托表更新失败:%+v", err)
			return err
		}
	}

	if err := dao.EntrustDaoInstance().Update(ctx, entrust); err != nil {
		log.Errorf("Update err:%+v", err)
		return err
	}

	if err := ContractServiceInstance().UpdateValMoneyByID(ctx, entrust.ContractID); err != nil {
		log.Errorf("更新可用资金失败:%+v", err)
		return serr.ErrBusiness("更新可用资金失败")
	}
	return nil
}

// brokerWithdraw 券商撤单
func (s *TradeService) brokerWithdraw(ctx context.Context, entrust *model.Entrust) error {
	brokerEntrusts, err := dao.BrokerEntrustDaoInstance().GetByEntrustID(ctx, entrust.ID)
	if err != nil {
		return err
	}
	if len(brokerEntrusts) == 0 {
		return nil
	}

	brokerMap := make(map[int64]*model.Broker)
	for _, broker := range BrokerServiceInstance().GetBrokers() {
		brokerMap[broker.ID] = broker
	}

	// 发送撤单委托
	withdrawBrokerEntrust := make([]*model.BrokerEntrust, 0)
	for _, it := range brokerEntrusts {
		brokerEntrust := it
		broker, ok := brokerMap[brokerEntrust.BrokerID]
		if !ok {
			log.Errorf("未找到有效券商,撤单券商ID:%+v", broker.ID)
			return serr.ErrBusiness("撤单失败")
		}
		if brokerEntrust.IsFinallyState() {
			continue
		}
		if err := BrokerServiceInstance().Withdraw(brokerEntrust, broker, brokerEntrust.BrokerEntrustNo); err != nil {
			log.Errorf("券商撤单失败:%+v", err)
			return serr.ErrBusiness("撤单失败")
			// TODO 多次撤单,看返回错误结果,目前是对撤单错误结果忽略
		}
		brokerEntrust.Status = model.EntrustStatusTypeWithdrawing
		withdrawBrokerEntrust = append(withdrawBrokerEntrust, brokerEntrust)
	}

	// 委托表、券商委托表变更状态:撤单中
	entrust.Status = model.EntrustStatusTypeWithdrawing
	if err := dao.EntrustDaoInstance().Update(ctx, entrust); err != nil {
		return err
	}

	if err := dao.BrokerEntrustDaoInstance().MCreate(ctx, withdrawBrokerEntrust); err != nil {
		return err
	}
	return nil
}

// GetEntrustList 委托记录
func (s *TradeService) GetEntrustList(ctx context.Context, contractID int64) ([]*model.Withdraw, error) {
	list, err := dao.EntrustDaoInstance().GetEntrustByContractID(ctx, contractID)
	if err != nil {
		return nil, err
	}
	result := make([]*model.Withdraw, 0)
	if len(list) == 0 {
		return result, nil
	}
	sort.SliceStable(list, func(i, j int) bool {
		return timeconv.TimeToInt64(list[i].OrderTime) > timeconv.TimeToInt64(list[j].OrderTime)
	})
	for _, it := range list {
		orderTime := it.OrderTime.Format("15:04:05")
		if timeconv.TimeToInt32(it.OrderTime) != timeconv.TimeToInt32(time.Now()) {
			orderTime = it.OrderTime.Format("2006-01-02")
		}

		w := &model.Withdraw{
			EntrustID:     it.ID,         // 委托id
			StockCode:     it.StockCode,  // 股票代码
			StockName:     it.StockName,  // 股票名称
			Time:          orderTime,     // 时间
			Type:          it.EntrustBS,  // 类型:1买入 2卖出
			EntrustPrice:  it.Price,      // 委托价格
			EntrustAmount: it.Amount,     // 委托数量
			DealAmount:    it.DealAmount, // 成交数量
			DealStatus:    true,          // 可撤单:true,不可撤单false
			StatusDesc:    model.EntrustStatusMap[it.Status],
		}
		if it.DealAmount > 0 {
			w.DealPrice = it.Price
		}
		result = append(result, w)
	}
	return result, nil
}

// GetSellOut 已清仓股票
func (s *TradeService) GetSellOut(ctx context.Context, contractID int64) ([]*model.SellOut, error) {
	wg := errgroup.GroupWithCount(5)
	var position []*model.Position
	wg.Go(func() error {
		list, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contractID)
		if err != nil {
			return err
		}
		position = list
		return nil
	})
	var entrust []*model.Entrust
	wg.Go(func() error {
		list, err := dao.EntrustDaoInstance().GetEntrustByContractID(ctx, contractID)
		if err != nil {
			return err
		}
		entrust = list
		return nil
	})
	var sell []*model.Sell
	wg.Go(func() error {
		list, err := dao.SellDaoInstance().GetByContractID(ctx, contractID)
		if err != nil {
			return err
		}
		sell = list
		return nil
	})
	var buy []*model.Buy
	wg.Go(func() error {
		list, err := dao.BuyDaoInstance().GetByContractID(ctx, contractID)
		if err != nil {
			return err
		}
		buy = list
		return nil
	})
	var dividend []*model.Dividend
	wg.Go(func() error {
		list, err := dao.DividendDaoInstance().GetByContractID(ctx, contractID)
		if err != nil {
			return err
		}
		dividend = list
		return nil
	})
	if err := wg.Wait(); err != nil {
		return nil, err
	}
	positionMap := make(map[int64]*model.Position) // map[持仓ID]*model.position
	for _, it := range position {
		positionMap[it.ID] = it
	}

	sellOutPositionIDsMap := make(map[int64]bool)
	for _, it := range entrust {
		// 委托:过滤掉非成交的
		if it.Status != model.EntrustStatusTypePartDeal && it.Status != model.EntrustStatusTypeDeal {
			continue
		}
		// 持仓:过滤掉持仓的记录
		if _, ok := positionMap[it.PositionID]; ok {
			continue
		}
		sellOutPositionIDsMap[it.PositionID] = true
	}

	list := make([]*model.SellOut, 0)
	if len(sellOutPositionIDsMap) == 0 {
		return list, nil
	}

	sellMap := make(map[int64][]*model.Sell) // sellMap[持仓ID][]*model.Sell
	for _, it := range sell {
		// 过滤掉非清仓的卖出记录
		if _, ok := sellOutPositionIDsMap[it.PositionID]; !ok {
			continue
		}
		s, ok := sellMap[it.PositionID]
		if !ok {
			s = make([]*model.Sell, 0)
		}
		s = append(s, it)
		sellMap[it.PositionID] = s
	}

	buyMap := make(map[int64][]*model.Buy)
	for _, it := range buy {
		// 过滤掉非清仓买入的记录
		if _, ok := sellOutPositionIDsMap[it.PositionID]; !ok {
			continue
		}
		b, ok := buyMap[it.PositionID]
		if !ok {
			b = make([]*model.Buy, 0)
		}
		b = append(b, it)
		buyMap[it.PositionID] = b
	}

	dividendMap := make(map[int64][]*model.Dividend)
	for _, it := range dividend {
		// 过滤掉非已清仓的分红
		if _, ok := sellOutPositionIDsMap[it.PositionID]; !ok {
			continue
		}
		d, ok := dividendMap[it.PositionID]
		if !ok {
			d = make([]*model.Dividend, 0)
		}
		d = append(d, it)
		dividendMap[it.PositionID] = d
	}

	for positionID := range sellOutPositionIDsMap {
		s, ok := sellMap[positionID]
		if !ok || len(s) == 0 {
			continue
		}
		b, ok := buyMap[positionID]
		if !ok || len(b) == 0 {
			continue
		}
		d, ok := dividendMap[positionID]
		if !ok {
			d = make([]*model.Dividend, 0)
		}
		// 排序,时间最前的放在第一个
		sort.SliceStable(s, func(i, j int) bool {
			return timeconv.TimeToInt64(s[i].OrderTime) > timeconv.TimeToInt64(s[j].OrderTime)
		})
		var investMoney, retrieveMoney, totalFee float64 // 投资资金 回收资金 交易手续费
		for _, it := range b {
			investMoney += it.Balance
			totalFee += it.Fee
		}
		for _, it := range s {
			retrieveMoney += it.Balance
			totalFee += it.Fee
		}
		// 投入资金 = 投入资金 + 买(卖)手续费
		investMoney += totalFee
		// 回收资金 = 回收资金 +派息金额
		for _, it := range d {
			retrieveMoney += it.DividendMoney
		}

		// 盈亏 =  回收资金 - 投入资金
		profit := retrieveMoney - investMoney

		// 计算卖出均价
		var sellAmount int64
		for _, it := range s {
			sellAmount += it.Amount
		}
		list = append(list, &model.SellOut{
			PositionID: positionID,
			StockCode:  s[0].StockCode,
			StockName:  s[0].StockName,
			Time:       s[0].OrderTime.Format("2006-01-02"),
			Profit:     util.FloatRound(profit, 2),
			ProfitPct:  util.ChangePctDesc(profit / investMoney),              // 盈亏比例 = 盈亏金额 / 总投入资金
			Price:      util.FloatRound(retrieveMoney/float64(sellAmount), 2), // 清仓均价
		})
	}
	return list, nil
}

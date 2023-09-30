package service

import (
	"context"
	"errors"
	"fmt"
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
	"strconv"
	"strings"
	"sync"
	"time"
)

// ContractService 合约服务
type ContractService struct {
}

var (
	contractService *ContractService
	contractOnce    sync.Once
)

// ContractServiceInstance ContractServiceInstance
func ContractServiceInstance() *ContractService {
	contractOnce.Do(func() {
		contractService = &ContractService{}

		ctx := context.Background()
		cs := CalendarServiceInstance()
		// 检测是否触达警戒线:触发警戒线,短信通知;一天仅通知一次
		go func() {
			for range time.Tick(5 * time.Second) {
				if !cs.IsTradeTime(ctx) {
					continue
				}
				// 合约检测:检测是否触发警戒线、平仓线
				if err := contractService.checkContact(ctx); err != nil {
					log.Errorf("checkContact err:%+v", err)
				}
			}
		}()

		// 扣取合约利息
		go func() {
			for range time.Tick(5 * time.Second) {
				// 3点15分开始收取利息费用
				if time.Now().Hour() >= 15 && time.Now().Minute() >= 15 {
					// 检查redis是否已经收取
					key := fmt.Sprintf("manage_fee_cache_key_%v", timeconv.TimeToInt32(time.Now()))
					if db.RedisClient().Get(ctx, key).Val() == "1" {
						continue
					}
					if err := contractService.contractFee(ctx); err != nil {
						log.Errorf("收取合约利息费用失败:%+v", err)
						continue
					}
					// 设置redis,已经收取合约利息
					if err := db.RedisClient().Set(ctx, key, "1", 12*time.Hour).Err(); err != nil {
						log.Errorf("设置redis收取管理利息费失败,err:%+v", err)
					}
				}
			}
		}()

	})
	return contractService
}

// contractFee 合约费用
func (s *ContractService) contractFee(ctx context.Context) error {
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		log.Errorf("GetSysParam err:%+v", err)
		return err
	}
	contracts, err := dao.ContractDaoInstance().GetContracts(ctx)
	if err != nil {
		log.Errorf("GetContracts err:%+v", err)
		return err
	}
	ok := CalendarServiceInstance().IsTradeDate(ctx)

	for _, it := range contracts {
		contract := it
		if it.Status != model.ContractStatusEnable {
			continue
		}
		// 判断今天是否应该收取管理费
		switch contract.Type {
		case model.ContractTypeDay: // 按天合约,节假日是否收取留仓费
			if !ok && !sys.HolidayCharge {
				log.Infof("非交易不收取留仓费")
				continue
			}
		case model.ContractTypeWeek:
			// 当天开的按周合约,不收取费用
			if timeconv.TimeToInt32(contract.OrderTime) == timeconv.TimeToInt32(time.Now()) {
				continue
			}
			// 合约是否到周期
			if contract.OrderTime.Weekday() != time.Now().Weekday() {
				continue
			}
		case model.ContractTypeMonth:
			// 当天开的按月合约,不收取费用
			if timeconv.TimeToInt32(contract.OrderTime) == timeconv.TimeToInt32(time.Now()) {
				continue
			}
			isCharge := false
			for index := 1; index <= 12; index++ {
				// 下个月提前一天扣费
				chargeDate := contract.OrderTime.AddDate(0, index, -1)
				if timeconv.TimeToInt32(chargeDate) > timeconv.TimeToInt32(time.Now()) {
					break
				}
				if timeconv.TimeToInt32(chargeDate) == timeconv.TimeToInt32(time.Now()) {
					isCharge = true
					break
				}
			}
			if !isCharge {
				continue
			}
		}

		interest := model.Interest(contract, sys, contract.InitMoney)
		it.Money -= interest
		if err := dao.ContractDaoInstance().UpdateContract(ctx, contract); err != nil {
			log.Errorf("收取合约[%v],金额:[%v]管理费失败:%+v", contract.ID, interest, err)
			continue
		}
		if err := dao.ContractFeeDaoInstance().Create(ctx, &model.ContractFee{
			UID:        contract.UID,
			ContractID: contract.ID,
			Code:       "",
			Name:       "",
			Amount:     0,
			OrderTime:  time.Now(),
			Direction:  model.ContractFeeDirectionPay,
			Money:      interest,
			Detail:     fmt.Sprintf("扣取合约资金利息费用:%0.2f", interest),
			Type:       model.ContractFeeTypeInterest,
		}); err != nil {
			log.Errorf("记录:收取合约[%v],金额:[%v],管理费失败:%+v", contract.ID, interest, err)
			continue
		}
		// 填写消息表
		if err := dao.MsgDaoInstance().Create(ctx, &model.Msg{
			UID:        contract.UID,
			Title:      "递延费",
			Content:    fmt.Sprintf("您合约[%d]申请借款资金:%0.2f元,扣取递延利息费用:%0.2f元,请留意您的资金变动。", contract.ID, contract.InitMoney*float64(contract.Lever), interest),
			CreateTime: time.Now(),
		}); err != nil {
			log.Errorf("填写消息表失败:%+v", err)
		}

		// 更新资金
		if err := s.UpdateValMoneyByID(ctx, contract.ID); err != nil {
			log.Errorf("更新合约资金失败:%+v", err)
		}
	}
	return nil
}

// GetWithdrawStatus 查询是合约是否可以撤单
func (s *ContractService) GetWithdrawStatus(ctx context.Context, contractID int64) model.ContractWithdrawStatus {
	if db.Get(ctx, s.withdrawOrderCacheKey(contractID)).Val() == "1" {
		return model.ContractStatusDisabled // 不可撤单
	}
	return model.ContractWithdrawStatusEnable // 可撤单
}

// withdrawOrderCacheKey 不允许撤单缓存key
func (s *ContractService) withdrawOrderCacheKey(contractID int64) string {
	return fmt.Sprintf("no_withdraw_order_contract_id_%d", contractID)
}

// closeContractCacheKey 爆仓发送短信
func (s *ContractService) closeContractCacheKey(contractID int64) string {
	return fmt.Sprintf("sms_to_user_reason:contract_close_id_%d", contractID)
}

// checkContact 合约检查:检查是否触发警戒线,平仓线
func (s *ContractService) checkContact(ctx context.Context) error {
	contracts, err := dao.ContractDaoInstance().GetContracts(ctx)
	if err != nil {
		log.Errorf("GetContracts err:%+v", err)
		return err
	}

	for _, contract := range contracts {

		// 过滤掉非操盘的合约
		if contract.Status != model.ContractStatusEnable {
			continue
		}
		level, err := s.GetContractRiskLevel(ctx, contract)
		if err != nil {
			log.Errorf("GetContractRiskLevel err:%+v", err)
			continue
		}
		switch level {
		case model.ContractRiskLevelHealth:
			{
				// 正常合约,检查redis是否设置了禁止撤单标志,是则删除掉
				if db.Get(ctx, s.withdrawOrderCacheKey(contract.ID)).Val() == "1" {
					log.Infof("删除撤单标识,合约编号:%+v", contract.ID)
					if err := db.Delete(ctx, s.withdrawOrderCacheKey(contract.ID)).Err(); err != nil {
						log.Errorf("删除禁止撤单标志失败:%+v", err)
						return err
					}
					log.Infof("删除撤单标识完毕!!!!!!合约编号:%+v", contract.ID)
				}
				// 正常合约,检查是否设置了爆仓短信提醒标志,是则删除掉
				if db.Get(ctx, s.closeContractCacheKey(contract.ID)).Val() == "1" {
					log.Infof("删除爆仓短信标志,合约编号:%+v", contract.ID)
					if err := db.Delete(ctx, s.closeContractCacheKey(contract.ID)).Err(); err != nil {
						log.Errorf("删除禁止撤单标志失败:%+v", err)
						return err
					}
				}
			}

		case model.ContractRiskLevelClose: // 触发平仓线:是否有股票,有股票则全部卖出;
			{
				// 检查合约是否已经设置了今日禁止撤单标志.
				if db.Get(ctx, s.withdrawOrderCacheKey(contract.ID)).Val() == "1" {
					continue
				}
				positions, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contract.ID)
				if err != nil {
					log.Errorf("GetPositionByContractID err:%+v", err)
					return nil
				}
				// 无持仓股票,则跳过
				if len(positions) == 0 {
					continue
				}
				log.Infof("合约爆仓:%+v", contract.ID)
				// 撤销所有的正在委托订单
				entrusts, err := dao.EntrustDaoInstance().GetEntrustByContractID(ctx, contract.ID)
				if err != nil {
					log.Errorf("GetEntrustByContractID err:%+v", err)
					return err
				}
				for _, it := range entrusts {
					// 今日未成交订单,发起委托撤单
					if timeconv.TimeToInt32(it.OrderTime) == timeconv.TimeToInt32(time.Now()) && it.Status == model.EntrustStatusTypeUnDeal {
						if err := TradeServiceInstance().Withdraw(ctx, it.ID); err != nil {
							log.Errorf("爆仓撤单失败,Withdraw err:%+v", err)
							return err
						}
						log.Infof("合约爆仓,撤销委托:%+v", it)
					}
				}

				// 持仓股票,以市价委托卖出
				for _, it := range positions {
					// 可卖股数小于等于0,则不卖出
					if it.Amount-it.FreezeAmount <= 0 {
						continue
					}
					if err := TradeServiceInstance().Sell(ctx, &model.EntrustPackage{
						UID:         it.UID,
						ContractID:  it.ContractID,
						Code:        it.StockCode,
						Price:       0,
						Amount:      it.Amount - it.FreezeAmount,
						EntrustProp: model.EntrustPropTypeMarketPrice,
						Mode:        model.SystemMode, // 委托方
					}); err != nil {
						log.Errorf("爆仓平仓卖出错误,Sell err:%+v", err)
						return err
					}
				}
				user, err := dao.UserDaoInstance().GetUserByUID(ctx, contract.UID)
				if err != nil {
					log.Errorf("GetUserByUID err:%+v", err)
					return err
				}
				// 设置禁止撤单标志
				if _, err := db.Set(ctx, s.withdrawOrderCacheKey(contract.ID), "1", 7*time.Hour).Result(); err != nil {
					log.Errorf("设置禁止撤单标志失败:%+v", err)
					return err
				}

				// 是否已经短信提醒
				if db.Get(ctx, s.closeContractCacheKey(contract.ID)).Val() == "1" {
					continue
				}
				if _, err := db.Set(ctx, s.closeContractCacheKey(contract.ID), "1", -1).Result(); err != nil {
					log.Errorf("爆仓设置短信redis提醒标志失败,err:%+v", err)
					return err
				}
				if err := SmsServiceInstance().SendSms(ctx, fmt.Sprintf("尊敬的客户,由于您的%s:[%d]保证金已触达平仓水平,合约持仓股票将按市况执行平仓处理，请知悉。", contract.FullName(), contract.ID), user.UserName); err != nil {
					log.Errorf("SendSms err:%+v", err)
				}
				log.Infof("合约爆仓完毕，合约编号:%+v", contract.ID)
			}

		case model.ContractRiskLevelWarn:
			{
				// 触发警戒线: 发送短信提醒,仅限交易日发送
				if CalendarServiceInstance().IsEntrustTime(ctx) {

					// 检测今天是否已经发送短信;
					cacheKey := fmt.Sprintf("sms_to_warn_contract_id_%d", contract.ID)
					if db.Get(ctx, cacheKey).Val() == "1" {
						continue
					}
					log.Infof("合约触发警戒线，合约编号:%+v", contract.ID)

					content := fmt.Sprintf("尊敬的客户,您的%s[%d]保证金已触达警戒水平,请知悉。", contract.FullName(), contract.ID)
					user, err := dao.UserDaoInstance().GetUserByUID(ctx, contract.UID)
					if err != nil {
						log.Errorf("GetUserByUID err:%+v", err)
						return err
					}
					if err := SmsServiceInstance().SendSms(ctx, content, user.UserName); err != nil {
						log.Errorf("SendSms err:%+v", err)
					}

					// 设置redis,为已发送短信
					if err := db.Set(ctx, cacheKey, "1", 8*time.Hour).Err(); err != nil {
						log.Errorf("set redis err:%+v", err)
					}

				}

			}

		}
	}
	return nil
}

// List 查询合约
func (s *ContractService) List(ctx context.Context, uid int64) ([]*model.ValidContract, error) {
	list, err := dao.ContractDaoInstance().GetContractsByUID(ctx, uid)
	if err != nil {
		return nil, serr.ErrBusiness("查询合约失败")
	}
	if len(list) == 0 {
		return make([]*model.ValidContract, 0), nil
	}
	result := make([]*model.ValidContract, 0)

	eg := errgroup.GroupWithCount(2)
	// 查询系统参数
	var sys *model.SysParam
	eg.Go(func() error {
		ret, err := dao.SysDaoInstance().GetSysParam(ctx)
		if err != nil {
			return err
		}
		sys = ret
		return nil
	})
	// 查询用户
	var user *model.User
	eg.Go(func() error {
		ret, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
		if err != nil {
			return err
		}
		user = ret
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	sort.SliceStable(list, func(i, j int) bool {
		return timeconv.TimeToInt64(list[i].OrderTime) > timeconv.TimeToInt64(list[j].OrderTime)
	})
	for _, contract := range list {
		// 返回生效的合约
		if contract.Status != model.ContractStatusEnable {
			continue
		}
		positions, err := PositionServiceInstance().GetPositionByContractID(ctx, contract.ID)
		if err != nil {
			return nil, err
		}
		profit := model.CalculatePositionProfit(positions)
		todayProfit, err := s.GetTodayProfitByContractID(ctx, contract.ID)
		if err != nil {
			log.Errorf("GetTodayProfitByContractID err:%+v", err)
			return nil, err
		}
		result = append(result, &model.ValidContract{
			ID:          contract.ID,                                                                                     // 合约id
			Name:        contract.FullName(),                                                                             // 合约名称
			TodayProfit: util.FloatRound(todayProfit, 2),                                                                 // 今日盈亏
			Profit:      util.FloatRound(profit, 2),                                                                      // 持仓盈亏
			ProfitPct:   util.FloatRound(profit/(contract.ValMoney+model.CalculatePositionMarketValue(positions)), 4),    // 持仓盈亏比率=持仓盈亏/总资产(可用资金+持仓市值)
			Money:       contract.Money,                                                                                  // 保证金
			MarketValue: util.FloatRound(model.CalculatePositionMarketValue(positions), 2),                               // 证券市值
			ValMoney:    contract.ValMoney,                                                                               // 可用资金
			Interest:    fmt.Sprintf("%.2f元/%s", model.Interest(contract, sys, contract.InitMoney), contract.TypeText()), // 管理费
			Warn:        s.Warn(contract, sys),                                                                           // 警戒参考值线
			Close:       s.Close(contract, sys),                                                                          // 平仓线参考值
			Risk:        int64(s.riskDesc(ctx, contract)),                                                                // 风险水平
			Select:      user.CurrentContractID == contract.ID,                                                           // 当前合约(true为选中)
		})
	}

	// 当前合约放在第一个
	for index, contract := range result {
		if contract.ID == user.CurrentContractID {
			item := result[index]
			result[index] = result[0]
			result[0] = item
			break
		}
	}
	return result, nil
}

// ApplyInit 合约申请初始化
func (s *ContractService) ApplyInit(ctx context.Context, uid int64) (*model.ContractConf, error) {
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		return nil, err
	}
	// 合约类型:按天、按周、按月
	contractTypes := make([]*model.ContractType, 0)
	if sys.IsSupportDayContract {
		contractTypes = append(contractTypes, model.ContractTypeMap[model.ContractTypeDay])
	}
	if sys.IsSupportWeekContract {
		contractTypes = append(contractTypes, model.ContractTypeMap[model.ContractTypeWeek])
	}
	if sys.IsSupportMonthContract {
		contractTypes = append(contractTypes, model.ContractTypeMap[model.ContractTypeMonth])
	}

	// 合约杠杆:2倍、3倍...
	contractLevers := make([]*model.ContractLever, 0)
	segs := strings.Split(sys.ContractLever, ",")
	for _, seg := range segs {
		lever, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			log.Errorf("parse err :%+v", err)
			continue
		}
		contractLevers = append(contractLevers, &model.ContractLever{
			Lever: lever,
			Name:  fmt.Sprintf("%d倍", lever),
		})
	}
	// 合约可用资金
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &model.ContractConf{
		Money: user.Money,     // 可用资金
		Type:  contractTypes,  // 类型
		Lever: contractLevers, // 合约杠杆
	}, nil
}

// ContractApply 创建合约
func (s *ContractService) ContractApply(ctx context.Context, uid int64, money float64, contractType, contractLever int64) (*model.ContractApply, error) {
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	contract, err := dao.ContractDaoInstance().CreateContract(ctx, &model.Contract{
		UID:       user.ID,                              // 用户ID
		InitMoney: money,                                // 原始保证金
		Money:     money,                                // 现保证金
		ValMoney:  money*float64(contractLever) + money, // 可用资金:可用资金 = 原始资金*杠杠 + 原始本金
		Lever:     contractLever,                        // 合约杠杠倍数
		Status:    1,                                    // 合约状态:1预申请 2操盘中 3操盘结束
		Type:      contractType,                         // 合约类型:1按天合约 2:按周合约 3:按月合约
		OrderTime: time.Now(),                           // 订单时间
		CloseTime: time.Now().AddDate(1, 0, 0),
	})
	if err != nil {
		log.Errorf("创建合约失败,err:%+v", err)
		return nil, serr.ErrBusiness("创建合约失败")
	}
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		return nil, err
	}

	return &model.ContractApply{
		ContractName: contract.FullName(),                                                                             // 合约名称
		Money:        contract.Money,                                                                                  // 投资本金
		ContractID:   contract.ID,                                                                                     // 合约id
		ContractType: fmt.Sprintf("按%s合约", contract.TypeText()),                                                       // 合约类型
		ValMoney:     contract.ValMoney,                                                                               // 操盘资金
		Period:       "自动续约",                                                                                          // 操盘期限
		Interest:     fmt.Sprintf("%.2f元/%s", model.Interest(contract, sys, contract.InitMoney), contract.TypeText()), // 利息
		Warn:         strconv.FormatFloat(s.Warn(contract, sys), 'f', 2, 64),                                          // 警戒线
		Close:        strconv.FormatFloat(s.Close(contract, sys), 'f', 2, 64),                                         // 平仓线
		Pay:          model.Interest(contract, sys, contract.InitMoney) + contract.Money,                              // 支付本金 = 利息+投资本金
		Wallet:       user.Money,                                                                                      // 钱包余额
	}, nil
}

// Warn 警戒线值:原始资金*杠杠 + 原始资金 - 原始资金*警戒比率
func (s *ContractService) Warn(contract *model.Contract, sys *model.SysParam) float64 {
	ValMoney := contract.InitMoney * float64(contract.Lever)
	return ValMoney + contract.InitMoney*sys.WarnPct
}

// Close 平仓线:原始资金*杠杠 + 原始资金 - 原始资金*平仓比率
func (s *ContractService) Close(contract *model.Contract, sys *model.SysParam) float64 {
	ValMoney := contract.InitMoney * float64(contract.Lever)
	return ValMoney + contract.InitMoney*sys.ClosePct
}

// ContractFee 合约费用
func (s *ContractService) ContractFee(contract *model.Contract, sys *model.SysParam) float64 {
	switch contract.Type {
	case 1: // 按日
		return sys.DayContractFee * (contract.InitMoney * float64(contract.Lever))
	case 2: // 按周
		return sys.WeekContractFee * (contract.InitMoney * float64(contract.Lever))
	case 3: // 按月
		return sys.MonthContractFee * (contract.InitMoney * float64(contract.Lever))
	}
	return 0.00
}

// UpdateValMoneyByID 刷新合约资金
func (s *ContractService) UpdateValMoneyByID(ctx context.Context, contractID int64) error {
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		return serr.ErrBusiness("合约不存在")
	}
	if contract.Status != model.ContractStatusEnable {
		return serr.ErrBusiness("合约非有效状态")
	}
	positions, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contract.ID)
	if err != nil {
		return err
	}
	// 1.合约市值
	positionAsset := model.CalculatePositionAsset(positions)

	// 2.查询委托
	entrusts, err := dao.EntrustDaoInstance().GetTodayEntrust(ctx, contract.ID)
	if err != nil {
		return err
	}
	entrustAsset := 0.00
	for _, it := range entrusts {
		if it.EntrustBS != model.EntrustBsTypeBuy {
			continue
		}
		if it.Status == model.EntrustStatusTypeUnDeal || it.Status == model.EntrustStatusTypeReported {
			entrustAsset += it.Price * float64(it.Amount)
		}
	}

	// 3.计算可用资金 = (InitMoney(现保证金)*lever(倍数) + 保证金) - 持仓市值 - 委托市值
	valMoney := (contract.InitMoney*float64(contract.Lever) + contract.Money) - positionAsset - entrustAsset
	if valMoney < 0 {
		valMoney = 0
	}
	if err := dao.ContractDaoInstance().UpdateContractValMoney(ctx, valMoney, contract.ID); err != nil {
		log.Errorf("刷新可用资金失败:%+v", err)
		return err
	}

	return nil
}

// Create 确认合约
func (s *ContractService) Create(ctx context.Context, uid, contractID int64) error {
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		return err
	}

	wg := errgroup.GroupWithCount(2)
	var contract *model.Contract
	wg.Go(func() error {
		ret, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
		if err != nil {
			log.Errorf("合约不存在 err:%+v", err)
			return serr.ErrBusiness("合约不存在")
		}
		contract = ret
		return nil
	})
	var user *model.User
	wg.Go(func() error {
		ret, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
		if err != nil {
			return serr.ErrBusiness("用户不存在")
		}
		user = ret
		return nil
	})
	if err := wg.Wait(); err != nil {
		return err
	}

	// 检查合约状态
	if contract.Status != model.ContractStatusApply {
		return serr.ErrBusiness("申请合约失败,合约状态错误")
	}
	// 申请合约资金不能小于2000
	if contract.Money < 2000 {
		return serr.ErrBusiness("合约申请保证金不能小于2000")
	}

	// 合约扣费
	payMoney := contract.Money + model.Interest(contract, sys, contract.InitMoney) // 费用=本金+利息
	if user.Money < payMoney {
		return serr.ErrBusiness("合约申请失败:账户资金不足")
	}
	user.Money = user.Money - payMoney   // 扣除用户资金
	user.CurrentContractID = contract.ID // 设置当前合约为选中合约
	eg := errgroup.GroupWithCount(3)
	eg.Go(func() error {
		if err := dao.UserDaoInstance().CreateUser(ctx, user); err != nil {
			log.Errorf("CreateUser err:%+v", err)
			return err
		}
		return nil
	})
	eg.Go(func() error {
		// 扣取合约费用
		if err := dao.ContractFeeDaoInstance().Create(ctx, &model.ContractFee{
			UID:        contract.UID,
			ContractID: contractID,
			Code:       "",
			Name:       "",
			Amount:     0,
			OrderTime:  time.Now(),
			Direction:  model.ContractFeeDirectionPay,                                                                                                         // 方向:1支出 2:收入
			Money:      model.Interest(contract, sys, contract.InitMoney),                                                                                     // 金额
			Detail:     fmt.Sprintf("%s[%d]申请成功,扣除合约费用:%0.2f元,请留意资金变动。", contract.FullName(), contract.ID, model.Interest(contract, sys, contract.InitMoney)), // 明细
			Type:       model.ContractFeeTypeInterest,                                                                                                         // 费用类型1:买入手续费 2:卖出手续费 3:合约利息 4:卖出盈亏 5:追加保证金 6:扩大资金 7:合约结算
		}); err != nil {
			log.Errorf("申请合约填写扣费信息失败:%+v", err)
			return serr.ErrBusiness("合约申请失败")
		}
		return nil
	})
	eg.Go(func() error {
		// 设置合约状态
		contract.Status = model.ContractStatusEnable
		if err := dao.ContractDaoInstance().UpdateContract(ctx, contract); err != nil {
			log.Errorf("UpdateContract err:%+v", err)
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	if err := dao.TransferDaoInstance().Create(ctx, &model.Transfer{
		UID:       contract.UID,
		OrderTime: time.Now(),
		Money:     contract.Money,
		Type:      model.TransferTypeCreateContract,
		Status:    model.TransferStatusSuccess,
		Name:      "",
		BankNo:    "",
		Channel:   "",
		OrderNo:   "",
	}); err != nil {
		log.Errorf("填写资金明细表失败,err:%+v", err)
		return err
	}
	log.Infof("开启合约:%+v 扣取费用:%f", contract, model.Interest(contract, sys, contract.InitMoney))
	return nil
}

// Detail 合约详情
func (s *ContractService) Detail(ctx context.Context, contractID int64) (*model.ContractDetail, error) {
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		return nil, serr.ErrBusiness("合约不存在")
	}
	// 获取持仓盈亏比例
	positions, err := PositionServiceInstance().GetPositionByContractID(ctx, contractID)
	if err != nil {
		log.Errorf("获取持仓盈亏失败:%+v", err)
		return nil, serr.ErrBusiness("查询持仓失败")
	}
	profit := model.CalculatePositionProfit(positions)
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		return nil, err
	}
	// 提盈金额
	getProfitMoney := 0.00
	list, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contractID)
	if err != nil {
		return nil, err
	}
	// 有持仓则不允许提盈:	提盈金额 = 合约保证金 - 初始资金
	if len(list) == 0 && contract.Money < contract.InitMoney {
		getProfitMoney = contract.Money - contract.InitMoney
	}

	result := &model.ContractDetail{
		Name:                  contract.FullName(),                                                                                        // 合约名称
		ID:                    contract.ID,                                                                                                // 合约id
		Profit:                util.FloatRound(profit, 2),                                                                                 // 持仓盈亏
		ProfitPct:             util.FloatRound(profit/(contract.InitMoney*float64(contract.Lever)+contract.InitMoney), 4),                 // 持仓盈亏比率
		MarketValue:           model.CalculatePositionMarketValue(positions),                                                              // 证券市值
		Money:                 contract.Money,                                                                                             // 保证金
		ValMoney:              contract.ValMoney,                                                                                          // 可用资金
		InterestBearingAmount: contract.InitMoney * float64(contract.Lever),                                                               // 计息金额
		Interest:              fmt.Sprintf("%0.2f/%s", model.Interest(contract, sys, contract.InitMoney), contract.TypeText()),            // 利息
		AppendMoney:           contract.AppendMoney,                                                                                       // 追加保证金
		Warn:                  util.FloatRound(s.Warn(contract, sys), 2),                                                                  // 警戒参考值线
		Close:                 util.FloatRound(s.Close(contract, sys), 2),                                                                 // 平仓线参考值
		Risk:                  int64(s.riskDesc(ctx, contract)),                                                                           // 风险水平
		CreateTime:            contract.OrderTime.Format("2006-01-02 15:04:05"),                                                           // 合约创建时间
		TotalAsset:            fmt.Sprintf("%0.2f元", util.FloatRound(model.CalculatePositionMarketValue(positions)+contract.ValMoney, 2)), // 总资产
		OriginalMoney:         fmt.Sprintf("%0.2f元", contract.InitMoney),                                                                  // 原始保证金
		Lever:                 fmt.Sprintf("%+v", contract.Lever),                                                                         // 合约杠杠
		GetProfit:             fmt.Sprintf("%0.2f元", getProfitMoney),                                                                      // 可提取利润
	}
	return result, nil
}

// GetContractRiskLevel 合约风险登记
func (s *ContractService) GetContractRiskLevel(ctx context.Context, contract *model.Contract) (model.ContractRiskLevel, error) {
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		return 0, err
	}
	positions, err := PositionServiceInstance().GetPositionByContractID(ctx, contract.ID)
	if err != nil {
		return 0, err
	}
	profit := 0.00
	for _, position := range positions {
		// 行情价格错误
		if position.CurPrice <= 0.1 {
			return 0, errors.New("行情价格错误")
		}
		profit += (position.CurPrice - position.Price) * float64(position.Amount)
	}

	if contract.InitMoney*sys.ClosePct > contract.Money+profit {
		// 低于平仓线
		return model.ContractRiskLevelClose, nil
	} else if contract.InitMoney*sys.WarnPct > contract.Money+profit {
		// 低于警戒线
		return model.ContractRiskLevelWarn, nil
	} else {
		// 正常值
		return model.ContractRiskLevelHealth, nil
	}
}

func (s *ContractService) riskDesc(ctx context.Context, contract *model.Contract) model.ContractRiskLevel {
	level, err := s.GetContractRiskLevel(ctx, contract)
	if err != nil {
		return 1
	}
	return level
}

// Settlement 合约结算
func (s *ContractService) Settlement(ctx context.Context, contractID int64) error {
	wg := errgroup.GroupWithCount(2)
	var contract *model.Contract
	wg.Go(func() error {
		ret, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
		if err != nil {
			log.Errorf("GetContractByID err:%+v", err)
			return serr.ErrBusiness("合约不存在")
		}
		contract = ret
		return nil
	})

	var positions []*model.Position
	wg.Go(func() error {
		ret, err := PositionServiceInstance().GetPositionByContractID(ctx, contractID)
		if err != nil {
			log.Errorf("查询持仓失败:%+v", err)
			return serr.ErrBusiness("合约结算失败")
		}
		positions = ret
		return nil
	})
	if err := wg.Wait(); err != nil {
		return err
	}

	// 1.检查合约状态
	if contract.Status != model.ContractStatusEnable {
		return serr.ErrBusiness("合约非操盘状态")
	}

	// 2.检查合约是否有持仓
	if len(positions) > 0 {
		return serr.ErrBusiness("合约结算失败:未清仓股票")
	}

	user, err := dao.UserDaoInstance().GetUserByUID(ctx, contract.UID)
	if err != nil {
		log.Errorf("GetUserByUID err:%+v", err)
		return serr.ErrBusiness("用户不存在")
	}

	// 设置合约状态 & 转移合约资金到钱包
	tx := db.StockDB().WithContext(ctx).Begin()
	defer tx.Rollback()
	contract.Status = model.ContractStatusDisabled
	contract.CloseTime = time.Now()
	contract.CloseExplain = "主动关闭"

	eg := errgroup.GroupWithCount(3)
	// 设置合约状态
	eg.Go(func() error {
		if err := dao.ContractDaoInstance().UpdateWithTx(tx, contract); err != nil {
			log.Errorf("合约设置状态失败:%+v", err)
			return serr.ErrBusiness("合约结算失败")
		}
		return nil
	})
	// 更新用户资金
	eg.Go(func() error {
		user.Money += contract.Money
		if err := dao.UserDaoInstance().UpdateUserWithTx(tx, user); err != nil {
			log.Errorf("UpdateUserWithTx err:%+v", err)
			return serr.ErrBusiness("合约结算失败")
		}
		return nil
	})
	// 填写合约结算费用
	eg.Go(func() error {
		if err := dao.ContractFeeDaoInstance().CreateWithTx(tx, &model.ContractFee{
			UID:        contract.UID,
			ContractID: contractID,
			Code:       "",
			Name:       "",
			Amount:     0,
			OrderTime:  time.Now(),
			Direction:  model.ContractFeeDirectionIncome,                // 方向:1支出 2:收入
			Money:      contract.Money,                                  // 金额
			Detail:     fmt.Sprintf("合约结算成功,结算资金:%.2f", contract.Money), // 明细
			Type:       model.ContractFeeTypeClose,                      // 费用类型1:买入手续费 2:卖出手续费 3:合约利息 4:卖出盈亏 5:追加保证金 6:扩大资金 7:合约结算
		}); err != nil {
			log.Errorf("合约结算资金填写失败:%+v", err)
			return serr.ErrBusiness("合约结算失败")
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}
	if err := tx.Commit().Error; err != nil {
		log.Errorf("事务提交失败:%+v", err)
		return err
	}

	// 发送消息
	if err := dao.MsgDaoInstance().Create(ctx, &model.Msg{
		UID:        contract.UID,                                                                           // 用户ID
		Title:      fmt.Sprintf("合约[%d]:结算成功", contract.ID),                                                // 标题
		Content:    fmt.Sprintf("合约[%d]:结算成功,结算金额:%0.2f", contract.ID, util.FloatRound(contract.Money, 2)), // 内容
		CreateTime: time.Now(),                                                                             // 创建时间
	}); err != nil {
		log.Errorf("合约结算失败:%+v", err)
		return nil
	}

	if err := dao.TransferDaoInstance().Create(ctx, &model.Transfer{
		UID:       contract.UID,
		OrderTime: time.Now(),
		Money:     contract.Money,
		Type:      model.TransferTypeCloseContract,
		Status:    model.TransferStatusSuccess,
		Name:      "",
		BankNo:    "",
		Channel:   "",
		OrderNo:   "",
	}); err != nil {
		log.Errorf("TransferDaoInstance().Create err:%+v", err)
		return nil
	}

	// 设置用户有效的合约为当前合约
	curContract, err := dao.ContractDaoInstance().GetEnableContractByUID(ctx, contract.UID)
	if err == nil {
		if err := dao.UserDaoInstance().UpdateCurrentContractID(ctx, curContract.UID, curContract.ID); err != nil {
			log.Errorf("更新用户的有效合约失败:%+v", err)
		}
	}

	return nil
}

// AppendMoney 追加保证金
func (s *ContractService) AppendMoney(ctx context.Context, contractID int64, money float64) error {
	// 1. 检查资金是否足够 & 合约状态是否正常
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		log.Errorf("合约不存在:%+v", err)
		return err
	}
	if contract.Status != model.ContractStatusEnable {
		log.Errorf("合约状态非操盘中")
		return serr.ErrBusiness("合约状态错误")
	}

	user, err := dao.UserDaoInstance().GetUserByUID(ctx, contract.UID)
	if err != nil {
		log.Errorf("GetUserByUID err:%+v", err)
		return serr.ErrBusiness("用户不存在")
	}
	if user.Money < money {
		return serr.ErrBusiness("账户余额不足")
	}

	user.Money -= money           // 用户钱包余额资金
	contract.Money += money       // 保证金
	contract.AppendMoney += money // 追加保证金

	tx := db.StockDB().WithContext(ctx).Begin()
	defer tx.Rollback()
	wg := errgroup.GroupWithCount(3)
	wg.Go(func() error {
		if err := dao.ContractDaoInstance().UpdateWithTx(tx, contract); err != nil {
			log.Errorf("更新合约失败:%+v", err)
			return err
		}
		return nil
	})
	wg.Go(func() error {
		if err := dao.UserDaoInstance().UpdateUserWithTx(tx, user); err != nil {
			log.Errorf("更新用户失败:%+v", err)
			return err
		}
		return nil
	})
	wg.Go(func() error {
		if err := dao.ContractFeeDaoInstance().CreateWithTx(tx, &model.ContractFee{
			UID:        contract.UID,
			ContractID: contract.ID,
			OrderTime:  time.Now(),                         // 订单时间
			Direction:  model.ContractFeeDirectionPay,      // 方向:1支出 2:收入
			Money:      money,                              // 金额
			Detail:     fmt.Sprintf("追加保证金:%0.2f元", money), // 明细
			Type:       model.ContractFeeTypeAppendMoney,   // 费用类型1:买入手续费 2:卖出手续费 3:合约利息 4:卖出盈亏 5:追加保证金 6:扩大资金 7:合约结算
		}); err != nil {
			log.Errorf("追加保证金失败:%+v", err)
			return err
		}
		return nil
	})
	if err := wg.Wait(); err != nil {
		return serr.ErrBusiness("追加保证金失败")
	}
	if err := tx.Commit().Error; err != nil {
		log.Errorf("事务提交失败:%+v", err)
		return err
	}

	if err := dao.MsgDaoInstance().Create(ctx, &model.Msg{
		UID:        contract.UID,                                                            // 用户ID
		Title:      "追加保证金成功",                                                               // 标题
		Content:    fmt.Sprintf("合约[%d],追加保证金:%f成功", contractID, util.FloatRound(money, 2)), // 内容
		CreateTime: time.Now(),                                                              // 创建时间
	}); err != nil {
		log.Errorf("创建消息失败:%+v", err)
		return nil
	}

	if err := dao.TransferDaoInstance().Create(ctx, &model.Transfer{
		UID:       contract.UID,
		OrderTime: time.Now(),
		Money:     money,
		Type:      model.TransferTypeAppendMoney,
		Status:    model.TransferStatusSuccess,
		Name:      "",
		BankNo:    "",
		Channel:   "",
		OrderNo:   "",
	}); err != nil {
		log.Errorf("TransferDaoInstance().Create err:%+v", err)
		return nil
	}

	if err := s.UpdateValMoneyByID(ctx, contract.ID); err != nil {
		log.Errorf("更新合约资金失败:%+v", err)
	}

	return nil
}

// GetAppendMoney 追加保证金页面初始化
func (s *ContractService) GetAppendMoney(ctx context.Context, contractID int64) (*model.AppendExpandContract, error) {
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		return nil, err
	}
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, contract.UID)
	if err != nil {
		return nil, err
	}

	return &model.AppendExpandContract{
		ID:    contractID,          // 合约ID
		Name:  contract.FullName(), // 合约名称
		Money: user.Money,          // 可追加的保证金
	}, nil
}

// ExpandMoney 扩大合约资金
func (s *ContractService) ExpandMoney(ctx context.Context, contractID int64, money float64) error {
	// 1. 检查资金是否足够 & 合约状态是否正常
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		log.Errorf("合约不存在:%+v", err)
		return err
	}
	if contract.Status != model.ContractStatusEnable {
		log.Errorf("合约状态非操盘中")
		return serr.ErrBusiness("合约状态错误")
	}

	user, err := dao.UserDaoInstance().GetUserByUID(ctx, contract.UID)
	if err != nil {
		log.Errorf("GetUserByUID err:%+v", err)
		return serr.ErrBusiness("用户不存在")
	}
	if user.Money < money {
		return serr.ErrBusiness("账户余额不足")
	}

	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		log.Errorf("GetSysParam err:%+v", err)
		return serr.ErrBusiness("网络错误")
	}
	// 扣取用户的资金
	user.Money -= money
	contract.InitMoney += money
	contract.Money += money
	// 扣取利息
	interest := model.Interest(contract, sys, money)
	contract.Money -= interest
	tx := db.StockDB().WithContext(ctx).Begin()
	defer tx.Rollback()
	wg := errgroup.GroupWithCount(3)
	wg.Go(func() error {
		if err := dao.ContractDaoInstance().UpdateWithTx(tx, contract); err != nil {
			log.Errorf("UpdateWithTx更新合约失败:%+v", err)
			return err
		}
		return nil
	})
	wg.Go(func() error {
		if err := dao.UserDaoInstance().UpdateUserWithTx(tx, user); err != nil {
			log.Errorf("更新用户失败:%+v", err)
			return err
		}
		return nil
	})
	wg.Go(func() error {
		if err := dao.ContractFeeDaoInstance().CreateWithTx(tx, &model.ContractFee{
			UID:        contract.UID,
			ContractID: contract.ID,
			Code:       "",
			Name:       "",
			Amount:     0,
			OrderTime:  time.Now(),
			Direction:  model.ContractFeeDirectionPay,
			Money:      money,
			Detail:     fmt.Sprintf("%s[%d]扩大资金成功:%0.2f元,收取扩大资金利息费用:%0.2f元", contract.FullName(), contract.ID, money, util.FloatRound(interest, 2)),
			Type:       model.ContractFeeTypeExpandMoney,
		}); err != nil {
			log.Errorf("CreateWithTx err:%+v", err)
			return err
		}
		return nil
	})
	if err := wg.Wait(); err != nil {
		return serr.ErrBusiness("扩大保证金失败")
	}
	if err := tx.Commit().Error; err != nil {
		log.Errorf("事务提交失败:%+v", err)
		return err
	}

	if err := dao.MsgDaoInstance().Create(ctx, &model.Msg{
		UID:        contract.UID,                                                                                                             // 用户ID
		Title:      "扩大资金成功",                                                                                                                 // 标题
		Content:    fmt.Sprintf("合约[%d],扩大保证金:%f元成功,收取扩大资金利息费用:%0.2f元", contractID, util.FloatRound(money, 2), util.FloatRound(interest, 2)), // 内容
		CreateTime: time.Now(),                                                                                                               // 创建时间
	}); err != nil {
		log.Errorf("创建消息失败:%+v", err)
		return err
	}

	if err := dao.TransferDaoInstance().Create(ctx, &model.Transfer{
		UID:       contract.UID,
		OrderTime: time.Now(),
		Money:     money,
		Type:      model.TransferTypeExpandMoney,
		Status:    model.TransferStatusSuccess,
		Name:      "",
		BankNo:    "",
		Channel:   "",
		OrderNo:   "",
	}); err != nil {
		log.Errorf("TransferDaoInstance().Create err:%+v", err)
		return nil
	}

	if err := s.UpdateValMoneyByID(ctx, contract.ID); err != nil {
		log.Errorf("更新合约资金失败:%+v", err)
	}
	return nil
}

// HisContract 查询历史合约
func (s *ContractService) HisContract(ctx context.Context, uid int64) ([]*model.HisContract, error) {
	contracts, err := dao.ContractDaoInstance().GetContractsByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	result := make([]*model.HisContract, 0)
	if len(contracts) == 0 {
		return result, nil
	}
	// 结算合约ID
	closeIDs := make([]int64, 0)
	for _, it := range contracts {
		if it.Status == model.ContractStatusDisabled {
			closeIDs = append(closeIDs, it.ID)
		}
	}
	// 没有结算的合约则直接返回
	if len(closeIDs) == 0 {
		return result, nil
	}
	list, err := dao.ContractFeeDaoInstance().GetContractFeeByIDs(ctx, closeIDs)
	if err != nil {
		log.Errorf("GetContractFeeByIDs 查询合约费用失败,err:%+v", err)
		return nil, err
	}
	contractFeeMap := make(map[int64][]*model.ContractFee) // map[contract.ID][[]*model.ContractFee
	for _, it := range list {
		contractFeeList, ok := contractFeeMap[it.ContractID]
		if !ok {
			contractFeeList = make([]*model.ContractFee, 0)
		}
		contractFeeList = append(contractFeeList, it)
		contractFeeMap[it.ContractID] = contractFeeList
	}

	for _, it := range contracts {
		if it.Status != model.ContractStatusDisabled {
			continue
		}
		contractFee, ok := contractFeeMap[it.ID]
		if !ok {
			log.Errorf("结算合约获取费用失败,合约ID:[%d]", it.ID)
			continue
		}
		// 回收资金
		recoverMoney := 0.00
		for _, fee := range contractFee {
			// 关闭合约收回资金
			if fee.Type == model.ContractFeeTypeClose {
				recoverMoney += it.Money
			}
			// 提盈收回资金
			if fee.Type == model.ContractFeeTypeGetProfit {
				recoverMoney += it.Money
			}
		}

		investMoney := it.InitMoney + it.AppendMoney
		result = append(result, &model.HisContract{
			ID:           it.ID,                                        // 合约id
			Name:         it.FullName(),                                // 合约名称
			InvestMoney:  investMoney,                                  // 投入资金:初始资金+扩大合约资金
			RecoverMoney: recoverMoney,                                 // 收回资金
			Profit:       util.FloatRound(recoverMoney-investMoney, 2), // 盈亏金额:投资资金 - 回收资金
			CreateTime:   it.OrderTime.Format("2006-01-02 15:04:05"),   // 创建时间
			CloseTime:    it.CloseTime.Format("2006-01-02 15:04:05"),   // 结算时间
		})
	}
	return result, nil
}

// GetContract 切换合约时,查询有效的合约列表
func (s *ContractService) GetContract(ctx context.Context, uid int64) ([]*model.GetContract, error) {
	wg := errgroup.GroupWithCount(2)
	var contracts []*model.Contract
	wg.Go(func() error {
		ret, err := dao.ContractDaoInstance().GetContractsByUID(ctx, uid)
		if err != nil {
			return err
		}
		contracts = ret
		return nil
	})
	var user *model.User
	wg.Go(func() error {
		ret, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
		if err != nil {
			return err
		}
		user = ret
		return nil
	})
	if err := wg.Wait(); err != nil {
		return nil, serr.ErrBusiness("查询合约失败")
	}

	if len(contracts) == 0 {
		return make([]*model.GetContract, 0), nil
	}
	result := make([]*model.GetContract, 0)
	for _, it := range contracts {
		if it.Status != model.ContractStatusEnable {
			continue
		}
		result = append(result, &model.GetContract{
			ID:       it.ID,                           // 合约id
			Name:     it.FullName(),                   // 合约名称
			Money:    it.Money,                        // 保证金
			ValMoney: it.ValMoney,                     // 可用资金
			Select:   it.ID == user.CurrentContractID, // 当前选中合约
		})
	}
	return result, nil
}

// Select 选中合约
func (s *ContractService) Select(ctx context.Context, uid, contractID int64) error {
	// 检查合约是否存在
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		return err
	}
	if contract.Status != model.ContractStatusEnable {
		return serr.ErrBusiness("非有效合约")
	}
	return dao.UserDaoInstance().UpdateCurrentContractID(ctx, uid, contractID)
}

// GetTodayProfitByContractID 当日盈亏=当前价格*当前持仓数量 - 昨日收盘价*昨日持仓数量+当日卖出金额（含费）-当日买入金额（含费）
func (s *ContractService) GetTodayProfitByContractID(ctx context.Context, contractID int64) (float64, error) {
	wg := errgroup.GroupWithCount(4)
	positions := make([]*model.Position, 0)
	// 查询合约持仓
	wg.Go(func() error {
		ret, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contractID)
		if err != nil {
			return err
		}
		positions = ret
		return nil
	})
	// 查询今日委托
	entrusts := make([]*model.Entrust, 0)
	wg.Go(func() error {
		ret, err := dao.EntrustDaoInstance().GetTodayEntrust(ctx, contractID)
		if err != nil {
			return err
		}
		entrusts = ret
		return nil
	})
	// 查询昨日持仓
	yesterdayPosition := make([]*model.Position, 0)
	wg.Go(func() error {
		ret, err := dao.HisPositionDaoInstance().GetYesterdayPositionByContractID(ctx, contractID)
		if err != nil {
			return err
		}
		yesterdayPosition = ret
		return nil
	})
	// 今日是否为交易日
	isTradeDate := false
	wg.Go(func() error {
		if ok := CalendarServiceInstance().IsTradeDate(ctx); ok {
			isTradeDate = true
		}
		return nil
	})
	if err := wg.Wait(); err != nil {
		return 0.00, err
	}
	// 非交易日 || 无持仓数据,则返回盈利金额为0
	if !isTradeDate || (len(positions) == 0 && len(entrusts) == 0) {
		return 0.00, nil
	}
	// 在09:25分前面,当日盈亏为0
	now := time.Now()
	if now.Hour() < 9 || (now.Hour() == 9 && now.Minute() < 25) {
		return 0.00, nil
	}
	// 获取行情
	codes := make([]string, 0)
	for _, it := range entrusts {
		codes = append(codes, it.StockCode)
	}
	for _, it := range positions {
		codes = append(codes, it.StockCode)
	}
	qts, err := quote.QtServiceInstance().GetQuoteByTencent(codes)
	if err != nil {
		return 0.00, err
	}

	profit := 0.00
	// 当前价格*当前持仓数量
	for _, it := range positions {
		profit += qts[it.StockCode].CurrentPrice * float64(it.Amount)
	}
	// 昨日收盘价*昨日持仓数量
	for _, it := range yesterdayPosition {
		qt, ok := qts[it.StockCode]
		if !ok {
			continue
		}
		profit -= qt.ClosePrice * float64(it.Amount)
	}
	// + 当日卖出金额（含费） && -当日买入金额（含费）
	for _, it := range entrusts {
		if it.Status == model.EntrustStatusTypeDeal && it.EntrustBS == model.EntrustBsTypeBuy {
			profit -= it.Balance
		} else if it.Status == model.EntrustStatusTypeDeal && it.EntrustBS == model.EntrustBsTypeSell {
			profit += it.Balance
		}
	}
	return profit, nil
}

// GetWithdrawProfit 查询合约提盈
func (s *ContractService) GetWithdrawProfit(ctx context.Context, contractID int64) (*model.AppendExpandContract, error) {
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		return nil, err
	}
	result := &model.AppendExpandContract{
		ID:    contractID,
		Name:  contract.FullName(),
		Money: 0.00,
	}
	list, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contractID)
	if err != nil {
		return nil, err
	}
	// 有持仓则不允许提盈
	if len(list) > 0 {
		return result, nil
	}
	if contract.Money < contract.InitMoney {
		return result, nil
	}
	// 提盈金额 = 合约保证金 - 初始资金
	result.Money = contract.Money - contract.InitMoney
	return result, nil
}

// WithdrawProfit 合约提盈
func (s *ContractService) WithdrawProfit(ctx context.Context, contractID int64, money float64) error {
	list, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contractID)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		return serr.ErrBusiness("合约持有股票,提盈失败")
	}
	contract, err := dao.ContractDaoInstance().GetContractByID(ctx, contractID)
	if err != nil {
		return err
	}
	if money > contract.Money-contract.InitMoney {
		return serr.ErrBusiness("提盈金额大于可提取金额")
	}
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, contract.UID)
	if err != nil {
		log.Errorf("GetUserByUID err:%+v", err)
		return serr.ErrBusiness("提盈失败:用户不存在")
	}

	tx := db.StockDB().WithContext(ctx).Begin()
	defer tx.Rollback()

	wg := errgroup.GroupWithCount(4)
	// 扣除合约保证金,更新合约
	contract.Money = contract.Money - money
	wg.Go(func() error {
		if err := dao.ContractDaoInstance().UpdateWithTx(tx, contract); err != nil {
			log.Errorf("UpdateContract err:%+v", err)
			return err
		}
		return nil
	})

	user.Money += money
	wg.Go(func() error {
		if err := dao.UserDaoInstance().UpdateUserWithTx(tx, user); err != nil {
			log.Errorf("更新用户资金失败:%+v", err)
			return err
		}
		return nil
	})

	// 填写合约费用表
	wg.Go(func() error {
		if err := dao.ContractFeeDaoInstance().CreateWithTx(tx, &model.ContractFee{
			UID:        contract.UID,
			ContractID: contract.ID,
			Code:       "",
			Name:       "",
			Amount:     0,
			OrderTime:  time.Now(),
			Direction:  model.ContractFeeDirectionIncome,
			Money:      money,
			Detail:     fmt.Sprintf("%s[%d]提盈成功,金额:%0.2f元", contract.FullName(), contract.ID, money),
			Type:       model.ContractFeeTypeGetProfit,
		}); err != nil {
			log.Errorf("Create contract fee err:%+v", err)
			return err
		}
		return nil
	})

	// 填写msg表
	wg.Go(func() error {
		if err := dao.MsgDaoInstance().CreateWithTx(tx, &model.Msg{
			UID:        contract.UID,
			Title:      "提盈成功",
			Content:    fmt.Sprintf("%s[%d]提盈成功,提取金额:%0.2f元,请留意资金变动。", contract.FullName(), contract.ID, money),
			CreateTime: time.Now(),
		}); err != nil {
			log.Errorf("Create msg err :%+v", err)
			return serr.ErrBusiness("提盈失败")
		}
		return nil
	})

	if err := wg.Wait(); err != nil {
		return err
	}
	if err := tx.Commit().Error; err != nil {
		log.Errorf("事务提交失败:%+v", err)
		return err
	}

	if err := s.UpdateValMoneyByID(ctx, contract.ID); err != nil {
		log.Errorf("更新合约资金失败:%+v", err)
	}

	if err := dao.TransferDaoInstance().Create(ctx, &model.Transfer{
		UID:       contract.UID,
		OrderTime: time.Now(),
		Money:     money,
		Type:      model.TransferTypeGetMoney,
		Status:    model.TransferStatusSuccess,
		Name:      "",
		BankNo:    "",
		Channel:   "",
		OrderNo:   "",
	}); err != nil {
		log.Errorf("TransferDaoInstance().Create err:%+v", err)
		return nil
	}

	return nil
}

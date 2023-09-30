package service

import (
	"context"
	"encoding/json"
	"fmt"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/api-gateway/util"
	"stock/common/log"
	"stock/common/timeconv"
	"sync"
	"time"
)

// DividendService 分红派息服务
type DividendService struct {
}

var (
	dividendService *DividendService
	dividendOnce    sync.Once
)

// DividendServiceInstance DividendServiceInstance实例
func DividendServiceInstance() *DividendService {
	dividendOnce.Do(func() {
		dividendService = &DividendService{}
		ctx := context.Background()
		go func() {
			for range time.Tick(10 * time.Second) {
				if err := dividendService.load(ctx); err != nil {
					log.Errorf("getDividendFromEastMoney err:%+v", err)
					continue
				}
			}
		}()
	})
	return dividendService
}

type eastMoneyDividendItem struct {
	StockCode    string   `json:"SECUCODE"`           // 证券代码,带后缀
	Ratio        *float64 `json:"BONUS_IT_RATIO"`     // 送股数量:10股送x股
	RMB          *float64 `json:"PRETAX_BONUS_RMB"`   // 10股派息x元
	DividendDate string   `json:"EQUITY_RECORD_DATE"` // 股权登记日
	Plan         string   `json:"IMPL_PLAN_PROFILE"`  // 分红送股方案
}

type eastMoneyDividend struct {
	Result struct {
		Pages int                      `json:"pages"`
		List  []*eastMoneyDividendItem `json:"data"`
	} `json:"result"`
	Code int `json:"code"`
}

// getDividendFromEastMoney 根据url获取东财分红信息
func (s *DividendService) getDividendFromEastMoney(ctx context.Context, url string) (*eastMoneyDividend, error) {
	resp, err := util.Http(url)
	if err != nil {
		log.Errorf("http get url:%+v err:%+v", url, err)
		return nil, err
	}
	var item eastMoneyDividend
	if err := json.Unmarshal([]byte(resp), &item); err != nil {
		log.Errorf("unmarshal err:%+v", err)
		return nil, err
	}
	if item.Code != 0 {
		log.Errorf("请求分红配送接口失败,请求结果:%+v", resp)
		return nil, serr.ErrBusiness("请求成功,但code解析失败")
	}

	return &item, nil
}

func (s *DividendService) cacheKey() string {
	return fmt.Sprintf("dividend_cache_key_date:%+v", timeconv.TimeToInt32(time.Now()))
}

// load load
func (s *DividendService) load(ctx context.Context) error {
	// 下午4点开始检查分红
	if time.Now().Hour() < 16 {
		return nil
	}

	// redis 检查:今日是否已经检查过分红
	if db.Get(ctx, s.cacheKey()).Val() == "1" {
		return nil
	}

	list, err := dao.PositionDaoInstance().GetPositions(ctx)
	if err != nil {
		log.Errorf("GetPositions err:%+v", err)
		return err
	}

	// 查询持仓
	for _, position := range list {
		url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?callback=&sortColumns=REPORT_DATE&sortTypes=-1&pageSize=50&pageNumber=1&reportName=RPT_SHAREBONUS_DET&columns=ALL&quoteColumns=&source=WEB&client=WEB&filter=(SECURITY_CODE=%s)", position.StockCode)
		dividends, err := s.getDividendFromEastMoney(ctx, url)
		if err != nil {
			log.Errorf("获取股票:%+v 分红除权除息失败:%+v", position.StockCode, err)
			continue
		}

		for _, item := range dividends.Result.List {
			dividendTime, err := time.Parse("2006-01-02 15:04:05", item.DividendDate)
			if err != nil {
				log.Errorf("解析时间错误:%+v", err)
				continue
			}

			if timeconv.TimeToInt32(dividendTime) != timeconv.TimeToInt32(time.Now()) {
				continue
			}

			// 分红除权除息
			if err := s.dividend(ctx, position, item); err != nil {
				log.Errorf("dividend err:%+v", err)
			}
		}
	}

	// redis 检查:今日是否已经检查过分红
	if err := db.Set(ctx, s.cacheKey(), "1", 7*24*time.Hour).Err(); err != nil {
		log.Errorf("设置redis失败:%+v", err)
		panic("设置redis失败")
	}

	return nil
}

func (s *DividendService) dividend(ctx context.Context, position *model.Position, it *eastMoneyDividendItem) error {
	if len(it.DividendDate) == 0 {
		return nil
	}

	var dividendType int64
	dividendAmount := 0.00 // 送股数量:10股送x股
	if it.Ratio != nil {
		dividendAmount = *it.Ratio * (float64(position.Amount) / 10)
		dividendType = model.DividendTypeStockConversion
	}
	dividendMoney := 0.00 // 10股派息金额
	if it.RMB != nil {
		dividendMoney = *it.RMB * (float64(position.Amount) / 10)
		dividendType = model.DividendTypeBonusShare
	}
	if it.Ratio != nil && it.RMB != nil {
		dividendType = model.DividendTypeBonusShareAndStockConversion
	}

	log.Infof("股票代码:%+v 送股数量:%+v 派息金额:%+v 登记日:%+v", it.StockCode, dividendAmount, dividendMoney, it.DividendDate)

	// 1. 填写dividend表
	if err := dao.DividendDaoInstance().Create(ctx, &model.Dividend{
		UID:            position.UID,
		ContractID:     position.ContractID,
		PositionID:     position.ID,
		OrderTime:      time.Now(),
		StockCode:      position.StockCode,
		StockName:      position.StockName,
		PositionPrice:  position.Price,
		PositionAmount: position.Amount,
		DividendMoney:  dividendMoney,
		DividendAmount: int64(dividendAmount),
		Type:           dividendType,
		PlanExplain:    it.Plan,
	}); err != nil {
		log.Errorf("Create err:%+v", err)
		return err
	}
	if dividendType == model.DividendTypeStockConversion || dividendType == model.DividendTypeBonusShareAndStockConversion {
		// 2. 送股分红,则填写持仓表
		position.Amount += int64(dividendAmount)
		if err := dao.PositionDaoInstance().Update(ctx, position); err != nil {
			log.Errorf("送股分红增加股票失败:%+v", err)
			return err
		}
	} else if dividendType == model.DividendTypeBonusShare || dividendType == model.DividendTypeBonusShareAndStockConversion {
		// 2.现金分红,汇入合约
		contract, err := dao.ContractDaoInstance().GetContractByID(ctx, position.ContractID)
		if err != nil {
			log.Errorf("GetContractByID err:%+v", err)
			return err
		}
		contract.Money += dividendMoney
		if err := dao.ContractDaoInstance().UpdateContract(ctx, contract); err != nil {
			log.Errorf("UpdateContract err:%+v", err)
			return err
		}
	}
	// 3.填写msg表
	if err := dao.MsgDaoInstance().Create(ctx, &model.Msg{
		UID:   position.UID,
		Title: "分红送配",
		Content: fmt.Sprintf("您的持仓%s(%s)今日(%s)分红送配,方案:%s,请留意您的账户资金分红。",
			position.StockName, position.StockCode, time.Now().Format("01月-02日"), it.Plan),
		CreateTime: time.Now(),
	}); err != nil {
		log.Errorf("Create err:%+v", err)
	}

	// 4. 短信通知
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, position.UID)
	if err != nil {
		log.Errorf("GetUserByUID err:%+v", err)
		return err
	}
	if err := SmsServiceInstance().SendSms(ctx,
		fmt.Sprintf("您的持仓%s(%s)今日(%s)分红送配,方案:%s,请留意您的账户资金分红。",
			position.StockName, position.StockCode, time.Now().Format("01月-02日"), it.Plan),
		user.UserName); err != nil {
		log.Errorf("短信发送失败")
	}

	return nil
}

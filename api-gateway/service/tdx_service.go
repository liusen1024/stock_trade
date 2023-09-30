package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"stock/api-gateway/dao"
	"stock/api-gateway/model"
	"stock/api-gateway/util"
	"stock/common/log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type tdxEntrustPropType int64
type tdxEntrustBsType int64

const (
	tdxEntrustPropTypeLimit         tdxEntrustPropType = 0 // 深圳限价,上海限价
	tdxEntrustPropTypeSZMarketPrice tdxEntrustPropType = 1 // 深圳市价
	tdxEntrustPropTypeSHMarketPrice tdxEntrustPropType = 6 // 上海市价

	tdxEntrustBsTypeBuy  tdxEntrustBsType = 0 // 买入
	tdxEntrustBsTypeSell tdxEntrustBsType = 1 // 卖出
)

// TDXService 通达信服务
type TDXService struct {
}

var (
	tdxService *TDXService
	tdxOnce    sync.Once
)

// TDXServiceInstance 实例
func TDXServiceInstance() *TDXService {
	tdxOnce.Do(func() {
		tdxService = &TDXService{}
	})
	return tdxService
}

// Login 通达信券商登录:登录成功返回客户id
func (s *TDXService) Login(broker *model.Broker) (int64, error) {
	brokerHost, err := util.BrokerHost()
	if err != nil {
		return 0, err
	}
	loginURL := fmt.Sprintf("http://%s/login?ip=%s&port=%d&version=%s&branch_no=%d&fund_account=%s&trade_account=%s&trade_password=%s&tx_password=%s",
		brokerHost, broker.IP, broker.Port, broker.Version, broker.BranchNo, broker.FundAccount, broker.TradeAccount, broker.TradePassword, broker.TxPassword)
	resp, err := util.HttpWithTimeout(loginURL, 60*time.Second)
	if err != nil {
		log.Errorf("HttpWithTimeout err:%+v", err)
		s.logErr(loginURL, err)
		return 0, err
	}
	item := struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}{}
	if err := json.Unmarshal([]byte(resp), &item); err != nil {
		log.Errorf("Unmarshal err:%+v", err)
		return 0, err
	}
	if len(item.Error) > 0 {
		log.Errorf("login err:%+v", err)
		return 0, errors.New("登录失败")
	}
	return strconv.ParseInt(item.Result, 10, 64)
}

// getHolderAccount 根据股票代码返回股东代码
func (s *TDXService) getHolderAccount(broker *model.Broker, stockCode string) string {
	if code, _ := strconv.ParseInt(stockCode, 10, 64); code >= 600000 {
		return broker.SHHolderAccount
	}
	return broker.SZHolderAccount
}

// Entrust 委托
func (s *TDXService) Entrust(entrust *model.BrokerEntrust) error {
	host, err := util.BrokerHost()
	if err != nil {
		return err
	}

	entrustBs := tdxEntrustBsTypeBuy
	if entrust.EntrustBs == model.EntrustBsTypeSell {
		entrustBs = tdxEntrustBsTypeSell
	}

	entrustProp := tdxEntrustPropTypeLimit                       // 限价
	if entrust.EntrustProp == model.EntrustPropTypeMarketPrice { // 市价委托
		if code, _ := strconv.ParseInt(entrust.StockCode, 10, 64); code >= 600000 {
			entrustProp = tdxEntrustPropTypeSHMarketPrice // 上海市价
		} else {
			entrustProp = tdxEntrustPropTypeSZMarketPrice // 深圳市价
		}
	}

	// type:0买入,1卖出  priceType:0限价,1市价 gddm:上海|深圳股东代码  price:委托价格
	req := fmt.Sprintf("http://%s/send_order?client_id=%d&type=%d&price_type=%d&gddm=%s&stock_code=%s&price=%0.2f&amount=%d",
		host, entrust.Broker.ClientID, entrustBs, entrustProp, s.getHolderAccount(entrust.Broker, entrust.StockCode), entrust.StockCode, entrust.EntrustPrice, entrust.EntrustAmount)
	res, err := s.requestTDX(req)
	if err != nil {
		log.Errorf("requestTDX err:%+v", err)
		s.logErr(req, err)
		return err
	}

	// 券商委托编号
	if len(res) > 0 && len(res[1]) > 0 {
		entrust.BrokerEntrustNo = res[1][0]
	}

	return nil
}

func (s *TDXService) requestTDX(url string) ([][]string, error) {
	resp, err := util.HttpWithTimeout(url, 30*time.Second)
	if err != nil {
		log.Errorf("HttpWithTimeout err:%+v", err)
		return nil, errors.New("http请求超时")
	}
	resp = util.ConvertToString(resp, "gbk", "utf-8")
	resp = strings.ReplaceAll(resp, "\n", "|")
	resp = strings.ReplaceAll(resp, "\t", "_")
	item := struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}{}
	if err := json.Unmarshal([]byte(resp), &item); err != nil {
		log.Errorf("Unmarshal err:%+v", err)
		return nil, errors.New(resp)
	}
	if len(item.Error) > 0 {
		log.Errorf("item err:%+v", item.Error)
		return nil, errors.New(item.Error)
	}
	columns := make([][]string, 0)
	for _, column := range strings.Split(item.Result, "|") {
		list := make([]string, 0)
		for _, it := range strings.Split(column, "_") {
			list = append(list, it)
		}
		columns = append(columns, list)
	}
	return columns, nil
}

// logErr 记录错误日志
func (s *TDXService) logErr(url string, err error) {
	ctx := context.Background()
	if e := dao.BrokerErrorLogDaoInstance().Create(ctx, &model.BrokerErrorLog{
		URL:   url,
		Error: err.Error(),
	}); e != nil {
		log.Errorf("Create err:%+v", err)
	}
}

// query 查询
func (s *TDXService) query(queryType model.TDXQueryType, broker *model.Broker) ([][]string, error) {
	host, err := util.BrokerHost()
	if err != nil {
		return nil, err
	}
	req := fmt.Sprintf("http://%s/query?client_id=%d&type=%d", host, broker.ClientID, queryType)
	res, err := s.requestTDX(req)
	if err != nil {
		s.logErr(req, err)
		return nil, err
	}
	return res, nil
}

// QueryFund 查询资金
func (s *TDXService) QueryFund(broker *model.Broker) (*model.TDXBrokerFund, error) {
	res, err := s.query(model.TDXQueryTypeFund, broker)
	if err != nil {
		return nil, err
	}
	itemMap := make(map[string]string, 0)
	for index, items := range res {
		for col, it := range items {
			if index == 0 {
				itemMap[it] = ""
			} else {
				itemMap[res[0][col]] = it
			}
		}
	}
	// 可用资金
	var valMoney float64
	if val, ok := itemMap["可用资金"]; ok {
		valMoney, err = strconv.ParseFloat(val, 10)
		if err != nil {
			return nil, err
		}
	}
	// 总资产
	var asset float64
	if val, ok := itemMap["总资产"]; ok {
		asset, err = strconv.ParseFloat(val, 10)
		if err != nil {
			return nil, err
		}
	}
	// 市值
	var marketVal float64
	if v, ok := itemMap["最新市值"]; ok {
		marketVal, err = strconv.ParseFloat(v, 10)
		if err != nil {
			return nil, err
		}
	}
	return &model.TDXBrokerFund{
		ClientID:    broker.ClientID,
		FundAccount: broker.FundAccount,
		ValMoney:    valMoney,
		Asset:       asset,
		MarketValue: marketVal,
	}, nil
}

// QueryTodayEntrust 查询今日委托
func (s *TDXService) QueryTodayEntrust(broker *model.Broker) ([]*model.TDXTodayEntrust, error) {
	res, err := s.query(model.TDXQueryTypeTodayEntrust, broker)
	if err != nil {
		return nil, err
	}
	return model.ParseTdxTodayEntrust(broker, res), nil
}

// QueryPosition 查询持仓
func (s *TDXService) QueryPosition(broker *model.Broker) ([]*model.TDXPosition, error) {
	res, err := s.query(model.TDXQueryTypePosition, broker)
	if err != nil {
		return nil, err
	}
	positions := make([]*model.TDXPosition, 0)
	for col, items := range res {
		// 跳过表头
		if col == 0 {
			continue
		}
		position := &model.TDXPosition{
			ClientID:    broker.ClientID,
			FundAccount: broker.FundAccount,
		}
		for index, v := range items {
			switch index {
			case 0: // 证券代码
				position.StockCode = v
			case 1: // 证券名称
				position.StockName = v
			case 2: // 证券数量
				if strings.Contains(v, ".") {
					amount, _ := strconv.ParseFloat(v, 64)
					position.Amount = int64(amount)
				} else {
					position.Amount, _ = strconv.ParseInt(v, 10, 64)
				}
			case 3: // 可卖数量
				if strings.Contains(v, ".") {
					amount, _ := strconv.ParseFloat(v, 64)
					position.FreezeAmount = position.Amount - int64(amount)
				} else {
					position.FreezeAmount, err = strconv.ParseInt(v, 10, 64)
					position.FreezeAmount = position.Amount - position.FreezeAmount
				}
			case 4: // 最新价
				position.CurrentPrice, _ = strconv.ParseFloat(v, 10)
			case 6: // 成本价
				position.PositionPrice, _ = strconv.ParseFloat(v, 10)
			}
		}
		// 跳过空数据
		if len(position.StockCode) == 0 {
			continue
		}
		positions = append(positions, position)
	}
	return positions, nil
}

// QueryWithdraw 查询可撤单
func (s *TDXService) QueryWithdraw(broker *model.Broker) ([]*model.TDXWithdraw, error) {
	res, err := s.query(model.TDXQueryTypeWithdraw, broker)
	if err != nil {
		return nil, err
	}
	list := make([]*model.TDXWithdraw, 0)
	for col, items := range res {
		// 跳过表头
		if col == 0 {
			continue
		}
		item := &model.TDXWithdraw{
			ClientID:    broker.ClientID,
			FundAccount: broker.FundAccount,
		}
		for index, v := range items {
			switch index {
			case 0: // 委托时间
				item.EntrustTime, _ = time.Parse("15:04:05", v)
			case 1: // 证券代码
				item.StockCode = v
			case 2: // 证券名称
				item.StockName = v
			case 4: // 买卖标志
				item.EntrustBs = model.EntrustBsTypeBuy
				if v == "1" {
					item.EntrustBs = model.EntrustBsTypeSell
				}
			case 6: // 委托价格
				item.EntrustPrice, _ = strconv.ParseFloat(v, 10)
			case 7: // 委托数量
				item.EntrustAmount, _ = strconv.ParseInt(v, 10, 64)
			case 8: // 委托编号
				item.EntrustNo = v
			case 9: // 成交数量
				item.DealAmount, _ = strconv.ParseInt(v, 10, 64)
			}
		}
		// 过滤掉空的数据
		if len(item.StockCode) == 0 {
			continue
		}
		list = append(list, item)
	}
	return list, nil
}

// CancelOrder 通达信撤单
func (s *TDXService) CancelOrder(entrust *model.BrokerEntrust, broker *model.Broker, entrustNo string) error {
	host, err := util.BrokerHost()
	if err != nil {
		return err
	}
	// 交易所:1上海 0深圳
	exchange := 0
	if code, _ := strconv.ParseInt(entrust.StockCode, 10, 64); code >= 600000 {
		exchange = 1 // 上海
	}

	req := fmt.Sprintf("http://%s/cancel_order?client_id=%d&exchange_id=%d&entrust_no=%s", host, broker.ClientID, exchange, entrustNo)
	if _, err := s.requestTDX(req); err != nil {
		log.Errorf("requestTDX err:%+v", err)
		s.logErr(req, err)
		return err
	}
	return nil
}

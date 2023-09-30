package service

import (
	"context"
	"stock/api-gateway/model"
	"stock/api-gateway/util"
	"sync"
)

// HomeService 首页服务
type HomeService struct {
}

var (
	homeService *HomeService
	homeOnce    sync.Once
)

const cacheKey = "home_indicator_cache"

// HomeServiceInstance 首页
func HomeServiceInstance() *HomeService {
	homeOnce.Do(func() {
		homeService = &HomeService{}
	})
	return homeService
}

// GetHome 首页
func (s *HomeService) GetHome(ctx context.Context) (interface{}, error) {
	data := DataServiceInstance().GetData()
	temperature := 0.5
	if data.MarketOverView.UpAmount != 0 && data.MarketOverView.DownAmount != 0 {
		temperature = float64(data.MarketOverView.UpAmount) / float64(data.MarketOverView.UpAmount+data.MarketOverView.DownAmount)
	}
	return &model.Home{
		UpAmount:               data.MarketOverView.UpAmount,
		FlatAmount:             data.MarketOverView.FlatAmount,                 // 平家数
		DownAmount:             data.MarketOverView.DownAmount,                 // 下跌家数
		MarketTemperature:      util.FloatRound(temperature, 4),                // 市场热度
		LimitUpCount:           data.XGBMarketIndicator.LimitUpCount,           // 涨停
		LimitDownCount:         data.XGBMarketIndicator.LimitDownCount,         // 跌停
		LimitUpBrokenRatio:     data.XGBMarketIndicator.LimitUpBrokenRatio,     // 炸板率
		LimitUpBrokenCount:     data.XGBMarketIndicator.LimitUpBrokenCount,     // 炸板数量
		YesterdayLimitUpAvgPcp: data.XGBMarketIndicator.YesterdayLimitUpAvgPcp, // 昨涨停今表现
		HotStock:               data.HotStock[0:8],                             // 热门股票
		UpRank:                 data.Rank.UpRank,                               // 涨幅榜
		DownRank:               data.Rank.DownRank,                             // 跌幅榜
		TurnoverRank:           data.Rank.TurnoverRank,                         // 换手榜
		VolumeRank:             data.Rank.VolumeRank,                           // 成交榜
		NetInBalanceRank:       data.Rank.NetInBalanceRank,
	}, nil
}

func (s *HomeService) AllHotStock(ctx context.Context) (interface{}, error) {
	data := DataServiceInstance().GetData()
	return map[string]interface{}{
		"list": data.HotStock,
	}, nil
}

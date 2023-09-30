package handler

// MinuteKline 雪球分时kline数据
type MinuteKline struct {
	Data []MinuteKlineItem `json:"data"`
}

type MinuteKlineItem struct {
	Current   float64     `json:"current"`
	Volume    int         `json:"volume"`
	AvgPrice  float64     `json:"avg_price"`
	Chg       float64     `json:"chg"`
	Percent   float64     `json:"percent"`
	Timestamp int64       `json:"timestamp"`
	Amount    float64     `json:"amount"`
	High      float64     `json:"high"`
	Low       float64     `json:"low"`
	Macd      interface{} `json:"macd"`
	Kdj       interface{} `json:"kdj"`
	Ratio     interface{} `json:"ratio"`
	Capital   *struct {
		Small  float64 `json:"small"`
		Medium float64 `json:"medium"`
		Large  float64 `json:"large"`
		Xlarge float64 `json:"xlarge"`
	} `json:"capital"`
	VolumeCompare struct {
		VolumeSum     int `json:"volume_sum"`
		VolumeSumLast int `json:"volume_sum_last"`
	} `json:"volume_compare"`
}

type MinuteKlineResp struct {
	Data      []*MinuteKlineItem `json:"data"`
	Code      string             `json:"code"`
	LastClose float64            `json:"last_close"`
}

type T struct {
	Data struct {
		Symbol string      `json:"symbol"`
		Column []string    `json:"column"`
		Item   [][]float64 `json:"item"`
	} `json:"data"`
}

type KlineResp struct {
}

// KlineItem 日k
type KlineItem struct {
	Item [][]float64 `json:"data"`
}

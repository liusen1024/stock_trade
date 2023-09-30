package quote

import (
	"context"
	"reflect"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/api-gateway/util"
	"stock/common/log"
	"strings"
	"sync"
	"time"

	"github.com/axgle/mahonia"
)

// QtService 行情服务
type QtService struct {
	m     map[string]*model.TencentQuote
	mutex sync.Mutex
}

var (
	qtService *QtService
	qtOnce    sync.Once
)

// QtServiceInstance 行情服务实例
func QtServiceInstance() *QtService {
	qtOnce.Do(func() {
		qtService = &QtService{
			m: map[string]*model.TencentQuote{},
		}

		ctx := context.Background()
		go func() {
			for range time.Tick(1 * time.Second) {
				if err := qtService.load(ctx); err != nil {
					log.Errorf("load err:%+v", err)
					continue
				}
			}
		}()

	})
	return qtService
}

// load 预先加载持仓、委托、昨日持仓股票行情
func (s *QtService) load(ctx context.Context) error {
	return nil
}

var constMap map[int]string = *initConstMap()

// initConstMap 初始化定义结构体
func initConstMap() *map[int]string {
	return &map[int]string{
		model.Name:             "Name",
		model.Code:             "Code",
		model.CurrentPrice:     "CurrentPrice",
		model.ClosePrice:       "ClosePrice",
		model.OpenPrice:        "OpenPrice",
		model.TotalVol:         "TotalVol",
		model.BuyPrice1:        "BuyPrice1",
		model.BuyVol1:          "BuyVol1",
		model.BuyPrice2:        "BuyPrice2",
		model.BuyVol2:          "BuyVol2",
		model.BuyPrice3:        "BuyPrice3",
		model.BuyVol3:          "BuyVol3",
		model.BuyPrice4:        "BuyPrice4",
		model.BuyVol4:          "BuyVol4",
		model.BuyPrice5:        "BuyPrice5",
		model.BuyVol5:          "BuyVol5",
		model.SellPrice1:       "SellPrice1",
		model.SellVol1:         "SellVol1",
		model.SellPrice2:       "SellPrice2",
		model.SellVol2:         "SellVol2",
		model.SellPrice3:       "SellPrice3",
		model.SellVol3:         "SellVol3",
		model.SellPrice4:       "SellPrice4",
		model.SellVol4:         "SellVol4",
		model.SellPrice5:       "SellPrice5",
		model.SellVol5:         "SellVol5",
		model.Time:             "Time",
		model.Chg:              "Chg",
		model.ChgPercent:       "ChgPercent",
		model.HighPx:           "HighPx",
		model.LowPx:            "LowPx",
		model.TotalAmount:      "TotalAmount",
		model.TurnOverRate:     "TurnOverRate",
		model.Pe:               "Pe",
		model.FloatMarketValue: "FloatMarketValue",
		model.TotalMarketValue: "TotalMarketValue",
		model.Pb:               "Pb",
		model.LimitUpPrice:     "LimitUpPrice",
		model.LimitDownPrice:   "LimitDownPrice",
	}
}

// cache 将qtMap结果存储
func (s *QtService) cache(qtMap map[string]*model.TencentQuote) {
	if len(qtMap) == 0 {
		return
	}
	for code, qt := range qtMap {
		s.mutex.Lock()
		s.m[code] = qt
		s.mutex.Unlock()
	}
}

// getFromRedis 从redis中查询缓存的行情
func (s *QtService) getFromCache(codes []string) (map[string]*model.TencentQuote, []string) {
	quoteMap := make(map[string]*model.TencentQuote)
	missCodes := make([]string, 0)
	for _, code := range codes {
		s.mutex.Lock()
		qt, ok := s.m[code]
		s.mutex.Unlock()
		if !ok {
			missCodes = append(missCodes, code)
			continue
		}

		// 与当前时间比价,超过3秒的则设置为失效
		if time.Now().Sub(qt.DataTime) > 2*time.Second {
			missCodes = append(missCodes, code)
			continue
		}
		quoteMap[code] = qt
	}

	return quoteMap, missCodes
}

// GetQuoteByTencent 查询腾讯行情
func (s *QtService) GetQuoteByTencent(codes []string) (map[string]*model.TencentQuote, error) {
	result := make(map[string]*model.TencentQuote)
	return result, nil
}

// mGetQuoteByTencent 获取腾讯行情
func (s *QtService) mGetQuoteByTencent(codes []string) (map[string]*model.TencentQuote, error) {
	resp, err := util.Http(s.genTencentURL(codes))
	if err != nil {
		return nil, serr.New(serr.ErrCodeBusinessFail, "行情请求失败")
	}

	result := s.parseTencent(resp)

	return result, nil
}

// genTencentURL 构造url
func (s *QtService) genTencentURL(codes []string) string {
	var quoteURL = "http://qt.gtimg.cn/q="
	return quoteURL
}

func duplicate(a interface{}) (ret []interface{}) {
	va := reflect.ValueOf(a)
	for i := 0; i < va.Len(); i++ {
		if i > 0 && reflect.DeepEqual(va.Index(i-1).Interface(), va.Index(i).Interface()) {
			continue
		}
		ret = append(ret, va.Index(i).Interface())
	}
	return ret
}

// parseTencent 解析url
func (s *QtService) parseTencent(data string) map[string]*model.TencentQuote {
	m := make(map[string]*model.TencentQuote)
	for -1 != strings.Index(data, "=") {
		// 获取单条股票数据
		data = data[strings.Index(data, "\"")+1:]
		src := data[:strings.Index(data, "\"")]
		data = data[strings.Index(data, ";"):]
		// 分割字符串
		arr := strings.Split(src, "~")
		code := arr[model.Code]
		quote := &model.TencentQuote{
			DataTime: time.Now(), // 设置当前时间
		}
		for index, value := range arr {
			field, ok := getFieldNameByIndex(index)
			if !ok {
				continue
			}
			if field == "Name" {
				value = mahonia.NewDecoder("gbk").ConvertString(value) // gbk转utf8
			}
			setFieldValue(quote, field, value)
		}
		m[code] = quote
	}
	return m
}

func setFieldValue(quote *model.TencentQuote, field, v string) {
	typ := reflect.ValueOf(quote).Elem().FieldByName(field).Type().Kind()
	switch typ {
	case reflect.String:
		reflect.ValueOf(quote).Elem().FieldByName(field).SetString(v)
	case reflect.Float64:
		reflect.ValueOf(quote).Elem().FieldByName(field).SetFloat(util.String2Float64(v))
	case reflect.Int64:
		reflect.ValueOf(quote).Elem().FieldByName(field).SetInt(util.String2Int64(v))
	}
}

// getFieldNameByIndex 通过结构体字段获取反射序号
func getFieldNameByIndex(index int) (string, bool) {
	name, ok := constMap[index]
	if !ok {
		return "", false
	}
	return name, true
}

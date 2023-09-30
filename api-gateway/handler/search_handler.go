package handler

import (
	"sort"
	"stock/api-gateway/model"
	"stock/api-gateway/service"
	"stock/api-gateway/util"
	"stock/common/errgroup"
	"sync"

	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/gin-gonic/gin"
)

// SearchHandler handler
type SearchHandler struct {
}

// NewSearchHandler 单例
func NewSearchHandler() *SearchHandler {
	return &SearchHandler{}
}

// Register 注册handler
func (h *SearchHandler) Register(e *gin.Engine) {
	// 热门搜索股票
	e.GET("/search/hot", JSONWrapper(h.GetHotStock))
	e.GET("/search", JSONWrapper(h.Search))
}

type stock struct {
	StockCode string `json:"stock_code"`
	StockName string `json:"stock_name"`
}

// Search 股票搜索
func (h *SearchHandler) Search(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	list := make([]*stock, 0)
	query, _ := String(c, "query")
	// 查询内容为空,则返回默认的热门股票
	if len(query) == 0 {
		for _, it := range service.DataServiceInstance().GetData().HotStock[0:20] {
			list = append(list, &stock{
				StockCode: it.Code,
				StockName: it.Name,
			})
		}
		return list, nil
	}

	stocks, err := service.StockDataServiceInstance().GetStocks(ctx)
	if err != nil {
		return make([]*stock, 0), nil
	}
	codes := make([]string, 0)
	names := make([]string, 0)
	//pinyinList := make([]string, 0)

	codeMap := make(map[string]*model.StockData) // 股票代码
	nameMap := make(map[string]*model.StockData) // 股票名称
	//pinyinMap := make(map[string]*model.StockData) // 股票拼音
	for _, it := range stocks {
		codes = append(codes, it.Code)
		names = append(names, it.Name)
		codeMap[it.Code] = it
		nameMap[it.Name] = it
	}
	var mutex sync.Mutex
	//eg := errgroup.GroupWithCount(10)
	//for _, it := range stocks {
	//	stock := it
	//	eg.Go(func() error {
	//		// 拼音
	//		stockNamePinYin := pinyin.LazyPinyin(stock.Name, pinyin.Args{
	//			Style:     pinyin.FirstLetter,
	//			Heteronym: false,
	//			Separator: "",
	//			Fallback:  pinyin.Fallback,
	//		})
	//		mutex.Lock()
	//		pinyinMap[strings.Join(stockNamePinYin, "")] = stock
	//		pinyinList = append(pinyinList, strings.Join(stockNamePinYin, ""))
	//		mutex.Unlock()
	//		return nil
	//	})
	//}
	//if err := eg.Wait(); err != nil {
	//	return nil, err
	//}

	// 股票代码搜索
	wgCode := errgroup.GroupWithCount(20)
	for _, it := range fuzzy.Find(query, codes) {
		code := it
		wgCode.Go(func() error {
			s, ok := codeMap[code]
			if !ok {
				return nil
			}
			mutex.Lock()
			list = append(list, &stock{
				StockCode: s.Code,
				StockName: s.Name,
			})
			mutex.Unlock()
			return nil
		})
	}
	if err := wgCode.Wait(); err != nil {
		return nil, err
	}

	// 股票名称搜索
	wgName := errgroup.GroupWithCount(20)
	for _, it := range fuzzy.Find(query, names) {
		name := it
		wgName.Go(func() error {
			s, ok := nameMap[name]
			if !ok {
				return nil
			}
			mutex.Lock()
			list = append(list, &stock{
				StockCode: s.Code,
				StockName: s.Name,
			})
			mutex.Unlock()
			return nil
		})
	}
	if err := wgName.Wait(); err != nil {
		return nil, err
	}

	// 股票拼音搜索
	//wgPinyin := errgroup.GroupWithCount(20)
	//for _, it := range fuzzy.Find(query, pinyinList) {
	//	namePinYin := it
	//	wgPinyin.Go(func() error {
	//		s, ok := pinyinMap[namePinYin]
	//		if !ok {
	//			return nil
	//		}
	//		mutex.Lock()
	//		list = append(list, &stock{
	//			StockCode: s.Code,
	//			StockName: s.Name,
	//		})
	//		mutex.Unlock()
	//		return nil
	//	})
	//}
	//if err := wgPinyin.Wait(); err != nil {
	//	return nil, err
	//}

	sort.SliceStable(list, func(i, j int) bool {
		return list[i].StockCode > list[j].StockCode
	})
	if len(list) > 50 {
		list = list[0:50]
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// GetHotStock 热门股票搜索
func (h *SearchHandler) GetHotStock(c *gin.Context) (interface{}, error) {
	data := service.DataServiceInstance().GetData()
	list := make([]*model.HotStock, 0)
	if len(data.HotStock) > 0 {
		list = data.HotStock[0:8]
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

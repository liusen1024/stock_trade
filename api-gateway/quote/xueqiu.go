package quote

import (
	"stock/api-gateway/model"
	"sync"
)

// XqQtService 雪球行情服务
type XqQtService struct {
}

var (
	xqQtService *XqQtService
	xqQtOnce    sync.Once
)

// GetQuoteByXueQiu 雪球行情查询
func (s *QtService) GetQuoteByXueQiu(codes []string) (map[string]*model.TencentQuote, error) {
	//if len(codes) == 0 {
	//	return make(map[string]*model.TencentQuote), nil
	//}
	//// 过滤掉重复的codes
	//tmp := make(map[string]bool)
	//for _, code := range codes {
	//	tmp[code] = true
	//}
	//codes = codes[:0]
	//for k := range tmp {
	//	codes = append(codes, k)
	//}
	//
	//result := make(map[string]*model.TencentQuote)
	//wg := errgroup.GroupWithCount(5)
	//var mutex sync.Mutex
	//for len(codes) > 0 {
	//	cnt := len(codes)
	//	if cnt > 100 {
	//		cnt = 100
	//	}
	//	tmp := codes[0:cnt]
	//	codes = codes[cnt:]
	//	wg.Go(func() error {
	//		m, err := s.mGetQuoteByTencent(tmp)
	//		if err != nil {
	//			return err
	//		}
	//		mutex.Lock()
	//		for k, v := range m {
	//			result[k] = v
	//		}
	//		mutex.Unlock()
	//		return nil
	//	})
	//}
	//if err := wg.Wait(); err != nil {
	//	log.Errorf("获取腾讯行情失败:%+v", err)
	//	return nil, serr.ErrBusiness("获取行情失败")
	//}
	//return result, nil
	return nil, nil
}

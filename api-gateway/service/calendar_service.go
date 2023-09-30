package service

import (
	"context"
	"fmt"
	"stock/api-gateway/dao"
	"stock/api-gateway/model"
	"stock/common/errgroup"
	"stock/common/log"
	"stock/common/timeconv"
	"sync"
	"time"
)

// CalendarService 交易日历服务
type CalendarService struct {
	calendar map[int32]bool
}

var (
	calendarService *CalendarService
	calendarOnce    sync.Once
)

// CalendarServiceInstance 实例
func CalendarServiceInstance() *CalendarService {
	calendarOnce.Do(func() {
		calendarService = &CalendarService{
			calendar: map[int32]bool{},
		}
		ctx := context.Background()
		if err := calendarService.update(ctx); err != nil {
			log.Errorf("update err:%+v", err)
		}

		if err := calendarService.load(ctx); err != nil {
			log.Errorf("load err:%+v", err)
			panic("导入交易日历失败")
		}

		// 定时器从交易所拉取交易日历
		go func() {
			m := make(map[int32]bool)
			for range time.Tick(24 * time.Hour) {
				if _, ok := m[timeconv.TimeToInt32(time.Now())]; ok {
					continue
				}
				if err := calendarService.update(ctx); err != nil {
					log.Errorf("loadCalendar err:%+v", err)
					continue
				}
				m[timeconv.TimeToInt32(time.Now())] = true
			}
		}()
		// 定时从数据库更新交易日历
		go func() {
			for range time.Tick(1 * time.Hour) {
				if err := calendarService.load(ctx); err != nil {
					log.Errorf("load err:%+v", err)
				}

			}
		}()
	})
	return calendarService
}

// load 从数据库导出交易日历
func (s *CalendarService) load(ctx context.Context) error {
	list, err := dao.TradeCalendarDaoInstance().GetCalendar(ctx)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return nil
	}
	m := make(map[int32]bool)
	for _, it := range list {
		m[timeconv.TimeToInt32(it.Date)] = it.Trade
	}
	s.calendar = m
	return nil
}

// IsEntrustTime 是否委托时间
func (s *CalendarService) IsEntrustTime(ctx context.Context) bool {
	// 1. 今天是否为交易日
	if !CalendarServiceInstance().IsTradeDate(ctx) {
		return false
	}
	// 2. 当前时间是否超过3点
	now := time.Now()
	et, _ := time.Parse("15:04:05", "14:59:59") // 收盘时间为:14:59:59
	if now.After(time.Date(now.Year(), now.Month(), now.Day(), et.Hour(), et.Minute(), et.Second(), 0, time.Local)) {
		return false
	}

	return true
}

// IsTradeTime 是否交易时间
func (s *CalendarService) IsTradeTime(ctx context.Context) bool {
	// 1 .优先检查今天是否交易日
	if !s.IsTradeDate(ctx) {
		return false
	}

	// 2. 检查现在时间是否处于交易时间
	now := time.Now()
	bt, _ := time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%s %s", now.Format("2006-01-02"), "09:30:01"))
	et, _ := time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%s %s", now.Format("2006-01-02"), "11:29:59"))

	if timeconv.TimeToInt64(now) > timeconv.TimeToInt64(bt) && timeconv.TimeToInt64(now) < timeconv.TimeToInt64(et) {
		return true
	}
	// 下午
	bt, _ = time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%s %s", now.Format("2006-01-02"), "13:00:01"))
	et, _ = time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%s %s", now.Format("2006-01-02"), "14:59:59"))
	if timeconv.TimeToInt64(now) > timeconv.TimeToInt64(bt) && timeconv.TimeToInt64(now) < timeconv.TimeToInt64(et) {
		return true
	}
	return false
}

// IsTradeDate date 是否为交易日:true为交易日,false为非交易日
func (s *CalendarService) IsTradeDate(ctx context.Context) bool {
	isTradeDate, ok := s.calendar[timeconv.TimeToInt32(time.Now())]
	if !ok {
		return false
	}
	return isTradeDate
}

// update 网络请求更新交易日历
func (s *CalendarService) update(ctx context.Context) error {
	calendar := make([]*model.TradeCalendar, 0)
	// 获取本月、下月的交易日历
	wg := errgroup.GroupWithCount(2)
	var mutex sync.Mutex
	for _, it := range []string{
		fmt.Sprintf("%s", time.Now().Format("2006-01")),
		fmt.Sprintf("%s", time.Now().AddDate(0, 1, 0).Format("2006-01")),
	} {
		month := it
		wg.Go(func() error {
			ret, err := s.fetchCalendar(month)
			if err != nil {
				log.Errorf("fetchCalendar month:%+v err:%+v", month, err)
				return err
			}
			mutex.Lock()
			calendar = append(calendar, ret...)
			mutex.Unlock()
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		log.Errorf("获取交易日历失败:%+v", err)
		return err
	}

	if err := dao.TradeCalendarDaoInstance().Create(ctx, calendar); err != nil {
		log.Errorf("创建交易日历失败:%+v", err)
		return err
	}

	return nil
}

func (s *CalendarService) fetchCalendar(month string) ([]*model.TradeCalendar, error) {
	calendar := make([]*model.TradeCalendar, 0)
	return calendar, nil
}

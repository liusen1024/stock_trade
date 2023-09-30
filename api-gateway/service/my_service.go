package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/id_gen"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"
	"stock/common/timeconv"
	"strconv"
	"sync"
	"time"
)

// MyService 服务
type MyService struct {
}

var (
	myService *MyService
	myOnce    sync.Once
)

// MyServiceInstance 实例
func MyServiceInstance() *MyService {
	myOnce.Do(func() {
		myService = &MyService{}
	})
	return myService
}

// WithdrawCommit 提交提现
func (s *MyService) WithdrawCommit(ctx context.Context, uid int64, money float64, name, bankNo, code string) error {
	// 检查验证码是否成功
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
	if err != nil {
		return err
	}
	ok, err := dao.SmsDaoInstance().VerifySms(ctx, user.UserName, code)
	if err != nil || !ok {
		return serr.ErrBusiness("验证码错误")
	}
	if user.Money < money {
		return serr.ErrBusiness("转出金额大于可提现金额")
	}
	// 冻结资金
	tx := db.StockDB().WithContext(ctx).Begin()
	defer tx.Rollback()

	user.Money = user.Money - money
	user.FreezeMoney += money
	if err := dao.UserDaoInstance().UpdateUserWithTx(tx, user); err != nil {
		log.Errorf("变更用户资金失败:%+v", err)
		return serr.ErrBusiness("资金转出失败")
	}

	if err := dao.TransferDaoInstance().CreateWithTx(tx, &model.Transfer{
		UID:       uid,                          // 用户ID
		OrderTime: time.Now(),                   // 订单时间
		Money:     money,                        // 金额
		Type:      model.TransferTypeWithdraw,   // 类型：1充值 2提现
		Status:    model.TransferStatusWaitExam, // 状态:0预插入 1待审核 2成功 3失败
		Name:      name,                         // 提现收款人
		BankNo:    bankNo,                       // 提现银行卡号
	}); err != nil {
		log.Errorf("CreateWithTx:%+v", err)
		return serr.ErrBusiness("转出失败")
	}
	if err := tx.Commit().Error; err != nil {
		log.Errorf("事务提交失败:%+v", err)
		return err
	}

	if err := dao.SmsDaoInstance().InvalidCode(ctx, user.UserName); err != nil {
		log.Errorf("InvalidCode err:%+v", err)
	}
	return nil
}

// Withdraw 提现初始化:收款人、收款银行卡号、可提现金额
func (s *MyService) Withdraw(ctx context.Context, uid int64) (string, string, float64, error) {
	// 查询以前是否提现成功过
	list, err := dao.TransferDaoInstance().GetByUid(ctx, uid)
	if err != nil {
		return "", "", 0, err
	}
	var name, bankNo string
	for _, it := range list {
		if it.Type == model.TransferTypeWithdraw && it.Status == model.TransferStatusSuccess {
			name = it.Name
			bankNo = it.BankNo
			break
		}
	}

	user, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
	if err != nil {
		return "", "", 0, err
	}
	return name, bankNo, user.Money, nil
}

func (s *MyService) RechargeQrcodeCommit(ctx context.Context, uid int64, money float64, orderNo string) error {
	if err := s.isExistsCommit(ctx, uid); err != nil {
		return err
	}
	if err := dao.TransferDaoInstance().Create(ctx, &model.Transfer{
		UID:       uid,
		OrderTime: time.Now(),
		Money:     money,
		Type:      model.TransferTypeRecharge,
		Channel:   "扫码支付",
		Status:    model.TransferStatusWaitExam,
	}); err != nil {
		return serr.ErrBusiness("转入错误")
	}
	return nil
}

// RechargeQrcode 二维码初始化
func (s *MyService) RechargeQrcode(ctx context.Context, uid int64, money float64) (string, string, error) {
	//key := fmt.Sprintf("qrcode_money_%d", int64(money*1000))
	//buff, err := db.Get(ctx, key).Bytes()
	//if err != nil {
	//	if err != redis.Nil {
	//		return "", "", serr.ErrBusiness("转入渠道错误")
	//	}
	//}
	//
	//result := &model.Data{}
	//err = json.Unmarshal(buff, result)
	//if err != nil {
	//	log.Errorf("json.Unmarshal error: %v", err)
	//	return &model.Data{}
	//}
	return "https://test-1252629308.cos.ap-guangzhou.myqcloud.com/122.png", strconv.FormatInt(id_gen.GetNextID(), 10), nil
}

// isExistsCommit 是否存在转入提交
func (s *MyService) isExistsCommit(ctx context.Context, uid int64) error {
	list, err := dao.TransferDaoInstance().GetByUid(ctx, uid)
	if err != nil {
		return err
	}
	// 检查是否允许用户提交
	for _, it := range list {
		// 非待审核状态,充值则跳过
		if it.Status != model.TransferStatusWaitExam || it.Type != model.TransferTypeRecharge {
			continue
		}
		if timeconv.TimeToInt32(it.OrderTime) == timeconv.TimeToInt32(time.Now()) {
			return serr.ErrBusiness("您有一笔订单处于待审核状态,请稍后再提交。")
		}
	}
	return nil
}

// RechargeBankCommit 银行转入资金提交
func (s *MyService) RechargeBankCommit(ctx context.Context, uid int64, money float64) error {
	if err := s.isExistsCommit(ctx, uid); err != nil {
		return err
	}
	if err := dao.TransferDaoInstance().Create(ctx, &model.Transfer{
		UID:       uid,
		OrderTime: time.Now(),
		Money:     money,
		Type:      model.TransferTypeRecharge,
		Channel:   "银行卡",
		Status:    model.TransferStatusWaitExam, // 等待审核
	}); err != nil {
		log.Errorf("Create err:%+v", err)
		return serr.ErrBusiness("转入资金错误")
	}
	return nil
}

func (s *MyService) RechargeAlipay(ctx context.Context, uid int64, money float64) (string, error) {
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		return "", err
	}
	if !sys.AlipayChannel {
		return "", serr.ErrBusiness("转入资金渠道未开放")
	}
	orderNo := strconv.FormatInt(id_gen.GetNextID(), 10)
	if err := dao.TransferDaoInstance().Create(ctx, &model.Transfer{
		UID:       uid,                        // 用户ID
		OrderTime: time.Now(),                 // 订单时间
		Money:     money,                      // 金额
		Type:      model.TransferTypeRecharge, // 类型：1充值 2提现
		Status:    model.TransferStatusPre,    // 状态:0预插入 1待审核 2成功 3失败
		Name:      "",                         // 提现收款人
		Channel:   "支付宝",
		BankNo:    "",      // 提现银行卡号
		OrderNo:   orderNo, // 订单号流水
	}); err != nil {
		log.Errorf("插入转账记录表失败:%+v", err)
		return "", serr.ErrBusiness("渠道错误")
	}
	return AlipayServiceInstance().GetAlipayURL(orderNo, money)
}

func (s *MyService) Balance(ctx context.Context, uid int64) ([]*model.MyBalance, float64, error) {
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
	if err != nil {
		return nil, 0, err
	}
	transfers, err := dao.TransferDaoInstance().GetByUid(ctx, uid)
	if err != nil {
		return nil, 0, err
	}
	sort.SliceStable(transfers, func(i, j int) bool {
		return timeconv.TimeToInt64(transfers[i].OrderTime) > timeconv.TimeToInt64(transfers[j].OrderTime)
	})
	balance := make([]*model.MyBalance, 0)
	for _, it := range transfers {
		if it.Status == model.TransferStatusPre {
			continue
		}
		switch it.Type {
		case model.TransferTypeWithdraw, model.TransferTypeAppendMoney, model.TransferTypeExpandMoney, model.TransferTypeCreateContract:
			it.Money *= -1
		}
		balance = append(balance, &model.MyBalance{
			Title:   it.TransferConvertTitle(),
			Balance: it.Money,
			Time:    it.OrderTime.Format("2006-01-02 15:04:05"),
		})
	}
	return balance, user.Money, nil
}

func (s *MyService) Authentication(ctx context.Context, uid int64, name, idNo string) error {
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
	if err != nil {
		return err
	}
	if len(user.Name) > 0 && len(user.ICCID) > 0 {
		return serr.ErrBusiness("已实名认证")
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://idcardcert.market.alicloudapi.com/idCardCert?idCard=%s&name=%s", idNo, name), nil)
	if err != nil {
		log.Errorf("NewRequest err:%+v", err)
		return fmt.Errorf("验证码发送失败")
	}
	req.Header.Add("Authorization", "APPCODE xxxxxxxxxxxxxxxxxxxxxgi")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	type t struct {
		Status string `json:"status"`
		Msg    string `json:"msg"`
	}
	result := &t{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	if result.Status != "01" {
		return serr.ErrBusiness(result.Msg)
	}
	// 更新用户信息到数据库表
	user.Name = name
	user.ICCID = idNo
	if err := dao.UserDaoInstance().CreateUser(ctx, user); err != nil {
		log.Errorf("CreateUser err:%+v", err)
		return serr.ErrBusiness("实名认证失败")
	}
	return nil
}

func (s *MyService) Msg(ctx context.Context, uid int64) ([]*model.MyMsg, error) {
	list, err := dao.MsgDaoInstance().GetByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(list, func(i, j int) bool {
		return timeconv.TimeToInt64(list[i].CreateTime) > timeconv.TimeToInt64(list[j].CreateTime)
	})
	result := make([]*model.MyMsg, 0)
	for _, it := range list {
		result = append(result, &model.MyMsg{
			Title:   it.Title,
			Content: it.Content,
			Time:    it.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

// My 我的首页
func (s *MyService) My(ctx context.Context, uid int64) (interface{}, error) {
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	contracts, err := dao.ContractDaoInstance().GetContractsByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	var contractAmount int64
	for _, it := range contracts {
		if it.Status == model.ContractStatusEnable {
			contractAmount++
		}
	}

	return struct {
		UserName string  `json:"user_name"`
		Balance  float64 `json:"balance"`
		Contract int64   `json:"contract"`
	}{
		UserName: user.UserName,
		Balance:  user.Money,
		Contract: contractAmount,
	}, nil

}

package service

import (
	"context"
	"fmt"
	"net"
	"stock/api-gateway/dao"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/env"
	"stock/common/log"
	"strconv"
	"sync"

	"github.com/smartwalle/alipay/v3"
)

// AlipayService 服务
type AlipayService struct {
}

var (
	alipayService *AlipayService
	alipayOnce    sync.Once
)

// AlipayServiceInstance 实例
func AlipayServiceInstance() *AlipayService {
	myOnce.Do(func() {
		alipayService = &AlipayService{}
	})
	return alipayService
}

// AliPayNotify 支付宝回调通知
func (s *AlipayService) AliPayNotify(ctx context.Context, orderNo, buyer string, amount float64) {
	transfer, err := dao.TransferDaoInstance().GetByOrderNo(ctx, orderNo)
	if err != nil {
		log.Errorf("订单号不存在:%+v", err)
		return
	}
	if transfer.Status != model.TransferStatusPre {
		return
	}
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, transfer.UID)
	if err != nil {
		return
	}

	transfer.Status = model.TransferStatusSuccess
	if err := dao.TransferDaoInstance().Create(ctx, transfer); err != nil {
		return
	}

	// 设置用户资金
	user.Money += transfer.Money
	if err := dao.UserDaoInstance().CreateUser(ctx, user); err != nil {
		return
	}
}

func (s *AlipayService) GetAlipayURL(orderNo string, money float64) (string, error) {
	appID, ok := env.GlobalEnv().Get("APPID")
	if !ok {
		log.Errorf("no APPID config")
		return "", serr.ErrBusiness("渠道错误")
	}
	privateKey, ok := env.GlobalEnv().Get("PRIVATEKEY") // 应用私钥
	if !ok {
		log.Errorf("no PRIVATEKEY config")
		return "", serr.ErrBusiness("渠道错误")
	}
	aliPublicKey, ok := env.GlobalEnv().Get("PUBLICKEY") // 支付宝公钥
	if !ok {
		log.Errorf("no PUBLICKEY config")
		return "", serr.ErrBusiness("渠道错误")
	}
	localIP, ok := env.GlobalEnv().Get("IP") // 获取本机外网IP
	if !ok {
		log.Errorf("no IP config")
		return "", serr.ErrBusiness("渠道错误")
	}
	client, err := alipay.New(appID, privateKey, env.GlobalEnv().IsProd())
	if err != nil {
		log.Errorf("alipay New err:%+v", err)
		return "", err
	}
	if err := client.LoadAliPayPublicKey(aliPublicKey); err != nil {
		return "", err
	}
	p := alipay.TradePagePay{}
	p.NotifyURL = fmt.Sprintf("http://%s:8080/alipay/callback", localIP) //支付结果通知
	p.Subject = "trade"
	p.OutTradeNo = orderNo
	p.TotalAmount = strconv.FormatFloat(money, 'f', 2, 64)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	alipayURL, err := client.TradePagePay(p)
	if err != nil {
		return "", err
	}
	return alipayURL.String(), nil
}

// getPublicIP 获取本地外网
func (s *AlipayService) getPublicIP() (string, error) {
	// 获取本机外网地址
	as, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, a := range as {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		ip := ipNet.IP.To4()
		if ip != nil &&
			(ip[0] != 10 && ip[0] != 172 && ip[0] != 192 && ip[1] != 168) {
			return ip.String(), nil
		}
	}
	return "", serr.ErrBusiness("no found valid ip")
}

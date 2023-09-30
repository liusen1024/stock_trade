package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/api-gateway/util"
	"stock/common/env"
	"stock/common/log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"gorm.io/gorm"
)

// SmsService 内容服务
type SmsService struct {
}

var (
	smsService *SmsService
	smsOnce    sync.Once
)

// SmsServiceInstance SmsServiceInstance实例
func SmsServiceInstance() *SmsService {
	smsOnce.Do(func() {
		smsService = &SmsService{}
	})
	return smsService
}

func (s *SmsService) limitCacheKey(phone string) string {
	return fmt.Sprintf("limit_times_phone:%+v", phone)
}

func (s *SmsService) SendSms(ctx context.Context, content string, phone string) error {
	times, err := db.RedisClient().Get(ctx, s.limitCacheKey(phone)).Int()
	if err != nil && err != redis.Nil {
		log.Errorf("get redis key:%+v err:%+v", s.limitCacheKey(phone), err)
		return err
	}
	if times > 10 {
		log.Errorf("发送频次大于10")
		return serr.ErrBusiness("发送频次大于10")
	}

	v, ok := env.GlobalEnv().Get("PLATFORM_NAME")
	if !ok {
		panic("no PLATFORM_NAME conf")
	}
	content = v + content
	if err := otherSend(phone, content); err != nil {
		log.Errorf("发送短信失败:%+v", err)
		return serr.ErrBusiness("发送短信失败")
	}
	// 保存短信
	if err := dao.SmsDaoInstance().Save(ctx, &model.Sms{
		Phone: phone,
		Msg:   content,
		Time:  time.Now(),
	}); err != nil {
		return err
	}

	pip := db.RedisClient().TxPipeline()
	pip.Expire(ctx, s.limitCacheKey(phone), 24*time.Hour)
	if err := pip.Incr(ctx, s.limitCacheKey(phone)).Err(); err != nil {
		log.Errorf("Incr err:%+v", err)
		return nil
	}
	if _, err := pip.Exec(ctx); err != nil {
		log.Errorf("Exec err:%+v", err)
		return nil
	}
	return nil
}

func (s *SmsService) Send(ctx context.Context, phone string) error {
	// 1. 根据手机号获取最近的一条短信
	sms, err := dao.SmsDaoInstance().GetLatestSms(ctx, phone)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("GetLatestSms err:%+v", err)
		return serr.ErrBusiness("验证码发送失败")
	}
	if sms != nil {
		ok, err := dao.SmsDaoInstance().VerifySms(ctx, sms.Phone, sms.Code)
		if err != nil {
			return serr.ErrBusiness("短信发送失败")
		}
		if ok {
			return serr.ErrBusiness("验证码已发送")
		}
	}
	// 2. 随机生成验证码
	code := randCode()
	//platform, ok := env.GlobalEnv().Get("PLATFORM_NAME")
	//if !ok {
	//	platform = "【验证码】"
	//}
	//content := fmt.Sprintf("%s您的验证码是:%s", platform, code)

	// 其他第三方短信平台发送
	//if err := otherSend(phone, content); err != nil {
	//	log.Errorf("漫道科技短信发送失败")
	//	return serr.ErrBusiness("验证码发送失败")
	//}
	content, err := smsBaoSend(phone, code)
	if err != nil {
		log.Errorf("")
		return serr.ErrBusiness("验证码发送失败")
	}

	// 保存验证码
	err = dao.SmsDaoInstance().Save(ctx, &model.Sms{
		Phone:  phone,
		Code:   code,
		Status: model.SmsStatusValid,
		Msg:    content,
		Time:   time.Now(),
	})
	if err != nil {
		log.Errorf("保存验证码失败:%+v", err)
		return err
	}
	return nil
}

// otherSend 第三方平台发送验证码
func otherSend(phone, content string) error {
	data := url.Values{
		"action":    {"send"},
		"userid":    {"280"},
		"timestamp": {"xxxxxxxxx"},
		"sign":      {"xxxxxx"},
		"mobile":    {phone},
		"content":   {content},
		"sendtime":  {""},
		"extno":     {""},
	}
	resp, err := http.PostForm("http://47.97.161.216:8088/v2sms.aspx", data)
	if err != nil {
		log.Errorf("验证码发送失败")
		return serr.ErrBusiness("验证码发送失败")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("读取短信验证码返回信息失败")
		return fmt.Errorf("验证码发送失败")
	}
	log.Infof("短信验证码发送成功:phone[%+v],content[%+v] return_body:%+v", phone, content, string(body))
	return nil
}

// yunPianSend 云片网发送验证码
func yunPianSend(phone, code string) error {
	apiKey, ok := env.GlobalEnv().Get("APIKEY")
	if !ok {
		log.Errorf("获取APIKEY失败")
		return fmt.Errorf("验证码发送失败")
	}
	text, ok := env.GlobalEnv().Get("SMS")
	if !ok {
		log.Errorf("获取SMS短信模板失败")
		return fmt.Errorf("验证码发送失败")
	}
	data := url.Values{"apikey": {apiKey}, "mobile": {phone}, "text": {fmt.Sprintf("%s%s", text, code)}}
	resp, err := http.PostForm("https://sms.yunpian.com/v2/sms/single_send.json", data)
	if err != nil {
		log.Errorf("验证码post失败")
		return fmt.Errorf("验证码发送失败")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("读取短信验证码返回信息失败")
		return fmt.Errorf("验证码发送失败")
	}
	log.Infof("短信验证码发送成功:phone[%+v],code[%+v] return_body:%+v", phone, code, string(body))
	return nil
}

// randCode 生层4个随机数
func randCode() string {
	buf := make([]byte, 4)
	var digit = []byte("0123456789")
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 4; i++ {
		buf[i] = digit[rand.Intn(10)]
	}
	return string(buf)
}

// smsBaoSend smsBao
func smsBaoSend(phone string, code string) (string, error) {
	userName := "xxxxxx"
	secKey := "xxxx"
	content := fmt.Sprintf("【鑫启点】您的验证码为%s，在5分钟内有效。", code)
	req := fmt.Sprintf("https://api.smsbao.com/sms?u=%s&p=%s&m=%s&c=%s", userName, secKey, phone, content)
	if _, err := util.Http(req); err != nil {
		return "", err
	}
	return content, nil
}

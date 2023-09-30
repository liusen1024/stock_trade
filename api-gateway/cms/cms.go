package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/serr"
	"stock/api-gateway/util"
	"stock/common/env"
	"stock/common/log"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"

	"github.com/gin-gonic/gin"
)

// CMSHandler 管理
type CMSHandler struct {
}

// NewCMSHandler 单例
func NewCMSHandler() *CMSHandler {
	return &CMSHandler{}
}

// Register 注册handler
func (h *CMSHandler) Register(e *gin.Engine) {
	// 登录
	e.POST("/cms/login", JSONWrapper(h.Login))
	e.POST("/cms/logout", JSONWrapper(h.Logout))
	e.GET("/cms/get_user_name", JSONWrapper(h.GetUserName))
	e.GET("/cms/agents", JSONWrapper(h.Agents))
	e.POST("/cms/file", JSONWrapper(h.UploadFile)) // 上传文件
}

// UploadFile 上传文件
func (h *CMSHandler) UploadFile(c *gin.Context) (interface{}, error) {
	//ctx := util.RPCContext(c)
	type request struct {
		FileKey string `form:"file_key" json:"file_key"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	file, err := c.FormFile("file")
	if err != nil {
		return nil, err
	}
	uploadDir := "./upload/"
	if err := createTempDir(uploadDir); err != nil {
		return nil, err
	}

	if err := c.SaveUploadedFile(file, uploadDir+file.Filename); err != nil {
		return nil, err
	}

	u, _ := url.Parse("https://test-1252629308.cos.ap-guangzhou.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u}
	cli := cos.NewClient(b, &http.Client{
		//设置超时时间
		Timeout: 100 * time.Second,
		Transport: &cos.AuthorizationTransport{
			//如实填写账号和密钥，也可以设置为环境变量
			SecretID:  "AKID1wjNq77v9B1g9ibcs20oS1QlMq9HryAj",
			SecretKey: "azCHBuwOhplXAcx77vX3NZlneNQM9lBR",
		},
	})

	name := uploadDir + file.Filename
	resp, err := cli.Object.PutFromFile(context.Background(), name, uploadDir+file.Filename, nil)
	if err != nil {
		log.Errorf("传输文件错误：%+v", err)
		return nil, err
	}
	defer resp.Body.Close()
	bs, _ := ioutil.ReadAll(resp.Body)
	log.Infof("%+v", string(bs))

	return map[string]interface{}{
		"result": true,
	}, nil
}

// Agents 用户代理-代理机构列表
func (h *CMSHandler) Agents(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type AgentName struct {
		Name string `json:"name"`
	}
	list := make([]*AgentName, 0)
	if IsAdmin(c) {
		roles, err := dao.RoleDaoInstance().GetRoles(ctx)
		if err != nil {
			return nil, err
		}
		for _, it := range roles {
			list = append(list, &AgentName{
				Name: it.UserName,
			})
		}
	} else {
		list = append(list, &AgentName{Name: Username(c)})
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// GetUserName 管理员昵称
func (h *CMSHandler) GetUserName(c *gin.Context) (interface{}, error) {
	return map[string]interface{}{
		"user_name": Username(c),
	}, nil
}

func (h *CMSHandler) Logout(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	if err := h.delete(ctx, Username(c)); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}
func (h *CMSHandler) delete(ctx context.Context, userName string) error {
	if err := db.RedisClient().Del(ctx, getUserSessionKey(userName)).Err(); err != nil {
		return err
	}
	return nil
}

// Login 登录
func (h *CMSHandler) Login(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		UserName string `form:"username" json:"username" binding:"required"`
		Password string `form:"password" json:"password" binding:"required"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	// 数据库检查是否存在该用户
	role, err := dao.RoleDaoInstance().GetRoleByUserName(ctx, req.UserName)
	if err != nil {
		return nil, serr.ErrBusiness("登录失败")
	}
	if role.Password != req.Password {
		return nil, serr.ErrBusiness("密码错误")
	}

	// 创建token
	token, err := Session(ctx, role)
	if err != nil {
		return nil, serr.ErrBusiness("登录失败")
	}

	ip, ok := env.GlobalEnv().Get("IP")
	if !ok {
		log.Errorf("获取配置信息错误:没有配置IP地址")
	}

	// 设置cookie
	c.SetCookie("x-token", token, maxExpireTTL/2, "/", fmt.Sprintf("%s:8890", ip), false, true)

	// 设置用户名
	c.Set("__USERNAME", role.UserName)
	return map[string]interface{}{
		"token": token,
	}, nil

}

func createTempDir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(dir, os.ModePerm); err != nil {
				log.Errorf("failed to create dir:%+v", err)
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

package env

import (
	"fmt"
	"os"
	"strings"

	sj "github.com/bitly/go-simplejson"
)

type Env struct {
	obj *sj.Json
	env string
}

const (
	EnvProd = "prod"
	EnvDev  = "dev"
)

func (e *Env) IsProd() bool {
	return e.env == EnvProd
}

func (e *Env) IsDev() bool {
	return e.env == EnvDev
}

// Get by key, 返回对应的环境变量，或者设置的默认值
func (e *Env) Get(key string) (string, bool) {
	// 如果json中没有配置对应的key，则无法找到
	jv, ok := e.obj.CheckGet(key)
	if !ok {
		return "", false
	}
	// 如果json中配置相关的key，则按照三个顺序读取
	// 1. json-key对于的环境变量,如果有设置则使用
	// 2. json-value如果是环境变量形式，读取对应的值,如果有设置则使用
	// 3. json-value的字面值
	v := os.Getenv(key)
	if len(v) > 0 {
		return v, true
	}
	s, _ := jv.String()
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		envKey := s[2 : len(s)-1]
		v = os.Getenv(envKey)
		if len(v) > 0 {
			return v, true
		}
	}
	if len(s) > 0 {
		return s, true
	}
	return "", false
}

// LoadEnv 从文件中加载环境变量
func LoadEnv(file string) (*Env, error) {
	e := &Env{}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	e.obj, err = sj.NewFromReader(f)
	if err != nil {
		return nil, err
	}
	// check 必须配置ENV值
	env, ok := e.Get("ENV")
	if !ok {
		return nil, fmt.Errorf("no ENV key in %s", file)
	}
	e.env = env
	return e, nil
}

var globalEnv *Env

// SetGlobalEnv 项目启动的时候设置一次
func SetGlobalEnv(e *Env) {
	globalEnv = e
}

// GlobalEnv 项目其他地方直接调用这个函数获取Env对象
func GlobalEnv() *Env {
	return globalEnv
}

func LoadGlobalEnv(conf string) {
	e, err := LoadEnv(conf)
	if err != nil {
		panic(err)
	}
	SetGlobalEnv(e)
	v, _ := e.Get("ENV")
	fmt.Printf("load global env from %s, ENV %s\n", conf, v)
}

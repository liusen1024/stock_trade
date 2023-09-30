module stock/api-gateway

go 1.16

require (
	github.com/PuerkitoBio/goquery v1.8.0 // indirect
	github.com/antchfx/htmlquery v1.2.5 // indirect
	github.com/antchfx/xmlquery v1.3.12 // indirect
	github.com/axgle/mahonia v0.0.0-20180208002826-3358181d7394
	github.com/gin-gonic/gin v1.7.4
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gocolly/colly v1.2.0
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/kennygrant/sanitize v1.2.4 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/lithammer/fuzzysearch v1.1.5
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mozillazg/go-httpheader v0.3.1 // indirect
	github.com/robfig/cron v1.2.0
	github.com/saintfish/chardet v0.0.0-20120816061221-3af4cd4741ca // indirect
	github.com/smartwalle/alipay/v3 v3.1.7
	github.com/sony/sonyflake v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/tealeg/xlsx v1.0.5
	github.com/temoto/robotstxt v1.1.2 // indirect
	github.com/tencentyun/cos-go-sdk-v5 v0.7.39
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/net v0.0.0-20221004154528-8021a29435af // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/appengine v1.6.7 // indirect
	gorm.io/driver/mysql v1.2.1
	gorm.io/gorm v1.22.4
	stock/common v0.0.0
)

replace stock/common => ../common

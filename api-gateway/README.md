# 安装go
1. 下载安装包
wget https://golang.google.cn/dl/go1.17.1.linux-amd64.tar.gz
2. 解压安装包
   tar -zxf go1.17.6.linux-amd64.tar.gz -C /usr/local
3. 配置环境变量
打开/etc/profile,粘贴以下内容到该文件
#golang env config
export GO111MODULE=on
export GOROOT=/usr/local/go
export GOPATH=/root/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
export GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

4. 执行
source /etc/profile

5. 创建目录
mkdir -p ~/go/src
6. 拷贝代码
git clone https://gitee.com/sen-liu/stock.git

# 安装redis
## 下载redis
yum install -y http://rpms.famillecollet.com/enterprise/remi-release-7.rpm
yum --enablerepo=remi install redis
## 安装完毕后，即可使用下面的命令启动redis服务
service redis start
或者
systemctl start redis

查看redis版本：
redis-cli --version

设置为开机自动启动：
chkconfig redis on
或者
systemctl enable redis.service

## 修改redis配置
vim /etc/redis.conf
bind 0.0.0.0
protected-mode no

## 添加redis密码
requirepass stock_redis_123456

## 重启redis服务
systemctl restart redis

# 安装mysql

# 安装nginx




# 拉取go源代码


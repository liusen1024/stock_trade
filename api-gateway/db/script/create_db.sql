create database if not exists stock default charset = utf8mb4;

use stock;

CREATE TABLE if not exists  `users` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID 主键',
    `user_name` VARCHAR(64) NOT NULL COMMENT '用户名 或 手机号',
    `password` VARCHAR(64) NOT NULL COMMENT '密码',
    `status` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态:1激活 2冻结',
    `current_contract_id` BIGINT(11)  COMMENT '当前合约ID',
    `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '姓名',
    `icc_id` VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '身份证号码',
    `bank_number` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '银行卡号码',
    `role_id` BIGINT(11)  NOT NULL DEFAULT 0 COMMENT '代理账号',
    `money` DECIMAL(15,2) DEFAULT 0 COMMENT '保证金',
    `freeze_money` DECIMAL(15,2) DEFAULT 0 COMMENT '冻结资金',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uniq_users_name` (`user_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 短信记录表
CREATE TABLE if not exists  `sms` (
    `phone` VARCHAR(32) NOT NULL COMMENT '手机号码',
    `code` VARCHAR(8) NOT NULL COMMENT '验证码',
    `send_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
    `status` INT(2) NOT NULL DEFAULT 1 COMMENT '状态:1生效 2失效',
    `msg` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '短信内容',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 系统配置表
CREATE TABLE if not exists  `sysparam` (
    `start_withdraw_time` time default '09:30:00' comment '提现开始时间',
    `stop_withdraw_time` time default '16:00:00' comment '提现结束时间',
    `limit_pct` DECIMAL(6,5) NOT NULL DEFAULT 8.5 comment '涨跌幅限制:股票涨跌达到涨幅限制买入',
    `cyb_limit_pct` DECIMAL(6,5) NOT NULL DEFAULT 8.5 comment '创业板涨跌幅限制',
    `st_limit_pct` DECIMAL(6,5) NOT NULL DEFAULT 3.5 comment 'ST涨跌幅限制:股票涨跌达到涨幅限制买入',
    `is_support_st_stock` BOOL DEFAULT TRUE COMMENT 'true:禁止st股票交易 false:允许st股交易',
    `is_support_sge_board` BOOL DEFAULT FALSE COMMENT '科创板允许交易:true:禁止交易 false:允许交易',
    `is_support_cyb_board` BOOL DEFAULT TRUE COMMENT '创业板允许交易:true:禁止交易 false:允许交易',
    `is_support_bj_board` BOOL DEFAULT FALSE COMMENT '北交所股票允许交易:true:禁止交易 false:允许交易',
    `single_position_pct` DECIMAL(6,5) DEFAULT 0 COMMENT '单只股票最大持仓比率',
    `recharge_notice` BOOL default  false COMMENT '充值通知管理员:true通知,false不通知',
    `register_notice` BOOL default  false COMMENT '注册通知管理员:true通知,false不通知',
    `withdraw_notice` BOOL default  false COMMENT '提现通知管理员:true通知,false不通知',
    `warn_pct` DECIMAL(6,5) DEFAULT 0.6 COMMENT '警戒线:亏损达到该比例则触发警告线 0:不启用',
    `low_warn_can_buy` BOOL DEFAULT TRUE COMMENT '低于警戒线是否允许开仓:true允许开仓 false禁止',
    `close_pct` DECIMAL(6,5) DEFAULT 0.8 COMMENT '平仓线:亏损达到该比例则触发平仓线 0:不启用',
    `holiday_charge` BOOL DEFAULT TRUE COMMENT '非交易日留仓收取管理费:true收取 false不收取',
    `buy_fee` DECIMAL(6,5) NOT NULL DEFAULT 0.005 COMMENT '买入手续费率',
    `sell_fee` DECIMAL(6,5) NOT NULL DEFAULT 0.004 COMMENT '卖出手续费率',
    `regist_code` BOOL DEFAULT FALSE COMMENT '是否启用推荐码(启用则必须要输入正确推荐码才能注册成功):true启用 false不启用',
    `bank_no` varchar(32) DEFAULT NULL COMMENT '收款行银行卡号',
    `bank_name` varchar(32) DEFAULT NULL COMMENT '收款人姓名',
    `bank_addr` varchar(33) DEFAULT NULL COMMENT '收款人开户行地址',
    `bank_channel` BOOL DEFAULT TRUE COMMENT '银行卡支付账户:true开启 false关闭',
    `qrcode_channel` BOOL DEFAULT TRUE COMMENT '支付宝支付通道:true开启 false关闭',
    `alipay_channel` BOOL DEFAULT TRUE COMMENT '支付宝唤醒支付:true开启 false关闭',
    `contract_lever` varchar(255) NOT NULL DEFAULT '1,2,3,4,5,6,7,8,9,10' COMMENT '合约倍数:1,2,3,4,5,6,7,8,10',
    `is_support_day_contract` BOOL NOT NULL DEFAULT TRUE COMMENT '合约类型:按天 true开启 false关闭',
    `is_support_week_contract` BOOL NOT NULL DEFAULT TRUE COMMENT '合约类型:按周 true开启 false关闭',
    `is_support_month_contract` BOOL NOT NULL DEFAULT TRUE COMMENT '合约类型:按月 true开启 false关闭',
    `day_contract_fee` DECIMAL(6,5) NOT NULL DEFAULT 0.0005 COMMENT '日管理费率',
    `week_contract_fee` DECIMAL(6,5) NOT NULL DEFAULT 0.0005 COMMENT '周管理费率',
    `month_contract_fee` DECIMAL(6,5) NOT NULL DEFAULT 0.0005 COMMENT '月管理费率',
    `mini_charge_fee` DECIMAL(6,5) NOT NULL DEFAULT 5 COMMENT '最低交易手续费',
    `is_support_broker` BOOL NOT NULL DEFAULT TRUE COMMENT '是否对接券商',
    `admin_phone` varchar(33) DEFAULT NULL COMMENT '管理员手机号码'
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 银行转账表
CREATE TABLE if not exists  `transfer` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `uid` BIGINT(11) NOT NULL COMMENT '用户ID',
    `order_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '订单时间',
    `money` DECIMAL(15,2) NOT NULL COMMENT '金额',
    `type` INT(2) NOT NULL COMMENT '类型:1充值 2提现',
    `status` INT(2) NOT NULL DEFAULT '0' COMMENT '状态:0预插入 1待审核 2成功 3失败',
    `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '提现收款人',
    `bank_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '提现银行卡号',
    `channel` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '渠道',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '订单流水号',
    INDEX `idx_transfer_order_no` (`order_no`),
    INDEX `idx_transfer_uid` (`uid`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 自选股
CREATE TABLE if not exists  `portfolio` (
    `uid` BIGINT(11) NOT NULL COMMENT '用户ID',
    `code` VARCHAR(8) NOT NULL COMMENT '股票代码',
    `name` VARCHAR(32) NOT NULL COMMENT '股票名称',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uniq_portfolio` (`uid`,`code`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 委托表(挂单表)
CREATE TABLE if not exists  `entrust` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `uid` BIGINT(11) NOT NULL COMMENT '用户ID',
    `contract_id` INT(11) NOT NULL COMMENT '合约编号',
    `order_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '订单时间',
    `stock_code` VARCHAR(8) NOT NULL COMMENT '股票代码',
    `stock_name` VARCHAR(32) NOT NULL COMMENT '股票名称',
    `amount` INT(11) NOT NULL COMMENT '数量(股)',
    `price` DECIMAL(15,2) NOT NULL COMMENT '委托价格',
    `balance` DECIMAL(15,2) NOT NULL COMMENT '委托金额',
    `deal_amount` INT(11) NOT NULL COMMENT '成交数量',
    `status` INT(2) NOT NULL COMMENT '委托状态:1未成交 2成交 3已撤单 4部分部撤 5等待撤单(用户发起撤单后的状态) 6已申报,未成交',
    `entrust_bs` INT(2) NOT NULL COMMENT '交易类型:1买入 2卖出',
    `entrust_prop` INT(2)  COMMENT '委托类型:1限价 2市价',
    `position_id` INT(11) COMMENT '持仓表id_卖出时需填写',
    `fee` DECIMAL(15,2) NOT NULL COMMENT '总交易手续费',
    `is_broker_entrust` BOOL NOT NULL DEFAULT TRUE COMMENT '是否券商委托',
    `remark` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '备注失败原因',
    INDEX `idx_entrust_uid` (`uid`),
    INDEX `idx_entrust_contract_id` (`contract_id`),
    INDEX `idx_entrust_order_time` (`order_time`),
    INDEX `idx_entrust_position_id` (`position_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 买入成交表
CREATE TABLE if not exists  `buy`(
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `entrust_id` INT(11) NOT NULL COMMENT '委托表ID',
    `position_id` INT(11) NOT NULL COMMENT '已持仓:持仓表序号',
    `uid` INT(11) NOT NULL COMMENT '用户ID',
    `contract_id` INT(11) NOT NULL COMMENT '合约编号',
    `order_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '订单时间',
    `stock_code` VARCHAR(16) NOT NULL COMMENT '股票代码',
    `stock_name` VARCHAR(32) NOT NULL COMMENT '股票名称',
    `price` DECIMAL(15,2) NOT NULL COMMENT '成交价格',
    `amount` INT(11) NOT NULL COMMENT '成交数量',
    `balance` DECIMAL(15,2) NOT NULL COMMENT '成交金额',
    `entrust_prop` INT(1)  COMMENT '委托类型:1限价 2市价',
    `fee` DECIMAL(15,2) NOT NULL COMMENT '买入费用总计',
    INDEX `idx_buy` (`entrust_id`,`uid`, `contract_id`),
    INDEX `idx_buy_uid` (`uid`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 卖出成交表
CREATE TABLE if not exists  `sell`(
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `entrust_id` INT(11) NOT NULL COMMENT '委托表ID',
    `position_id` INT(11) NOT NULL COMMENT '持仓表序号',
    `uid` INT(11) NOT NULL COMMENT '用户ID',
    `contract_id` INT(11) NOT NULL COMMENT '合约编号',
    `order_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '订单时间',
    `stock_code` VARCHAR(16) NOT NULL COMMENT '股票代码',
    `stock_name` VARCHAR(32) NOT NULL COMMENT '股票名称',
    `price` DECIMAL(15,2) NOT NULL COMMENT '成交价格',
    `amount` INT(11) NOT NULL COMMENT '成交数量',
    `balance` DECIMAL(15,2) NOT NULL COMMENT '成交金额',
    `position_price` DECIMAL(15,2) NOT NULL COMMENT '持仓价格',
    `profit` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '盈亏金额',
    `entrust_prop` INT(2)  COMMENT '委托类型:1限价 2市价',
    `fee` DECIMAL(15,2) NOT NULL COMMENT '卖出费用总计',
    `mode` INT(1) NOT NULL DEFAULT 1 COMMENT '类型:1主动卖出 2系统平仓',
    `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '系统平仓原因',
    INDEX `idx_sell` (`entrust_id`,`uid`, `contract_id`, `order_time`),
    INDEX `idx_sell_uid` (`uid`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 持仓表
create table if not exists `position`(
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `uid` INT(11) NOT NULL COMMENT '用户ID',
    `entrust_id` INT(11) NOT NULL COMMENT '委托ID',
    `contract_id` INT(11) NOT NULL COMMENT '合约编号',
    `order_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '订单时间',
    `stock_code` VARCHAR(16) NOT NULL COMMENT '股票代码',
    `stock_name` VARCHAR(32) NOT NULL COMMENT '股票名称',
    `price` DECIMAL(15,2) NOT NULL COMMENT '成交价格',
    `amount` INT(11) NOT NULL COMMENT '成交数量',
    `balance` DECIMAL(15,2) NOT NULL COMMENT '成交金额',
    `freeze_amount` INT(11) NOT NULL DEFAULT 0 COMMENT '冻结股数',
    INDEX `idx_position_uid` (`uid`),
    INDEX `idx_position_entrust_id` (`entrust_id`),
    INDEX `idx_position_contract_id` (`contract_id`),
    INDEX `idx_position` (`stock_code`,`order_time`)
)engine=InnoDB,DEFAULT CHARSET=utf8mb4;

-- 历史持仓表
create table if not exists `his_position`(
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `uid` INT(11) NOT NULL COMMENT '用户ID',
    `entrust_id` INT(11) NOT NULL COMMENT '委托ID',
    `contract_id` INT(11) NOT NULL COMMENT '合约编号',
    `order_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '订单时间',
    `stock_code` VARCHAR(16) NOT NULL COMMENT '股票代码',
    `stock_name` VARCHAR(32) NOT NULL COMMENT '股票名称',
    `price` DECIMAL(15,2) NOT NULL COMMENT '成交价格',
    `amount` INT(11) NOT NULL COMMENT '成交数量',
    `balance` DECIMAL(15,2) NOT NULL COMMENT '成交金额',
    `freeze_amount` INT(11) NOT NULL DEFAULT 0 COMMENT '冻结股数',
    INDEX `idx_his_position_contract_id` (`contract_id`, `order_time`),
    INDEX `idx_his_position_uid` (`uid`)
)engine=InnoDB,DEFAULT CHARSET=utf8mb4;

-- 合约表
CREATE TABLE if not exists  `contract`(
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `uid` INT(11) NOT NULL COMMENT '用户ID',
    `init_money` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '原始保证金',
    `money` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '现保证金',
    `val_money` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '可用资金',
    `lever` INT(2) NOT NULL DEFAULT 10 COMMENT '杠杆倍数',
    `status` INT(1) NOT NULL DEFAULT 1 COMMENT '合约状态:1预申请 2操盘中 3操盘结束',
    `type` INT(1) NOT NULL COMMENT '合约类型:1按天合约 2:按周合约 3:按月合约',
    `append_money` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '追加保证金',
    `order_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '合约时间',
    `close_explain` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关闭说明',
    `close_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '合约时间',
    INDEX `idx_contract_uid` (`uid`),
    INDEX `idx_contract_id` (`id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 信息表
CREATE TABLE if not exists  `msg`(
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `uid` BIGINT(11) NOT NULL COMMENT '用户id',
    `title` VARCHAR(255) NOT NULL COMMENT '标题',
    `content` VARCHAR(255) NOT NULL DEFAULT 1 COMMENT '内容',
    `create_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '时间',
    INDEX `idx_msg`(`uid`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 日志表
CREATE TABLE if not exists  `log`(
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `title` VARCHAR(16) NOT NULL COMMENT '标题',
    `content` INT(1) NOT NULL DEFAULT 1 COMMENT '内容',
    `create_at` INT(11) NOT NULL DEFAULT 0 COMMENT '时间'
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 分红表
CREATE TABLE if not exists  `dividend` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `uid` BIGINT(11) NOT NULL COMMENT '用户ID',
    `contract_id` INT(11) NOT NULL COMMENT '合约编号',
    `position_id` INT(11) NOT NULL COMMENT '持仓表ID',
    `order_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '订单时间',
    `stock_code` VARCHAR(8) NOT NULL COMMENT '股票代码',
    `stock_name` VARCHAR(32) NOT NULL COMMENT '股票名称',
    `position_price` DECIMAL(15,2) NOT NULL COMMENT '持仓均价',
    `position_amount` INT(11) NOT NULL COMMENT '持仓股数',
    `is_buy_back` BOOL NOT NULL DEFAULT FALSE COMMENT '是否零股回购',
    `buy_back_amount` INT(11) NOT NULL COMMENT '零股回购数量',
    `buy_back_price` DECIMAL(15,2) NOT NULL COMMENT '零股回购价格',
    `dividend_amount` INT(11) NOT NULL COMMENT '派息数量(股)',
    `dividend_money` DECIMAL(15,2) NOT NULL COMMENT '派息金额',
    `type` INT(11) NOT NULL COMMENT '1送股 2:现金分红',
    `plan_explain` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '方案说明',
    INDEX `idx_dividend_uid` (`uid`),
    INDEX `idx_dividend_position_id` (`position_id`),
    INDEX `idx_dividend_contract_id` (`contract_id`,`stock_code`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 股票列表
CREATE TABLE if not exists  `stock_data` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `code` CHAR(32) NOT NULL COMMENT '股票代码',
    `name` CHAR(64) NOT NULL COMMENT '股票代码',
    `ipo_day` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上市日期',
    `status` int(1) DEFAULT NULL COMMENT '1:允许交易 2:不允许交易',
    `created_at`     DATETIME              default current_timestamp,
    `updated_at`     DATETIME              default current_timestamp on update current_timestamp,
    UNIQUE KEY `uniq_stock_data_code` (`code`),
    UNIQUE KEY `uniq_stock_data_name` (`name`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 交易日表
CREATE TABLE if not exists  `trade_calendar` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `date` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '日期',
    `trade` bool NOT NULL COMMENT '类型:true为交易日;false非交易日',
    `create_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '时间',
    UNIQUE KEY `uniq_user_name` (`date`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 合约费用表
CREATE TABLE if not exists  `contract_fee` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `uid` BIGINT(11) NOT NULL COMMENT '用户ID',
    `contract_id` INT(11) NOT NULL COMMENT '合约编号',
    `code` CHAR(32) NOT NULL DEFAULT '' COMMENT '股票代码',
    `name` CHAR(32) NOT NULL DEFAULT  '' COMMENT '股票名称',
    `amount` INT(11) NOT NULL DEFAULT 0 COMMENT '股票交易数量',
    `order_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '订单时间',
    `direction` INT(2) NOT NULL COMMENT '方向:1支出 2:收入',
    `money` DECIMAL(15,2) DEFAULT 0 COMMENT '金额',
    `type` INT(2) NOT NULL COMMENT '费用类型 1:买入手续费 2:卖出手续费 3:合约利息 4:卖出盈亏 5:追加保证金 6:扩大资金 7:合约结算',
    `detail` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '费用明细说明',
    INDEX `contract_fee_contract_id` (`contract_id`),
    INDEX `contract_fee_uid` (`uid`)
    )ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- 合约记录表,保存昨天的合约
CREATE TABLE if not exists  `contract_record`(
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `uid` INT(11) NOT NULL COMMENT '用户ID',
    `init_money` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '原始保证金',
    `money` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '现保证金',
    `val_money` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '可用资金',
    `lever` INT(2) NOT NULL DEFAULT 10 COMMENT '杠杆倍数',
    `status` INT(1) NOT NULL DEFAULT 1 COMMENT '合约状态:1预申请 2操盘中 3操盘结束',
    `type` INT(1) NOT NULL COMMENT '合约类型:1按天合约 2:按周合约 3:按月合约',
    `append_money` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '追加保证金',
    `order_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '合约时间',
    `close_explain` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关闭说明',
    `close_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '合约时间',
    INDEX `idx_contract_uid` (`uid`),
    INDEX `idx_contract_id` (`id`)
    )ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- 券商列表
CREATE TABLE if not exists  `broker` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '券商交易服务器IP地址',
    `port` INT(11) NOT NULL DEFAULT 0 COMMENT '券商交易服务器端口号',
    `version` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '通达信客户端版本号',
    `branch_no` INT(11) NOT NULL DEFAULT 0 COMMENT '营业部代码',
    `account` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '资金账号',
    `trade_account` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '交易账号,一般与登录账号一致',
    `trade_password` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '交易密码',
    `tx_password` varchar(64) NOT NULL DEFAULT '' COMMENT '通讯密码',
    `sh_holder_account` varchar(64) NOT NULL DEFAULT '' COMMENT '上海股东账号',
    `sz_holder_account` varchar(64) NOT NULL DEFAULT '' COMMENT '深证股东账号',
    `priority` INT(11) NOT NULL DEFAULT 0 COMMENT '顺序,数字越大,优先级越高',
    `status` INT(2) NOT NULL DEFAULT 0 COMMENT '状态:1激活 2冻结',
    `broker_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '券商名称',
    `val_money` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '可用资金',
    `asset` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '总资产',
    `create_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '时间'
    )ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- 券商委托记录表
CREATE TABLE if not exists  `broker_entrust` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `uid` BIGINT(11) NOT NULL COMMENT '用户ID',
    `contract_id` INT(11) NOT NULL COMMENT '合约编号',
    `entrust_id` INT(11) NOT NULL COMMENT '委托表ID',
    `order_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '订单时间',
    `stock_code` VARCHAR(8) NOT NULL COMMENT '股票代码',
    `stock_name` VARCHAR(32) NOT NULL COMMENT '股票名称',
    `entrust_amount` INT(11) NOT NULL COMMENT '委托总数量(股)',
    `entrust_price` DECIMAL(15,2) NOT NULL COMMENT '委托价格',
    `entrust_balance` DECIMAL(15,2) NOT NULL COMMENT '委托金额',
    `deal_amount` INT(11) NOT NULL DEFAULT 0 COMMENT '成交数量(股)',
    `deal_price` DECIMAL(15,2) NOT NULL DEFAULT 0 COMMENT '成交价格',
    `deal_balance` DECIMAL(15,2) NOT NULL DEFAULT 0  COMMENT '成交金额',
    `status` INT(2) NOT NULL COMMENT '委托状态:1未成交 2成交 3已撤单',
    `entrust_bs` INT(2) NOT NULL COMMENT '交易类型:1买入 2卖出',
    `entrust_prop` INT(2)  COMMENT '委托类型:1限价 2市价',
    `fee` DECIMAL(15,2) NOT NULL COMMENT '券商交易手续费',
    `broker_id` INT(11) NOT NULL DEFAULT 0 COMMENT '券商列表ID',
    `broker_entrust_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '券商委托编号',
    `broker_withdraw_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '券商撤单编号',
    INDEX `idx_entrust_uid` (`uid`),
    INDEX `idx_entrust_contract_id` (`contract_id`),
    INDEX `idx_broker_entrust_id` (`entrust_id`),
    INDEX `idx_entrust_idx_entrust_no_order_time` (`idx_entrust_no`,`order_time`),
    INDEX `idx_entrust_broker_id` (`broker_id`)
    )ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 角色
CREATE TABLE if not exists  `role` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `user_name` VARCHAR(254) NOT NULL COMMENT '用户名',
    `password` VARCHAR(254) NOT NULL COMMENT '密码',
    `status` bool NOT NULL DEFAULT TRUE COMMENT '状态:true有效 false无效',
    `is_admin` bool NOT NULL DEFAULT false COMMENT '是否超级管理员',
    `phone` VARCHAR(254) NOT NULL COMMENT '手机号码',
    INDEX `idx_role_user_name` (`user_name`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 插入超级管理员
insert into `role`(user_name,password,status) values('admin','123456',true);

-- 角色
CREATE TABLE if not exists  `role_module` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `role_id` BIGINT(11)  NOT NULL COMMENT '角色ID',
    `module` VARCHAR(254) NOT NULL COMMENT '模块名称',
    `module_id` BIGINT(11)  NOT NULL COMMENT '模块ID',
    INDEX `idx_role_module_role_id` (`role_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- 券商请求错误记录
CREATE TABLE if not exists  `broker_error_log` (
    `id` BIGINT(11) PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    `url` VARCHAR(1024) NOT NULL COMMENT '请求地址',
    `error` VARCHAR(2048) NOT NULL COMMENT '错误信息',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    )ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
use stock;
alter table sysparam add `kcb_limit_pct` DECIMAL(6,4) NOT NULL DEFAULT 17.5 comment '科创板涨跌幅限制' after cyb_limit_pct;

alter table users add `is_delete` bool NOT NULL default false comment '是否已经删除' ;
alter table users add index user_id_is_del(`id`,`is_delete`);

alter table entrust add `is_delete` bool NOT NULL default false comment '是否已经删除' ;
alter table entrust add index entrust_id_is_del(`id`,`is_delete`);

alter table entrust add `mode` INT NOT NULL DEFAULT 0 COMMENT '委托方式:0用户委托,1系统委托';

-- 买入卖出记录，根据entrust的is_delete标志筛选出;
-- 禁止非超级管理员用户直到用户密码

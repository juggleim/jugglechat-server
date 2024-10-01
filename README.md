# JuggleChat

一个基于 JuggleIM 的开源即时通讯软件，覆盖全平台。
快速体验：https://www.jugglechat.com/docs/download/integrate/

## 架构图
待补充

## 快速部署
注意，部署此demo前，请先安装JuggleIM，部署文档参考：https://github.com/juggleim/im-server/blob/master/README.md

### 1. 安装并初始化 MySQL

#### 1) 安装 MySQL
略

#### 2) 创建DB实例
```
CREATE SCHEMA `app_db` ;
```

#### 3) 初始化表结构
初始化表结构的sql文件在  jugglechat-server/docs/appdb.sql , 导入命令如下：
```
mysql -u{db_user} -p{db_password} app_db < appdb.sql
```


### 2. 启动jugglechat-server

#### 1) 运行目录
运行目录为 jugglechat-server，其中 conf 目录下存放的是配置文件。

#### 2) 编辑配置文件

配置文件位置：jugglechat-server/conf/config.yml
```
port: 8091                # jugglechat-server 的监听端口

log:                      # 日志目录
  logPath: ./logs      
  logName: app-server

mysql:                    # db 配置
  user: <db_user>
  password: <db_password>
  address: 127.0.0.1:3306
  name: app_db

qiniu:                    # 文件存储配置，demo中使用七牛作为文件存储，用于存储用户头像，群头像等
  accessKey: <qiniu_ak>
  secretKey: <qiniu_sk>
  bucket: <qiniu_bucket>
  domain: <bucket_domain>

baidusms:                   # demo中使用百度的短信服务，用于短验登录
  apiKey: <baidu_sms_ak>
  secretKey: <baidu_sms_sk>

im:                          # demo 所使用的 IM 服务器地址和 租户的app_key/app_secret
  appKey: <juggleim_appkey>
  appSecret: <juggleim_secret>
  apiUrl: https://api.juggleim.com

```

#### 3) 启动jugglechat-server

在 jugglechat-server 目录下，执行如下命令：
```
go run main.go
```

#### 4) 部署/打包JuggleChat的客户端

各端部署/打包文档地址：

| 端类型 | 文档地址| 备注 |
| ----:|:-------:|:-----|
|Web端|https://github.com/juggleim/jugglechat-web||
|桌面端|https://github.com/juggleim/jugglechat-desktop||
|Android端|https://github.com/juggleim/jugglechat-android||
|iOS端|https://github.com/juggleim/jugglechat-ios||
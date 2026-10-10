# JuggleChat Server API 接口文档

本文档根据当前仓库的路由、Handler、请求模型和 Service 实现整理，覆盖所有已注册的 HTTP 业务接口。

- 主业务接口：85 个，统一使用 `/jim` 前缀。
- 助手回调接口：1 个，使用独立的 `/callback` 前缀和端口。
- 合计：86 个 `method + path` 接口。
- 文档反映当前代码行为；代码中没有显式实现的校验、权限或数据清理能力，不应视为接口保证。

## 1. 服务地址

### 1.1 主业务服务

默认地址：

```text
http://localhost:8070
```

业务 API 根路径：

```text
http://localhost:8070/jim
```

端口由配置项 `port` 决定，默认值为 `8070`。

### 1.2 助手回调服务

回调服务使用独立端口，仅当 `callbackPort > 0` 时启动：

```text
http://localhost:<callbackPort>/callback
```

## 2. 通用请求约定

### 2.1 请求头

| Header | 必填范围 | 说明 |
| --- | --- | --- |
| `appkey` | 除服务发现接口外 | 应用 AppKey |
| `Authorization` | 受保护接口 | 登录成功返回的用户 Token，或 `Bearer <API_KEY>` |
| `version` | 否 | 客户端版本；部分搜索、好友和群接口会进行版本阻断 |
| `Content-Type` | JSON POST 请求 | `application/json` |

普通用户 Token 应直接传递：

```http
Authorization: USER_AUTHORIZATION_TOKEN
```

不要给用户 Token 添加 `Bearer `。`Bearer <API_KEY>` 是应用 API Key 鉴权方式，但它不会写入当前用户 ID，因此不适用于大多数依赖当前用户身份的接口。

服务端会在主业务 API 的响应头中添加：

```http
request-id: <服务端生成的短 UUID>
```

### 2.2 鉴权等级

本文使用以下标记：

| 标记 | 含义 |
| --- | --- |
| 公开 | 不要求 `appkey` 和 `Authorization` |
| AppKey | 要求 `appkey`，不要求用户登录 |
| 登录 | 要求 `appkey` 和 `Authorization` |
| 回调公开 | 独立回调服务当前未配置鉴权中间件 |

完全公开接口：

- `GET /jim/serverinfos`
- `GET /jim/servers/info`
- `GET /jim/servers/list`

只要求 AppKey 的接口：

- `POST /jim/login`
- `POST /jim/register`
- `GET /jim/login/qrcode`
- `POST /jim/login/qrcode/check`
- `POST /jim/sms/send`
- `POST /jim/sms_login`
- `POST /jim/sms/login`
- `POST /jim/email/send`
- `POST /jim/email/login`

除上述接口外，其他 `/jim` 接口均要求登录。

## 3. 通用响应

### 3.1 成功响应

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

无返回数据时：

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

### 3.2 错误响应

```json
{
  "code": 17003,
  "msg": ""
}
```

当前错误映射除成功码外没有配置 HTTP 状态和错误消息，因此多数业务错误仍返回 HTTP `200`，且 `msg` 为空。客户端必须以响应 JSON 的 `code` 判断业务结果。

当 JSON 语法错误时，Gin 的 `BindJSON` 可能先写入 HTTP `400`。客户端应同时兼顾 HTTP 状态和业务 `code`。

### 3.3 通用错误码

| code | 说明 |
| ---: | --- |
| `0` | 成功 |
| `17000` | 通用业务、数据库或 SDK 错误 |
| `17001` | 缺少 `appkey` |
| `17002` | 应用不存在或 IM SDK 不可用 |
| `17003` | 请求体、JSON、Query 或必要参数非法 |
| `17004` | 内部或下游调用失败 |
| `17005` | 未登录或认证失败 |
| `17006` | 二维码登录尚未确认，应继续轮询 |
| `17007` | 登录二维码过期 |
| `17008` | 短信或邮件验证码发送失败 |
| `17009` | 验证码错误、无记录或过期 |
| `17010` | 未配置翻译引擎 |
| `17011` | 用户或账号已存在 |
| `17012` | 用户不存在 |
| `17013` | 密码错误 |
| `17014` | 手机号已绑定 |
| `17015` | 邮箱已绑定 |
| `17016` | 内容命中敏感词 |
| `17017` | 参数错误 |
| `17018` | 无权限或当前版本被阻断 |
| `17101` | 对方拒绝好友申请 |
| `17200` | 群组通用错误 |
| `17201` | 用户已经是群成员 |
| `17202` | 无群主或群管理员权限 |
| `17401` | 未配置文件存储服务 |
| `17402` | 文件上传签名生成失败 |

部分接口会直接透传 JuggleIM SDK 的错误码，因此运行时可能出现本表未列出的错误码。

## 4. 公共数据结构

### 4.1 登录用户 `LoginUserResp`

```json
{
  "user_id": "user_123",
  "authorization": "business-token",
  "nickname": "张三",
  "avatar": "https://example.com/avatar.png",
  "status": 0,
  "im_token": "im-token"
}
```

### 4.2 用户 `UserObj`

```json
{
  "user_id": "user_123",
  "nickname": "张三",
  "avatar": "https://example.com/avatar.png",
  "pinyin": "zhangsan",
  "user_type": 0,
  "phone": "",
  "email": "",
  "account": "zhangsan",
  "status": 0,
  "is_friend": false,
  "friend_info": {
    "is_friend": false,
    "display_name": ""
  },
  "is_block": false,
  "settings": {
    "language": "zh",
    "friend_verify_type": 0,
    "grp_verify_type": 0,
    "undisturb": "",
    "exts": {}
  }
}
```

### 4.3 群信息 `GrpInfo`

```json
{
  "group_id": "group_123",
  "group_name": "项目组",
  "group_portrait": "https://example.com/group.png",
  "members": [],
  "member_count": 0,
  "owner": {},
  "my_role": 0,
  "group_management": {},
  "grp_display_name": "",
  "member_offset": ""
}
```

`my_role`：`0` 普通成员，`1` 群主，`2` 管理员，`3` 非群成员。

### 4.4 群成员 `GroupMemberInfo`

```json
{
  "user_id": "user_123",
  "nickname": "张三",
  "avatar": "",
  "grp_display_name": "群昵称",
  "member_type": 0,
  "role": 0,
  "is_mute": 0,
  "friend_info": {
    "is_friend": false,
    "display_name": ""
  }
}
```

## 5. 服务发现接口（3 个）

### 5.1 查询服务信息

```http
GET /jim/serverinfos
GET /jim/servers/info
```

两个路径为兼容别名，行为完全相同。

- 鉴权：公开。
- Query：`app_key`、`no`，至少传一个；优先使用 `app_key`。
- 成功数据：找到应用时返回 `server_info_plain`，其值是 JSON 字符串；未找到时当前实现返回 `server_info` 加密字符串。
- 主要错误：`17017`。

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "server_info_plain": "{\"app_key\":\"...\",\"im_servers\":[\"...\"],\"app_servers\":[\"...\"]}"
  }
}
```

### 5.2 查询服务列表

```http
GET /jim/servers/list?limit=100&offset=
```

- 鉴权：公开。
- Query：`limit` 默认 `100`；`offset` 为编码后的数据库游标。
- 返回：`items` 服务配置列表和下一页 `offset`。
- 主要错误：`17004`。

## 6. 登录与注册接口（10 个）

### 6.1 密码登录

```http
POST /jim/login
```

- 鉴权：AppKey。
- Body：`account`、`phone`、`email` 三者至少一个非空；`password` 必填；`code` 不使用。
- 查找优先级：`account` > `phone` > `email`。
- 返回：`LoginUserResp`。
- 主要错误：`17003`、`17012`、`17013`、`17002`、`17004`。

```json
{
  "account": "zhangsan",
  "phone": "",
  "email": "",
  "password": "password"
}
```

### 6.2 注册

```http
POST /jim/register
```

- 鉴权：AppKey。
- Body：同密码登录，`password` 必填，三种标识至少一个非空。
- 账号注册时，`account` 必须匹配 `^[a-zA-Z0-9]{6,20}$`。
- 手机号注册会校验 `code`；邮箱注册当前不校验 `code`。
- 返回：`data: null`。
- 主要错误：`17003`、`17009`、`17011`。

### 6.3 发送短信验证码

```http
POST /jim/sms/send
```

- 鉴权：AppKey。
- Body：`phone` 必填，`code` 不使用。
- 返回：`data: null`。
- 主要错误：`17003`、`17008`。

```json
{"phone":"13800000000"}
```

### 6.4 短信验证码登录

```http
POST /jim/sms_login
POST /jim/sms/login
```

两个路径为兼容别名。

- 鉴权：AppKey。
- Body：`phone` 必填，`code` 用于验证码校验。
- 用户不存在时自动注册。
- 返回：`LoginUserResp`。
- 主要错误：`17003`、`17009`、`17005`、`17002`、`17004`。

```json
{"phone":"13800000000","code":"123456"}
```

### 6.5 发送邮件验证码

```http
POST /jim/email/send
```

- 鉴权：AppKey。
- Body：`email` 必填，当前没有格式校验。
- 返回：`data: null`。
- 主要错误：`17003`、`17008`。

### 6.6 邮箱验证码登录

```http
POST /jim/email/login
```

- 鉴权：AppKey。
- Body：`email` 必填，`code` 用于验证码校验。
- 用户不存在时自动注册。
- 返回：`LoginUserResp`。
- 主要错误：`17003`、`17009`、`17005`、`17002`、`17004`。

### 6.7 生成登录二维码

```http
GET /jim/login/qrcode
```

- 鉴权：AppKey。
- 无参数。
- 二维码记录有效期为 10 分钟。
- 返回：二维码记录 `id` 和 PNG Base64 字符串 `qr_code`。

```json
{"id":"uuid","qr_code":"<PNG base64>"}
```

### 6.8 检查登录二维码

```http
POST /jim/login/qrcode/check
```

- 鉴权：AppKey。
- Body：`id` 必填。
- 已确认时返回 `LoginUserResp`。
- 主要状态：`17006` 尚未确认，`17007` 已过期。

### 6.9 确认登录二维码

```http
POST /jim/login/qrcode/confirm
```

- 鉴权：登录。
- Body：`id` 必填。
- 将二维码记录绑定到当前登录用户。
- 返回：`data: null`。

## 7. 文件、翻译和同步接口（3 个）

### 7.1 获取文件上传凭证

```http
POST /jim/file_cred
```

- 鉴权：登录。
- Body：`file_type`、`ext`；Handler 只校验 JSON。
- `file_type`：`0` 默认、`1` 图片、`2` 音频、`3` 视频、`4` 文件、`5` 日志。
- 返回：按存储类型返回七牛凭证 `qiniu_resp` 或预签名数据 `pre_sign_resp`。
- 主要错误：`17003`、`17401`、`17402`。

```json
{"file_type":1,"ext":"png"}
```

### 7.2 翻译

```http
POST /jim/translate
```

- 鉴权：登录。
- Body：`items` 和 `target_lang` 必填；`source_lang` 可空。
- 返回：同一结构，其中 `items[].content` 被替换为翻译结果。
- 主要错误：`17003`、`17010`。

```json
{
  "items":[{"key":"title","content":"hello"}],
  "source_lang":"en",
  "target_lang":"zh"
}
```

### 7.3 同步用户配置

```http
GET /jim/syncconfs
```

- 鉴权：登录。
- 无参数。
- 当前固定返回空对象 `{}`。

## 8. 用户接口（16 个）

### 8.1 查询用户信息

```http
GET /jim/users/info?user_id=user_123
```

- 鉴权：登录。
- Query：`user_id` 必填。
- 返回：`UserObj`。查询自己时包含设置；查询他人时包含好友和黑名单状态。
- 当前用户不存在时可能返回仅包含 `user_id` 的空壳对象。
- 主要错误：`17003`。

### 8.2 设置登录账号

```http
POST /jim/users/setaccount
```

- 鉴权：登录。
- Body：`account` 必填；结构中的 `password` 当前不使用。
- 返回：`data: null`。
- 主要错误：`17003`、`17011`。

```json
{"account":"zhangsan","password":""}
```

### 8.3 更新用户资料

```http
POST /jim/users/update
```

- 鉴权：登录。
- Body：实际使用 `user_id`、`nickname`、`avatar`。
- 昵称会进行敏感词检查。
- 当前 Service 未校验 `user_id` 是否等于当前用户。
- 返回：`data: null`。
- 主要错误：`17003`、`17016`。

```json
{"user_id":"user_123","nickname":"张三","avatar":"https://example.com/a.png"}
```

### 8.4 删除当前用户

```http
POST /jim/users/del
```

- 鉴权：登录，必须使用用户 Token。
- Body：`user_id` 必填，且必须等于 Token 中的当前用户 ID。
- 返回：`data: null`。
- 主要错误：`17003`、`17018`、`17000`。

```json
{"user_id":"user_123"}
```

当前实现仅物理删除 `users` 表中的当前应用用户记录，不会级联清理好友、群组、设置、消息或 JuggleIM 服务数据，也不会撤销已签发的业务 Token 和 IM Token。

### 8.5 修改密码

```http
POST /jim/users/updpass
```

- 鉴权：登录。
- Body：`user_id`、`password`、`new_password` 均必填。
- 当前 Service 未校验 `user_id` 是否等于当前用户。
- 返回：`data: null`。
- 主要错误：`17003`、`17012`、`17013`、`17000`。

### 8.6 更新用户设置

```http
POST /jim/users/updsettings
```

- 鉴权：登录。
- Body：`language`、`friend_verify_type`、`grp_verify_type`、`undisturb`、`exts`。
- `exts` 不能覆盖内置键：`phone`、`language`、`undisturb`、`friend_verify_type`、`grp_verify_type`。
- 返回：`data: null`。

```json
{
  "language":"zh",
  "friend_verify_type":1,
  "grp_verify_type":0,
  "undisturb":"",
  "exts":{"theme":"dark"}
}
```

### 8.7 搜索用户

```http
POST /jim/users/search
```

- 鉴权：登录。
- Body：`keyword`、`phone` 至少一个非空；有 `phone` 时优先精确搜索。
- `limit`、`offset` 当前未被 Service 使用。
- 返回：`{"items":[UserObj],"offset":""}`。

### 8.8 获取个人二维码

```http
GET /jim/users/qrcode
```

- 鉴权：登录。
- 无参数。
- 二维码动作是 `add_friend`，用户 ID 为当前登录用户。
- 返回：`{"qr_code":"<PNG base64>"}`。

### 8.9 添加黑名单

```http
POST /jim/users/blockusers/add
```

- 鉴权：登录。
- Body：非空数组 `block_user_ids`。
- 返回：`data: null`。

```json
{"block_user_ids":["user_456"]}
```

### 8.10 移除黑名单

```http
POST /jim/users/blockusers/del
```

- 鉴权：登录。
- Body：非空数组 `block_user_ids`。
- 返回：`data: null`。

### 8.11 查询黑名单

```http
GET /jim/users/blockusers/list?offset=&count=20
```

- 鉴权：登录。
- Query：`offset` 游标；`count` 默认 `20`。
- 返回：`{"items":[UserObj],"offset":"..."}`，列表项的 `is_block` 为 `true`。

### 8.12 发送绑定邮箱验证码

```http
POST /jim/users/bindemail/send
```

- 鉴权：登录。
- Body：当前 Handler 要求 `email` 和 `code` 都非空，但发送逻辑不使用传入的 `code`。
- 主要错误：`17003`、`17000`、`17015`、`17008`。

### 8.13 绑定邮箱

```http
POST /jim/users/bindemail
```

- 鉴权：登录。
- Body：`email`、`code` 均必填。
- 返回：`data: null`。
- 主要错误：`17003`、`17009`、`17000`。

### 8.14 发送绑定手机号验证码

```http
POST /jim/users/bindphone/send
```

- 鉴权：登录。
- Body：当前 Handler 要求 `phone` 和 `code` 都非空，但发送逻辑不使用传入的 `code`。
- 主要错误：`17003`、`17000`、`17014`、`17008`。

### 8.15 绑定手机号

```http
POST /jim/users/bindphone
```

- 鉴权：登录。
- Body：`phone`、`code` 均必填。
- 返回：`data: null`。
- 主要错误：`17003`、`17009`、`17000`。

### 8.16 批量查询在线状态

```http
POST /jim/users/onlinestatus
```

- 鉴权：登录。
- Body：非空数组 `user_ids`。
- 返回：JuggleIM SDK 的在线状态响应。
- 主要错误：`17003`、`17002`、`17004` 及 SDK 错误码。

## 9. Telegram Bot 接口（4 个）

### 9.1 添加 Telegram Bot

```http
POST /jim/telegrambots/add
```

- 鉴权：登录。
- Body：实际使用 `bot_name`、`bot_token`；当前只校验 JSON。
- 返回：`bot_id` 和 `bot_name`。

```json
{"bot_name":"通知机器人","bot_token":"telegram-token"}
```

### 9.2 删除 Telegram Bot

```http
POST /jim/telegrambots/del
```

- 鉴权：登录。
- Body：`bot_id`。
- 无效或无法解码的 ID 当前可能被静默忽略。
- 返回：`data: null`。

### 9.3 批量删除 Telegram Bot

```http
POST /jim/telegrambots/batchdel
```

- 鉴权：登录。
- Body：`bot_ids` 数组。
- 返回：`data: null`。

### 9.4 查询 Telegram Bot 列表

```http
GET /jim/telegrambots/list?offset=&count=20
```

- 鉴权：登录。
- Query：`offset`、`count`，`count` 默认 `20`。
- 返回：`items` 和下一页 `offset`。
- 当前返回项包含完整 `bot_token`，调用方应注意敏感信息保护。

## 10. 群组基本接口（17 个）

### 10.1 创建群组

```http
POST /jim/groups/add
POST /jim/groups/create
```

两个路径为兼容别名。

- 鉴权：登录。
- Body：`group_name`、`group_portrait`；成员可通过 `member_ids` 或 `members[].user_id` 传入。
- 当前用户会自动加入并成为创建者。
- 返回：`GrpInfo`。
- 主要错误：`17003`、`17016` 及 SDK 错误码。

```json
{
  "group_name":"项目组",
  "group_portrait":"",
  "member_ids":["user_456"]
}
```

### 10.2 更新群组

```http
POST /jim/groups/update
```

- 鉴权：登录。
- Body：`group_id`、`group_name`、`group_portrait`。
- 群名进行敏感词检查；当前 Service 未执行群管理权限校验。
- 返回：`data: null`。

### 10.3 解散群组

```http
POST /jim/groups/dissolve
```

- 鉴权：登录。
- Body：`group_id`。
- 仅群主可以解散。
- 主要错误：`17200`。

### 10.4 退出群组

```http
POST /jim/groups/quit
```

- 鉴权：登录。
- Body：`group_id`。
- 返回：`data: null`。

### 10.5 添加群成员

```http
POST /jim/groups/members/add
```

- 鉴权：登录。
- Body：`group_id` 和 `members[].user_id`；该 Handler 不读取 `member_ids`。
- 当前 Service 未执行操作者权限校验。

```json
{"group_id":"group_123","members":[{"user_id":"user_456"}]}
```

### 10.6 删除群成员

```http
POST /jim/groups/members/del
```

- 鉴权：登录。
- Body：`group_id`、非空 `member_ids`。
- 当前 Service 未执行操作者权限校验。

### 10.7 查询群信息

```http
GET /jim/groups/info?group_id=group_123
```

- 鉴权：登录。
- Query：`group_id` 必填。
- 返回：`GrpInfo`；非群成员仍可获得基本信息，但 `my_role=3`。
- 内嵌成员最多 20 个，后续游标为 `member_offset`。

### 10.8 查询我的群组

```http
GET /jim/groups/mygroups?offset=&count=20
```

- 鉴权：登录。
- Query：`offset`、`count`，默认 `20`。
- 返回：`items` 和下一页 `offset`。

### 10.9 搜索我的群组

```http
POST /jim/groups/mygroups/search
```

- 鉴权：登录。
- Body：`keyword` 必填；`limit` 默认 `100`；`offset` 可空。
- 返回：群组分页列表。

### 10.10 查询群成员列表

```http
GET /jim/groups/members/list?group_id=group_123&offset=&limit=100
```

- 鉴权：登录。
- Query：`group_id`、`offset`、`limit`；`limit` 默认 `100`。
- 返回：`GroupMemberInfo` 列表和下一页 `offset`。

### 10.11 检查群成员

```http
POST /jim/groups/members/check
```

- 鉴权：登录。
- Body：`group_id`、`member_ids`。
- 返回：成员存在映射。

```json
{
  "group_id":"group_123",
  "member_exist_map":{"user_456":true}
}
```

### 10.12 搜索群成员

```http
POST /jim/groups/members/search
```

- 鉴权：登录。
- Body：`group_id`、`key` 必填；`offset` 可空；`limit` 默认 `100`。
- 受 `version` Header 的应用版本阻断策略影响。
- 主要错误：`17003`、`17018`。

### 10.13 设置群公告

```http
POST /jim/groups/setgrpannouncement
```

- 鉴权：登录。
- Body：`group_id`、`content`。
- 非空公告会发送一条 `@all` 群消息。
- 当前 Service 未执行群管理权限校验。

### 10.14 获取群公告

```http
GET /jim/groups/getgrpannouncement?group_id=group_123
```

- 鉴权：登录。
- Query：`group_id`。
- 返回：`{"group_id":"...","content":"..."}`；未找到时 `content` 为空。

### 10.15 设置群昵称

```http
POST /jim/groups/setdisplayname
```

- 鉴权：登录。
- Body：`group_id`、`grp_display_name`。
- 修改当前用户在该群的昵称。

### 10.16 获取群二维码

```http
GET /jim/groups/qrcode?group_id=group_123
```

- 鉴权：登录。
- Query：`group_id` 必填。
- 返回：`{"qr_code":"<PNG base64>"}`。
- 二维码动作是 `join_group`，并包含群 ID 和当前用户 ID。

## 11. 群申请与邀请接口（8 个）

### 11.1 申请加入群组

```http
POST /jim/groups/apply
```

- 鉴权：登录。
- Body：`group_id` 必填。
- 当前实现直接加入群组，没有根据群审核配置创建待审核申请。
- 主要错误：`17003`、`17201`。

### 11.2 邀请加入群组

```http
POST /jim/groups/invite
```

- 鉴权：登录。
- Body：`group_id`、非空 `member_ids`。
- 返回：`reason` 和逐用户 `results`。
- `reason`：`0` 直接邀请成功，`1` 待审核，`2` 拒绝邀请，`3` 重复邀请。
- 当前 Service 未执行邀请权限校验。

### 11.3 确认群申请或邀请

```http
POST /jim/groups/grpapplications/confirm
```

- 鉴权：登录。
- Body：`application_id`、`is_agree`。
- 当前 Service 未执行管理员权限校验，且主要处理邀请类型记录。

```json
{"application_id":"encoded-id","is_agree":true}
```

### 11.4 群申请列表分页约定

以下 5 个列表接口共享分页参数：

| Query | 说明 |
| --- | --- |
| `start` | 必须能转换为整数；当前实现中缺失或空值也会返回 `17003` |
| `count` | 默认 `20`，有效范围 `1`–`50` |
| `order` | `0` 倒序，`1` 正序，非法值回退 `0` |

统一返回 `items`，列表项可能包含 `application_id`、`grp_info`、`apply_type`、`sponsor`、`recipient`、`inviter`、`operator`、`apply_time`、`status`。

### 11.5 查询我发起的群申请

```http
GET /jim/groups/myapplications?start=0&count=20&order=0
```

- 鉴权：登录。

### 11.6 查询我的待处理群邀请

```http
GET /jim/groups/mypendinginvitations?start=0&count=20&order=0
```

- 鉴权：登录。

### 11.7 查询群邀请记录

```http
GET /jim/groups/grpinvitations?group_id=group_123&start=0&count=20&order=0
```

- 鉴权：登录。

### 11.8 查询群待处理申请

```http
GET /jim/groups/grppendingapplications?group_id=group_123&start=0&count=20&order=0
```

- 鉴权：登录。

### 11.9 查询群申请记录

```http
GET /jim/groups/grpapplications?group_id=group_123&start=0&count=20&order=0
```

- 鉴权：登录。
- 返回的信息比其他申请列表更完整。

## 12. 群管理接口（9 个）

> 当前多个群管理 Service 留有权限检查 TODO。要求登录不等于已经执行群主或管理员对象级权限校验，调用方和部署方应按当前代码审查结果评估风险。

### 12.1 转让群主

```http
POST /jim/groups/management/chgowner
```

- Body：`group_id`、`owner_id`。
- 返回：`data: null`。

### 12.2 添加群管理员

```http
POST /jim/groups/management/administrators/add
```

- Body：`group_id`、`admin_ids`。

### 12.3 删除群管理员

```http
POST /jim/groups/management/administrators/del
```

- Body：`group_id`、`admin_ids`。

### 12.4 查询群管理员

```http
GET /jim/groups/management/administrators/list?group_id=group_123
```

- Query：`group_id`。
- 返回：`{"group_id":"...","items":[GroupMemberInfo]}`。

### 12.5 设置全群禁言

```http
POST /jim/groups/management/setmute
```

- Body：`group_id`、`is_mute`。

### 12.6 设置成员禁言

```http
POST /jim/groups/management/setgrpmembersmute
```

- Body：`group_id`、非空 `member_ids`、`is_mute`。

### 12.7 设置入群审核类型

```http
POST /jim/groups/management/setgrpverifytype
```

- Body：`group_id`、`verify_type`。
- 预期值：`0` 无需审核，`1` 需要审核，`2` 拒绝入群；当前没有枚举值校验。

### 12.8 设置历史消息可见

```http
POST /jim/groups/management/sethismsgvisible
```

- Body：`group_id`、`group_his_msg_visible`。
- 大于 `0` 表示可见，否则不可见。

### 12.9 批量设置群管理权限

```http
POST /jim/groups/management/set
```

- Body：`group_id` 和需要修改的权限字段。
- 可选权限字段：
  - `group_edit_msg_right`
  - `group_add_member_right`
  - `group_mention_all_right`
  - `group_top_msg_right`
  - `group_send_msg_right`
  - `group_set_msg_life_right`
  - `group_apply_friend_right`
- 只有 JSON 中实际出现的指针字段会被更新。
- 主要错误：`17003`、`17004`。

本节全部接口均要求登录。

## 13. 好友接口（9 个）

### 13.1 查询好友列表

```http
GET /jim/friends/list?page=1&size=20&order_tag=
```

- 鉴权：登录。
- Query：`page` 默认 `1`，`size` 默认 `20`，`order_tag` 可空。
- 返回：`{"items":[UserObj]}`。

### 13.2 设置好友备注

```http
POST /jim/friends/setdisplayname
```

- 鉴权：登录。
- Body：`friend_id`、`friend_display_name`。
- 返回：`data: null`。

### 13.3 搜索好友

```http
POST /jim/friends/search
```

- 鉴权：登录。
- Body：`key` 必填；`offset` 可空；`limit` 默认 `100`。
- 受 `version` Header 的版本阻断策略影响。
- 主要错误：`17003`、`17018`。

### 13.4 申请添加好友

```http
POST /jim/friends/apply
```

- 鉴权：登录。
- Body：`friend_id`。
- 根据对方设置执行拒绝、创建申请或直接建立双向好友关系。
- 受 `version` Header 的版本阻断策略影响。
- 主要错误：`17003`、`17018`、`17101`。

### 13.5 确认好友申请

```http
POST /jim/friends/confirm
```

- 鉴权：登录。
- Body：`sponsor_id`、`is_agree`。
- 同意时建立双向好友关系。

### 13.6 删除好友

```http
POST /jim/friends/del
```

- 鉴权：登录。
- Body：`friend_ids` 数组。
- 删除双向好友关系。

```json
{"friend_ids":["user_456"]}
```

### 13.7 好友申请列表分页约定

以下 3 个接口共享参数：

| Query | 说明 |
| --- | --- |
| `start` | 必须为整数；缺失或非法返回 `17003` |
| `count` | 默认 `20`，有效范围 `1`–`50` |
| `order` | `0` 倒序，`1` 正序 |

返回项可能包含 `recipient`、`sponsor`、`target_user`、`is_sponsor`、`status`、`apply_time`。

### 13.8 查询全部相关好友申请

```http
GET /jim/friends/applications?start=0&count=20&order=0
```

### 13.9 查询我发起的好友申请

```http
GET /jim/friends/myapplications?start=0&count=20&order=0
```

### 13.10 查询我的待处理好友申请

```http
GET /jim/friends/mypendingapplications?start=0&count=20&order=0
```

## 14. 消息管理接口（2 个）

两个接口当前只支持群会话，要求 `channel_type == 2`，并检查当前用户是否为群主或管理员。

### 14.1 管理员撤回消息

```http
POST /jim/messages/recall
```

- 鉴权：登录。
- Body：`target_id` 必填、`channel_type` 固定为 `2`、`msg_id` 必填；还可传 `from_id`、`sub_channel`、`msg_time`、`exts`。
- 返回：`data: null`。
- 主要错误：`17003`、`17202`、`17000`。

```json
{
  "from_id":"user_456",
  "target_id":"group_123",
  "channel_type":2,
  "msg_id":"msg_123",
  "msg_time":0
}
```

### 14.2 管理员删除消息

```http
POST /jim/messages/del
```

- 鉴权：登录。
- Body：`target_id` 必填、`channel_type` 固定为 `2`、`msgs` 非空。
- SDK 删除范围固定为 `DelScope=1`。
- 返回：`data: null`。
- 主要错误：`17003`、`17202`、`17000`。

```json
{
  "target_id":"group_123",
  "channel_type":2,
  "msgs":[{"msg_id":"msg_123","msg_time":0,"msg_read_index":0}]
}
```

## 15. 会话配置接口（2 个）

当前实际支持的配置键是 `msg_life_time`，代码注释单位为毫秒。

### 15.1 获取会话配置

```http
GET /jim/converconfs/get?target_id=user_456&conver_type=1&sub_channel=
```

- 鉴权：登录。
- Query：`target_id` 必填；`conver_type` 必须大于 `0`；`sub_channel` 可空。
- 返回：动态对象，例如 `{"msg_life_time":86400000}`；无配置时返回 `{}`。

### 15.2 设置会话配置

```http
POST /jim/converconfs/set
```

- 鉴权：登录。
- Body：`target_id` 必填、`conver_type > 0`、`confs` 非空。
- 当前只保存 `msg_life_time`，其他键会被忽略。
- 返回：`data: null`。

```json
{
  "target_id":"user_456",
  "sub_channel":"",
  "conver_type":1,
  "confs":{"msg_life_time":86400000}
}
```

## 16. 应用接口（1 个）

### 16.1 查询应用列表

```http
GET /jim/applications/list?page=1&size=20&order=0
```

- 鉴权：登录。
- Query：`page` 默认 `1`、`size` 默认 `20`、`order > 0` 表示正序。
- 返回：应用 `items`、`page`、`size`。
- 应用项字段：`app_id`、`app_name`、`app_icon`、`app_desc`、`app_url`、`app_order`、`created_time`、`updated_time`。
- 当前数据库错误会返回数值 `1014`。

## 17. 反馈接口（1 个）

### 17.1 新增反馈

```http
POST /jim/feedbacks/add
```

- 鉴权：登录。
- Body：`category` 必填；可传 `text`、`images`、`videos`。
- 请求模型中的 `app_key`、`user`、`created_time`、`updated_time` 不从客户端使用。
- 返回：`data: null`。
- 主要错误：`17003`、`17000`。

```json
{
  "category":"功能建议",
  "text":"建议内容",
  "images":["https://example.com/1.png"],
  "videos":[]
}
```

## 18. 助手回调接口（1 个）

### 18.1 助手消息回调

```http
POST /callback/assistants/msgcallback
```

- 服务：独立 callback 端口。
- 鉴权：回调公开；当前未校验 Header、签名或来源。
- Body：

```json
{
  "app_key":"app-key",
  "sender":"user_123",
  "receiver":"assistant_123",
  "conver_type":1,
  "msg_type":"jg:text",
  "msg_content":"{\"content\":\"hello\"}",
  "msg_id":"msg_123",
  "msg_time":0,
  "mention_info":{
    "mention_type":"",
    "target_user_ids":[]
  },
  "token":""
}
```

- 服务端忽略请求中的 `token`，使用 `app_key + sender` 重新生成 Token，然后转发到配置的 Assistant Agent URL。
- 当前无论读取、解析、配置或转发是否失败，接口都会返回成功：

```json
{"code":0,"msg":"success","data":null}
```

## 19. 接口总览

| 模块 | 接口数 |
| --- | ---: |
| 服务发现 | 3 |
| 登录与注册 | 10 |
| 文件、翻译和同步 | 3 |
| 用户 | 16 |
| Telegram Bot | 4 |
| 群组基本操作 | 17 |
| 群申请与邀请 | 8 |
| 群管理 | 9 |
| 好友 | 9 |
| 消息管理 | 2 |
| 会话配置 | 2 |
| 应用 | 1 |
| 反馈 | 1 |
| 助手回调 | 1 |
| **总计** | **86** |

## 20. 兼容路径

| 接口 | 兼容路径 | Handler 行为 |
| --- | --- | --- |
| `GET /jim/servers/info` | `GET /jim/serverinfos` | 完全相同 |
| `POST /jim/sms/login` | `POST /jim/sms_login` | 完全相同 |
| `POST /jim/groups/create` | `POST /jim/groups/add` | 完全相同 |

## 21. 当前实现注意事项

1. 多数业务错误使用 HTTP `200` 返回，客户端必须检查 JSON `code`。
2. 用户 Token 没有 `Bearer ` 前缀；Bearer 模式是 API Key 模式。
3. Bearer API Key 鉴权不会设置当前用户 ID，不适合依赖当前用户身份的接口。
4. 当前多个用户和群管理接口缺少对象级权限校验，不能仅因“需要登录”就认为已经安全授权。
5. 验证码逻辑中存在特殊测试码行为；生产部署应审查并移除不安全的绕过逻辑。
6. Telegram Bot 列表当前返回完整 Token，应限制接口访问并避免日志泄漏。
7. Callback 接口当前没有签名和来源校验，并且失败时仍返回成功。
8. 分页接口混用了 `page/size`、`offset/count`、`offset/limit` 和 `start/count/order`，调用时应以各接口说明为准。
9. 仓库中还存在未注册的 `QryFriends`、`AddFriend` Handler，它们不属于当前可调用 API。
10. 静态 Web 的 `GET /`、`HEAD /` 和 SPA fallback 不计入本文 86 个业务接口。

## 22. 代码位置

| 内容 | 文件 |
| --- | --- |
| 主业务路由 | `routers/router.go` |
| Callback 路由 | `routers/callback.go` |
| 统一鉴权 | `apis/validate.go` |
| Handler | `apis/*.go` |
| 请求与响应模型 | `apis/models/*.go` |
| 统一响应 | `commons/responses/response.go` |
| 错误码 | `commons/errs/errorcode.go` |
| 业务实现 | `services/*.go` |

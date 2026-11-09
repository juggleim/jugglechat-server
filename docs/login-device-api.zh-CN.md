# 登录与设备管理 API（前端接入）

本文依据当前工作树实际实现整理（2026-11-05），包括尚未提交的实现；不是线上已部署证明。基础地址由部署方提供，以下路径包含 `/jim`。示例全部为虚构值/占位符。

## 1. 请求、响应与凭证

- 所有下列接口都必须传请求头 `appkey: <应用标识>`；POST 使用 `Content-Type: application/json`。
- 需登录接口传 `Authorization: <登录返回的 authorization 原文>`，**不要加 `Bearer `**。`Bearer ` 在本服务表示 API Key，不能用于设备管理或扫码确认；内部 `svc1` 服务凭证也不能用于这些接口。
- `authorization` 用于业务 HTTP；`im_token` 用于 IM SDK 连接，不能相互替代。前端将凭证视作不透明字符串，不自行拼装、解码后修改或推算有效期。
- `version` 请求头可保留，但它不决定是否要求设备字段。
- 通常 HTTP 200 也可能是业务失败，必须检查 JSON 的 `code`。现有 handler 使用 `BindJSON` 时，畸形 JSON 可能先设置 HTTP 400；同时处理非 2xx、网络失败和业务码。

成功 envelope（包括验证码发送、注册、扫码确认成功的 `data: null`）：

```json
{"code":0,"msg":"success","data":null}
```

原有通用错误通常**没有 `data`**，`msg` 目前通常为空：

```json
{"code":17005,"msg":""}
```

设备/新版 QR 错误由 `deviceError` 输出 HTTP 200，始终含 `data`（可为 null）：

```json
{"code":17601,"msg":"invalid device parameters","data":null}
```

不要根据 `msg` 判断状态，也不要假定每种错误都有相同 envelope 字段。

## 2. 设备字段：四类登录共用

密码、短信（两个别名）、邮箱登录，以及**待登录端的二维码 check**均把下面字段直接放在 JSON 顶层，不包在 `device` 或 `DeviceParams` 对象里。扫码确认端不替待登录端提交设备身份。

| 字段 | 类型/限制 | 必填条件及含义 |
| --- | --- | --- |
| `device_id` | string，1–128 个 Unicode 字符；无控制字符、无首尾空白 | 新设备会话必填，与 IM SDK 实际连接使用的设备 ID 完全一致 |
| `platform` | string，精确小写枚举 | 与 `device_id` 成对必填：`ios`、`android`、`web`、`windows`、`macos`、`linux` |
| `device_name` | string，最多 128 字符，无控制字符 | 选填，设备展示名 |
| `os_version` | string，最多 64 字符，无控制字符 | 选填，系统版本 |
| `app_version` | string，最多 64 字符，无控制字符 | 选填，客户端版本 |

新客户端始终传有效的 `device_id` 和 `platform`。通过对应平台 IM SDK 的 `getDeviceId` 能力取得当前 SDK 设备标识（具体调用对象、同步/异步形式以所集成 SDK 版本为准），再提交业务登录；不要另外随机生成一套业务 ID，也不要每次登录更换。此 Go 仓库不提供各端 SDK `getDeviceId` 的统一调用签名。Web/桌面容器也须使用实际 IM 连接侧的同一标识。含逗号的合法 ID 同样可管理：kick 使用 `device_ids` 数组中的完整字符串精确匹配，不按逗号拆分。

兼容规则按服务端当前应用配置决定：

- `device_auth_required_at[appkey]` 为 0/缺省，或当前时间尚未到正数截止值：仅当 `device_id`、`platform` **两个键均未提交**时允许旧登录；仍校验选填展示字段。
- 到达截止值（Unix 毫秒，含相等）后，两键均未提交返回 `17605`。
- 只传一个、空字符串、显式 `null`、错误 JSON 类型或非法枚举都不是兼容请求；身份校验返回 `17601`。展示字段类型不匹配可能先触发 handler 的 `17003`。
- 原有账号、密码、手机号、邮箱、验证码字段保持原名和原语义。可选展示字段缺失不会阻断新登录，但同设备再次登录时缺失值会覆盖为默认空字符串。
- 新登录返回设备会话 `authorization`（当前格式 `ds1...`），同一应用/用户/设备再次成功登录会更换会话，使该设备之前的业务会话失效。
- 旧登录返回旧格式 token，不创建受管理设备记录；旧端可能不会出现在设备列表中。兼容期不能承诺严格设备身份防绕过，安装标识也不是不可伪造的硬件身份。

## 3. 四类登录及验证码

除二维码 confirm 外，本节接口不要求 Authorization，仍需 appkey。

| 方式 | 方法/路径 | 保留的原始请求字段 | 新设备字段提交位置 |
| --- | --- | --- | --- |
| 密码 | `POST /jim/login` | `password` 非空；`account` / `phone` / `email` 至少一个非空 | 本请求顶层 |
| 短信 | `POST /jim/sms/login` | `phone` 非空、`code` 验证码 | 本请求顶层 |
| 短信兼容别名 | `POST /jim/sms_login` | 与 `/jim/sms/login` 完全相同，共用 handler | 本请求顶层 |
| 邮箱 | `POST /jim/email/login` | `email` 非空、`code` 验证码 | 本请求顶层 |
| 二维码 | `GET /jim/login/qrcode` → `POST /jim/login/qrcode/check` | 获取无 body；check 的 `id` 非空 | check 顶层，另带 `poll_secret`；详见第 6 节 |

密码登录同时传多个身份字段时按 `account` → `phone` → `email` 的优先级选一个查询，不会依次尝试。密码按既有协议传入 `password`；服务端现有逻辑对其做 SHA1 后比对，不因增加设备字段而更改。

短信/邮箱 `code` 是验证码校验输入；handler 虽只对 phone/email 判空，前端仍须提交实际下发的验证码。普通验证码记录按创建时间校验五分钟有效期。短信登录和换绑手机号已去除硬编码万能码放行，000000 也必须匹配真实未过期记录；这是安全行为变化，不改变请求字段或登录响应协议。邮箱校验的测试分支不在本次修复范围。验证码通过后可自动创建尚未存在的用户。

```json
{
  "account":"demo001",
  "password":"<用户输入的密码>",
  "device_id":"sdk-device-demo",
  "platform":"web",
  "device_name":"示例浏览器",
  "os_version":"示例系统版本",
  "app_version":"1.0.0"
}
```

短信将账号/密码替换为 `"phone":"<手机号>","code":"<验证码>"`；邮箱替换为 `"email":"<邮箱>","code":"<验证码>"`。设备字段规则完全一致。

### 3.1 获取验证码及注册边界

| 方法/路径 | 请求 JSON | 成功 data |
| --- | --- | --- |
| `POST /jim/sms/send` | `{"phone":"<手机号>"}` | null |
| `POST /jim/email/send` | `{"email":"<邮箱>"}` | null |
| `POST /jim/register` | `password` 非空，`account` / `phone` / `email` 至少一个；手机号注册还使用 `code` | null，不自动登录 |

注册也按 account → phone → email 选择身份；账号注册要求 6–20 位 ASCII 字母/数字。注册模型虽复用设备字段，但当前 Register handler 不签发设备会话、不处理设备登记；注册后仍须调用登录接口。

### 3.2 登录成功（四类一致）

```json
{
  "code":0,
  "msg":"success",
  "data":{
    "user_id":"demo-user",
    "authorization":"<业务凭证>",
    "nickname":"示例用户",
    "avatar":"",
    "status":0,
    "im_token":"<IM连接凭证>"
  }
}
```

`user_id`、`authorization`、`nickname`、`avatar` 为 string；`status` 为 int，当前这些登录实现没有赋其他状态，返回 0，不要解读为设备在线状态。`im_token` 为 string，模型带 `omitempty`，空值时不输出。新旧登录保留同一组响应字段。

凭证只用于本人的登录/连接，应受保护存储，不写埋点、URL 或普通日志。服务端不会通过设备列表暴露 `authorization`、`im_token`、`session_id`、应用密钥等敏感内部字段。

## 4. 设备列表

`GET /jim/devices/list`，无 body；需要有效的新用户设备会话 Authorization。

```json
{
  "code":0,
  "msg":"success",
  "data":{
    "online_devices":[{
      "device_id":"sdk-device-demo",
      "device_name":"示例浏览器",
      "platform":"web",
      "os_version":"示例系统版本",
      "app_version":"1.0.0",
      "last_login_time":1793750400000,
      "last_active_time":null,
      "is_current_device":true
    }],
    "offline_devices":[]
  }
}
```

- 仅列当前 appkey、当前用户下 `active` 且未删除的已登记设备；撤销中/已删除设备不返回。IM 在线但未登记的旧端也不会凭空加入列表。
- IM 按精确 `device_id` 判断分组，不是用户整体在线状态。两数组均为空时返回 `[]`，不是 null。
- **各组内**按 `last_login_time` 降序，相同时间按 `device_id` 字符串升序；不是按活跃时间，也不强制当前设备置顶。
- `is_current_device` 根据已验证的业务会话比较，不由请求 body 指定；当前设备若尚未连接 IM，也可在离线组。
- 时间为 Unix 毫秒。`last_active_time` 可为 null，是业务接口鉴权的近似活跃时间（写入节流约一分钟），不是 IM 消息活跃时间/下线时间；新登录会重置为 null。
- 业务鉴权在设备行锁事务内校验当前会话并按需更新活动时间，与撤销/会话轮换串行化；节流导致不更新不会拒绝合法 token。鉴权完成后的在途业务执行不承诺随后被撤销时回滚。
- 设备对象仅包含示例列出的 8 个字段。内部行 ID、appkey、用户 ID、会话 ID、撤销任务、封禁状态等不在设备对象中输出。
- IM 查询失败/响应不完整返回错误（通常 `17606`），前端应展示“暂不可获取”并重试，不能把失败渲染成“全部离线”。

## 5. 删除其他设备

`POST /jim/devices/del`；需要有效的新用户设备会话。不是 HTTP DELETE，也不是普通退出登录接口。

```json
{"device_id":"other-sdk-device-demo"}
```

`device_id` 必须是有效非空 string，规则与身份 ID 相同。只删除当前应用/用户的设备；其他用户的同名设备不受影响。UI 根据 `is_current_device` 禁用当前设备删除，服务端也拒绝该操作。

### 5.1 确认完成

```json
{"code":0,"msg":"success","data":{"operation_id":"<不透明操作ID>","status":"completed"}}
```

首次事务提交立即撤销目标业务会话并隐藏列表记录。不设置登录封禁，不调用 IM ban/unban，也不清理运营封禁。完整在线响应确认精确目标设备已离线时可完成；若仍在线，只 kick 该设备，再核查离线。此条件不是“调用 kick 返回 0”即成功，也不能证明某个特定 IM session 确已被踢。此前已离线的原设备任务允许成功。

`status` 为 `pending`、`completed` 或 `superseded`。若处理旧任务时发现设备已完成新登录、目标 session/operation 已变，任务完成为 `superseded`，不再发 kick，也不修改新会话；这不表示踢线成功。业务旧 `ds1` token 永久失效。

### 5.2 已受理但 IM 未确认（pending）

```json
{
  "code":17606,
  "msg":"IM device kick unconfirmed or unavailable: device IM kick integration disabled; business session revoked, IM disconnect pending",
  "data":{"operation_id":"<不透明操作ID>","status":"pending"}
}
```

这是开关未开启时的一种真实错误形状；其他 IM 故障的 msg 会不同。删除没有独立 pending 成功码或 `17609`。`17006` 是 QR 等待确认。`17606` 配合非空 `operation_id` 表示已提交/复用任务后 IM 未确认，可提示“业务会话已撤销，IM 下线待确认”，不能宣称踢线成功。数据库异常可能返回 `17600`；即使带非空操作 ID，也不能仅凭 ID 宣称事务已提交。

- 请求失败时即使 data 含 `operation_id`、`status` 也不代表完成。事务尚未创建任务的错误可能带 `{"operation_id":"","status":"pending"}`；中间件拒绝请求可能没有 data 或 data 为 null。
- 同一目标会话的重试复用操作 ID；没有按 operation_id 查询状态的公开路由。接口删除的是调用执行时的当前设备会话：目标重新登录后再次请求属于新删除、新操作 ID，不能当成旧请求的安全重放。观察到重新登录后应停止旧删除的自动重试。
- 响应丢失时也可按同设备重试。设备从列表消失只说明本地状态已非 active，不能作为 IM 操作完成证明。
- 服务端恢复 worker 每约 15 秒扫描到期重试任务；失败通常安排约 60 秒后再试。前端应退避重试，不高频轮询。
- 删除后立即允许同账号、同设备通过凭证验证重新登录，revoking 也不阻断；不读取历史 `banned_until` 限制登录。新登录生成新 session，旧任务在下一次处理时 superseded。历史任务即使旧期限已过仍可恢复，不再因十分钟 deadline 过期停止。
- 登录和 IM 处理锁同一 MySQL 设备行；处理按“设备行 → 任务行”加锁，在有界 IM 调用和核查期间持锁。正在 kick 时，同设备新登录短暂等待（处理 context 最多约 5 秒，数据库/网络调度可能另有延迟）；不承诺无阻塞。登录先提交则旧任务不发 kick。
- IM kick 只有 deviceid，无 session 条件。DB 锁消除本业务跨进程查后发送竞态，但不能约束上游异步/超时后延迟执行、旧 IM 凭证自行重连；离线核查只满足当前契约，不是零风险证明。客户端收到 kick 通知应停止自动重连并清理本地旧业务会话，让用户主动登录；本轮无需修改客户端代码，需提醒接入方核对该行为。
- 已有 IM 封禁不会自动解除，可能仍阻止 IM 连接；由运营按原规则处理，不自动 unban。

## 6. 二维码：获取、确认与领取

### 6.1 待登录端获取

`GET /jim/login/qrcode`，无需 Authorization：

```json
{"code":0,"msg":"success","data":{"id":"<二维码ID>","poll_secret":"<轮询秘密>","qr_code":"<PNG的Base64>"}}
```

`qr_code` 为 400×400 PNG 的裸 Base64，不含 data URL 前缀；需要时前端加 `data:image/png;base64,` 展示。二维码内实际 JSON 是 `{"action":"login","code":"<二维码ID>"}`，其中 `code` 对应后续请求的 `id`，**不包含 poll_secret**。秘密只由获取响应交给待登录端，勿放入二维码、分享链接、埋点或日志。

有效期为服务端创建起十分钟，获取响应没有 `expires_at` 字段；客户端倒计时仅作提示，以服务端错误为准。无需获取时提交设备字段。

### 6.2 已登录扫码端确认

解析二维码后由已登录用户确认，调用 `POST /jim/login/qrcode/confirm`：

```json
{"id":"<二维码ID>"}
```

需 appkey 和用户 Authorization；用户身份来自凭证，不接收 body 中指定其他用户。新设备会话可确认；兼容期旧用户 token 也可确认，严格模式旧 token 被拒绝；API Key/服务凭证不允许。扫码端不需要 poll_secret，也不提交待登录设备的设备字段。

成功为 `{"code":0,"msg":"success","data":null}`；有效期内同一确认用户重复确认可成功，不同用户抢占或无效/过期记录等 ConfirmQR 错误当前统一映射为 `17000`（通用错误 envelope），**不是 check 的 `17007`/`17608` 映射**。确认成功不是待登录端已完成领取。

### 6.3 待登录端 check（新流程）

`POST /jim/login/qrcode/check`，无需 Authorization：

```json
{
  "id":"<二维码ID>",
  "poll_secret":"<获取时返回的秘密>",
  "device_id":"sdk-device-demo",
  "platform":"web",
  "device_name":"示例浏览器",
  "os_version":"示例系统版本",
  "app_version":"1.0.0"
}
```

- 未确认返回 `17006`，data 为 null，继续间隔轮询（建议前端退避/限频；协议没有指定固定轮询频率）。
- 确认后首次领取原子绑定设备和会话，返回第 3.2 节的完整登录数据。首次成功后停止轮询并连接 IM。
- 网络失败/响应丢失，在二维码创建起十分钟内，以相同 id、poll_secret、device_id 和合法设备参数重试，可返回缓存的同一登录结果，不重复创建会话；不是“领取后任何重试都失败”。
- 已领取后更换 device_id 或秘密、切换领取模式不可领取；对应设备会话已撤销或被新登录替换，缓存也不能复活旧会话（返回 `17604`）。
- `17007` 表示已过期，停止轮询并重新获取；`17608` 表示记录不存在/秘密不匹配/模式冲突/其他无效领取，应重新获取。`17601` 修正设备参数；`17605` 升级使用设备流程。已消费的旧二维码结果不会创建新会话，重新登录须使用有效的新领取流程。
- 上游 IM 注册等暂时错误可能发生在领取前，可在有效期内保留同一请求参数退避重试；不要因一次网络失败丢弃秘密或随机换 ID。

### 6.4 旧流程差异与兼容限制

兼容期允许旧待登录端仅发 `{"id":"<二维码ID>"}`，**同时省略两个设备身份键且不带 poll_secret**，领取旧 token，不登记设备。新待登录端必须同时带秘密和设备身份；仅带秘密但不带身份不是新流程。

第一次成功领取原子决定 legacy/device 模式；之前未确认的 check 不锁定模式。device 模式已成功领取后无秘密旧请求不能再领取；legacy 模式已领取后新设备请求也不能领取。兼容期在首次领取前，知道二维码 ID 的旧请求仍可能走旧领取路径，不能声称 poll_secret 已为整个兼容系统提供严格防抢领/防绕过。严格截止后旧 check 缺设备字段返回 `17605`，旧用户 token 也不能再 confirm。

## 7. 错误码与前端动作

| code | 实际含义/适用范围 | 前端处理 |
| --- | --- | --- |
| `0` | 成功 | 按 data 处理 |
| `17000` | 通用失败；包括当前 QR confirm 的存储/状态错误 | 提示确认失败，必要时刷新二维码 |
| `17001` / `17002` | appkey 缺失 / 应用不存在（或登录 IM SDK 不可用） | 检查应用配置 |
| `17003` | 原有请求 JSON/账号凭证字段不合法 | 修正请求 |
| `17004` | 登录调用 IM 注册出错 | 退避重试 |
| `17005` | 未登录/旧 token 无效或被策略拒绝 | 重新以设备参数登录 |
| `17006` | QR check 尚待确认 | 继续轮询 |
| `17007` | QR check 已过期 | 重新获取 |
| `17008` | 短信/邮箱发送失败 | 提示重试 |
| `17009` | 短信/邮箱验证码失效或校验未通过 | 重新获取/输入验证码 |
| `17011` / `17012` / `17013` | 注册用户已存在 / 用户不存在 / 密码错误 | 对应提示 |
| `17600` | 未分类设备/存储操作错误 | 保留可能的 operation_id，稍后重试 |
| `17601` | 设备参数错误 | 修正成对身份、类型、长度、枚举等 |
| `17602` | 本人范围内找不到目标设备 | 刷新列表；不能据此推断其他用户的设备 |
| `17603` | 禁止删除当前设备 | 禁用当前设备删除 |
| `17604` | 设备会话无效/已撤销，或凭证类型不允许 | 用户设备会话重新登录；检查凭证用法 |
| `17605` | 需升级并携带设备参数重新登录 | 重新走新登录 |
| `17606` | IM 适配器故障、禁用、查询/撤销未确认或超时等 | 列表显示不可用；删除按 pending 分支处理 |
| `17607` | 历史设备封禁码，当前设备登录逻辑不再产生 | 不作为当前等待重新登录的依据 |
| `17608` | QR check 无效/秘密或领取模式冲突等 | 停止当前领取，重新获取二维码 |

当前 API 不再输出 `banned_until` 或 `retry_after_seconds`；历史数据列仅保留兼容，不参与登录和恢复判断。

鉴权优先于 handler：无 Authorization 一般 `17005`；旧 token 请求 `/devices/*` 在兼容期也被中间件以 `17005` 拒绝；API Key/内部服务凭证访问敏感路径一般 `17604`。`17605` 是缺少新设备身份/设备会话时的升级错误，不能把所有旧 token 请求都写成一定返回它。登录也先执行原有账号/验证码和 IM 注册逻辑，参数不合法时不保证优先于这些错误返回。IM SDK 非零返回码可能原样透传，不限于上表；前端保留未知码兜底。

## 8. 部署配置与 IM 边界（联调前确认）

以下是**代码配置语义/默认值**，不是当前线上配置；本文未读取真实配置、密钥或用户数据，也未验证部署状态：

| 服务端配置 | 代码行为 |
| --- | --- |
| `device_auth_required_at` | 按 appkey 配置 Unix 毫秒截止；缺省/0 为兼容，达到正数值后严格要求设备登录；不根据客户端 version 判断 |
| `device_im_kick_enabled` | bool，缺省 false；控制踢出和离线核查（含 worker）。关闭仍先撤销业务会话并持久化任务，随后返回 `17606`，不静默成功；在线列表不受影响 |
| `device_im_ban_enabled` | 弃用兼容别名；仅新键未配置时生效。旧线上 true 等价于 kick enabled，不会调用 ban；新键显式 false 优先于旧 true。建议部署时迁移为新键 |
| `assistant_service_credentials` | bool，缺省 false；内部助手凭证迁移开关，不是前端登录参数；严格截止需协调助手依赖适配 |
| `imApiDomain` | 服务端 IM 管理 API origin，需部署方正确接入；前端不直连管理 API、不持有管理密钥 |

已实现仅业务会话撤销与精确设备 kick/离线核查，不撤销 IM 凭证、不调用 ban/unban。上线需统一切换旧实例和 worker，避免旧二进制继续 ban 或拒绝登录；还需真实 IM 验收异步执行、查询完整性、设备隔离和 kick 通知。详情见 [IM 接入与剩余验收项](deviceauth-im-integration.zh-CN.md)，不要把测试当作已部署保证。

## 9. 维护与源码对照

- 路由：`routers/router.go`；公共鉴权/凭证允许范围：`apis/validate.go`、`services/deviceauth/policy.go`、`service.go`。
- 登录原字段/响应：`apis/models/login.go`、`apis/login.go`；设备参数/敏感字段屏蔽：`services/deviceauth/auth.go`。
- 设备列表/分组/排序：`apis/devices.go`、`services/deviceauth/im.go`、`store.go`。
- 删除任务/重试/期限：`store.go`、`http_im.go`、`main.go`；错误映射：`apis/devices.go`。
- QR 秘密/确认/领取/缓存：`apis/login.go`、`services/deviceauth/qr.go`。
- envelope/通用错误：`commons/responses/response.go`、`commons/errs/errorcode.go`；配置定义/严格判断：`commons/configures/configure.go`、`services/deviceauth/config.go`。
- 仓库原有资料为 Markdown：`docs/superpowers/specs/2026-11-02-login-device-management-design.md`（历史设计）、`docs/superpowers/plans/`（实施/验证计划）、`docs/deviceauth-im-integration.zh-CN.md`（IM 专项说明）。它们不替代本文的前端协议说明。

# 登录设备管理设计

## 目标与范围

为一个用户管理多台登录设备，满足以下能力：

- 按最后登录时间倒序列出登录设备。
- 将设备按 IM 在线和离线状态分组返回。
- 删除非当前设备时，立即使其业务登录态失效、强制其 IM 连接下线，并封禁该设备 10 分钟。
- 沿用既有登录 URL，仅扩展登录请求参数。

不在本次范围内：删除当前设备、批量删除设备、设备重命名、历史登录地点/IP 展示、可配置封禁时长。

## 现状与问题

现有业务 `Authorization` token 已有 `device_id` 编码字段，但 `GenerateToken` 未赋值；服务端也没有设备或会话的持久化状态。IM 在线查询当前只支持按用户查询，无法从业务层可靠地完成设备维度排序、吊销或 10 分钟封禁。因此设备状态必须由应用服务持久化，并将 `device_id` 绑定进业务 token。

## 客户端登录请求

以下真正签发登录态的既有接口增加统一的设备字段，保留原有字段和 URL：

- `POST /jim/login`
- `POST /jim/sms/login`
- `POST /jim/email/login`
- `POST /jim/login/qrcode/check`

请求字段：

```json
{
  "device_id": "稳定的客户端安装标识",
  "device_name": "Fei 的 iPhone",
  "platform": "ios",
  "os_version": "iOS 18.1",
  "app_version": "1.0.0"
}
```

`device_id` 必填，且必须由客户端在同一次安装生命周期内稳定保存。其余字段也必填；服务端限制长度、清理首尾空白，避免不可信客户端写入无界数据。二维码场景采用已确认的方案 A：待登录的新设备在每次 `POST /jim/login/qrcode/check` 请求中提交这组字段；二维码被确认且状态转为成功时，服务端使用该次请求设备信息创建登录态。`POST /jim/login/qrcode/confirm` 不创建新设备记录。

## 数据模型

新增 `user_devices` 表，每行表示 `(app_key, user_id, device_id)` 的一个受管理设备：

| 字段 | 说明 |
| --- | --- |
| `id` | 自增主键 |
| `app_key` | 应用隔离键 |
| `user_id` | 所属用户 |
| `device_id` | 客户端设备/安装标识 |
| `device_name` | 设备展示名 |
| `platform` | 平台，例如 `ios` |
| `os_version` | 系统版本 |
| `app_version` | 客户端版本 |
| `last_login_time` | 最近成功登录的毫秒时间戳 |
| `last_active_time` | 最近业务请求通过鉴权的毫秒时间戳；若未实现触摸更新，则等于最后登录时间 |
| `banned_until` | 删除后封禁截止的毫秒时间戳，0 表示未封禁 |
| `deleted_at` | 删除时间；保留记录以支持封禁和审计 |
| `created_time` / `updated_time` | 数据库时间 |

唯一索引为 `(app_key, user_id, device_id)`；列表索引为 `(app_key, user_id, deleted_at, last_login_time DESC)`。

删除设备不是物理删除：将 `deleted_at` 设置为当前时间、`banned_until` 设为当前时间加 10 分钟。后续成功登录同一 `device_id` 时，若仍在封禁期内必须拒绝；过期后可复用该记录，清除删除标记并刷新设备资料和登录时间。

## 认证与登录流程

1. 所有四个登录成功路径均调用统一的 `IssueDeviceLogin` 服务：账号密码、短信、邮箱，以及二维码状态检查成功。
2. 服务查询 `(app_key, user_id, device_id)`；若 `banned_until > now`，返回新增的“设备暂时封禁”业务错误及可选的剩余秒数。
3. 服务 upsert 活跃设备记录，更新设备元数据、`last_login_time`、`last_active_time`，清除已过期的删除/封禁状态。
4. 使用 `GenerateToken(appKey, userId, deviceId)` 签发 `Authorization`，把 `device_id` 编码入现有 token payload。
5. 每一个需要登录态的业务请求，鉴权中在解析 token 后验证相应设备记录：记录必须属于 token 用户、未删除、未封禁。验证失败时中止请求并返回未登录/设备已删除或设备封禁错误。可按节流策略更新 `last_active_time`，避免每个请求均写库。

兼容性：设备参数发布后采用“必填”策略。旧客户端登录会得到请求体非法错误；已签发的旧 token 缺少 `device_id`，在启用设备校验后不再作为有效登录态。发布需与客户端升级和用户重新登录窗口协调。

## 新增业务接口

### `GET /jim/devices/list`

仅查询当前 `Authorization` 所属用户的未删除设备。不接收其他 `user_id`，禁止跨用户读取。

服务端逐台向 IM 的**设备维度**在线状态能力查询，按结果分组为 `online_devices` 和 `offline_devices`。每组均按 `last_login_time` 倒序排列；`is_current_device` 由 token 的 `device_id` 与记录比较得出。

响应 `data`：

```json
{
  "online_devices": [
    {
      "device_id": "...",
      "device_name": "Fei 的 iPhone",
      "platform": "ios",
      "os_version": "iOS 18.1",
      "app_version": "1.0.0",
      "last_login_time": 1760000000000,
      "last_active_time": 1760000100000,
      "is_current_device": true
    }
  ],
  "offline_devices": []
}
```

IM 在线状态查询失败时整个接口返回可重试的 IM/内部超时错误，不返回可能错误的在线分组。

### `POST /jim/devices/del`

请求：

```json
{ "device_id": "要删除的设备 ID" }
```

仅可操作当前用户的设备。目标不存在、已删除，或与当前 token 的 `device_id` 相同，均返回明确的业务错误；当前设备不能通过该接口删除，普通登出仍走客户端本地退出流程。

成功删除的顺序：

1. 对目标设备设置 `deleted_at=now`、`banned_until=now+10min`，使业务 token 即刻在后续请求中失效。
2. 调用 IM Server 的按 `user_id + device_id` 强制下线 API，使已建立的 IM 连接立即断开。
3. 仅在 IM API 成功确认下线后返回成功；若调用失败，返回可重试错误，设备记录仍维持删除与封禁状态，防止业务 token 继续使用。

该操作应是幂等的：同一删除请求重复到达时，只要目标仍在封禁/已删除状态，返回成功或专用“已删除”错误，客户端按成功处理，不得再次延长 10 分钟封禁期。

## IM SDK 集成前置条件

当前仓库使用 `github.com/juggleim/imserver-sdk-go v1.0.16`，源码未调用设备级状态或踢下线 API。实施前必须在 SDK 文档或其可访问源码中确认下列能力与准确方法签名：

1. 基于 `user_id + device_id` 查询在线状态（不能只返回用户整体在线）。
2. 基于 `user_id + device_id` 强制断开该设备的 IM 长连接。
3. 客户端连接 IM 时能携带与业务 `device_id` 一致的值，使 IM 识别为同一设备。

若 SDK 不具备任一能力，本设计不能声称实现“立即 IM 下线”或准确在线/离线分组。届时需要升级 SDK/IM Server，或由 IM Server 增加对应管理 API；不得降级为仅吊销业务 token。

## 错误与安全规则

- 新增错误码：设备参数非法、设备不存在、禁止删除当前设备、设备已删除、设备封禁中、IM 设备操作失败。
- 服务端从 token 推导用户和当前设备，绝不相信删除接口中的用户 ID。
- `device_id` 不作为认证凭据；认证仍以签名业务 token 和持久化设备状态为准。
- 删除设备不删除用户账号，也不影响同一用户其他设备。
- 不记录或输出 `Authorization`、IM token、手机号等敏感凭据。

## 测试范围

- 四个登录接口：缺失设备参数拒绝；正常登录创建记录；同设备再次登录更新时间；封禁期内拒绝；封禁过期后允许重新登录。
- 鉴权：有效设备 token 通过；已删除、封禁、缺失设备 ID、跨用户设备记录的 token 均拒绝。
- 列表：仅返回当前用户设备；在线与离线分组正确；两组均按最后登录时间倒序；正确标记当前设备；IM 状态查询失败时返回错误。
- 删除：删除其他设备会写入 10 分钟封禁并调用 IM 下线；删除当前设备、其他用户设备、未知设备均拒绝；重复请求不延长封禁。
- IM 集成测试：确认目标设备连接被即时断开，并且重新连接在封禁期间被拒绝。

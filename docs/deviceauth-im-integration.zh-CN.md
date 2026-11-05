# 设备管理 IM 接入说明（2026-11-05 已批准变更）

## 当前契约

删除只永久撤销当前业务 `ds1` 会话并尝试精确设备 IM 强制下线。绝不调用 IM ban/unban（包括历史带 ban 期限任务），不查询或清理运营封禁。不检查/设置新 `banned_until` 登录限制；历史列保留兼容，新任务期限列为零。删除后包括 revoking 期间均可立即凭有效账号凭证登录，生成全新 session，旧 token 永不复活。

## 串行化及任务恢复

业务 Authenticate 在设备行锁事务内校验当前 session、按需更新活动时间，与删除/轮换串行化。节流无需更新仍成功，不用 RowsAffected=0 判断鉴权失败。锁仅覆盖鉴权事务，鉴权完成后的在途业务执行不承诺随撤销回滚。

1. Delete 事务锁设备行、撤销当前会话并保存 operation ID 与 target session。
2. Process 从持久化任务读取目标，按设备行 → 任务行 `SELECT FOR UPDATE`，与 Issue/QR 的设备行锁共享同一 MySQL 主库一致性。
3. 新 session/operation 已替换目标，或目标行消失/不再 revoking：done=true、last_error=superseded，不调用 IM，不修改新会话。
4. 目标仍匹配时持锁完成有界 IM 调用和离线核查，期间同设备登录可能短暂等待。Process 不自动重试包含远端副作用的事务；失败/未知结果保留任务，下一轮先核对。正常 IM 错误持久化 attempts/next_attempt；context 超时可能回滚该轮诊断，任务仍保留供恢复。
5. 不因旧十分钟 deadline 过期停止恢复。worker 约每 15 秒扫描最多 20 个任务，正常失败约 60 秒后重试。旧 done 任务仍幂等，不改历史审计。

所有写实例必须使用同一支持 InnoDB 行锁的主库，不可把此路径放到延迟只读副本。无需新表/列迁移，20261103 中历史 banned_until 列保留；不自动改写既有值或 IM 封禁。发布须先停旧版实例/worker 再统一切换，混部旧版本不满足新契约。

## IM 调用与完成条件

复用此前审查的接口契约，本轮未重新断言 IM dev SHA 最新：

- POST `/apigateway/users/onlinestatus/query`，完整响应必须有唯一目标用户、is_online、非 null devices，拒绝矛盾状态/重复或非法设备。
- Reconcile 在锁内核查目标设备；原设备已离线允许完成，不要求此前必定发过 kick。这只说明当前离线，不能证明特定 session 被踢。
- 仍在线则 POST `/apigateway/users/kick`，仅 `user_id` 与单个 `device_ids`，不传 platforms/ext；拒绝逗号 ID，避免旧 IM 拆分扩大目标。
- 仍在线则 POST `/apigateway/users/kick`，仅 `user_id` 与单个 `device_ids`，不传 platforms/ext。合法含逗号 ID 保留为数组中的一个完整字符串精确匹配，不使用旧 ban scope_value 或逗号拆分。
- kick 后重新查询目标离线；仅异步 code=0 不完成。错码、不完整状态、仍在线、超时均不成功。无 ban/unban/query-ban 请求。

**能力边界：** IM 只有 deviceid，没有 target session 条件。数据库行锁防止业务跨进程“查旧会话后、新登录先提交、再发送旧 kick”的竞态；不能撤回 IM 已接收但延迟执行的请求，也不能阻止旧 IM token 自行重连。超时释放锁后或远端异步执行时仍可能误伤新连接；无 session fencing 不能保证零风险。绝对隔离需要上游条件 kick/可靠执行屏障。不能靠永久阻塞业务登录替代这一能力。

客户端应在 kick 通知后停止自动重连、丢弃旧业务会话，用户主动登录可立即新建会话。本轮不改客户端代码，需要接入方核对通知处理。既有 IM 运营封禁不会自动解除，若阻止连接须按原运营规则处理，不能批量 unban。

## 配置迁移

```yaml
# 默认 false；显式开启 kick/离线核查，在线列表不受影响。
device_im_kick_enabled: true
```

`device_im_ban_enabled` 为弃用别名，仅新键未配置时作为 fallback：旧 true → 仅开启 kick，绝不 ban；新 false 优先旧 true。两键均缺省为 false。关闭时先持久化业务撤销，再返回 17606/IM pending，绝不静默成功。实际配置文件未读取或修改；部署负责人迁移键名。

沿用 imApiDomain 和当前应用 app_secret，由运行时配置解析；appkey/nonce/timestamp/SHA1 签名。每轮处理 context 最多 5 秒（含锁等待、HTTP 与核查），无重定向/HTTP 自动重试，响应限 1 MiB，诊断不保存凭据或上游正文。device_auth_required_at/assistant_service_credentials 语义不变。

## 验收

以 [本轮验证记录](superpowers/plans/2026-11-05-device-kick-only-validation.md) 为实际测试证据。真实 IM 尚需验收设备 ID 一致性、精准断连、其他设备隔离、异步/超时延迟 kick、查询失败传播/缓存一致性与客户端停止重连。离线状态回读不替代真实连接观察。未读取真实配置凭据、未调用真实管理 API、未部署、未提交。

历史：此前版本设计有固定十分钟 ban。该约定已被本变更取代；保留旧实施计划用于历史追溯，不据此执行 ban/unban 或清理历史数据。

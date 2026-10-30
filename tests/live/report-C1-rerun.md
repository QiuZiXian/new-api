# Seedance 计费探测报告

- 环境：http://www.qsykai.com:33000
- 时间：2026-10-29 15:21:45
- 换算式：`¥ = quota / 500000`
- 说明：异步任务消费日志为「预扣 + 差额补扣」，下表 quota 已按 task_id 求和。

| 用例 | 分组 | 模型 | 档位 | 视频实际时长 | 公式 tokens | 平台 tokens | token 偏差 | 档位价 | 实扣 quota | 金额 | 金额偏差 | 结论 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| C1 | 默认 | doubao-seedance-2.0 | 720p 16:9 | -s | 108000 | - | - | - | - | - | - | FAIL |

## 明细

### C1 @ 默认组
```json
{
  "case": "C1",
  "account": "默认组",
  "group": "默认",
  "model": "doubao-seedance-2.0",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "has_video": false,
  "status": "FAIL",
  "expected_tokens": 108000,
  "submit_http": 404,
  "ratios_header": null,
  "reason": "提交失败: {\"code\": \"fail_to_fetch_task\", \"message\": \"\", \"status_code\": 404, \"error\": {\"code\": \"fail_to_fetch_task\", \"message\": \"\", \"type\": \"new_api_error\"}}"
}
```

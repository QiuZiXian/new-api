# Seedance 计费探测报告

- 环境：http://www.qsykai.com:33000
- 时间：2026-10-29 13:00:10
- 换算式：`¥ = quota / 500000`
- 说明：异步任务消费日志为「预扣 + 差额补扣」，下表 quota 已按 task_id 求和。

| 用例 | 分组 | 模型 | 档位 | 视频实际时长 | 公式 tokens | 平台 tokens | token 偏差 | 档位价 | 实扣 quota | 金额 | 金额偏差 | 结论 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| C1 | 默认 | doubao-seedance-2.0 | 720p 16:9 | 5.088s | 108000 | 217392 | +101.29% | 46.0 | 5000016 | ¥10.0000 | +101.29% | FAIL |

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
  "task_id": "task_Z7Yx4GAqhkpX9xGqEcCj6tZmJkMR7dF9",
  "submit_http": 200,
  "state": "completed",
  "content_ok": true,
  "video_seconds": 5.088,
  "log_lines": 2,
  "total_quota": 5000016,
  "pre_consumed_quota": 2484000,
  "actual_quota_field": 5000016,
  "log_model_ratio": 23,
  "log_group_ratio": 1,
  "log_content": [
    "操作 generate, 预估 tokens：108000",
    "token重算：tokens=217392, modelRatio=23.00, groupRatio=1.00, otherMultiplier=1.0000"
  ],
  "tokens_source": "log",
  "actual_tokens": 217392,
  "log_prompt_tokens": 0,
  "log_completion_tokens": 217392,
  "tier_price": 46.0,
  "group_ratio": 1.0,
  "expected_quota": 5000016.0,
  "quota_dev": 0.0,
  "token_dev": 1.0129,
  "amount_cny": 10.0,
  "amount_expected": 4.968,
  "amount_dev": 1.0129,
  "platform_recompute": 5000016.0,
  "platform_dev": 0.0,
  "token_ok": false,
  "amount_ok": false
}
```

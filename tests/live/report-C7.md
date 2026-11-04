# Seedance 计费探测报告

- 环境：http://www.qsykai.com:33000
- 时间：2026-10-30 11:30:32
- 换算式：`¥ = quota / 500000`
- 说明：异步任务消费日志为「预扣 + 差额补扣」，下表 quota 已按 task_id 求和。

| 用例 | 分组 | 模型 | 档位 | 请求时长 | 成片时长 | 时长偏差 | 公式 tokens | 平台 tokens | token 偏差 | 档位价 | 净扣 quota | 金额 | 金额偏差 | 结论 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| C7 | 默认 | doubao-seedance-2.0 | 480p 16:9 | 5s | 5.062s | 0.062s | 48576 | 50638 | +4.24% | 46.0 | 1164674 | ¥2.3293 | +4.24% | FAIL |

## 明细

### C7 @ 默认组
```json
{
  "case": "C7",
  "account": "默认组",
  "group": "默认",
  "model": "doubao-seedance-2.0",
  "resolution": "480p",
  "ratio": "16:9",
  "duration": 5,
  "has_video": false,
  "status": "FAIL",
  "expected_tokens": 47981,
  "task_id": "task_FDIKqkSJpPkJj6CtwzYkO5DPCTLsEqtX",
  "submit_http": 200,
  "state": "completed",
  "content_ok": true,
  "video_seconds": 5.062,
  "log_lines": 2,
  "consume_quota": 2207126,
  "refund_quota": 1042452,
  "net_quota": 1164674,
  "total_quota": 1164674,
  "pre_consumed_quota": 2207126,
  "actual_quota_field": 1164674,
  "log_model_ratio": 23,
  "log_group_ratio": 1,
  "log_content": [
    "操作 generate, 预估 tokens：95962",
    "token重算：tokens=50638, modelRatio=23.00, groupRatio=1.00, otherMultiplier=1.0000"
  ],
  "log_types": [
    2,
    6
  ],
  "tokens_source": "log",
  "actual_tokens": 50638,
  "log_prompt_tokens": 0,
  "log_completion_tokens": 50638,
  "tier_price": 46.0,
  "group_ratio": 1.0,
  "formula_tokens": 48576,
  "formula_basis_seconds": 5.062,
  "duration_dev": 0.062,
  "duration_ok": true,
  "expected_quota": 1164674.0,
  "token_dev": 0.0424,
  "amount_cny": 2.3293,
  "amount_expected": 2.2345,
  "amount_dev": 0.0424,
  "platform_recompute": 1164674.0,
  "platform_dev": 0.0,
  "token_ok": true,
  "amount_ok": false
}
```

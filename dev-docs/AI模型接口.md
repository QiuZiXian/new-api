---
title: AI模型接口
language_tabs:
  - shell: Shell
  - http: HTTP
  - javascript: JavaScript
  - ruby: Ruby
  - python: Python
  - php: PHP
  - java: Java
  - go: Go
toc_footers: []
includes: []
search: true
code_clipboard: true
highlight_theme: darkula
headingLevel: 2
generator: "@tarslib/widdershins v4.0.30"

---

# AI模型接口

Base URLs:

# Authentication

- HTTP Authentication, scheme: bearer<br/>使用 Bearer Token 认证。
格式: `Authorization: Bearer sk-xxxxxx`

- HTTP Authentication, scheme: bearer

# 自建/素材

## DELETE 删除素材

DELETE /v1/assets/{asset_id}

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|asset_id|path|string| 是 |none|

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## POST 创建素材

POST /v1/asset-groups/group_58eyg2x6tj8dzgQjwIW6kMXUm01XDAI2/assets

> Body 请求参数

```json
{
  "AssetGroupId": "group-20260907145308-xntbl",
  "AssetType": "Image",
  "Name": "4",
  "image_url": "http://116.205.230.111:9366/api/667865bd-0218-46a8-aa2a-b35345e3e33d.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=U47IZ8SLXAM5FLG5AHUO%2F20260909%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20260909T102723Z&X-Amz-Expires=3600&X-Amz-Security-Token=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJleHAiOjE3ODg5ODU0MDMsInBhcmVudCI6InJ1c3Rmc2FkbWluIn0.7m69fgaUP5x_gRbBPlEDRR6EU4prmdWX2j0E97Fr4Ooouf-_LCPCLKtdk510XLxu0SH3JuCwSgXIcS4H0xu5Rw&X-Amz-Signature=49d29db5e7e96d72445c48243bee7adf971782f34babe4042cf7167ebed45123&X-Amz-SignedHeaders=host&x-amz-checksum-mode=ENABLED&x-id=GetObject"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|

> 返回示例

> 200 Response

```json
{"ResponseMetadata":{"RequestId":"202609071530121D2BCF0178D22A63729B","Action":"CreateAsset","Version":"2024-01-01","Service":"ark","Region":"cn-beijing"},"Result":{"Id":"asset-20260907153012-rjh2c"}}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## GET 素材列表 

GET /v1/asset-groups/group_58eyg2x6tj8dzgQjwIW6kMXUm01XDAI2/assets

> Body 请求参数

```json
{
  "AssetGroupId": "string",
  "Filter": {
    "Status": [
      "string"
    ]
  },
  "PageNum": 0,
  "PageSize": 0
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» AssetGroupId|body|string| 是 |none|
|» Filter|body|object| 是 |none|
|»» Status|body|[string]| 是 |none|
|» PageNum|body|integer| 是 |none|
|» PageSize|body|integer| 是 |none|

> 返回示例

> 200 Response

```json
{"ResponseMetadata":{"RequestId":""},"Result":{"Items":[{"Id":"group-20260907145308-xntbl","Name":"我的素材组","Description":"用于存放视频生成的参考图片素材","GroupType":"AIGC","CreatedAt":"2026-09-07T14:53:09Z","UpdatedAt":"2026-09-07T14:53:09Z"}],"TotalCount":1}}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## GET 素材详情

GET /v1/assets/{assetId}

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|assetId|path|string| 是 |none|

> 返回示例

> 200 Response

```json
{"ResponseMetadata":{"RequestId":"20260907153929B7418632BF376C4D413D","Action":"GetAsset","Version":"2024-01-01","Service":"ark","Region":"cn-beijing"},"Result":{"Id":"asset-20260907153012-rjh2c","Name":"焦恩俊","URL":"https://ark-media-asset.tos-cn-beijing.volces.com/2101594641/090715301275077076.png?X-Tos-Algorithm=TOS4-HMAC-SHA256&X-Tos-Credential=<REDACTED>&X-Tos-Date=20260907T073929Z&X-Tos-Expires=41400&X-Tos-Security-Token=<REDACTED>&X-Tos-Signature=<REDACTED>&X-Tos-SignedHeaders=host","AssetType":"Image","GroupId":"group-20260907145308-xntbl","Status":"Active","Moderation":{"Strategy":"Default"},"CreateTime":"2026-09-07T07:30:12Z","UpdateTime":"2026-09-07T07:30:13Z","ProjectName":"xinmenhudoubaohuiju"}}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

# 自建/素材组管理

## PUT 修改素材组 

PUT /v1/asset-groups/group_q2PdSXTJRuPR4n5IyqejFxirLu9DkWZL

> Body 请求参数

```json
{
  "Name": "我的素材组22",
  "Description": "用于存放视频生成的参考图片素材"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|

> 返回示例

> 200 Response

```json
{"ResponseMetadata":{"RequestId":"2026090714530873132E48ABCAC199153D","Action":"CreateAssetGroup","Version":"2024-01-01","Service":"ark","Region":"cn-beijing"},"Result":{"Id":"group-20260907145308-xntbl"}}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## DELETE 删除素材组  

DELETE /v1/asset-groups/group_q2PdSXTJRuPR4n5IyqejFxirLu9DkWZL

> Body 请求参数

```json
{
  "Name": "我的素材组22",
  "Description": "用于存放视频生成的参考图片素材"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|

> 返回示例

> 200 Response

```json
{"ResponseMetadata":{"RequestId":"2026090714530873132E48ABCAC199153D","Action":"CreateAssetGroup","Version":"2024-01-01","Service":"ark","Region":"cn-beijing"},"Result":{"Id":"group-20260907145308-xntbl"}}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## POST 创建素材组

POST /v1/asset-groups

> Body 请求参数

```json
{
  "Name": "我的素材组3-del",
  "Description": "用于存放视频生成的参考图片素材"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|

> 返回示例

> 200 Response

```json
{"ResponseMetadata":{"RequestId":"2026090714530873132E48ABCAC199153D","Action":"CreateAssetGroup","Version":"2024-01-01","Service":"ark","Region":"cn-beijing"},"Result":{"Id":"group-20260907145308-xntbl"}}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## GET 素材组列表

GET /v1/asset-groups

> 返回示例

> 200 Response

```json
{"ResponseMetadata":{"RequestId":""},"Result":{"Items":[{"Id":"group-20260907145308-xntbl","Name":"我的素材组","Description":"用于存放视频生成的参考图片素材","GroupType":"AIGC","CreatedAt":"2026-09-07T14:53:09Z","UpdatedAt":"2026-09-07T14:53:09Z"}],"TotalCount":1}}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

# 自建/智能擦除

## POST 创建字幕擦除任务

POST /v1/videos/subtitle-erase/tasks

> Body 请求参数

```json
{
     "model": "cii-subtitle-erase",
     "prompt": "remove subtitles",
     "metadata": {"video_url": "https://ark-content-generation-cn-beijing.tos-cn-beijing.volces.com/doubao-seedance-1-5-pro/02178946630732800000000000000000000ffffac1413ec6a110c.mp4?X-Tos-Algorithm=TOS4-HMAC-SHA256&X-Tos-Credential=<REDACTED>&X-Tos-Date=20260915T095924Z&X-Tos-Expires=86400&X-Tos-Signature=<REDACTED>&X-Tos-SignedHeaders=host"}
   }
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## GET 查询字幕擦除任务 

GET /v1/videos/subtitle-erase/tasks/{task_id}

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|task_id|path|string| 是 |none|

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

# 自建/Seedance

## POST 创建视频任务

POST /v1/videos/generations

> Body 请求参数

```json
{
    "model": "doubao-seedance-1.5-pro",
    "prompt": "小猫对着镜头打哈欠；视频字幕：猫咪觉得很无聊，一直在打哈欠",
    "content": [
        {
            "type": "text",
            "text": "小猫对着镜头打哈欠；视频字幕：猫咪觉得很无聊，一直在打哈欠"
        }
    ],
    "resolution": "720p",
    "ratio": "16:9",
    "duration": 5,
    "seed": 11,
    "camera_fixed": false,
    "watermark": true
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## GET 查询视频任务

GET /v1/videos/{task_id}

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|task_id|path|string| 是 |none|

> 返回示例

> 401 Response

```json
{
  "error": {
    "code": "string",
    "message": "string",
    "type": "string"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|

### 返回数据结构

状态码 **401**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» error|object|true|none||none|
|»» code|string|true|none||none|
|»» message|string|true|none||none|
|»» type|string|true|none||none|

# 自建/聊天

## POST openai风格

POST /v1/chat/completions

> Body 请求参数

```json
{
        "model": "claude-opus-4-6",
        "messages": [
          {"role": "user", "content": "Hello!"}
        ],
        "stream": false
      }
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» model|body|string| 是 |none|
|» messages|body|[object]| 是 |none|
|»» role|body|string| 是 |none|
|»» content|body|string| 是 |none|
|» stream|body|boolean| 是 |none|

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## POST claude风格

POST /v1/messages

> Body 请求参数

```json
{
        "model": "claude-opus-4-6",
        "messages": [
          {"role": "user", "content": "Hello!"}
        ],
        "stream": false
      }
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» model|body|string| 是 |none|
|» messages|body|[object]| 是 |none|
|»» role|body|string| 否 |none|
|»» content|body|string| 否 |none|
|» stream|body|boolean| 是 |none|

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

# 数据模型


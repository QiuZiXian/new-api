# cii-group 仓山区人工智能公共服务平台 — API 文档

> 本文档从 https://www.cii-group.com/docs.html 解析。
> 所有端点以 `https://www.cii-group.com/app-api` 为网关域名。
> 鉴权方式：所有接口均通过 `Authorization: Bearer <API Key>` 头鉴权。

## 目录

- [Video 生成任务 (Seedance)](#Video 生成任务 (Seedance))
- [视频去字幕任务 (Subtitle Erase)](#视频去字幕任务 (Subtitle Erase))
- [图片生成 (Image Generations)](#图片生成 (Image Generations))
- [Gemini 转发 (3rd party)](#Gemini 转发 (3rd party))
- [素材组 (Asset Groups)](#素材组 (Asset Groups))
- [素材 (Assets)](#素材 (Assets))
- [真人审核 (Visual Validate)](#真人审核 (Visual Validate))
- [错误码 (Error Codes)](#错误码 (Error Codes))

---

## Video 生成任务 (Seedance)

### POST https://www.cii-group.com/app-api/api/v1/contents/generations/tasks

```http
POST https://www.cii-group.com/app-api/api/v1/contents/generations/tasks
```

本文介绍创建视频生成任务 API 的输入输出参数，供您使用接口时查阅字段含义。模型会依据传入的图片及文本信息生成视频，待生成完成后，您可以按条件查询任务并获取生成的视频。

**模型能力**

* **doubao-seedance-2.5**（有声视频 / 无声视频）
   * **全模态参考生视频**：输入参考图片（0-30 张）+ 参考视频（0-10 个）+ 参考音频（0-10 个）+ 文本提示词（可选）生成 1 个目标视频。支持仅传入音频。支持生成全新视频、编辑视频、延长视频，支持 30 秒视频连贯直出。
   * **图生视频-首尾帧**：输入首帧图片 + 尾帧图片 + 文本提示词（可选）生成 1 个目标视频。
   * **图生视频-首帧**：输入首帧图片 + 文本提示词（可选）生成 1 个目标视频。
   * **文生视频**：输入文本提示词生成 1 个目标视频。
* **doubao-seedance-2.0 系列（doubao-seedance-2.0 / 2.0-fast / 2.0-mini）**（有声视频 / 无声视频）
   * **全模态参考生视频**：输入参考图片（0-9 张）+ 参考视频（0-3 个）+ 参考音频（0-3 个）+ 文本提示词（可选）生成 1 个目标视频。注意不可单独输入音频，应至少包含 1 个参考视频或图片。支持生成全新视频、编辑视频、延长视频。
   * **图生视频-首尾帧**：输入首帧图片 + 尾帧图片 + 文本提示词（可选）生成 1 个目标视频。
   * **图生视频-首帧**：输入首帧图片 + 文本提示词（可选）生成 1 个目标视频。
   * **文生视频**：输入文本提示词生成 1 个目标视频。
* **doubao-seedance-1.5-pro**（有声视频 / 无声视频）
   * 支持【图生视频-首尾帧】【图生视频-首帧】【文生视频】。

:::danger
doubao-seedance-2.5 已全面公开，您可在火山方舟平台调用 API 及在线体验。调用模型前，**请务必仔细查阅使用必读**，以确保正确设置任务类型和配置参数。
:::

**参数传入方式说明**

对于 `resolution`、`ratio`、`duration`、`seed`、`camera_fixed`、`watermark` 参数，除了在 request body 中直接传入，也支持在文本提示词后追加 `--[parameters]` 的弱校验方式传参，所有模型均兼容。

不同模型对应支持不同的参数与取值，详见[输出视频格式](https://docs.volcengine.com/docs/82379/2298881#9fe4cce0)。当输入的参数或取值不符合所选的模型时，该参数将被忽略或触发报错：

* **常规方式（推荐）**：在 request body 中直接传入参数。此方式为**强校验**，若参数填写错误，模型会返回错误提示。
* **弱校验方式**：在文本提示词后追加 `--[parameters]`。若参数填写错误，该参数将被忽略或触发报错。

**常规方式（推荐）：在 request body 中直接传入参数**

```JSON
{
    "model": "doubao-seedance-1.5-pro",
    "content": [
        {
            "type": "text",
            "text": "小猫对着镜头打哈欠"
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

**弱校验方式：在文本提示词后追加 `--[parameters]`**

```JSON
{
    "model": "doubao-seedance-1.5-pro",
    "content": [
        {
            "type": "text",
            "text": "小猫对着镜头打哈欠 --rs 720p --rt 16:9 --dur 5 --seed 11 --cf false --wm true"
        }
    ]
}
```

---

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求体

---

**content** `object[]` %%require%%
输入内容列表。

输入给模型生成视频的信息，支持文本、图片、音频、视频、样片任务 ID 等多种类型的元素。

支持以下几种组合：

* 纯文本
* 文本（可选）+ 图片
* 文本（可选）+ 视频
* 文本（可选）+ 音频（仅 doubao-seedance-2.5 支持单独传入音频）
* 文本（可选）+ 图片 + 音频
* 文本（可选）+ 图片 + 视频
* 文本（可选）+ 视频 + 音频
* 文本（可选）+ 图片 + 视频 + 音频
* 样片任务 ID：样片指使用 Seedance 模型成功生成的样片视频，模型可基于样片生成高质量正式视频

:::warning
doubao-seedance-2.5、doubao-seedance-2.0 系列模型不支持直接上传含有真人人脸的参考图 / 视频。为了便利创作者对肖像的使用，平台推出了以下解决方案：

* 支持使用部分模型的含人脸原始产物作为输入素材
* 支持使用预置虚拟人像作为输入素材
* 支持使用已授权真人素材作为输入
:::

**文本信息** `object`
文本部分，作为生成内容的文本提示词。

content.**type** `string` %%require%%
输入内容的类型，此处固定为 `text`。

content.**text** `string` %%require%%
输入给模型的文本提示词，描述期望生成的视频。

:::tip
**提示词语言支持**：所有模型均支持中英文提示词；doubao-seedance-2.5 额外支持西班牙语、印度尼西亚语、葡萄牙语、日语、马来语、泰语、阿拉伯语、越南语、韩语；doubao-seedance-2.0 系列额外支持西班牙语、印度尼西亚语、葡萄牙语、日语。

**提示词字数建议**：中文提示词不超过 500 字，英文提示词不超过 1000 词。字数过多易导致信息分散，模型可能忽略细节、仅关注重点，进而造成视频缺失部分元素。

**更多使用技巧**：提示词的详细使用技巧，请参见 [Seedance 提示词指南](https://docs.volcengine.com/docs/82379/2222480)。
:::

---

**图片信息** `object`
输入给模型的图片信息。

content.**type** `string` %%require%%
输入内容的类型，此处固定为 `image_url`。

content.**image_url** `object` %%require%%
输入给模型的图片对象。

content.image_url.**url** `string` %%require%%
图片 URL、图片 Base64 编码、素材 ID。

* **图片 URL**：填入图片的公网 URL。
* **Base64 编码**：将本地文件转换为 Base64 编码字符串后提交给大模型，遵循格式 `data:image/<图片格式>;base64,<Base64 编码>`，注意 `<图片格式>` 需小写，如 `data:image/png;base64,{base64_image}`。
* **素材 ID**：用于视频生成的预置素材及虚拟人像的 ID，遵循格式 `asset://<ASSET_ID>`。可从素材 & 虚拟人像库获取。

:::tip
**传入单张图片要求**

* **格式**：`jpeg`、`png`、`webp`、`bmp`、`tiff`、`gif`。其中，doubao-seedance-1.5-pro 及以上模型版本额外支持 `heic`、`heif`。
* **宽高比（宽/高）**：`[0.4, 2.5]`
* **宽高长度（px）**：`[300, 6000]`
* **大小**：单张图片小于 30 MB。请求体大小不超过 64 MB。大文件请勿使用 Base64 编码。
* **图片数量**：
   * **图生视频-首帧**：1 张
   * **图生视频-首尾帧**：2 张
   * **doubao-seedance-2.5 全模态参考生视频**：1-30 张
   * **doubao-seedance-2.0 系列全模态参考生视频**：1-9 张
:::

content.**role** `string`
图片的位置或用途。

**图生视频-首帧**：需要传入 1 个 `image_url` 对象，`role` 为 `first_frame` 或不填。

**图生视频-首尾帧**：需要传入 2 个 `image_url` 对象，且 `role` 必填。

* 首帧图片对应的 `role` 为 `first_frame`
* 尾帧图片对应的 `role` 为 `last_frame`

:::tip
传入的首尾帧图片可相同。首尾帧图片的宽高比不一致时，以首帧图片为主，尾帧图片会自动裁剪适配。
:::

**模型支持**：`doubao-seedance-2.5`、`doubao-seedance-2.0 系列`、`doubao-seedance-1.5-pro`

**图生视频-参考图**：必填，每张参考图对应的 `role` 均为 `reference_image`。

**模型支持**：`doubao-seedance-2.5`、`doubao-seedance-2.0 系列`

:::warning
* **图生视频-首帧**、**图生视频-首尾帧**、**全模态参考生视频**（包括参考图、视频、音频）为 3 种互斥场景，**不可混用**。
* **全模态参考生视频**可通过提示词指定参考图片作为首帧 / 尾帧，间接实现「首尾帧 + 全模态参考」效果。若需严格保障首尾帧和指定图片一致，**优先使用图生视频-首尾帧**（配置 `role` 为 `first_frame` / `last_frame`）。
:::

---

**视频信息** `object`
输入给模型的视频信息。

方舟平台信任 doubao-seedance-2.5、doubao-seedance-2.0 系列模型生成的含人脸视频，您可使用**本账号下近 30 天内**由上述模型生成的含人脸原始视频，作为输入素材进行二次创作。

**模型支持**：`doubao-seedance-2.5`、`doubao-seedance-2.0 系列`

content.**type** `string` %%require%%
输入内容的类型，此处固定为 `video_url`。

content.**video_url** `object` %%require%%
输入给模型的视频对象。

content.video_url.**url** `string` %%require%%
视频 URL、素材 ID。

* **视频 URL**：填入视频的公网 URL。
* **素材 ID**：用于视频生成的预置素材及虚拟人像视频的 ID，遵循格式 `asset://<ASSET_ID>`。可从素材 & 虚拟人像库获取。

:::tip
**传入单个视频要求**

* **视频格式**：`mp4`、`mov`，支持编码格式见下表。
* **分辨率**：`480p`、`720p`、`1080p`、`4k`
* **时长**：
   * **doubao-seedance-2.5**：
     * 非视频编辑任务：单个视频时长 `[2, 30]` s。
     * 视频编辑任务：单个视频时长 `[4, 30]` s。
     * 最多传入 10 个参考视频，所有视频总时长不超过 30 s。
   * **doubao-seedance-2.0 系列**：单个视频时长 `[2, 15]` s，最多传入 3 个参考视频，所有视频总时长不超过 15 s。
* **尺寸**：
   * **宽高比（宽/高）**：`[0.4, 2.5]`
   * **宽高长度（px）**：`[300, 6000]`
   * **总像素数**：`[614×664=407696, 3326×2494=8295044]`，即宽和高的乘积符合 `[407696, 8295044]` 的区间要求。
* **大小**：单个视频不超过 200 MB。
* **帧率（FPS）**：`[24, 60]`
:::

**支持的视频编码格式**：

| 容器格式 | 常用文件扩展名 | MIME | 支持编码 |
|---|---|---|---|
| MP4 | .mp4 | video/mp4 | 视频：H.264/AVC、H.265/HEVC；音频：AAC、MP3 |
| QuickTime | .mov | video/quicktime | 视频：H.264/AVC、H.265/HEVC；音频：AAC、MP3、PCM |

content.**role** `string`
视频的位置或用途，此处固定为 `reference_video`。

---

**音频信息** `object`
用作参考音频。

**模型支持**：

* **doubao-seedance-2.5**：可仅传入音频，无需搭配图片 / 视频；也可配合图片 / 视频一起传入
* **doubao-seedance-2.0 系列**：不可单独输入音频，应至少包含 1 个参考视频或图片。

content.**type** `string` %%require%%
输入内容的类型，此处固定为 `audio_url`。

content.**audio_url** `object` %%require%%
输入给模型的音频对象。

content.audio_url.**url** `string` %%require%%
音频 URL、音频 Base64 编码、素材 ID。

* **音频 URL**：填入音频的公网 URL。
* **Base64 编码**：将本地文件转换为 Base64 编码字符串后提交给大模型，遵循格式 `data:audio/<音频格式>;base64,<Base64 编码>`，注意 `<音频格式>` 需小写，如 `data:audio/wav;base64,{base64_audio}`。
* **素材 ID**：用于视频生成的虚拟人的音频素材 ID，遵循格式 `asset://<ASSET_ID>`。可从素材 & 虚拟人像库获取。

:::tip
**传入单个音频要求**

* **格式**：`wav`、`mp3`
* **时长**：
   * **doubao-seedance-2.5**：单个音频时长 `[2, 30]` s，最多传入 10 段参考音频，所有音频总时长不超过 30 s。
   * **doubao-seedance-2.0 系列**：单个音频时长 `[2, 15]` s，最多传入 3 段参考音频，所有音频总时长不超过 15 s。
* **大小**：单个音频不超过 15 MB，请求体大小不超过 64 MB。大文件请勿使用 Base64 编码。
:::

content.**role** `string`
音频的位置或用途，此处固定为 `reference_audio`。

---

**样片信息** `object`
样片信息。基于样片任务 ID，生成正式视频。[阅读文档](https://docs.volcengine.com/docs/82379/2298881#5acd28c8) 获取 draft 功能的使用教程和注意事项。

**模型支持**：`doubao-seedance-1.5-pro`

content.**type** `string` %%require%%
输入内容的类型，此处固定为 `draft_task`。

content.**draft_task** `object` %%require%%
输入给模型的样片任务对象。

content.draft_task.**id** `string` %%require%%
样片任务 ID。

平台将自动复用 Draft 视频使用的用户输入（**model**、content.**text**、content.**image_url**、**generate_audio**、**seed**、**ratio**、**duration**、**camera_fixed**），生成正式视频。其余参数支持指定，不指定将使用本模型的默认值。

使用分为两步：

* **Step 1**：调用本接口生成 Draft 视频。
* **Step 2**：如果确认 Draft 视频符合预期，可基于 Step 1 返回的 Draft 视频任务 ID 调用本接口生成最终视频。

---

**model** `string` %%require%%
模型 ID。

您需要调用的模型的 ID（Model ID）。[开通模型服务](https://console.volcengine.com/ark/region:cn-beijing/openManagement?LLM=%7B%7D&OpenTokenDrawer=false) 后可查询 Model ID。

您也可通过 Endpoint ID 来调用模型，获得限流、计费类型（前付费 / 后付费）、运行状态查询、监控、安全等高级能力。

---

**callback_url** `string`
回调地址。

填写本次生成任务结果的回调通知地址。当视频生成任务有状态变化时，方舟将向此地址推送 POST 请求，请求内容结构与查询任务 API 的返回体一致。

:::tip
回调返回的 `status` 包括以下状态：

* `queued`：排队中
* `running`：任务运行中
* `succeeded`：任务成功（如发送失败，即 5 秒内未接收到成功发送的信息，会回调三次）
* `failed`：任务失败（如发送失败，即 5 秒内未接收到成功发送的信息，会回调三次）
* `expired`：任务超时，即任务处于「运行中或排队中」状态超过过期时间，可通过 `execution_expires_after` 字段设置过期时间
:::

---

**camera_fixed** `boolean` `默认值 false`
固定摄像头。

是否固定摄像头。

* `true`：固定摄像头。平台会在用户提示词中追加固定摄像头，实际效果不保证。
* `false`：不固定摄像头。

参考图场景不支持。

**模型支持**：`doubao-seedance-1.5-pro`

---

**draft** `boolean` `默认值 false`
样片模式。

控制是否开启样片模式。[阅读文档](https://docs.volcengine.com/docs/82379/2298881#5acd28c8) 获取使用教程和注意事项。

* `true`：开启样片模式，生成一段预览视频，快速验证场景结构、镜头调度、主体动作与 Prompt 意图是否符合预期。消耗 token 数较正常视频更少，使用成本更低。
* `false`：关闭样片模式，正常生成一段视频。

:::tip
开启样片模式后，将使用 480p 分辨率生成 Draft 视频（使用其他分辨率会报错），不支持返回尾帧功能，不支持离线推理功能。
:::

**模型支持**：`doubao-seedance-1.5-pro`

---

**duration** `integer`
视频时长。

生成视频时长（单位：秒）。

如果您希望生成整数秒的视频，建议指定 `duration`。

:::warning
**duration 适配规则**

当 `duration` 设置为 `-1` 时，实际生成视频的时长可通过查询视频生成任务 API 返回的 `duration` 字段获取。视频时长与计费相关，请谨慎设置。
:::

| 模型 | `duration = -1` 取值规则 |
|---|---|
| doubao-seedance-2.0 系列、doubao-seedance-1.5-pro | 模型在 `duration` 的有效取值范围内，自主选择合适的视频长度（整数秒）。 |
| doubao-seedance-2.5（视频编辑任务） | 模型根据提示词意图选定待编辑视频，并自动保持输出视频时长和待编辑视频基本一致，不支持另行设置。输出时长可能为非整数秒，且可能略短于待编辑视频，误差约 0.4 秒。 |
| doubao-seedance-2.5（其他任务类型） | 模型在 `duration` 的有效取值范围内，自主选择合适的视频长度（整数秒）。 |

:::tip
**duration 返回值说明**

通过查询视频生成任务 API 接口返回的 `duration` 值为视频的时长约数（整数秒），与实际的视频时长可能不同。具体返回规则如下：

* 计算公式：接口返回的 `duration` = 实际总帧数 / 24（向下取整）
* 举例：若最终生成视频为 133 帧，视频实际时长为 133 / 24 = 5.54 秒；向下取整后，`duration` 最终返回值为 5
:::

:::warning
doubao-seedance-2.5 模型在**视频编辑任务**的限制条件：

* 仅支持配置 `duration` 为 `-1`，不支持指定具体输出时长。
* 传入的待编辑视频时长需在 `[4, 30]` s 内，否则将触发报错。
:::

**模型支持**：

* **doubao-seedance-2.5**：默认值 `-1`；取值范围 `[4, 30]`；或设置为 `-1`（智能选择）
* **doubao-seedance-2.0 系列**：取值范围 `[4, 15]`；或设置为 `-1`（智能选择）
* **doubao-seedance-1.5-pro**：取值范围 `[4, 12]`；或设置为 `-1`（智能选择）

---

**execution_expires_after** `integer` `默认值 172800`
任务超时阈值。

指定任务提交后的过期时间（单位：秒），从 `created_at` 时间戳开始计算，默认 48 小时。

超过该时间后任务会被自动终止，并标记为 `expired` 状态。

不论使用哪种 `service_tier`，都建议根据业务场景设置合适的超时时间。

**取值范围**：`[3600, 259200]`

---

**generate_audio** `boolean` `默认值 true`
生成有声视频。

控制生成的视频是否包含与画面同步的声音。

* `true`：模型输出的视频包含同步音频，模型会基于文本提示词与视觉内容自动生成匹配的人声、音效及背景音乐。建议将对话部分置于双引号内以优化音频生成效果，例如：男人叫住女人说："你记住，以后不可以用手指指月亮。"
* `false`：模型输出的视频为无声视频。

:::warning
生成的有声视频均为单声道，和传入的音频声道数无关。
:::

**模型支持**：`doubao-seedance-2.5`、`doubao-seedance-2.0 系列`、`doubao-seedance-1.5-pro`

---

**omni_reference_task_type** `string` `默认值 auto`
任务类型引导。

doubao-seedance-2.5 **全模态参考生视频任务**包括参考生视频、视频编辑和视频延长 3 类子任务。不同任务类型对参数有特殊限制，为减少任务创建后异步报错的情况，可通过本参数指定子任务类型，以提前校验对应限制。

* 默认情况下，即 `omni_reference_task_type=auto`：模型根据输入素材和提示词自动判定任务类型，再校验参数取值。如果参数与实际任务类型不兼容，任务将触发异步报错。
* 显式指定任务类型，即 `omni_reference_task_type` 为 `reference`、`edit` 或 `extend`：接口在提交任务时提前校验对应任务的特殊参数限制。不符合要求时，接口立即报错，任务不会创建。

:::warning
实际处理任务时，模型仍会进一步结合提示词判断任务类型。若实际判定的任务类型和指定的不一致，仍会触发异步报错。建议遵循各任务类型的提示词写法，降低报错概率。
:::

**可选值**：

* `auto`：由模型根据输入素材和提示词自动判定任务类型。
* `reference`：参考生视频任务，即基于参考图片、参考视频或参考音频生成新视频。设置 `reference` 时，`ratio` 或 `duration` 无特殊限制。
* `edit`：视频编辑任务，即对原视频的画面或音频进行编辑操作。设置 `edit` 时，`content` 中必须至少包含一个 `reference_video`，且视频时长必须为 4–30 秒；`ratio` 必须为 `adaptive`；`duration` 必须为 `-1`。
* `extend`：视频延长任务，即对原视频向前或向后延长。设置 `extend` 时，`content` 中必须至少包含一个 `reference_video`；`ratio` 必须为 `adaptive`。

**模型支持**：`doubao-seedance-2.5`

---

**output_format** `string` `默认值 mp4`
输出格式。

输出视频的格式。

* `mp4`：通用格式，兼容性最好，采用标准色彩精度，可在网页、移动端、各类播放器及分发平台直接播放。
* `mov`：面向专业场景的高色彩精度格式，更好地保持画面色彩与亮度一致性、适用于调色、抠像、合成等对色彩还原要求高的专业后期加工。推荐在视频编辑、视频延长场景使用 mov 格式作为输入和输出。

:::tip
**mov 格式播放兼容性**

mov 格式采用专业编码（H.264 视频编码 + yuv444p 色度采样 + PCM 音频编码），部分播放器可能不兼容。以下为常见的支持播放 mov 格式的播放器：

| 播放器 | macOS | Windows |
|---|---|---|
| IINA | ✓ | ✕ |
| VLC | ✓ | ✓ |
| mpv | ✓ | ✓ |
| ffplay | ✓ | ✓ |
:::

**模型支持**：`doubao-seedance-2.5`

---

**priority** `integer` `默认值 0`
执行优先级。

设置当前请求的执行优先级，决定其在队列中的排序位置。数值越大，优先级越高。

默认情况下，请求按 FIFO（First In, First Out，先进先出）顺序执行；设置较高优先级后，该请求将插队到同 Endpoint（推理接入点）下所有低优先级请求之前。

**示例**：某 Endpoint 当前队列中有 3 个排队中（`status=queued`）任务，优先级均为 0（默认）：

```Plain Text
队列：[任务A: priority=0] → [任务B: priority=0] → [任务C: priority=0]
```

此时提交一个 `priority=5` 的新请求，该请求将直接排到队首：

```Plain Text
队列：[新请求: priority=5] → [任务A: priority=0] → [任务B: priority=0] → [任务C: priority=0]
```

:::tip
* 相同优先级的请求之间仍按 FIFO 排序。
* 优先级仅影响排队顺序，不会中断正在执行中（`status=running`）的任务。
* 优先级仅在同一 Endpoint 内生效，不影响其他 Endpoint。
* 离线推理模式（`service_tier=flex`）不支持配置优先级。
:::

**取值范围**：`[0, 9]`

**模型支持**：`doubao-seedance-2.5`、`doubao-seedance-2.0 系列`

---

**ratio** `string`
视频宽高比。

生成视频的宽高比例。

可选值：`16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive`（根据任务类型和输入内容自动适配宽高比）

**不同模型的取值限制和 `adaptive` 适配规则**：

| 任务类型 | | doubao-seedance-2.5 | doubao-seedance-2.0 系列 | doubao-seedance-1.5-pro |
|---|---|---|---|---|
| 文生视频 | | 支持 `adaptive` 或指定宽高比 | 支持 `adaptive` 或指定宽高比 | 支持 `adaptive` 或指定宽高比 |
| 首帧或首尾帧生视频 | | 自动保持输出视频宽高比和 `first_frame` 指定的首帧图片一致，默认且仅支持 `adaptive` | 支持 `adaptive` 或指定宽高比 | 支持 `adaptive` 或指定宽高比 |
| 全模态参考生视频 | 视频编辑 / 视频延长 | 模型根据提示词意图选定待编辑 / 待延长视频，并自动保持输出视频宽高比和该视频一致，不支持另行设置；默认且仅支持 `adaptive` | 支持 `adaptive` 或指定宽高比 | — |
| 全模态参考生视频 | 参考生视频 | 支持 `adaptive` 或指定宽高比 | 支持 `adaptive` 或指定宽高比 | — |

**不同宽高比对应的宽高像素值**：

图生视频，选择的宽高比与您上传的图片宽高比不一致时，方舟会对您的图片进行裁剪，裁剪时会居中裁剪，详细规则见 [图片裁剪规则](https://docs.volcengine.com/docs/82379/2298881#f76aafc8)。

| 分辨率 | 宽高比 | doubao-seedance-2.5 | doubao-seedance-2.0 系列 | doubao-seedance-1.5-pro |
|---|---|---|---|---|
| 480p | `16:9` | 854×480 | 864×496 | 864×496 |
| 480p | `4:3` | 752×560 | 752×560 | 752×560 |
| 480p | `1:1` | 640×640 | 640×640 | 640×640 |
| 480p | `3:4` | 560×752 | 560×752 | 560×752 |
| 480p | `9:16` | 480×854 | 496×864 | 496×864 |
| 480p | `21:9` | 992×432 | 992×432 | 992×432 |
| 720p | `16:9` | 1280×720 | 1280×720 | 1280×720 |
| 720p | `4:3` | 1112×834 | 1112×834 | 1112×834 |
| 720p | `1:1` | 960×960 | 960×960 | 960×960 |
| 720p | `3:4` | 834×1112 | 834×1112 | 834×1112 |
| 720p | `9:16` | 720×1280 | 720×1280 | 720×1280 |
| 720p | `21:9` | 1470×630 | 1470×630 | 1470×630 |
| 1080p（doubao-seedance-2.0-fast / 2.0-mini 暂不支持） | `16:9` | 1920×1080 | 1920×1080 | 1920×1080 |
| 1080p（doubao-seedance-2.0-fast / 2.0-mini 暂不支持） | `4:3` | 1664×1248 | 1664×1248 | 1664×1248 |
| 1080p（doubao-seedance-2.0-fast / 2.0-mini 暂不支持） | `1:1` | 1440×1440 | 1440×1440 | 1440×1440 |
| 1080p（doubao-seedance-2.0-fast / 2.0-mini 暂不支持） | `3:4` | 1248×1664 | 1248×1664 | 1248×1664 |
| 1080p（doubao-seedance-2.0-fast / 2.0-mini 暂不支持） | `9:16` | 1080×1920 | 1080×1920 | 1080×1920 |
| 1080p（doubao-seedance-2.0-fast / 2.0-mini 暂不支持） | `21:9` | 2206×946 | 2206×946 | 2206×946 |
| 4k（仅 doubao-seedance-2.0 支持） | `16:9` | — | 3840×2160 | — |
| 4k（仅 doubao-seedance-2.0 支持） | `4:3` | — | 3326×2494 | — |
| 4k（仅 doubao-seedance-2.0 支持） | `1:1` | — | 2880×2880 | — |
| 4k（仅 doubao-seedance-2.0 支持） | `3:4` | — | 2494×3326 | — |
| 4k（仅 doubao-seedance-2.0 支持） | `9:16` | — | 2160×3840 | — |
| 4k（仅 doubao-seedance-2.0 支持） | `21:9` | — | 4398×1886 | — |

:::warning
doubao-seedance-2.5 模型在**视频编辑、视频延长、首帧 / 首尾帧生视频任务**的限制条件：仅支持配置 `ratio` 为 `adaptive`，不支持指定具体宽高比。

* 视频编辑 / 视频延长任务中，模型根据提示词意图选定待编辑视频或待延长视频，并自动保持输出视频宽高比和待编辑视频或待延长视频一致。
* 首帧 / 首尾帧生视频任务中，模型自动保持输出视频宽高比和 `first_frame` 指定的首帧图片一致。
:::

**模型支持**：

* **doubao-seedance-2.5**：默认值 `adaptive`
* **doubao-seedance-2.0 系列**：默认值 `adaptive`
* **doubao-seedance-1.5-pro**：默认值 `adaptive`

---

**resolution** `string`
视频分辨率。

视频分辨率。可选值：`480p`、`720p`、`1080p`、`4k`。

:::tip
* **doubao-seedance-2.5** 输出的 1080p 视频和 **doubao-seedance-2.0** 输出的 4k 视频均采用 10bit 位深与 H.265/HEVC 编码。
  * 相较于一般的 8bit 位深，10bit 位深能够保留更丰富的色彩层次与更平滑的渐变过渡，满足专业影视制作与 HDR 视频内容的要求。
  * H.265/HEVC 编码在少数播放环境中可能不兼容。如遇问题，建议升级系统、更换设备，或使用 VLC、MPV、QuickTime Player 等播放器查看。
:::

**模型支持**：

* **doubao-seedance-2.5**：默认值 `720p`；可选值 `480p`、`720p`、`1080p`
* **doubao-seedance-2.0**：默认值 `720p`；可选值 `480p`、`720p`、`1080p`、`4k`
* **doubao-seedance-2.0-fast**：默认值 `720p`；可选值 `480p`、`720p`
* **doubao-seedance-2.0-mini**：默认值 `720p`；可选值 `480p`、`720p`
* **doubao-seedance-1.5-pro**：默认值 `720p`；可选值 `480p`、`720p`、`1080p`

---

**return_last_frame** `boolean` `默认值 false`
返回尾帧。

是否返回生成视频的尾帧图像。

* `true`：返回生成视频的尾帧图像，可通过查询视频生成任务接口获取，尾帧图像的格式为 png，宽高像素值与生成的视频保持一致，无水印。
* `false`：不返回生成视频的尾帧图像。

使用该参数可实现生成多个连续视频：以上一个生成视频的尾帧作为下一个视频任务的首帧，快速生成多个连续视频，调用示例详见 [教程](https://docs.volcengine.com/docs/82379/2298881#141cf7fa)。

---

**safety_identifier** `string`
用户标识。

终端用户的唯一标识符，用于协助平台检测您的应用中可能违反火山方舟使用政策的用户。该标识符为英文字符串，需保证对单个用户固定且唯一，长度不超过 64 个字符。

推荐传入对用户名、用户 ID 或邮箱进行哈希处理后生成的字符串，避免泄露用户隐私信息。

---

**seed** `integer` `默认值 -1`
随机种子。

种子整数，用于控制生成内容的随机性。`-1` 表示使用随机数替代。

:::warning
* 相同的请求下，模型收到不同的 `seed` 值（不指定、令 `seed=-1` 或手动变更），将生成不同的结果。
* 相同的请求下，模型收到相同的 `seed` 值，会生成类似的结果，但不保证完全一致。
:::

**取值范围**：`[-1, 2147483647]`

**模型支持**：`doubao-seedance-1.5-pro`

---

**service_tier** `string` `默认值 default`
服务等级。

指定处理本次请求的服务等级类型。

* `default`：在线推理模式，RPM 和并发数配额较低，适合对推理时效性要求较高的场景。
* `flex`：离线推理模式，TPD 配额更高，价格为在线推理的 50%，适合对推理时延要求不高的场景。`doubao-seedance-2.5`、`doubao-seedance-2.0 系列`暂不支持。

:::warning
不支持修改已提交任务的服务等级。
:::

**模型支持**：`doubao-seedance-1.5-pro`

---

**tools** `object[]`
工具配置。

配置模型要调用的工具。

**模型支持**：`doubao-seedance-2.5`、`doubao-seedance-2.0 系列`

tools.**type** `string` %%require%%
指定使用的工具类型。

* `web_search`：联网搜索工具。开启联网搜索后，模型会根据用户的提示词自主判断是否搜索互联网内容（如商品、天气等）。可提升生成视频的时效性，但也会增加一定的时延。[阅读教程](https://docs.volcengine.com/docs/82379/2291680#c40ed3ef) 获取详细代码示例。

:::tip
实际搜索次数可通过查询视频生成任务 API 返回的 `usage.tool_usage.web_search` 字段获取，如果为 0 表示未搜索。
:::

---

**watermark** `boolean` `默认值 false`
视频水印。

生成视频是否包含水印。

* `true`：生成视频右下角会展示 `AI 生成` 水印。
* `false`：生成视频不含水印。

---

#### 请求体示例

```JSON
{
    "model": "doubao-seedance-1.5-pro",
    "content": [
        {
            "type": "text",
            "text": "小猫对着镜头打哈欠"
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

---

### 响应参数

---

**id** `string`
任务 ID。

视频生成任务 ID。仅保存 7 天（从 `created_at` 时间戳开始计算），超时后将自动清除。

* 设置 `"draft": true`，为 Draft 视频任务 ID。
* 设置 `"draft": false`，为正常视频任务 ID。

创建视频生成任务为**异步接口**，获取 ID 后需要通过查询视频生成任务 API 来查询任务状态。任务成功后会输出生成视频的 `video_url`。

### GET https://www.cii-group.com/app-api/api/v1/contents/generations/tasks/{id}

```http
GET https://www.cii-group.com/app-api/api/v1/contents/generations/tasks/{id}
```

查询视频生成任务的状态。

:::
```mixin-react
return (<Tabs>
<Tabs.TabPane title="快速入口" key="fq9yXaKY"><RenderMd content={` [ ](#)[体验中心](https://console.volcengine.com/ark/region:ark+cn-beijing/experience/vision)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_2abecd05ca2779567c6d32f0ddc7874d.png =20x) </span>[模型列表](https://www.volcengine.com/docs/82379/1330310)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_a5fdd3028d35cc512a10bd71b982b6eb.png =20x) </span>[模型计费](https://www.volcengine.com/docs/82379/1099320#%E8%A7%86%E9%A2%91%E7%94%9F%E6%88%90%E6%A8%A1%E5%9E%8B)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_afbcf38bdec05c05089d5de5c3fd8fc8.png =20x) </span>[API Key](https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey?apikey=%7B%7D)
 <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_57d0bca8e0d122ab1191b40101b5df75.png =20x) </span>[调用教程](https://www.volcengine.com/docs/82379/1366799)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_f45b5cd5863d1eed3bc3c81b9af54407.png =20x) </span>[接口文档](https://www.volcengine.com/docs/82379/1521309)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_1609c71a747f84df24be1e6421ce58f0.png =20x) </span>[常见问题](https://www.volcengine.com/docs/82379/1359411)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_bef4bc3de3535ee19d0c5d6c37b0ffdd.png =20x) </span>[开通模型](https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement?LLM=%7B%7D&OpenTokenDrawer=false)
`}></RenderMd></Tabs.TabPane>
<Tabs.TabPane title="鉴权说明" key="3vCxpwty"><RenderMd content={`本接口支持 API Key 鉴权，详见[鉴权认证方式](https://www.volcengine.com/docs/82379/1298459)。
`}></RenderMd></Tabs.TabPane></Tabs>);
```

---

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

<span id="RxN8G2nH"></span>
### 请求参数 

---

**id** `string` %%require%%
您需要查询的视频生成任务的 ID 。
:::tip
上面参数为Query String Parameters，在URL String中传入。

:::
---

&nbsp;
<span id="7mi8G8RI"></span>
### 响应参数

---

**id ** `string`
视频生成任务 ID 。

---

**model** `string`
任务使用的模型名称，`模型名称`。

---

**status** `string`
任务状态，以及相关的信息：

* `queued`：排队中。
* `running`：任务运行中。
* `cancelled`：取消任务，取消状态24h自动删除（只支持排队中状态的任务被取消）。
* `succeeded`： 任务成功。
* `failed`：任务失败。
* `expired`：任务超时。

---

**error** `object / null`
错误提示信息，任务成功返回`null`，任务失败时返回错误数据，错误信息具体参见 [错误处理](https://www.volcengine.com/docs/82379/1299023#.5pa56Iif6ZSZ6K-v56CB)。

属性

---

error.**code** `string`
错误码。

---

error.**message** `string`
错误提示信息。

---

**created_at** `integer`
任务创建时间的 Unix 时间戳（秒）。

---

**updated_at** `integer`
任务当前状态更新时间的 Unix 时间戳（秒）。

---

**content** `object`
视频生成任务的输出内容。

属性

---

content.**video_url** `string`
生成视频的 URL，格式为 mp4。为保障信息安全，生成的视频会在24小时后被清理，请及时转存。
推荐配置火山引擎 TOS 提供的数据订阅功能，将您的模型推理产物自动转存到自己的 TOS 桶中，便于长期备份或二次加工。详细介绍请参见 [TOS 数据订阅](https://www.volcengine.com/docs/6349/2280949?lang=zh)。
content.**last_frame_url ** `string`
视频的尾帧图像 URL。有效期为 24小时，请及时转存。
说明：[创建视频生成任务](https://www.volcengine.com/docs/82379/1520757) 时设置 `"return_last_frame": true` 时，会返回该参数。

---

**seed** `integer`
本次请求使用的种子整数值。

---

**resolution **  `string` 
生成视频的分辨率。

---

**ratio ** `string`
生成视频的宽高比。

---

**duration** `integer` 
生成视频的时长，单位：秒。
说明：**duration 和 frames 参数只会返回一个**。[创建视频生成任务](https://www.volcengine.com/docs/82379/1520757) 时未指定 frames，会返回 duration。

---

**frames** `integer`  
生成视频的帧数。
说明：**duration 和 frames 参数只会返回一个**。[创建视频生成任务](https://www.volcengine.com/docs/82379/1520757) 时指定了 frames，会返回 frames。

---

**framespersecond**  `integer` 
生成视频的帧率。

---

**generate_audio** `boolean`
生成的视频是否包含与画面同步的声音。仅 doubao-seedance-2.0 / doubao-seedance-2.0-fast、doubao-seedance-1.5-pro 会返回该参数。

* `true`：模型输出的视频包含同步音频。
* `false`：模型输出的视频为无声视频。

---

**tools==^new^==** ** ** `object[]` 
本次请求模型实际使用的工具。未使用工具时不返回。

属性
tools.**type ** `string`
实际使用的工具类型

* web_search：联网搜索工具。

---

**safety_identifier==^new^==** `string`
终端用户的唯一标识符。若 [创建视频生成任务](https://www.volcengine.com/docs/82379/1520757) 时设置了该参数，接口会原样返回此信息。

---

**draft** `boolean`
生成的视频是否为 Draft 视频。仅 doubao-seedance-1.5-pro 会返回该参数。

* `true`：表示当前输出为 Draft 视频。
* `false`：表示当前输出为正常视频。

---

**draft_task_id ** `string`
Draft 视频任务 ID。基于 Draft 视频生成正式视频时，会返回该参数。

---

**service_tier  ** `string`
实际处理任务使用的服务等级。

---

**execution_expires_after** ** ** `integer`
任务超时阈值，单位：秒。

---

**usage** `object`
本次请求的 token 用量。

属性

---

usage.**completion_tokens** `integer`
模型输出视频花费的 token 数量，可作为计费对账依据。
:::tip
doubao-seedance-2.0 系列模型存在最低 token 用量限制，如果实际 token 用量 ＜ 最低 token 用量，本字段会返回最低 token 用量，平台按最低 token 用量计费。

:::
---

usage.**total_tokens** `integer`
本次请求消耗的总 token 数量。视频生成模型不统计输入 token，输入 token 为 0，故 **total_tokens**=**completion_tokens**。

---

usage.**tool_usage==^new^==** ** ** `object`
使用工具的用量信息。

属性
usage.tool_usage.**web_search ** `integer`
实际调用联网搜索工具的次数，仅开启联网搜索时返回。

### GET https://www.cii-group.com/app-api/api/v1/contents/generations/tasks?page_num={page_num}&page_size={page_size}&filter.status={filter.status}&filter.task_ids={filter.task_ids}&filter.model={filter.model}

```http
GET https://www.cii-group.com/app-api/api/v1/contents/generations/tasks?page_num={page_num}&page_size={page_size}&filter.status={filter.status}&filter.task_ids={filter.task_ids}&filter.model={filter.model}
```

当您要查询符合条件的任务，您可以传入条件筛选参数，返回符合要求的任务。

:::
```mixin-react
return (<Tabs>
<Tabs.TabPane title="快速入口" key="opV4RT2k"><RenderMd content={` [ ](#)[体验中心](https://console.volcengine.com/ark/region:ark+cn-beijing/experience/vision)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_2abecd05ca2779567c6d32f0ddc7874d.png =20x) </span>[模型列表](https://www.volcengine.com/docs/82379/1330310)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_a5fdd3028d35cc512a10bd71b982b6eb.png =20x) </span>[模型计费](https://www.volcengine.com/docs/82379/1099320#%E8%A7%86%E9%A2%91%E7%94%9F%E6%88%90%E6%A8%A1%E5%9E%8B)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_afbcf38bdec05c05089d5de5c3fd8fc8.png =20x) </span>[API Key](https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey?apikey=%7B%7D)
 <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_57d0bca8e0d122ab1191b40101b5df75.png =20x) </span>[调用教程](https://www.volcengine.com/docs/82379/1366799)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_f45b5cd5863d1eed3bc3c81b9af54407.png =20x) </span>[接口文档](https://www.volcengine.com/docs/82379/1521675)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_1609c71a747f84df24be1e6421ce58f0.png =20x) </span>[常见问题](https://www.volcengine.com/docs/82379/1359411)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_bef4bc3de3535ee19d0c5d6c37b0ffdd.png =20x) </span>[开通模型](https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement?LLM=%7B%7D&OpenTokenDrawer=false)
`}></RenderMd></Tabs.TabPane>
<Tabs.TabPane title="鉴权说明" key="CPeW5vNl"><RenderMd content={`本接口支持 API Key 鉴权，详见[鉴权认证方式](https://www.volcengine.com/docs/82379/1298459)。
`}></RenderMd></Tabs.TabPane></Tabs>);
```

---

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

<span id="RxN8G2nH"></span>
### 请求参数 

:::tip
下面参数为Query String Parameters，在URL String中传入。

:::
---

**page_num** `integer / null` 
取值范围：[1, 500]
返回结果的页码。

---

**page_size ** `integer / null`
取值范围：[1, 500]
返回结果的每页的结果数量。

---

**filter.status ** `string / null`
过滤参数，查询某个任务状态。

* `queued`：排队中的任务。
* `running`：运行中任务。
* `cancelled`：取消的任务。
* `succeeded`： 成功的任务。
* `failed`：失败的任务。

---

**filter.task_ids ** `string[] / null`
视频生成任务 ID，精确搜索，支持同时搜索多个任务 ID。多个任务 ID 之间通过 `&`连接。示例：`filter.task_ids=id1&filter.task_ids=id2`。

---

**filter.model ** `string / null`
与返回参数不同，该字段为任务使用的推理接入点 ID，精确搜索。

---

**filter.service_tier ** `string / null` `默认值 default`
 处理任务使用的服务等级。

* `default`：在线推理模式
* `flex`：离线推理模式

<span id="7mi8G8RI"></span>
### 响应参数

---

**items ** `object[]`
查询到的视频生成任务列表。

属性

---

items.**id ** `string`
视频生成任务 ID 。

---

items.**model** `string`
任务使用的模型名称和版本，`模型名称`。

---

items.**status** `string`
任务状态，以及相关的信息：

* `queued`：排队中。
* `running`：任务运行中。
* `cancelled`：取消任务（只支持排队中状态的任务被取消）。
* `succeeded`： 任务成功。
* `failed`：任务失败。
* `expired`：任务超时。

---

items.**error** `object / null`
错误提示信息，任务成功返回`null`，任务失败时返回错误数据，错误信息具体参见 [错误处理](https://www.volcengine.com/docs/82379/1393047#653d2c40)。

属性

---

error.**code** `string`
错误码。

---

error.**message** `string`
错误提示信息。

---

items.**created_at** `integer`
任务创建时间的 Unix 时间戳（秒）。

---

items.**updated_at** `integer`
任务当前状态更新时间的 Unix 时间戳（秒）。

---

items.**content** `object`
当视频生成任务完成，会输出该字段，包含生成视频下载的 URL。

属性

---

content.**video_url** `string`
生成视频的URL。为保障信息安全，生成的视频会在24小时后被清理，请及时转存。

---

content.**last_frame_url ** `string`
视频的尾帧图像 URL。有效期为 24小时，请及时转存。
说明：[创建视频生成任务](https://www.volcengine.com/docs/82379/1520757) 时设置 `"return_last_frame": true` 时，会返回参数。

---

items.**seed** `integer`
本次请求使用的种子整数值。

---

items.**resolution **  `string` 
生成视频的分辨率。

---

items.**ratio ** `string`
生成视频的宽高比。

---

items.**duration** `integer` 
生成视频的时长，单位：秒。
说明：**duration 和 frames 参数只会返回一个**。[创建视频生成任务](https://www.volcengine.com/docs/82379/1520757) 时未指定 frames，会返回 duration。

---

items.**frames ** `integer`  
生成视频的帧数。
说明：**duration 和 frames 参数只会返回一个**。[创建视频生成任务](https://www.volcengine.com/docs/82379/1520757) 时指定了 frames，会返回 frames。

---

items.**framespersecond**  `integer` 
生成视频的帧率。

---

items.**generate_audio** `boolean`
生成的视频是否包含与画面同步的声音。仅 doubao-seedance-2.0 / doubao-seedance-2.0-fast、doubao-seedance-1.5-pro 会返回该参数。

* `true`：模型输出的视频包含同步音频。
* `false`：模型输出的视频为无声视频。

---

items.**tools==^new^==** ** ** `object[]` 
本次请求模型实际使用的工具。未使用工具时不返回。

属性
items.tools.**type ** `string`
实际使用的工具类型

* web_search：联网搜索工具。

---

items.**safety_identifier==^new^==** `string`
终端用户的唯一标识符。若 [创建视频生成任务](https://www.volcengine.com/docs/82379/1520757) 时设置了该参数，接口会原样返回此信息。

---

items.**draft** `boolean`
生成的视频是否为 Draft 视频。仅 doubao-seedance-1.5-pro 会返回该参数。

* `true`：表示当前输出为 Draft 视频。
* `false`：表示当前输出为正常视频。

---

items.**draft_task_id ** `string`
Draft 视频任务 ID。基于 Draft 视频生成正式视频时，会返回该参数。

---

items.**service_tier ** `string`
实际处理任务使用的服务等级。

---

items.**execution_expires_after** ** ** `integer`
任务超时阈值，单位：秒。

---

items.**usage** `object`
本次请求的 token 用量。

属性

---

items.usage.**completion_tokens** `integer`
模型输出视频花费的 token 数量，可作为计费对账依据。
:::tip
doubao-seedance-2.0 系列模型存在最低 token 用量限制，如果实际 token 用量 ＜ 最低 token 用量，本字段会返回最低 token 用量，平台按最低 token 用量计费。

:::
---

items.usage.**total_tokens**`integer`
本次请求消耗的总 token 数量。视频生成模型不统计输入 token，输入 token 为 0，故 **total_tokens**=**completion_tokens**。

---

items.usage.**tool_usage==^new^==** ** ** `object`
使用工具的用量信息。

属性
items.usage.tool_usage.**web_search ** `integer`
实际调用联网搜索工具的次数，仅开启联网搜索时返回。

---

**total ** `integer`
符合筛选条件的任务数量。

### DELETE https://www.cii-group.com/app-api/api/v1/contents/generations/tasks/{id}

```http
DELETE https://www.cii-group.com/app-api/api/v1/contents/generations/tasks/{id}
```

取消排队中的视频生成任务，或者删除视频生成任务记录。

```mixin-react
return (<Tabs>
<Tabs.TabPane title="快速入口" key="vI631gwS"><RenderMd content={` [ ](#)[体验中心](https://console.volcengine.com/ark/region:ark+cn-beijing/experience/vision)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_2abecd05ca2779567c6d32f0ddc7874d.png =20x) </span>[模型列表](https://www.volcengine.com/docs/82379/1330310)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_a5fdd3028d35cc512a10bd71b982b6eb.png =20x) </span>[模型计费](https://www.volcengine.com/docs/82379/1099320#%E8%A7%86%E9%A2%91%E7%94%9F%E6%88%90%E6%A8%A1%E5%9E%8B)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_afbcf38bdec05c05089d5de5c3fd8fc8.png =20x) </span>[API Key](https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey?apikey=%7B%7D)
 <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_57d0bca8e0d122ab1191b40101b5df75.png =20x) </span>[调用教程](https://www.volcengine.com/docs/82379/1366799)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_f45b5cd5863d1eed3bc3c81b9af54407.png =20x) </span>[接口文档](https://www.volcengine.com/docs/82379/1521675)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_1609c71a747f84df24be1e6421ce58f0.png =20x) </span>[常见问题](https://www.volcengine.com/docs/82379/1359411)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_bef4bc3de3535ee19d0c5d6c37b0ffdd.png =20x) </span>[开通模型](https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement?LLM=%7B%7D&OpenTokenDrawer=false)
`}></RenderMd></Tabs.TabPane>
<Tabs.TabPane title="鉴权说明" key="L8aMwmZD"><RenderMd content={`本接口支持 API Key 鉴权，详见[鉴权认证方式](https://www.volcengine.com/docs/82379/1298459)。
`}></RenderMd></Tabs.TabPane></Tabs>);
```

---

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

<span id="RxN8G2nH"></span>
### 请求参数 

:::tip
下面参数为Query String Parameters，在URL String中传入。

:::
---

**id** `string` %%require%%
需要取消或者删除的视频生成任务。
任务状态不同，调用`DELETE`接口，执行的操作有所不同，具体说明如下：

|当前任务状态  |是否支持DELETE操作 |操作含义  |DELETE操作后任务状态 |
|---|---|---|---|
|queued  |是  |任务取消排队，任务状态被变更为cancelled。  |cancelled  |
|running  |否 |\- |\- |
|succeeded  |是  |删除视频生成任务记录，后续将不支持查询。  |\- |
|failed  |是  |删除视频生成任务记录，后续将不支持查询。  |\- |
|cancelled  |否  |\- |\- |
|expired |是 |删除视频生成任务记录，后续将不支持查询。  |\- |

&nbsp;

---

&nbsp;
<span id="7mi8G8RI"></span>
### 响应参数

本接口无返回参数。


---

## 视频去字幕任务 (Subtitle Erase)

### POST https://www.cii-group.com/app-api/api/v1/videos/subtitle-erase/tasks

```http
POST https://www.cii-group.com/app-api/api/v1/videos/subtitle-erase/tasks
```

# 创建字幕擦除任务

提交可访问的视频地址，创建异步字幕擦除任务。创建成功后使用 `task_id` 查询任务状态和结果视频地址。

---

### 鉴权方式

请求头必须携带 API Key：

```http
Authorization: Bearer <API Key>
```

---

### 请求参数

#### 请求体

**video_url** `string` %%require%%

待处理视频的公网 HTTP(S) 地址。原始视频文件大小不能超过 **100G**。该地址需要在任务处理期间保持可访问；使用签名地址时，请确保其有效期足够覆盖任务处理时间。

除 `video_url` 外不需要传递任务类型或其他处理参数。

---

### 请求幂等

系统会根据 API Key 和规范化后的请求内容生成幂等指纹。相同请求在已有进行中任务时会返回已创建的任务，不会重复创建任务或重复计费。

任务已进入终态后，再次提交相同视频会创建一条新任务。

---

### 计费

平台生成的视频提交原始有效地址时，提供免费的字幕擦除支持。第三方视频按照原始视频时长计费，**每秒 1 积分**，不足 1 秒按 1 秒计算；账户积分不足时返回 `402 InsufficientBalance`。

---

### 常见问题

#### 1. 本接口能做什么？

本接口使用 AI 检测并擦除视频中的字幕，同时对字幕区域进行画面修复。当前能力主要适用于以下字幕：

| 条件 | 适用范围 |
| --- | --- |
| 字幕位置 | 视频画面下方 50% 区域内的横向字幕。 |
| 字幕大小 | 单个文字的竖向高度约占视频高度的 1% 至 10%。 |
| 字幕语言 | 中文、英文。 |

以上范围表示模型主要优化方向，不代表每个视频都能完全无痕擦除。字体、背景、运动和画面复杂度不同，实际效果可能存在差异，完成后应检查输出视频。

#### 2. 哪些内容不能处理或不能保证效果？

- 视频画面上方 50% 区域内的字幕不在主要擦除范围内。
- 竖向字幕、文字高度小于视频高度 1% 或大于 10% 的字幕不保证擦除。
- 中文和英文以外的字幕，以及无法识别语种的字幕，不保证正常擦除。
- 接口不提供区域框选参数，不能指定只擦除某个局部区域。
- 位置和大小符合擦除条件的横幅、招牌等场景文字可能被误判并擦除。
- 本接口用于字幕擦除，不应作为水印、Logo、人物或其他画面对象的通用移除工具。

#### 3. 完成任务需要满足什么条件？

1. 使用有效且已启用的 API Key 发起请求，并保存创建成功后返回的 `task_id`。
2. `video_url` 必须是公网可访问的 HTTP(S) 地址，不能依赖登录状态、浏览器 Cookie、内网地址或本地地址。
3. 原始视频文件大小必须小于或等于 **100G**；超过限制时无法完成处理。
4. 使用带签名的临时地址时，地址在提交时不能已经过期，并且需要在任务开始处理前持续有效。
5. 字幕擦除是异步任务。创建成功不代表擦除已经完成，需要通过获取字幕擦除结果接口轮询至 `completed` 或 `failed`。
6. 如果提交的是第三方视频，账户需要有足够积分支付整个源视频时长的处理费用。

#### 4. 平台生成视频是否免费？

通过本平台生成的视频，在提交其原始、未转存、仍有效的视频地址时，支持免费字幕擦除。

如果视频来自第三方，或平台生成视频的原始地址已过期、已转存、无法确认属于平台生成内容，则按照第三方视频规则计费。

#### 5. 完成任务需要多少积分？

- 平台生成视频提交原始有效地址时，目前不扣除积分。
- 第三方视频按照原始视频时长计费，每秒 1 积分，不足 1 秒按 1 秒计算。
- 是否免费以平台对视频来源的识别结果为准，不能仅根据文件名或域名认定为免费任务。
- 任务处理失败且已经发生预扣时，系统会按任务最终状态执行退款。

例如，第三方视频时长为 62 秒，则需要 62 积分。文件大小不会直接改变计费规则，但文件仍必须满足不超过 100G 的限制。

#### 6. 一般需要等待多久？

根据能力提供方的参考数据，在不排队时通常可在 20 至 30 分钟内完成；短时间集中提交大量任务时可能进入队列，一般可能需要 2 至 3 小时。该时间为经验值而非服务时限承诺，实际耗时还会受到视频时长、源地址下载速度和共享资源负载影响，建议平滑提交任务。

首次使用字幕擦除能力时，服务能力准备可能需要约 5 分钟。遇到初始化提示时，应稍后重试；不要在短时间内高频创建大量任务。

#### 7. 重复提交会不会重复处理和扣费？

系统会对同一 API Key 下规范化后的请求内容生成幂等指纹。相同请求已经存在进行中任务时，将返回该任务而不会重复创建任务或重复计费；原任务进入终态后再次提交，会创建新任务并按新任务重新判断费用。

---

### 请求示例

```json
{
  "video_url": "https://example.com/source.mp4"
}
```

---

### 成功响应

```json
{
  "success": true,
  "task_id": "se-24023a2230ef49df886825cab738e4ae",
  "request_id": "c8cb70b2-bad1-4b73-b5bb-86b956d39070"
}
```

**success** `boolean`

是否创建成功。成功时为 `true`。

---

**task_id** `string`

字幕擦除任务 ID，用于查询任务状态和结果。

---

**request_id** `string`

本次创建请求的追踪 ID，可用于问题排查。

---

### 失败响应

```json
{
  "success": false,
  "task_id": "",
  "request_id": "c8cb70b2-bad1-4b73-b5bb-86b956d39070",
  "error": {
    "code": "InvalidParameter",
    "message": "仅支持 video_url 参数",
    "param": "video_url",
    "type": "BadRequest"
  }
}
```

**error.code** `string`

机器可读的错误码。

---

**error.message** `string`

错误说明。

---

**error.param** `string`

相关参数名称；没有特定参数时为空字符串。

---

**error.type** `string`

错误类型。

---

### 常见错误

| HTTP 状态码 | error.code | 说明 |
| --- | --- | --- |
| 400 | `InvalidParameter` | 请求体不符合要求，例如缺少或传入了空的 `video_url`。 |
| 401 | `AuthenticationError` | API Key 无效或已禁用。 |
| 402 | `InsufficientBalance` | 第三方视频计费场景下账户积分不足。 |
| 500 | `InternalServerError` | 服务端内部错误。 |

### GET https://www.cii-group.com/app-api/api/v1/videos/subtitle-erase/tasks/{task_id}

```http
GET https://www.cii-group.com/app-api/api/v1/videos/subtitle-erase/tasks/{task_id}
```

# 获取字幕擦除结果

根据创建任务返回的 `task_id` 查询字幕擦除状态和结果。任务未结束时建议采用间隔轮询。

---

### 鉴权方式

请求头必须携带创建任务所用 API Key：

```http
Authorization: Bearer <API Key>
```

仅任务所属用户和 API Key 可查询该任务；无权访问时按任务不存在处理。

---

### 路径参数

**task_id** `string` %%require%%

创建字幕擦除任务后返回的任务 ID。

---

### 任务状态

| status | 说明 |
| --- | --- |
| `running` | 任务正在处理中。 |
| `completed` | 任务已完成，结果位于 `result.video_url`。 |
| `failed` | 任务失败，失败原因位于 `error`。 |

---

### 常见问题

#### 1. 为什么任务长时间处于 `running`？

`running` 表示任务仍在排队或处理中。根据能力提供方的参考数据，不排队时通常需要 20 至 30 分钟；集中提交导致排队时可能需要 2 至 3 小时。该时间为经验值，不是服务时限承诺。

应使用创建接口返回的同一个 `task_id` 继续间隔轮询。相同视频已经存在进行中任务时，重复调用创建接口仍会返回已有任务，不能通过重复创建绕过队列。

#### 2. `success` 为 `true` 是否代表字幕擦除成功？

不一定。`success` 仅表示本次查询请求被正常处理，任务结果必须以 `status` 为准：

- `completed`：任务完成，从 `result.video_url` 获取结果。
- `failed`：任务失败，从 `error.code` 和 `error.message` 获取原因。
- `running`：任务仍在处理中，稍后继续查询。

#### 3. 结果视频地址可以永久访问吗？

不能。`result.video_url` 是临时下载地址，真实失效时间由 `expires_at` 给出。不同视频的结果地址有效期可能不同，不应固定按 24 小时计算，也不要将源视频保存周期视为结果下载地址有效期。

客户端应在当前时间早于 `expires_at` 时完成下载或转存，并以下载成功作为业务侧保存完成的依据。接口不保证已经失效的结果地址可以通过重复查询续期。

#### 4. 任务失败后如何处理？

先记录 `task_id`、`request_id`、`error.code` 和 `error.message`。源地址无法访问、签名过期、文件超过 100G、视频处理失败或结果地址生成失败，都可能使任务进入 `failed`。修正源地址或输入文件后重新调用创建接口会生成新任务；已预扣但最终失败的计费任务会按任务状态执行退款。

---

### 成功完成响应

```json
{
  "success": true,
  "task_id": "se-24023a2230ef49df886825cab738e4ae",
  "request_id": "5ed13f9e-4904-49cc-9147-b9f22674eb60",
  "status": "completed",
  "result": {
    "video_url": "https://example.com/result.mp4?auth_key=..."
  },
  "expires_at": 1785127299,
  "error": null,
  "created_at": 1785123639,
  "finished_at": 1785123699
}
```

**success** `boolean`

本次查询是否成功。任务处理失败时，该字段仍为 `true`，请以 `status` 和 `error` 判断任务结果。

---

**task_id** `string`

字幕擦除任务 ID。

---

**request_id** `string`

本次查询请求的追踪 ID。

---

**status** `string`

任务状态，取值为 `running`、`completed` 或 `failed`。

---

**result** `object / null`

仅在 `status` 为 `completed` 时返回结果对象，其他状态为 `null`。

result.**video_url** `string`

去字幕后的视频下载地址。

---

**expires_at** `integer / null`

结果视频地址的真实失效时间，Unix 秒级时间戳（UTC）。仅在任务完成且存在结果地址时返回；客户端应在该时间前下载或使用视频地址。

---

**error** `object / null`

仅在 `status` 为 `failed` 时返回。

error.**code** `string`

机器可读的失败原因：

| error.code | 说明 |
| --- | --- |
| `SourceVideoUnavailable` | 无法读取原始视频。请检查 `video_url` 的公网可访问性和有效期。 |
| `InvalidSourceVideo` | 无法获取原始视频的有效媒体信息。请确认视频文件可正常播放。 |
| `ResultUnavailable` | 任务已完成处理，但结果视频暂不可用。 |
| `InsufficientBalance` | 账户积分不足，无法完成本次任务。 |
| `TaskFailed` | 字幕擦除处理失败，且没有可进一步公开的具体原因。 |

error.**message** `string`

错误说明。

---

error.**param** `string`

相关参数名称；没有特定参数时为空字符串。

---

error.**type** `string`

错误类型，任务处理失败时为 `TaskFailed`。

---

**created_at** `integer`

任务创建时间，Unix 秒级时间戳。

---

**finished_at** `integer / null`

任务完成或失败时间，Unix 秒级时间戳；`running` 状态时为 `null`。

---

### 处理中响应

```json
{
  "success": true,
  "task_id": "se-24023a2230ef49df886825cab738e4ae",
  "request_id": "5ed13f9e-4904-49cc-9147-b9f22674eb60",
  "status": "running",
  "result": null,
  "expires_at": null,
  "error": null,
  "created_at": 1785123639,
  "finished_at": null
}
```

---

### 任务失败响应

```json
{
  "success": true,
  "task_id": "se-24023a2230ef49df886825cab738e4ae",
  "request_id": "5ed13f9e-4904-49cc-9147-b9f22674eb60",
  "status": "failed",
  "result": null,
  "expires_at": null,
  "error": {
    "code": "SourceVideoUnavailable",
    "message": "无法读取原始视频，请确认 video_url 可公网访问且在任务处理期间持续有效后重新提交",
    "param": "",
    "type": "TaskFailed"
  },
  "created_at": 1785123639,
  "finished_at": 1785123699
}
```

---

### 查询失败响应

```json
{
  "success": false,
  "task_id": "",
  "request_id": "5ed13f9e-4904-49cc-9147-b9f22674eb60",
  "error": {
    "code": "NotFound",
    "message": "任务不存在或无权访问",
    "param": "",
    "type": "NotFound"
  }
}
```

### 常见错误

| HTTP 状态码 | error.code | 说明 |
| --- | --- | --- |
| 401 | `AuthenticationError` | API Key 无效或已禁用。 |
| 404 | `NotFound` | 任务不存在或无访问权限。 |
| 500 | `InternalServerError` | 服务端内部错误。 |


---

## 图片生成 (Image Generations)

### POST https://www.cii-group.com/app-api/api/v1/images/generations

```http
POST https://www.cii-group.com/app-api/api/v1/images/generations
```

本文介绍图片生成模型如 doubao-seedream-5.0-lite 的调用 API ，包括输入输出参数，取值范围，注意事项等信息，供您使用接口时查阅字段含义。

**不同模型支持的图片生成能力简介**

* **doubao\-seedream\-5.0\-pro==^new^==** **doubao\-seedream\-5.0\-lite==^new^==** **、doubao\-seedream\-4.5/4.0**
   * 生成组图（组图：基于您输入的内容，生成的一组内容关联的图片；需配置 **sequential_image_generation ** 为`auto` **）** 
      * 多图生组图，根据您输入的 **++多张参考图片（2\-14）++ **  +++文本提示词++ 生成一组内容关联的图片（输入的参考图数量+最终生成的图片数量≤15张）。
      * 单图生组图，根据您输入的 ++单张参考图片+文本提示词++ 生成一组内容关联的图片（最多生成14张图片）。
      * 文生组图，根据您输入的 ++文本提示词++ 生成一组内容关联的图片（最多生成15张图片）。
   * 生成单图（配置 **sequential_image_generation ** 为`disabled` **）** 
      * 多图生图，根据您输入的 **++多张参考图片（2\-14）++ **  +++文本提示词++ 生成单张图片。
      * 单图生图，根据您输入的 ++单张参考图片+文本提示词++ 生成单张图片。
      * 文生图，根据您输入的 ++文本提示词++ 生成单张图片。
* **doubao\-seedream\-3.0**
   * 文生图，根据您输入的 ++文本提示词++ 生成单张图片。

&nbsp;

```mixin-react
return (<Tabs>
<Tabs.TabPane title="鉴权说明" key="oOTdY3Sn"><RenderMd content={`本接口仅支持 API Key 鉴权，请在 [获取 API Key](https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey) 页面，获取长效 API Key。
`}></RenderMd></Tabs.TabPane>
<Tabs.TabPane title="快速入门" key="HHCpvO5jKo"><RenderMd content={` [ ](#)[体验中心](https://console.volcengine.com/ark/region:ark+cn-beijing/experience/vision?type=GenImage)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_2abecd05ca2779567c6d32f0ddc7874d.png =20x) </span>[模型列表](https://www.volcengine.com/docs/82379/1330310?lang=zh#d3e5e0eb)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_a5fdd3028d35cc512a10bd71b982b6eb.png =20x) </span>[模型计费](https://www.volcengine.com/docs/82379/1544106?lang=zh#457edfd0)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_afbcf38bdec05c05089d5de5c3fd8fc8.png =20x) </span>[API Key](https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey?apikey=%7B%7D)
 <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_57d0bca8e0d122ab1191b40101b5df75.png =20x) </span>[调用教程](https://www.volcengine.com/docs/82379/1548482)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_f45b5cd5863d1eed3bc3c81b9af54407.png =20x) </span>[接口文档](https://www.volcengine.com/docs/82379/1666945)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_1609c71a747f84df24be1e6421ce58f0.png =20x) </span>[常见问题](https://www.volcengine.com/docs/82379/1359411)       <span>![图片](https://portal.volccdn.com/obj/volcfe/cloud-universal-doc/upload_bef4bc3de3535ee19d0c5d6c37b0ffdd.png =20x) </span>[开通模型](https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement?LLM=%7B%7D&OpenTokenDrawer=false)
`}></RenderMd></Tabs.TabPane></Tabs>);
```

---

<span id="7thx2dVa"></span>
### 请求参数 
<span id="BFVUvDi6"></span>
#### 请求体

---

**model** `string` %%require%%
本次请求使用模型的 [Model ID](https://www.volcengine.com/docs/82379/1099455?lang=zh#fc299dc6) 或[推理接入点](https://www.volcengine.com/docs/82379/1099522) (Endpoint ID)。

---

**prompt ** `string` %%require%%
用于生成图像的提示词，支持中英文。（查看提示词指南：[doubao-seedream-4.0](https://www.volcengine.com/docs/82379/1829186) 、[doubao-seedream-3.0](https://www.volcengine.com/docs/82379/1795150)）
建议不超过300个汉字或600个英文单词。字数过多信息容易分散，模型可能因此忽略细节，只关注重点，造成图片缺失部分元素。

---

**image** `string/array` 
> doubao\-seedream\-3.0 不支持该参数

输入的图片信息，支持 URL 或 Base64 编码。其中，doubao\-seedream\-5.0\-lite/4.5/4.0 支持单图或多图输入（[查看多图融合示例](https://www.volcengine.com/docs/82379/1824121?lang=zh#4a35e28f)）。

* 图片URL：请确保图片URL可被访问。
* Base64编码：请遵循此格式`data:image/<图片格式>;base64,<Base64编码>`。注意 `<图片格式>` 需小写，如 `data:image/png;base64,<base64_image>`。

:::tip

* 传入单张图片要求：
   * 图片格式：jpeg、png（doubao\-seedream\-5.0\-lite/4.5/4.0 模型新增支持 webp、bmp、tiff、gif 格式**==^new^==**）
   * 宽高比（宽/高）范围：
      * [1/16, 16] (适用模型：doubao\-seedream\-5.0\-lite/4.5/4.0）
      * [1/3, 3] (适用模型：doubao\-seedream\-3.0）
   * 宽高长度（px） \> 14
   * 大小：不超过 10MB
   * 总像素：不超过 `6000x6000=36000000` px （对单张图宽度和高度的像素乘积限制，而不是对宽度或高度的单独值进行限制）
* doubao\-seedream\-5.0\-lite/4.5/4.0 最多支持传入 14 张参考图。

:::
---

**size **  `string` 

```mixin-react
return (<Tabs>
<Tabs.TabPane title="doubao-seedream-5.0-lite" key="BMB6AP1M"><RenderMd content={`指定生成图像的尺寸信息，支持以下两种方式，不可混用。

* 方式 1 | 指定生成图像的分辨率，并在prompt中用自然语言描述图片宽高比、图片形状或图片用途，最终由模型判断生成图片的大小。
   * 可选值：\`2K\`、\`3K\`
* 方式 2 | 指定生成图像的宽高像素值：
   * 默认值：\`2048x2048\`
   * 总像素取值范围：[\`2560x1440=3686400\`, \`3072x3072x1.1025=10404496\`] 
   * 宽高比取值范围：[1/16, 16]

:::tip
采用方式 2 时，需同时满足总像素取值范围和宽高比取值范围。其中，总像素是对单张图宽度和高度的像素乘积限制，而不是对宽度或高度的单独值进行限制。

* **有效示例**：\`3750x1250\`

总像素值 3750x1250=4687500，符合 [3686400, 10404496] 的区间要求；宽高比 3750/1250=3，符合 [1/16, 16] 的区间要求，故该示例值有效。

* **无效示例**：\`1500x1500\`

总像素值 1500x1500=2250000，未达到 3686400 的最低要求；宽高 1500/1500=1，虽符合 [1/16, 16] 的区间要求，但因其未同时满足两项限制，故该示例值无效。
:::
推荐的宽高像素值：

|分辨率 |宽高比 |宽高像素值 |
|---|---|---|
|<div style="text-align: center">|1:1 |2048x2048 |\\
|2K</div>| | |\\
| | | |
|^^|4:3 |2304x1728 |
|^^|3:4 |1728x2304 |
|^^|16:9 |2848x1600 |
|^^|9:16 |1600x2848 |
|^^|3:2 |2496x1664 |
|^^|2:3 |1664x2496 |
|^^|21:9 |3136x1344 |
|<div style="text-align: center">|1:1 |3072x3072 |\\
|3K</div>| | |\\
| | | |
|^^|4:3 |3456x2592 |
|^^|3:4 |2592x3456 |
|^^|16:9  |4096x2304 |
|^^|9:16 |2304x4096 |
|^^|2:3 |2496x3744 |
|^^|3:2 |3744x2496 |
|^^|21:9 |4704x2016 |

`}></RenderMd></Tabs.TabPane>
<Tabs.TabPane title="doubao-seedream-4.5" key="kghENadO"><RenderMd content={`指定生成图像的尺寸信息，支持以下两种方式，不可混用。

* 方式 1 | 指定生成图像的分辨率，并在prompt中用自然语言描述图片宽高比、图片形状或图片用途，最终由模型判断生成图片的大小。
   * 可选值：\`2K\`、\`4K\`
* 方式 2 | 指定生成图像的宽高像素值：
   * 默认值：\`2048x2048\`
   * 总像素取值范围：[\`2560x1440=3686400\`, \`4096x4096=16777216\`] 
   * 宽高比取值范围：[1/16, 16]

:::tip
采用方式 2 时，需同时满足总像素取值范围和宽高比取值范围。其中，总像素是对单张图宽度和高度的像素乘积限制，而不是对宽度或高度的单独值进行限制。

* **有效示例**：\`3750x1250\`

总像素值 3750x1250=4687500，符合 [3686400, 16777216] 的区间要求；宽高比 3750/1250=3，符合 [1/16, 16] 的区间要求，故该示例值有效。

* **无效示例**：\`1500x1500\`

总像素值 1500x1500=2250000，未达到 3686400 的最低要求；宽高 1500/1500=1，虽符合 [1/16, 16] 的区间要求，但因其未同时满足两项限制，故该示例值无效。
:::
推荐的宽高像素值：

|分辨率 |宽高比 |宽高像素值 |
|---|---|---|
|<div style="text-align: center">|1:1 |2048x2048 |\\
|2K</div>| | |\\
| | | |
|^^|4:3 |2304x1728 |
|^^|3:4 |1728x2304 |
|^^|16:9 |2848x1600 |
|^^|9:16 |1600x2848 |
|^^|3:2 |2496x1664 |
|^^|2:3 |1664x2496 |
|^^|21:9 |3136x1344 |
|<div style="text-align: center">|1:1 |4096x4096 |\\
|4K</div>| | |\\
| | | |
|^^|3:4 |3520x4704 |
|^^|4:3 |4704x3520 |
|^^|16:9 |5504x3040 |
|^^|9:16 |3040x5504 |
|^^|2:3  |3328x4992 |
|^^|3:2  |4992x3328 |
|^^|21:9  |6240x2656 |

`}></RenderMd></Tabs.TabPane>
<Tabs.TabPane title="doubao-seedream-4.0" key="MKsftGMr"><RenderMd content={`指定生成图像的尺寸信息，支持以下两种方式，不可混用。

* 方式 1 | 指定生成图像的分辨率，并在prompt中用自然语言描述图片宽高比、图片形状或图片用途，最终由模型判断生成图片的大小。
   * 可选值：\`1K\`、\`2K\`、\`4K\`
* 方式 2 | 指定生成图像的宽高像素值：
   * 默认值：\`2048x2048\`
   * 总像素取值范围：[\`1280x720=921600\`, \`4096x4096=16777216\`] 
   * 宽高比取值范围：[1/16, 16]

:::tip
采用方式 2 时，需同时满足总像素取值范围和宽高比取值范围。其中，总像素是对单张图宽度和高度的像素乘积限制，而不是对宽度或高度的单独值进行限制。

* **有效示例**：\`1600x600\`

总像素值 1600x600=960000，符合 [921600, 16777216] 的区间要求；宽高比 1600/600=8/3，符合 [1/16, 16] 的区间要求，故该示例值有效。

* **无效示例**：\`800x800\`

总像素值 800x800=640000，未达到 921600 的最低要求；宽高 800/800=1，虽符合 [1/16, 16] 的区间要求，但因其未同时满足两项限制，故该示例值无效。
:::
推荐的宽高像素值：

|分辨率 |宽高比 |宽高像素值 |
|---|---|---|
|<div style="text-align: center">|1:1 |1024x1024 |\\
|1K</div>| | |\\
| | | |
|^^|4:3 |1152x864 |
|^^|3:4 |864x1152 |
|^^|16:9 |1280x720 |
|^^|9:16 |720x1280 |
|^^|3:2 |1248x832 |
|^^|2:3 |832x1248  |
|^^|21:9 |1512x648 |
|<div style="text-align: center">|1:1 |2048x2048 |\\
|2K</div>| | |\\
| | | |
|^^|4:3 |2304x1728 |
|^^|3:4 |1728x2304 |
|^^|16:9 |2848x1600 |
|^^|9:16 |1600x2848 |
|^^|3:2 |2496x1664 |
|^^|2:3 |1664x2496 |
|^^|21:9 |3136x1344 |
|<div style="text-align: center">|1:1 |4096x4096 |\\
|4K</div>| | |\\
| | | |
|^^|3:4 |3520x4704 |
|^^|4:3 |4704x3520 |
|^^|16:9 |5504x3040 |
|^^|9:16 |3040x5504 |
|^^|2:3  |3328x4992 |
|^^|3:2  |4992x3328 |
|^^|21:9  |6240x2656 |

`}></RenderMd></Tabs.TabPane>
<Tabs.TabPane title="doubao-seedream-3.0" key="dUuqsxPhNL"><RenderMd content={`指定生成图像的宽高像素值。

* 默认值：\`1024x1024\`
* 单张图片像素取值范围： [\`512x512\`, \`2048x2048\`] 

推荐的宽高像素值：

|宽高比 |宽高像素值 |
|---|---|
|1:1 |1024x1024 |
|4:3 |864x1152 |
|3:4 |1152x864 |
|16:9 |1280x720 |
|9:16 |720x1280 |
|3:2 |832x1248  |
|2:3 |1248x832 |
|21:9 |1512x648 |

`}></RenderMd></Tabs.TabPane></Tabs>);
```

---

**seed** `integer`  `默认值 -1`
> 仅 doubao\-seedream\-3.0 支持该参数

随机数种子，用于控制模型生成内容的随机性。取值范围为 [\-1, 2147483647]。
:::warning

* 相同的请求下，模型收到不同的seed值，如：不指定seed值或令seed取值为\-1（会使用随机数替代）、或手动变更seed值，将生成不同的结果。
* 相同的请求下，模型收到相同的seed值，会生成类似的结果，但不保证完全一致。

:::
---

**sequential_image_generation** `string` `默认值 disabled`
> 仅 doubao\-seedream\-5.0\-lite/4.5/4.0 支持该参数 | [查看组图输出示例](https://www.volcengine.com/docs/82379/1824121?lang=zh#fc9f85e4)

控制是否关闭组图功能。
:::tip
组图：基于您输入的内容，生成的一组内容关联的图片。

:::
* `auto`：自动判断模式，模型会根据用户提供的提示词自主判断是否返回组图以及组图包含的图片数量。
* `disabled`：关闭组图功能，模型只会生成一张图。

---

**sequential_image_generation_options ** `object`
> 仅 doubao\-seedream\-5.0\-lite/4.5/4.0 支持该参数

组图功能的配置。仅当 **sequential_image_generation ** 为 `auto` 时生效。

属性

---

sequential_image_generation_options.**max_images **  ** ** `integer` `默认值 15`
指定本次请求，最多可生成的图片数量。

* 取值范围： [1, 15]

:::tip
实际可生成的图片数量，除受到 **max_images ** 影响外 **，** 还受到输入的参考图数量影响。**输入的参考图数量+最终生成的图片数量≤15张**。

:::

---

**tools==^new^==** ** **  `array of object`
> 仅 doubao\-seedream\-5.0\-lite 支持该参数

配置模型要调用的工具。

属性

---

tools.**type ** `string`  
指定使用的工具类型。

* `web_search`：联网搜索功能。

:::tip

* 开启联网搜索后，模型会根据用户的提示词自主判断是否搜索互联网内容（如商品、天气等），提升生成图片的时效性，但也会增加一定的时延。
* 实际搜索次数可通过字段 usage.tool_usage.**web_search** 查询，如果为 0 表示未搜索。

:::

---

**stream**  `Boolean` `默认值 false`
> 仅 doubao\-seedream\-5.0\-lite/4.5/4.0 支持该参数 | [查看流式输出示例](https://www.volcengine.com/docs/82379/1824121?lang=zh#e5bef0d7)

控制是否开启流式输出模式。

* `false`：非流式输出模式，等待所有图片全部生成结束后再一次性返回所有信息。
* `true`：流式输出模式，即时返回每张图片输出的结果。在生成单图和组图的场景下，流式输出模式均生效。

---

**guidance_scale **  `Float` 
> doubao\-seedream\-3.0 默认值 2.5
> doubao\-seedream\-5.0\-lite/4.5/4.0 不支持

模型输出结果与prompt的一致程度，生成图像的自由度，又称为文本权重；值越大，模型自由度越小，与用户输入的提示词相关性越强。
取值范围：[`1`, `10`] 。

---

**output_format==^new^==**`string` `默认值 jpeg`
> 仅 doubao\-seedream\-5.0\-lite 支持该参数

指定生成图像的文件格式。可选值：

* `png`
* `jpeg`

:::tip
doubao\-seedream\-4.5/4.0、doubao\-seedream\-3.0 模型生成图像的文件格式默认为 jpeg，不支持自定义设置。

:::
---

**response_format** `string` `默认值 url`
指定生成图像的返回格式。支持以下两种返回方式：

* `url`：返回图片下载链接；**链接在图片生成后24小时内有效，请及时下载图片。** 
* `b64_json`：以 Base64 编码字符串的 JSON 格式返回图像数据。

---

**watermark**  `Boolean` `默认值 true`
是否在生成的图片中添加水印。

* `false`：不添加水印。
* `true`：在图片右下角添加“AI生成”字样的水印标识。

---

**optimize_prompt_options ** `object` 
> 仅 doubao\-seedream\-5.0\-lite/4.5/4.0 支持该参数

提示词优化功能的配置。

属性
optimize_prompt_options.**mode ** `string`  `默认值 standard`
设置提示词优化功能使用的模式。

* `standard`：标准模式，生成内容的质量更高，耗时较长。
* `fast`：快速模式，生成内容的耗时更短，质量一般；doubao\-seedream\-5.0\-lite/4.5 当前不支持。

---

&nbsp;
<span id="7P96iLnc"></span>
### 响应参数
<span id="Hrya4y9k"></span>
#### 流式响应参数
请参见[文档](https://www.volcengine.com/docs/82379/1824137?lang=zh)。
&nbsp;
<span id="1AxnwQZN"></span>
#### 非流式响应参数

---

**model** `string`
本次请求使用的模型 ID （`模型名称-版本`）。

---

**created** `integer`
本次请求创建时间的 Unix 时间戳（秒）。

---

**data** `array`
输出图像的信息。
:::tip
doubao\-seedream\-5.0\-lite/4.5/4.0 模型生成组图场景下，组图生成过程中某张图生成失败时：

* 若失败原因为审核不通过：仍会继续请求下一个图片生成任务，即不影响同请求内其他图片的生成流程。
* 若失败原因为内部服务异常（500）：不会继续请求下一个图片生成任务。

:::
可能类型
图片信息 `object`
生成成功的图片信息。

属性
data.**url ** `string`
图片的 url 信息，当 **response_format ** 指定为 `url` 时返回。该链接将在生成后 **24 小时内失效**，请务必及时保存图像。
推荐配置火山引擎 TOS 提供的数据订阅功能，将您的模型推理产物自动转存到自己的 TOS 桶中，便于长期备份或二次加工。详细介绍请参见 [TOS 数据订阅](https://www.volcengine.com/docs/6349/2280949?lang=zh)。

---

data.**b64_json** `string`
图片的 base64 信息，当 **response_format ** 指定为 `b64_json` 时返回。

---

data.**size** `string`
> 仅 doubao\-seedream\-5.0\-lite/4.5/4.0 支持该字段。

图像的宽高像素值，格式 `<宽像素>x<高像素>`，如`2048×2048`。

---

错误信息 `object`
某张图片生成失败，错误信息。

属性
data.**error** `object`
错误信息结构体。

属性

---

data.error.**code**
某张图片生成错误的错误码，请参见[错误码](https://www.volcengine.com/docs/82379/1299023)。

---

data.error.**message**
某张图片生成错误的提示信息。

---

**tools**  `array of object` 
本次请求，配置的模型调用工具

属性

---

tools.**type ** `string` 
配置的调用工具类型。

* web_search：联网搜索工具。

---

**usage** `object`
本次请求的用量信息。

属性

---

usage.**generated_images ** `integer`
模型成功生成的图片张数，不包含生成失败的图片。
仅对成功生成图片按张数进行计费。

---

usage.**output_tokens** `integer`
模型生成的图片花费的 token 数量。
计算逻辑为：计算 `sum(图片长*图片宽)/256` ，然后取整。

---

usage.**total_tokens** `integer`
本次请求消耗的总 token 数量。
当前不计算输入 token，故与 **output_tokens** 值一致。

---

usage.**tool_usage ** `object`
使用工具的用量信息。

属性

---

usage.tool_usage.**web_search ** `integer`
调用联网搜索工具次数，仅开启联网搜索时返回。

---

**error**  `object`
本次请求，如发生错误，对应的错误信息。 

属性

---

error.**code** `string` 
请参见[错误码](https://www.volcengine.com/docs/82379/1299023)。

---

error.**message** `string`
错误提示信息

&nbsp;

### POST https://www.cii-group.com/app-api/v1/images/edits

```http
POST https://www.cii-group.com/app-api/v1/images/edits
```

本文介绍 doubao-image-2-flatfee、doubao-image-2-flatfee-2k、doubao-image-2-flatfee-4k编辑图片 API 的输入输出参数，供您使用接口时查阅字段含义。通过传入原始图片及编辑提示词，模型将对图片进行编辑处理。

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### Content-Type

本接口使用 `multipart/form-data` 格式提交请求。

---

#### 请求体

---

**model** `string` %%require%%
调用的模型名称。如 `doubao-image-2-flatfee`。

---

**image** `binary` %%require%%
需要编辑的原始图片文件。

---

**prompt** `string` %%require%%
编辑图片的提示词，描述期望的编辑效果。

---

**n** `string` %%require%%
生成图片的数量。必须为整数，介于 1 和 10 之间。

---

**size** `string` %%require%%
生成图片的尺寸，如 `1024x1024`。

---

#### 请求体示例

```
model: doubao-image-2-flatfee
image: <图片文件二进制数据>
prompt: 将图片换成黑色背景
n: 1
size: 1024x1024
```

---

### 响应参数

**created** `integer`
请求创建时间的 Unix 时间戳。

---

**data** `array`
生成的图片信息数组。

属性

---

data[].**revised_prompt** `string`
模型优化后的提示词。

---

data[].**url** `string`
生成图片的 URL 地址。

---

#### 成功响应示例

```json
{
  "created": 1777010901,
  "data": [
    {
      "revised_prompt": "将图片换成黑色背景",
      "url": "https://pro.filesystem.site/cdn/20260424/8615b17125c841bf7eac69d77369b9.png"
    }
  ]
}
```

---

#### 错误响应

| HTTP 状态码 | 说明 |
|---|---|
| 400 | 请求参数错误 |
| 401 | API Key 无效或已禁用 |
| 403 | 权限不足或操作被拒绝 |
| 500 | 服务内部错误 |


---

## Gemini 转发 (3rd party)

### POST https://www.cii-group.com/app-api/3rd/v1beta/models/{model}:generateContent

```http
POST https://www.cii-group.com/app-api/3rd/v1beta/models/{model}:generateContent
```

本文介绍图生图模型 API 的输入输出参数，供您使用接口时查阅字段含义。通过传入参考图片及文本提示词，模型将基于参考图片生成新的图片。
doubao-3.1-flash-image-preview、doubao-3.1-flash-image-preview-2k、doubao-3.1-flash-image-preview-4k、doubao-3.5-flash、doubao-banana-2、doubao-banana-2-2k、doubao-banana-2-4k、doubao-banana-pro、doubao-banana-pro-2k、doubao-banana-pro-4k

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 请求体

---

**contents** `array` %%require%%
输入给模型的内容数组，包含参考图片和文本提示词信息。

属性

---

contents[].**parts** `array` %%require%%
内容片段数组，可包含图片数据和文本提示词。

属性

---

contents[].parts[].**inline_data** `object`
输入的参考图片信息。

属性

---

contents[].parts[].inline_data.**mime_type** `string` %%require%%
图片的 MIME 类型，常见取值：

* `image/png`
* `image/jpeg`
* `image/webp`

---

contents[].parts[].inline_data.**data** `string` %%require%%
图片文件的 Base64 编码内容（必须是纯 Base64）。

---

contents[].parts[].**text** `string`
输入生成的内容描述（文本提示词）。

---

**generationConfig** `object` %%require%%
生成配置信息。

属性

---

generationConfig.**imageConfig** `object` %%require%%
图片生成配置。

属性

---

generationConfig.imageConfig.**aspectRatio** `string` %%require%%
生成图片的纵横比例。可选值：

* `"1:1"`
* `"1:4"`
* `"1:8"`
* `"2:3"`
* `"3:2"`
* `"3:4"`
* `"4:1"`
* `"4:3"`
* `"4:5"`
* `"5:4"`
* `"8:1"`
* `"9:16"`
* `"16:9"`
* `"21:9"`

---

generationConfig.imageConfig.**imageSize** `string` %%require%%
生成图片的分辨率。可选值：

* `"512px"`
* `"1K"`
* `"2K"`
* `"4K"`

---

#### 请求体示例

```json
{
    "contents": [
      {
        "parts": [
          {
            "inline_data": {
              "mime_type": "image/png",
              "data": "<Base64编码的参考图片数据>"
            }
          },
          {
            "text": "将这张图片转换为水彩画风格"
          }
        ]
      }
    ],
    "generationConfig": {
      "imageConfig": {
        "aspectRatio": "16:9",
        "imageSize": "2K"
      }
    }
}
```

---

### 响应参数

**candidates** `array`
模型生成的图片结果数组。

属性

---

candidates[].**content** `object`
生成的内容信息。

属性

---

candidates[].content.**parts** `array`
内容片段数组，包含生成的图片数据。

属性

---

candidates[].content.parts[].**inlineData** `object`
生成的图片数据。

属性

---

candidates[].content.parts[].inlineData.**mimeType** `string`
图片的 MIME 类型，如 `image/png`。

---

candidates[].content.parts[].inlineData.**data** `string`
图片的 Base64 编码数据。

---

candidates[].content.parts[].**thoughtSignature** `string`
思考签名信息。

---

candidates[].content.**role** `string`
角色标识，如 `model`。

---

candidates[].**finishReason** `string`
结束原因，如 `STOP`。

---

candidates[].**index** `integer`
结果索引。

---

**usageMetadata** `object`
本次请求的用量信息。

属性

---

usageMetadata.**promptTokenCount** `integer`
输入提示词的 Token 数量。

---

usageMetadata.**candidatesTokenCount** `integer`
生成内容的 Token 数量。

---

usageMetadata.**totalTokenCount** `integer`
消耗的总 Token 数量。

---

usageMetadata.**promptTokensDetails** `array`
输入 Token 明细。

属性

---

usageMetadata.promptTokensDetails[].**modality** `string`
模态类型，如 `TEXT`。

---

usageMetadata.promptTokensDetails[].**tokenCount** `integer`
该模态的 Token 数量。

---

usageMetadata.**candidatesTokensDetails** `array`
输出 Token 明细。

属性

---

usageMetadata.candidatesTokensDetails[].**modality** `string`
模态类型，如 `IMAGE`。

---

usageMetadata.candidatesTokensDetails[].**tokenCount** `integer`
该模态的 Token 数量。

---

**modelVersion** `string`
本次请求使用的模型版本号。

---

**responseId** `string`
响应的唯一标识 ID。

---

#### 成功响应示例

```json
{
  "candidates": [
    {
      "content": {
        "parts": [
          {
            "inlineData": {
              "mimeType": "image/png",
              "data": "<Base64编码图片数据>"
            },
            "thoughtSignature": "aaaaaaaaaa"
          }
        ],
        "role": "model"
      },
      "finishReason": "STOP",
      "index": 0
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 17,
    "candidatesTokenCount": 1551,
    "totalTokenCount": 1568,
    "promptTokensDetails": [
      {
        "modality": "TEXT",
        "tokenCount": 17
      }
    ],
    "candidatesTokensDetails": [
      {
        "modality": "IMAGE",
        "tokenCount": 1120
      }
    ]
  },
  "responseId": "lkWiaeiSLdumjMcP4vfT4Q4"
}
```

### POST https://www.cii-group.com/app-api/3rd/v1beta/models/{model}:generateContent

```http
POST https://www.cii-group.com/app-api/3rd/v1beta/models/{model}:generateContent
```

本文介绍文生图模型 API 的输入输出参数，供您使用接口时查阅字段含义。通过传入文本提示词，模型将生成对应的图片。
doubao-3.1-flash-image-preview、doubao-3.1-flash-image-preview-2k、doubao-3.1-flash-image-preview-4k、doubao-3.5-flash、doubao-banana-2、doubao-banana-2-2k、doubao-banana-2-4k、doubao-banana-pro、doubao-banana-pro-2k、doubao-banana-pro-4k
**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 请求体

---

**contents** `array` %%require%%
输入给模型的内容数组，包含文本提示词信息。

属性

---

contents[].**parts** `array` %%require%%
内容片段数组。

属性

---

contents[].parts[].**text** `string` %%require%%
输入想要生成的内容描述（文本提示词）。

---

**generationConfig** `object` %%require%%
生成配置信息。

属性

---

generationConfig.**imageConfig** `object` %%require%%
图片生成配置。

属性

---

generationConfig.imageConfig.**aspectRatio** `string` %%require%%
生成图片的纵横比例。可选值：

* `"1:1"`
* `"1:4"`
* `"1:8"`
* `"2:3"`
* `"3:2"`
* `"3:4"`
* `"4:1"`
* `"4:3"`
* `"4:5"`
* `"5:4"`
* `"8:1"`
* `"9:16"`
* `"16:9"`
* `"21:9"`

---

generationConfig.imageConfig.**imageSize** `string` %%require%%
生成图片的分辨率。可选值：

* `"512px"`
* `"1K"`
* `"2K"`
* `"4K"`

---

#### 请求体示例

```json
{
    "contents": [{
      "parts": [
        {"text": "Create a picture of a nano banana dish in a fancy restaurant with a Gemini theme"}
      ]
    }],
    "generationConfig": {
      "imageConfig": {
        "aspectRatio": "16:9",
        "imageSize": "1K"
      }
    }
}
```

---

### 响应参数

**candidates** `array`
模型生成的图片结果数组。

属性

---

candidates[].**content** `object`
生成的内容信息。

属性

---

candidates[].content.**parts** `array`
内容片段数组，包含生成的图片数据。

属性

---

candidates[].content.parts[].**inlineData** `object`
生成的图片数据。

属性

---

candidates[].content.parts[].inlineData.**mimeType** `string`
图片的 MIME 类型，如 `image/png`。

---

candidates[].content.parts[].inlineData.**data** `string`
图片的 Base64 编码数据。

---

candidates[].content.parts[].**thoughtSignature** `string`
思考签名信息。

---

candidates[].content.**role** `string`
角色标识，如 `model`。

---

candidates[].**finishReason** `string`
结束原因，如 `STOP`。

---

candidates[].**index** `integer`
结果索引。

---

**usageMetadata** `object`
本次请求的用量信息。

属性

---

usageMetadata.**promptTokenCount** `integer`
输入提示词的 Token 数量。

---

usageMetadata.**candidatesTokenCount** `integer`
生成内容的 Token 数量。

---

usageMetadata.**totalTokenCount** `integer`
消耗的总 Token 数量。

---

usageMetadata.**promptTokensDetails** `array`
输入 Token 明细。

属性

---

usageMetadata.promptTokensDetails[].**modality** `string`
模态类型，如 `TEXT`。

---

usageMetadata.promptTokensDetails[].**tokenCount** `integer`
该模态的 Token 数量。

---

usageMetadata.**candidatesTokensDetails** `array`
输出 Token 明细。

属性

---

usageMetadata.candidatesTokensDetails[].**modality** `string`
模态类型，如 `IMAGE`。

---

usageMetadata.candidatesTokensDetails[].**tokenCount** `integer`
该模态的 Token 数量。

---

**modelVersion** `string`
本次请求使用的模型版本号。

---

**responseId** `string`
响应的唯一标识 ID。

---

#### 成功响应示例

```json
{
  "candidates": [
    {
      "content": {
        "parts": [
          {
            "inlineData": {
              "mimeType": "image/png",
              "data": "<Base64编码图片数据>"
            },
            "thoughtSignature": "aaaaaaaaaa"
          }
        ],
        "role": "model"
      },
      "finishReason": "STOP",
      "index": 0
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 17,
    "candidatesTokenCount": 1551,
    "totalTokenCount": 1568,
    "promptTokensDetails": [
      {
        "modality": "TEXT",
        "tokenCount": 17
      }
    ],
    "candidatesTokensDetails": [
      {
        "modality": "IMAGE",
        "tokenCount": 1120
      }
    ]
  },

  "responseId": "lkWiaeiSLdumjMcP4vfT4Q4"
}
```


---

## 素材组 (Asset Groups)

### POST https://www.cii-group.com/app-api/api/v1/asset-groups

```http
POST https://www.cii-group.com/app-api/api/v1/asset-groups
```

本文介绍创建素材组 API 的输入输出参数，供您使用接口时查阅字段含义。调用本接口可在当前用户下创建一个新的素材组，用于管理图片、视频等素材资源。

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 请求体

---

**Name** `string` **必填**
素材组名称。

---

**Description** `string` `可选`
素材组描述信息。不传时默认为空字符串。

---

#### 请求体示例

```json
{
  "Name": "我的素材组",
  "Description": "用于存放视频生成的参考图片素材"
}
```

---

<span id="响应参数"></span>
### 响应参数

成功时返回 HTTP 200，响应体为火山引擎素材管理 API 返回的素材组信息（JSON 格式透传）。

#### 成功响应示例

```json
{
  "AssetGroupId": "ag-xxxxxxxx",
  "Name": "我的素材组",
  "Description": "用于存放视频生成的参考图片素材",
  "CreatedAt": "2026-04-14T10:00:00Z",
  "UpdatedAt": "2026-04-14T10:00:00Z"
}
```

---

#### 错误响应

错误时返回统一格式：

**error.code** `integer`
HTTP 状态码。

**error.message** `string`
错误描述信息。

---

| HTTP 状态码 | 说明 |
|---|---|
| 400 | 请求参数错误，如 Name 为空 |
| 401 | API Key 无效或已禁用 |
| 403 | 权限不足或操作被拒绝 |
| 500 | 服务内部错误 |

#### 错误响应示例

```json
{
  "error": {
    "code": 400,
    "message": "Name is required"
  }
}
```

```json
{
  "error": {
    "code": 401,
    "message": "API Key 无效或已禁用"
  }
}
```

### PUT https://www.cii-group.com/app-api/api/v1/asset-groups/{groupId}

```http
PUT https://www.cii-group.com/app-api/api/v1/asset-groups/{groupId}
```

本文介绍更新素材组 API 的输入输出参数，供您使用接口时查阅字段含义。调用本接口可修改指定素材组的名称和描述信息。

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 路径参数

---

**groupId** `string` **必填**
需要更新的素材组 ID。

---

#### 请求体

---

**Name** `string` `可选`
新的素材组名称。

---

**Description** `string` `可选`
新的素材组描述信息。

---

#### 请求体示例

```json
{
  "Name": "新名称",
  "Description": "新描述"
}
```

---

#### 请求示例

```
PUT /api/v1/asset-groups/ag-xxxxxxxx
Authorization: Bearer <your-api-key>
Content-Type: application/json

{
  "Name": "新名称",
  "Description": "新描述"
}
```

---

<span id="响应参数"></span>
### 响应参数

成功时返回 HTTP 200，响应体为火山引擎素材管理 API 返回的更新结果（JSON 格式透传）。

#### 成功响应示例

```json
{
  "AssetGroupId": "ag-xxxxxxxx",
  "Name": "新名称",
  "Description": "新描述",
  "CreatedAt": "2026-04-14T10:00:00Z",
  "UpdatedAt": "2026-04-14T12:00:00Z"
}
```

---

#### 错误响应

| HTTP 状态码 | 说明 |
|---|---|
| 401 | API Key 无效或已禁用 |
| 403 | 权限不足或操作被拒绝 |
| 404 | 素材组不存在 |
| 500 | 服务内部错误 |

### GET https://www.cii-group.com/app-api/api/v1/asset-groups/{groupId}

```http
GET https://www.cii-group.com/app-api/api/v1/asset-groups/{groupId}
```

本文介绍查询素材组详情 API 的输入输出参数，供您使用接口时查阅字段含义。调用本接口可根据素材组 ID 查询该素材组的详细信息。

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 路径参数

---

**groupId** `string` **必填**
素材组 ID。

---

#### 请求示例

```
GET /api/v1/asset-groups/ag-xxxxxxxx
Authorization: Bearer <your-api-key>
```

---

<span id="响应参数"></span>
### 响应参数

成功时返回 HTTP 200，响应体为火山引擎素材管理 API 返回的素材组详情（JSON 格式透传）。

#### 成功响应示例

```json
{
  "AssetGroupId": "ag-xxxxxxxx",
  "Name": "我的素材组",
  "Description": "用于存放视频生成的参考图片素材",
  "CreatedAt": "2026-04-14T10:00:00Z",
  "UpdatedAt": "2026-04-14T10:00:00Z"
}
```

---

#### 错误响应

| HTTP 状态码 | 说明 |
|---|---|
| 401 | API Key 无效或已禁用 |
| 403 | 权限不足或操作被拒绝 |
| 404 | 素材组不存在 |
| 500 | 服务内部错误 |

#### 错误响应示例

```json
{
  "error": {
    "code": 404,
    "message": "素材组不存在"
  }
}
```

### POST https://www.cii-group.com/app-api/api/v1/assets/list

```http
POST https://www.cii-group.com/app-api/api/v1/assets/list
```

本文介绍列表查询素材 API 的输入输出参数，供您使用接口时查阅字段含义。调用本接口可分页查询指定素材组下的所有素材，支持按状态进行过滤。

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 请求体

---

**AssetGroupId** `string` **必填**
需要查询的素材组 ID。

---

**Filter** `object` `可选`
过滤条件。

属性

---

Filter.**Status** `string[]` `可选`
按素材状态过滤，支持传入多个状态值。

* `Active`：正常可用的素材
* 其他状态值以火山引擎素材管理 API 支持的状态为准

---

**PageNum** `integer` `可选` `默认值 1`
返回结果的页码。

---

**PageSize** `integer` `可选` `默认值 20`
每页返回的结果数量。

---

#### 请求体示例

```json
{
  "AssetGroupId": "ag-xxxxxxxx",
  "Filter": {
    "Status": ["Active"]
  },
  "PageNum": 1,
  "PageSize": 20
}
```

---

<span id="响应参数"></span>
### 响应参数

成功时返回 HTTP 200，响应体为火山引擎素材管理 API 返回的素材列表（JSON 格式透传）。

#### 成功响应示例

```json
{
  "Items": [
    {
      "AssetId": "ast-xxxxxxxx",
      "AssetGroupId": "ag-xxxxxxxx",
      "AssetType": "Image",
      "Name": "我的图片",
      "Status": "Active",
      "CreatedAt": "2026-04-14T10:00:00Z",
      "UpdatedAt": "2026-04-14T10:00:00Z"
    }
  ],
  "Total": 1,
  "PageNum": 1,
  "PageSize": 20
}
```

---

#### 错误响应

| HTTP 状态码 | 说明 |
|---|---|
| 400 | 请求参数错误，如 AssetGroupId 为空 |
| 401 | API Key 无效或已禁用 |
| 403 | 权限不足或操作被拒绝 |
| 404 | 素材组不存在 |
| 500 | 服务内部错误 |


---

## 素材 (Assets)

### POST https://www.cii-group.com/app-api/api/v1/assets

```http
POST https://www.cii-group.com/app-api/api/v1/assets
```

本文介绍创建素材 API 的输入输出参数，供您使用接口时查阅字段含义。调用本接口可在指定素材组中创建一个新的素材（图片/视频等），通过公网 URL 上传素材资源。

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 请求体

---

**AssetGroupId** `string` **必填**
素材所属的素材组 ID。

---

**ImageUrl** `string` **必填**
素材的公网 URL 地址。支持图片、视频等资源的公网可访问 URL。

---

**AssetType** `string` `可选` `默认值 "Image"`
素材类型。枚举值：

* `Image`：图片素材
* 其他类型以火山引擎素材管理 API 支持的类型为准

---

**Name** `string` `可选` `默认值 ""`
素材名称。不传时默认为空字符串。

---

#### 请求体示例

```json
{
  "AssetGroupId": "ag-xxxxxxxx",
  "AssetType": "Image",
  "Name": "我的图片",
  "ImageUrl": "https://example.com/image.jpg"
}
```

---

<span id="响应参数"></span>
### 响应参数

成功时返回 HTTP 200，响应体为火山引擎素材管理 API 返回的素材信息（JSON 格式透传）。

#### 成功响应示例

```json
{
  "AssetId": "ast-xxxxxxxx",
  "AssetGroupId": "ag-xxxxxxxx",
  "AssetType": "Image",
  "Name": "我的图片",
  "Status": "Active",
  "CreatedAt": "2026-04-14T10:00:00Z",
  "UpdatedAt": "2026-04-14T10:00:00Z"
}
```

---

#### 错误响应

| HTTP 状态码 | 说明 |
|---|---|
| 400 | 请求参数错误，如 AssetGroupId 或 ImageUrl 为空 |
| 401 | API Key 无效或已禁用 |
| 403 | 权限不足或操作被拒绝 |
| 404 | 素材组不存在 |
| 500 | 服务内部错误 |

#### 错误响应示例

```json
{
  "error": {
    "code": 400,
    "message": "AssetGroupId is required"
  }
}
```

```json
{
  "error": {
    "code": 400,
    "message": "ImageUrl is required"
  }
}
```

### PUT https://www.cii-group.com/app-api/api/v1/assets/{assetId}

```http
PUT https://www.cii-group.com/app-api/api/v1/assets/{assetId}
```

本文介绍更新素材 API 的输入输出参数，供您使用接口时查阅字段含义。调用本接口可修改指定素材的名称。

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 路径参数

---

**assetId** `string` **必填**
需要更新的素材 ID。

---

#### 请求体

---

**Name** `string` `可选`
新的素材名称。

---

#### 请求体示例

```json
{
  "Name": "新名称"
}
```

---

#### 请求示例

```
PUT /api/v1/assets/ast-xxxxxxxx
Authorization: Bearer <your-api-key>
Content-Type: application/json

{
  "Name": "新名称"
}
```

---

<span id="响应参数"></span>
### 响应参数

成功时返回 HTTP 200，响应体为火山引擎素材管理 API 返回的更新结果（JSON 格式透传）。

#### 成功响应示例

```json
{
  "AssetId": "ast-xxxxxxxx",
  "AssetGroupId": "ag-xxxxxxxx",
  "AssetType": "Image",
  "Name": "新名称",
  "Status": "Active",
  "CreatedAt": "2026-04-14T10:00:00Z",
  "UpdatedAt": "2026-04-14T12:00:00Z"
}
```

---

#### 错误响应

| HTTP 状态码 | 说明 |
|---|---|
| 401 | API Key 无效或已禁用 |
| 403 | 权限不足或操作被拒绝 |
| 404 | 素材不存在 |
| 500 | 服务内部错误 |

### GET https://www.cii-group.com/app-api/api/v1/assets/{assetId}

```http
GET https://www.cii-group.com/app-api/api/v1/assets/{assetId}
```

本文介绍查询素材详情 API 的输入输出参数，供您使用接口时查阅字段含义。调用本接口可根据素材 ID 查询该素材的详细信息。

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 路径参数

---

**assetId** `string` **必填**
素材 ID。

---

#### 请求示例

```
GET /api/v1/assets/ast-xxxxxxxx
Authorization: Bearer <your-api-key>
```

---

<span id="响应参数"></span>
### 响应参数

成功时返回 HTTP 200，响应体为火山引擎素材管理 API 返回的素材详情（JSON 格式透传）。

#### 成功响应示例

```json
{
  "AssetId": "ast-xxxxxxxx",
  "AssetGroupId": "ag-xxxxxxxx",
  "AssetType": "Image",
  "Name": "我的图片",
  "Status": "Active",
  "CreatedAt": "2026-04-14T10:00:00Z",
  "UpdatedAt": "2026-04-14T10:00:00Z"
}
```

---

#### 错误响应

| HTTP 状态码 | 说明 |
|---|---|
| 401 | API Key 无效或已禁用 |
| 403 | 权限不足或操作被拒绝 |
| 404 | 素材不存在 |
| 500 | 服务内部错误 |

#### 错误响应示例

```json
{
  "error": {
    "code": 404,
    "message": "素材不存在"
  }
}
```

### POST https://www.cii-group.com/app-api/api/v1/assets/list

```http
POST https://www.cii-group.com/app-api/api/v1/assets/list
```

本文介绍列表查询素材 API 的输入输出参数，供您使用接口时查阅字段含义。调用本接口可分页查询指定素材组下的所有素材，支持按状态进行过滤。

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 请求体

---

**AssetGroupId** `string` **必填**
需要查询的素材组 ID。

---

**Filter** `object` `可选`
过滤条件。

属性

---

Filter.**Status** `string[]` `可选`
按素材状态过滤，支持传入多个状态值。

* `Active`：正常可用的素材
* 其他状态值以火山引擎素材管理 API 支持的状态为准

---

**PageNum** `integer` `可选` `默认值 1`
返回结果的页码。

---

**PageSize** `integer` `可选` `默认值 20`
每页返回的结果数量。

---

#### 请求体示例

```json
{
  "AssetGroupId": "ag-xxxxxxxx",
  "Filter": {
    "Status": ["Active"]
  },
  "PageNum": 1,
  "PageSize": 20
}
```

---

<span id="响应参数"></span>
### 响应参数

成功时返回 HTTP 200，响应体为火山引擎素材管理 API 返回的素材列表（JSON 格式透传）。

#### 成功响应示例

```json
{
  "Items": [
    {
      "AssetId": "ast-xxxxxxxx",
      "AssetGroupId": "ag-xxxxxxxx",
      "AssetType": "Image",
      "Name": "我的图片",
      "Status": "Active",
      "CreatedAt": "2026-04-14T10:00:00Z",
      "UpdatedAt": "2026-04-14T10:00:00Z"
    }
  ],
  "Total": 1,
  "PageNum": 1,
  "PageSize": 20
}
```

---

#### 错误响应

| HTTP 状态码 | 说明 |
|---|---|
| 400 | 请求参数错误，如 AssetGroupId 为空 |
| 401 | API Key 无效或已禁用 |
| 403 | 权限不足或操作被拒绝 |
| 404 | 素材组不存在 |
| 500 | 服务内部错误 |


---

## 真人审核 (Visual Validate)

### POST https://www.cii-group.com/app-api/api/v1/visual-validate/sessions

```http
POST https://www.cii-group.com/app-api/api/v1/visual-validate/sessions
```

拉起端上 H5 真人认证页面链接。终端客户使用 H5Link 完成真人认证，点击"完成"按钮后，将打开 CallbackURL 链接，您可通过解析 CallbackURL 地址后拼接的 resultCode 参数获取真人认证结果。

终端客户通过真人认证（**resultCode** 为 **10000**）后，您可使用 API 返回的 BytedToken 查询该终端客户对应的 Asset Group ID。

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 请求体

---

**CallbackURL** `string` **必填**
用于认证结束后跳转的可访问 URL。

---

**ProjectName** `string` `可选` `默认值 "default"`
资源所属的项目名称（大小写敏感）。若资源不在默认项目中，需填写正确的项目名称。

---

#### 请求体示例

```json
{
  "CallbackURL": "https://www.example.com/callback"
}
```

---

### 响应参数

---

**BytedToken** `string`
本次认证的唯一凭证标识，用来在 GetVisualValidateResult 获取本次创建的 Group ID。

:::warning
byted_token 有效期为 30 分钟，请及时使用（仅支持认证一次，禁止重复认证）。
:::

---

**H5Link** `string`
真人认证 H5 页面链接，链接在使用后失效，再次检测需重新调用本接口生成链接。

:::tip
可通过 H5Link 链接后缀的 **lng** 字段指定页面语言。目前支持简体中文 `zh`、英文 `en` 和繁体 `zh-Hant`，默认值为 `zh`。
:::

---

**CallbackURL** `string`
用于认证结束后跳转的公网可访问 URL（原样回传）。

---

#### 回调说明

终端客户使用 H5Link 完成真人认证并点击"完成"按钮后，浏览器将打开 CallbackURL 链接，并拼接以下查询参数：

```
<CallbackURL>?bytedToken=...&resultCode=10000&algorithmBaseRespCode=0&reqMeasureInfoValue=1&verify_type=real_time
```

| 参数 | 说明 |
|---|---|
| bytedToken | 本次认证的唯一凭证标识，用于调用 GetVisualValidateResult 获取 Group ID |
| resultCode | **10000** 表示认证成功，其他值表示失败，详见错误码文档 |
| algorithmBaseRespCode | 服务端子错误码，建议 resultCode 为服务端错误码时再检查该字段 |
| reqMeasureInfoValue | 本次操作是否计费，0 为不计费，1 为计费。**当前真人认证相关服务限时免费** |
| verify_type | 认证类型，当前为固定值 `real_time` |

---

#### 成功响应示例

```json
{
  "BytedToken": "********22041234D89A4DDA08**************",
  "H5Link": "https://ark.volcengine.com/region:cn-beijing/mobile/livenees-face-manage/authorization?pl=******************",
  "CallbackURL": "https://www.example.com/callback"
}
```

---

#### 错误响应

| HTTP 状态码 | 说明 |
|---|---|
| 400 | 请求参数错误，如 CallbackURL 为空 |
| 401 | API Key 无效或已禁用 |
| 403 | 权限不足或操作被拒绝 |
| 500 | 服务内部错误 |

#### 错误响应示例

```json
{
  "error": {
    "code": 400,
    "message": "CallbackURL is required"
  }
}
```

### POST https://www.cii-group.com/app-api/api/v1/visual-validate/results

```http
POST https://www.cii-group.com/app-api/api/v1/visual-validate/results
```

真人认证通过后（即 CallbackURL 回调中 **resultCode** 参数值为 **10000** 时），可通过本接口获取本次真人认证创建的 Asset Group ID。

**鉴权方式**

本接口通过 `Authorization` 请求头进行 API Key 鉴权，格式为 `Bearer <API Key>`。请在平台获取有效的 API Key 后发起调用。

---

### 请求参数

#### 请求头

---

**Authorization** `string` **必填**
API Key 鉴权凭证，格式为 `Bearer <your-api-key>`。

---

#### 请求体

---

**BytedToken** `string` **必填**
本次认证的唯一凭证标识，从 CreateVisualValidateSession 的返回体中获取。

:::warning
byted_token 有效期为 30 分钟，请及时使用（仅支持认证一次，禁止重复认证）。
:::

---

**ProjectName** `string` `可选` `默认值 "default"`
资源所属的项目名称（大小写敏感）。若资源不在默认项目中，需填写正确的项目名称。

---

#### 请求体示例

```json
{
  "BytedToken": "202603311449168C23BA26**************"
}
```

---

### 响应参数

---

**GroupId** `string`
新创建的真人人像素材组 ID，格式示例：`group-20260331145705-*****`。

---

#### 成功响应示例

```json
{
  "GroupId": "group-20260331145705-*****"
}
```

---

#### 错误响应

| HTTP 状态码 | 说明 |
|---|---|
| 400 | 请求参数错误，如 BytedToken 为空 |
| 401 | API Key 无效或已禁用 |
| 403 | 权限不足或操作被拒绝 |
| 404 | 认证记录不存在或已过期 |
| 500 | 服务内部错误 |

#### 错误响应示例

```json
{
  "error": {
    "code": 400,
    "message": "BytedToken is required"
  }
}
```


---

## 错误码 (Error Codes)

### 本文列举您调用火山方舟 API 可能会涉及的错误码信息，包含方舟错误码和公共错误码。

```http
本文列举您调用火山方舟 API 可能会涉及的错误码信息，包含方舟错误码和公共错误码。
```

<span id=".5qih5Z6LLeaZuuiDveS9k-iwg-eUqC1hcGkt5YWs5YWx6ZSZ6K-v56CB"></span>
# 推理错误码

<span aceTableMode="list" aceTableWidth="2,2,3,4,4"></span>
|HTTP|错误类型|错误码|错误信息|含义 |\
|状态码 |Type |Code |Message | |
|---|---|---|---|---|
|400 |BadRequest |MissingParameter |The request failed because it is missing one or multiple required parameters. Request ID: |请求缺少必要参数，请查阅 API 文档。 |
|400 |BadRequest |InvalidParameter |One or more parameters specified in the request are not valid. Request ID: |请求包含非法参数，请查阅 API 文档。 |
|400 |BadRequest |InvalidParameter |The parameter `instructions` specified in the request are not valid: caching is not supported for instructions. Request id: |Responses API 中，当配置过 **instructions** 字段信息，后续轮次无法配置 **Caching** 字段。 |
|400 |BadRequest |InvalidEndpoint.ClosedEndpoint |The request targeted an endpoint that is currently closed or temporarily unavailable. Request ID: |推理接入点处于已被关闭或暂时不可用， 请稍后重试，或联系推理接入点管理员。 |
|400 |BadRequest |SensitiveContentDetected |The request failed because the input text may contain sensitive information. |输入文本可能包含敏感信息，请您使用其他 prompt。 |
|400 |BadRequest |SensitiveContentDetected.SevereViolation |The request failed because the input text may contain severe violation information. |输入文本可能包含严重违规相关信息，请您使用其他 prompt |
|400 |BadRequest |SensitiveContentDetected.Violence |The request failed because the input text may contain violence information. |输入文本可能包含激进行为相关信息，请您使用其他 prompt |
|400 |BadRequest |InputTextSensitiveContentDetected |The request failed because the input text may contain sensitive information.Request ID: |输入文本可能包含敏感信息，请您更换后重试。 |
|400 |BadRequest |InputImageSensitiveContentDetected |The request failed because the input image may contain sensitive information.Request ID: |输入图像可能包含敏感信息，请您更换后重试。 |
|400 |BadRequest |InputVideoSensitiveContentDetected |The request failed because the input video may contain sensitive information. |输入视频可能包含敏感信息，请您更换后重试。 |
|400 |BadRequest |InputAudioSensitiveContentDetected |The request failed because the input audio may contain sensitive information.Request ID: |输入音频可能包含敏感信息，请您更换后重试 |
|400 |BadRequest |OutputTextSensitiveContentDetected |The request failed because the output may contain sensitive information. |生成的文字可能包含敏感信息，请您更换输入内容后重试 |
|400 |BadRequest |OutputImageSensitiveContentDetected |The request failed because the output image may contain sensitive information. |生成的图像可能包含敏感信息，请您更换输入内容后重试。 |
|400 |BadRequest |OutputVideoSensitiveContentDetected |The request failed because the output video may contain sensitive information.Request ID: |生成的视频可能包含敏感信息，请您更换输入内容后重试。 |
|400 |BadRequest |OutputAudioSensitiveContentDetected |The request failed because the output audio may contain sensitive information.Request ID: |生成的音频可能包含敏感信息，请您更换输入内容后重试。 |
|400 |BadRequest |InputTextSensitiveContentDetected.PolicyViolation |The request failed because the input text may violate platform rules.Request ID: |输入文本可能违反平台规定，请您更换后重试。 |
|400 |BadRequest |InputImageSensitiveContentDetected.PolicyViolation |The request failed because the input image may violate platform rules.Request ID: |输入图片可能违反平台规定，请您更换后重试。 |
|400 |BadRequest |InputVideoSensitiveContentDetected.PolicyViolation |The request failed because the input video may violate platform rules.Request ID: |输入视频可能违反平台规定，请您更换后重试。 |
|400 |BadRequest |InputAudioSensitiveContentDetected.PolicyViolation |The request failed because the input audio may violate platform rules.Request ID: |输入音频可能违反平台规定，请您更换后重试。 |
|400 |BadRequest |InputImageSensitiveContentDetected.PrivacyInformation |The request failed because the input image may contain real person.Request ID: |输入图片可能包含真人，请您更换后重试。 |
|400 |BadRequest |InputVideoSensitiveContentDetected.PrivacyInformation |The request failed because the input video may contain real person.Request ID: |输入视频可能包含真人，请您更换后重试。 |
|400 |BadRequest |InputTextRiskDetection |The request could not be processed because the input text includes sensitive content that violates ContentSecurityDetection.ARKRequest ID:{id};CSDRequestId:{RequestId};Label:{Label};SubLabel: |火山引擎风险识别产品检测到输入文本可能包含敏感信息，请您更换后重试。 |
|400 |BadRequest |InputImageRiskDetection |The request could not be processed because the input image includes sensitive content that violates ContentSecurityDetection.ARKRequest ID:{id};CSDRequestId:{RequestId};Label:{Label};SubLabel: |火山引擎风险识别产品检测到输入图片可能包含敏感信息，请您更换后重试。 |
|400 |BadRequest |OutputTextRiskDetection |The request could not be processed because the output text includes sensitive content that violates ContentSecurityDetection.ARKRequest ID:{id};CSDRequestId:{RequestId};Label:{Label};SubLabel: |火山引擎风险识别产品检测到输出文本可能包含敏感信息，请您更换后重试。 |
|400 |BadRequest |OutputImageRiskDetection |The request could not be processed because the output image includes sensitive content that violates ContentSecurityDetection.ARKRequest ID:{id};CSDRequestId:{RequestId};Label:{Label};SubLabel: |火山引擎风险识别产品检测到输出图片可能包含敏感信息，请您更换后重试。 |
|400 |BadRequest |ContentSecurityDetectionError |Internal error.ARKRequest ID:{id};CSDRequestId:{RequestId};CSDcode:{};CSDmessage:{} |火山引擎风险识别产品请求失败。 |
|400 |BadRequest |InvalidParameter.{{Parameter}} |The specified parameter {{Parameter}} is invalid. |请求参数值不合法。请检查参数值的正确性后重试。 |
|400 |BadRequest |MissingParameter.{{Parameter}} |The required parameter {{Parameter}} is missing. |缺少必要的请求参数。请确认请求参数后重试。 |
|400 |BadRequest |Duplicate.Tags.Key |The specified object of tags contains duplicate keys. |对象的标签存在重复Key。 |
|400 |BadRequest |InvalidArgumentError |MissingRole：Invalid message: {{Message}} |请求中的 messages 列表里，有消息体缺少 role 字段 |
|400 |BadRequest |InvalidArgumentError.UnknownRole |Unknow the role of message: {{Role}} |消息体中的 role 值不被支持，如`user_`。 |
|400 |BadRequest |InvalidArgumentError.UnknownRole |The Inference role not found: {{Role}} |指定的 inference_role 未在配置中定义。 |
|400 |BadRequest |InvalidArgumentError.InvalidImageDetail |Invalid image detail: {{Parameter}} |image_url 中的 detail 参数值无效，只接受 "auto", "high", "low" |
|400 |BadRequest |InvalidArgumentError.InvalidPixelLimit |Customized min_pixels 100 is greater than max_pixels 50 |用户自定义的图片像素限制（min_pixels, max_pixels）无效（例如 min_pixels \> max_pixels，或超出了服务配置的范围） |
|400 |BadRequest |InvalidImageURL.EmptyURL |Empty base64 image url |传入的图片 URL 为空 |
|400 |BadRequest |InvalidImageURL.InvalidFormat |Invalid base64 image url |无法解析或处理图片，可能是 Base64 格式不正确、图片数据损坏或格式不支持 |
|400 |BadRequest |OutofContextError |Total tokens of image and text exceed max message tokens. |当请求中包含图片时，文本和图片编码后的总 token 数超过了模型上下文长度限制 |
|400 |BadRequest |InvalidParameter |The given Lean code is not compilable under Lean version %s |输入的不是一个合法的Lean Code。 |
|400 |BadRequest |InvalidParameter |The format of the given lean code is not supported so far. |输入的Lean code格式暂不支持。 |
|400 |BadRequest |InvalidParameter |/ |输入的Lean code必须包含theorem |
|400 |Forbidden |InvalidSubscription |Your account ({{account_identifier}}) does not have a valid coding plan subscription, or your subscription has expired. Please visit {{subscription_check_url}} to review your subscription status or complete the subscription or renewal process. |Coding Plan 套餐未订阅或已过期。 |
|401 |Unauthorized |AuthenticationError |The API key or AK/SK in the request is missing or invalid. Request ID: |请求携带的 API Key 或 AK/SK 校验未通过，请您重新检查设置的 鉴权凭证，或者查看 API 调用文档来排查问题。 |
|401 |Forbidden |InvalidAccountStatus |There is an issue with your account status. If you need assistance, please contact the platform administrators. |当前使用的账号异常。 |
|403 |Forbidden |OperationDenied.InvalidState |The specified context is in invalid state: InProgress.Request ID: |请求所关联的Context ID处于非空闲状态，不可调用。 |
|403 |Forbidden |OperationDenied.ConflictedValidationSet |Operation is denied because it is not supported to configure ValidationSet and ValidationPercentage at the same time. |无法同时上传验证集和设置训练集取样为验证集百分比，不支持该操作。 |
|403 |Forbidden |OperationDenied.PermissionDenied |Operation is denied because you are not permitted to access the specified configuration of the FoundationModel. |您没有权限访问基础模型的配置，不支持该操作。 |
|403 |Forbidden |OperationDenied.UnsupportedCustomizationType |Operation is denied because the specified CustomizationType is not supported by the CustomModel. |模型不支持该训练方法，不支持该操作。 |
|403 |Forbidden |OperationDenied.CustomizationNotSupported |Operation is denied because the specified version of the FoundationModel is not configured for the specified type of customization. |基础模型的版本不支持该训练方法，不支持该操作。 |
|403 |Forbidden |OperationDenied.ServiceNotOpen |Operation is denied because the model service is unavailable, please go to the Volcano Ark console activation management page to activate the corresponding model service, or submit a work order to contact us. |模型服务不可用，不支持该操作。请前往火山方舟控制台激活模型服务，或提交工单联系我们。 |
|403 |Forbidden |OperationDenied.ServiceOverdue |Operation is denied because your account balance is overdue, please go to the Volc Trading Center to recharge in order to continue using the service. |您的账单已逾期，不支持该操作。请前往火山费用中心充值。 |
|403 |Forbidden |AccountOverdueError |The request failed because your account has an overdue balance. Request ID: |当前账号欠费（余额<0），如需继续调用，请前往 [火山引擎费用中心](https://console.volcengine.com/finance/fund/recharge) 进行充值，详细操作参见 [充值操作指引](https://www.volcengine.com/docs/6269/100434)。 |
|403 |Forbidden |AccessDenied |The request failed because you do not have access to the requested resource. Request ID: |没有访问该资源的权限，请检查权限设置，或联系管理员添加白名单。 |
|403 |Forbidden |OperationDenied.InvalidState |Operation is denied because the specified context is in invalid state: InProgress. Request id: |请求的缓存信息状态是不可用状态。请查看缓存信息是否正在被更新中。 |
|403 |Forbidden |OperationDenied.UnsupportedPhase |Operation is denied because operation is not supported while the target is in its current phase. |操作失败，操作目标在特殊状态，请检查目标是否存在或者被锁定等特殊状态中。 |
|403 |Forbidden |OperationDenied.FileQuotaExceeded |Your account %s has exhausted its file storage quota. To continue using the service, please delete historical files. |当前账号 %s 已耗尽文件存储额度，如需继续使用，请删除历史文件。 |
|403 |Forbidden |OperationDenied.InvalidState |The specified file is in invalid state: InProgress.Request ID: |请求所关联的File ID处于非可用状态，不可调用。 |
|404 |NotFound |InvalidEndpointOrModel.NotFound |The model or endpoint %s does not exist or you do not have access to it. |模型或者推理接入点 %s 不存在或者您无权访问它。 |
|404 |NotFound |ModelNotOpen |Your account %s has not activated the model %s. Please activate the model service in the Ark Console. |当前账号 %s 暂未开通 %s 模型服务，请前往火山方舟控制台开通管理页开通对应模型服务。 |
|404 |NotFound |NotFound.{{Parameter}} |The specified {{ResourceType}} {{ResourceContent}} is not found. |指定资源找不到。请确认参数后重试。 |
|404 |NotFound |InvalidEndpointOrModel.ModelIDAccessDisabled |Accessing the model via Model ID is not allowed for your account. Please use a custom endpoint ID instead. Request id: |未能找到指定的模型ID。你的账号不允许使用模型ID来调用模型，请确认你账号权限或者使用有权限的推理接入点 ID 来调用模型服务。 |
|404 |NotFound |UnsupportedModel |The {{model_name}} model does not support the coding plan feature. Please refer to the documentation at {{doc_url}} to select a compatible model. |当前模型不支持 Coding Plan。 |
|429 |TooManyRequests |RateLimitExceeded.EndpointRPMExceeded |The Requests Per Minute (RPM) limit of the associated endpoint for your account has been exceeded. Request ID: |请求所关联的推理接入点已超过 RPM (Requests Per Minute) 限制, 请稍后重试。 |
|429 |TooManyRequests |RateLimitExceeded.EndpointTPMExceeded |The Tokens Per Minute (TPM) limit of the associated endpoint for your account has been exceeded. Request ID: |请求所关联的推理接入点已超过 TPM (Tokens Per Minute) 限制, 请稍后重试。 |
|429 |TooManyRequests |ModelAccountRpmRateLimitExceeded |RPM (Requests Per Minute) limit of the model is exceeded. Request ID: |请求已超过帐户模型 RPM (Requests Per Minute) 限制: 请您稍后重试, 或者联系平台技术同学进行解决 |
|429 |TooManyRequests |ModelAccountTpmRateLimitExceeded |TPM (Tokens Per Minute) limit of the model is exceeded. Request ID: |请求已超过帐户模型 TPM (Tokens Per Minute) 限制: 请您稍后重试, 或者联系平台技术同学进行解决 |
|429 |TooManyRequests |APIAccountRpmRateLimitExceeded |The RPM (Requests Per Minute) limit for the API on your account has been exceeded. Request ID: |当前账号该接口的RPM (Requests Per Minute)限制已超出，请稍后重试。 |
|429 |TooManyRequests |ModelAccountIpmRateLimitExceeded |IPM (Images Per Minute) limit of the model is exceeded. |请求已超过账户模型 IPM (Images Per Minute) 限制: 请您稍后重试, 或者联系平台技术同学进行解决 |
|429 |TooManyRequests |QuotaExceeded |Your account [%s] has exhausted its free trial quota for the [%s] model. Request ID: |当前账号 %s 对 %s 模型的免费试用额度已消耗完毕，如需继续调用，请前往火山方舟控制台开通管理页开通对应模型服务。 |
|429 |TooManyRequests |QuotaExceeded |The request has exceeded the quota. Request ID: |当前账号处于排队中状态的任务数已超过限制，请稍后重试。 |
|429 |TooManyRequests |ServerOverloaded |The service is currently unable to handle additional requests due to server overload. Please retry later. Request ID: |服务资源紧张，请您稍后重试。常出现在调用流量突增或刚开始调用长时间未使用的推理接入点。|\
| | | | |:::tip|\
| | | | |调用`doubao-seed-1.8`及之前版本模型触发突增流量限制时，返回此错误码。可参考[突发流量处理最佳实践](/docs/82379/1848593)处理。|\
| | | | ||\
| | | | |:::|
|429 |TooManyRequests |RequestBurstTooFast |System protection triggered by request burst. Please slow down traffic growth and increase requests gradually before retrying. |请求量激增触发系统保护，请放缓流量提升速度，逐步增加请求量后再尝试|\
| | | | |:::tip|\
| | | | |调用 `doubao-seed-2.0`及之后版本模型触发突增流量限制时，返回此错误码，可参考[突发流量处理最佳实践](/docs/82379/1848593)处理。|\
| | | | ||\
| | | | |:::|
|429 |TooManyRequests |SetLimitExceeded |Your account [%s] has reached the set inference limit for the [%s] model, and the model service has been paused. To continue using this model, please visit the Model Activation page to adjust or close the "Safe Experience Mode".|当前账号 %s 对 %s 模型已达到设置的推理限额值，如需继续调用，请前往火山方舟控制台开通管理页修改限额值或关闭安心体验模式。 |\
| | | |Request ID: | |
|429 |TooManyRequests |InflightBatchsizeExceeded |The Inflight Batchsize limit has been exceeded.Request ID: |您已经达到当前充值金额下的最大并发数限制，您可以充值解锁更大并发额度或降低并发数。 |
|429 |TooManyRequests |AccountRateLimitExceeded |Requests are too frequent. Please reduce your request frequency, wait a short moment, and retry your request. |请求超出RPM / TPM限制。 |
|429 |TooManyRequests |QuotaExceeded |You have exceeded the 5\-hour/weekly/monthly usage quota. It will reset at {{reset_time}}. |使用的额度超出5小时/周/月限额。 |
|500 |InternalServerError |InternalServiceError |The service encountered an unexpected internal error. Please retry later. Request ID: |内部系统异常，请您稍后重试。 |

<span id="9b1548df"></span>
# 精调错误码
本部分将介绍模型精调相关的错误码信息及其建议的解决方案。

<span aceTableMode="list" aceTableWidth="2,2,3"></span>
|错误码 |示例错误信息 |说明与建议解决方案 |
|---|---|---|
|InvalidData.MissingKey |Data format is not expected:column not found |数据格式不符合预期, 未找到名为的列, 建议检查并补全数据集中相应键值. |
|InvalidData.UnknownKey |Wrong Key, parsing sample failed |数据中有错误的Key, 解析样本失败, 建议检查错误信息中对应样本的键值 |
|InvalidData.InvalidValue |Unsupported data type:, only pretrain, dialog, dialog\-dpo and multimodal supported |不支持的数据集类型, 仅支持"pretrain", "dialog", "dialog\-dpo", 和"multimodal"类型, 建议检查填入的数据集类型 |
|^^|Content is empty, original text content is: |content字段内容为空, 原文本内容为, 建议检查原文本的数据完整性. |
|InvalidData.InvalidJsonl |not supported |数据集文件格式不支持, 建议调整对应文件为`.jsonl`格式. |
|^^|No jsonl file available |无可用的jsonl文件, 建议检查数据集包含的文件列表. |
|InvalidData.InvalidJson |Expecting value: line 1 column 1 (char 0) in fileatrow |在<文件名\>中第行, JSON解析失败, 建议检查数据集文件内容是否符合JSON规范. |
|InvalidData |failed to init data preprocess builder: training tos bucket: maas\-data\-test, tos path:|提供的数据集地址在TOS中不存在, 建议检查TOS中数据集文件是否出现缺失或地址错误. |\
| |: tos objects do not exist | |
|UnknownError |service error occur, please contact customer service for help |平台服务错误, 不可重试, 建议发起工单协助排查 |
|InternalError |task failed, please check the logs |训练失败, 建议检查日志后发起重试, 如仍无法解决问题, 建议发起工单协助排查. |

<span id="d674f4be"></span>
# 公共错误码
查询火山引擎的[公共错误码](https://www.volcengine.com/docs/6369/68677)。


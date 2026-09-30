# UI 风格规范（New API / Seedance）

> 统一前台与后台、品牌页与功能页的视觉语言。任何新增页面 / 组件上线前应自检本规范。

---

## 1. 设计目标与适用场景

- **品牌定位**：面向 OPC（一人公司）与中小团队的 AI API 网关，对外页面要显得「克制、可靠、面向开发者」；管理后台要显得「信息密度高、操作直接」。
- **设计原则**
  1. **一致性优先**：相同语义用同一份令牌；禁止在页面里写死 `bg-#xxx`。
  2. **层级清晰**：用字号、字重、颜色、阴影四件套构建层级，避免依赖纯色块。
  3. **克制留白**：卡片之间纵向 16/24/32px 三档节奏，不要堆 12 个像素+的小碎块。
  4. **小图标是亮点**：IconBadge 走 vibrant 色板，承担视觉锚点；正文与菜单保持中性色。
  5. **暗色模式默认开启**：所有色值必须有 dark 反义，深色版不只换背景。

---

## 2. 色彩系统

> 所有色值都已在 `src/styles/theme.css` 用 OKLCH 定义；Tailwind 直接消费 `--color-*` 变量。
> 不要再写裸 hex / rgb。

### 2.1 语义色（默认 Light）

| 角色 | Light | Dark | 用途 |
| --- | --- | --- | --- |
| `background` | `oklch(1 0 0)` | `oklch(0.145 0 0)` | 页面底 |
| `foreground` | `oklch(0.145 0 0)` | `oklch(0.985 0 0)` | 主文本 |
| `card` / `popover` | 白 | `oklch(0.205 0 0)` | 卡片表面 |
| `muted` | 浅灰 | 深灰 | 分隔、背景块 |
| `muted-foreground` | `oklch(0.49 0 0)` | `oklch(0.71 0 0)` | 副文本、说明 |
| `primary` | `oklch(0.692 0.141 243.716)` | `oklch(0.7 0.16 250)` | 主按钮、品牌强调 |
| `primary-foreground` | 白 | 白 | 主按钮文字 |
| `destructive` | 红 | 红 | 删除、错误 |
| `success` / `warning` / `info` | 见 theme.css | 同 | 状态色 |
| `border` / `ring` | `oklch(0.92 0 0)` | `oklch(1 0 0 / 10%)` | 边框、聚焦环 |

### 2.2 图表色板（chart-1 ~ chart-5）

| 名称 | Light 值 | 用途 |
| --- | --- | --- |
| `chart-1` | `oklch(0.72 0.18 250)` | 主蓝色，趋势/折线首选 |
| `chart-2` | `oklch(0.65 0.15 200)` | 青蓝 |
| `chart-3` | `oklch(0.7 0.12 280)` | 紫蓝 |
| `chart-4` | `oklch(0.68 0.19 325)` | 品红 |
| `chart-5` | `oklch(0.68 0.16 155)` | 翠绿 |

使用规则：饼图、环形图不超过 5 个扇区；同一系列内从 1 开始递增，不要跳号。

### 2.3 IconBadge vibrant 色板（**小图标专用**）

> 语义：每个 IconBadge 必须绑一种 `tone`，不允许「无色 IconBadge」。
> 写法见 `src/components/ui/icon-badge.tsx`，已在 Features / HowItWorks 落地。

| tone | Tailwind 类 | 语义建议 |
| --- | --- | --- |
| `vibrant-blue` | `bg-blue-500/15 text-blue-600 dark:text-blue-400` | 接口、连接、入口 |
| `vibrant-cyan` | `bg-cyan-500/15 text-cyan-600 dark:text-cyan-400` | 速率、传输 |
| `vibrant-emerald` | `bg-emerald-500/15 text-emerald-600 dark:text-emerald-400` | 成功、增长、节省 |
| `vibrant-amber` | `bg-amber-500/15 text-amber-600 dark:text-amber-400` | 成本、计费、提醒 |
| `vibrant-violet` | `bg-violet-500/15 text-violet-600 dark:text-violet-400` | 权限、安全、策略 |
| `vibrant-pink` | `bg-pink-500/15 text-pink-600 dark:text-pink-400` | 创作、内容、媒体 |
| `vibrant-orange` | `bg-orange-500/15 text-orange-600 dark:text-orange-400` | 警告、流量 |
| `vibrant-indigo` | `bg-indigo-500/15 text-indigo-600 dark:text-indigo-400` | 团队、组织、平台 |

**已固化映射**（见主页第 3、4 屏，不要再乱改）：
- 统一接口 → `vibrant-blue`
- 可控成本 → `vibrant-amber`
- 权限管控 → `vibrant-violet`
- 稳定路由 → `vibrant-emerald`
- 开发团队 → `vibrant-blue`
- 内容团队 → `vibrant-pink`
- 企业管理 → `vibrant-indigo`

### 2.4 渐变色（标题用）

```text
from-blue-600 via-indigo-500 to-violet-500
```
Dark 端：`from-blue-400 via-indigo-400 to-violet-400`。**仅用于 H1 / H2 标题文字**，禁止用于卡片底色。

---

## 3. 字体

- **基础**：`--font-sans: 'Public Sans Variable', sans-serif`（已通过 `@fontsource-variable/public-sans` 内联）。
- **衬线（标题氛围）**：`'Lora'` 变量字体，由 `theme-presets.css` 在特定主题下激活。
- **正文家族**：默认走 `--font-body`，由 `data-theme-font` 切换；不要在组件里硬写 `font-sans`。

### 3.1 字号 / 字重 / 行高尺度

| 用途 | 类 | 字号 | 字重 | 行高 | letter-spacing |
| --- | --- | --- | --- | --- | --- |
| 页面大标题 H1 | `text-[clamp(2rem,4vw,3rem)]` | 32–48 | `font-bold` | `leading-[1.1]` | `tracking-tight` |
| 区块标题 H2 | `text-[clamp(1.75rem,3vw,2.5rem)]` | 28–40 | `font-bold` | `leading-[1.18]` | `tracking-tight` |
| 卡片标题 H3 | `text-[17px]` | 17 | `font-semibold` | `leading-normal` | — |
| 段落正文 | `text-[15px]` | 15 | `font-normal` | `leading-relaxed` (1.625) | — |
| 辅助说明 | `text-sm` (14) | 14 | `font-normal` | `leading-relaxed` | — |
| 标签 / chip | `text-[11px]` | 11 | `font-medium` | `leading-none` | `tracking-[0.18em] uppercase` |
| 数字 / 价格 | `text-[18px] font-bold tabular-nums` | 18 | `font-bold` | `leading-none` | — |
| 菜单项 | `text-sm` | 14 | `font-medium` | `leading-none` | — |

### 3.2 标题颜色

- 浅色 H1/H2 → `text-slate-900 dark:text-slate-50`；或使用 §2.4 渐变文字。
- 副标题/段落 → `text-slate-500 dark:text-slate-400`。

---

## 4. 圆角与阴影

### 4.1 圆角阶梯（基于 `--radius: 1rem`）

| 名字 | 公式 | 实用值 |
| --- | --- | --- |
| `rounded-sm` | `radius * 0.6` | 6px |
| `rounded-md` | `radius * 0.8` | 8px |
| `rounded-lg` | `radius` | 12px |
| `rounded-xl` | `radius * 1.4` | 18px |
| `rounded-2xl` | `radius * 1.8` | 22px |
| `rounded-3xl` | `radius * 2.2` | 28px |

**用法**：
- 按钮 / 输入框：`rounded-md` ~ `rounded-lg`（8–12px）。
- 卡片：`rounded-xl` ~ `rounded-2xl`（18–22px）。
- 营销区块容器：`rounded-2xl` ~ `rounded-3xl`（22–28px）。

### 4.2 阴影阶梯

```text
card-rest   : 0 18px 40px -30px rgba(15,23,42,0.35)
card-hover  : 0 28px 60px -32px rgba(79,70,229,0.45)
hero-banner : 0 30px 70px -52px rgba(37,99,235,0.55)
focus-ring  : 0 0 0 3px var(--ring) / 50%
```

### 4.3 深度组合

- 浮层 / Dialog：阴影 `0 24px 60px -28px rgba(15,23,42,0.45)` + `ring-1 ring-slate-200/70`。
- 暗色模式用同公式但源颜色 `0 0 0`：`0 28px 60px -32px rgba(0,0,0,0.6)`。

---

## 5. 间距与容器

- **基础栅格**：4px。所有间距必须是 4 的倍数（4 / 8 / 12 / 16 / 24 / 32 / 48 / 64）。
- **页面横向内边距**：`<main>` `px-4 md:px-6`；区块容器 `max-w-[1400px] mx-auto`（首页大区块统一用这个值）。
- **卡片内边距**：基础 `p-6`；密集字段 `p-5`；强调卡 `p-7`。
- **段落间距**：正文段落 `mt-3`；段落集合 `space-y-4` 或 `space-y-6`。
- **区块节奏**：垂直 `py-12 md:py-16` 或 `md:py-18`；同一页里至少错落（12 / 16 / 20）。

---

## 6. 组件规范

### 6.1 标题（H2 标题样板）

```tsx
<span className="inline-flex items-center rounded-full border border-indigo-200/80 bg-white/70 px-3 py-1 text-[11px] font-medium tracking-[0.18em] text-indigo-600 uppercase backdrop-blur-sm dark:border-white/15 dark:bg-white/10 dark:text-indigo-300">
  {chip}
</span>
<h2 className="mt-5 text-[clamp(1.75rem,3vw,2.5rem)] leading-[1.18] font-bold tracking-tight">
  <span className="bg-gradient-to-r from-blue-600 via-indigo-500 to-violet-500 bg-clip-text text-transparent dark:from-blue-400 dark:via-indigo-400 dark:to-violet-400">
    {title}
  </span>
</h2>
<p className="mx-auto mt-4 max-w-2xl text-[15px] leading-relaxed text-slate-500 dark:text-slate-400">
  {desc}
</p>
```

规则：chip 文案**英文大写 + 中文短词**；标题允许两行，使用渐变文字。

### 6.2 正文

```tsx
<p className="text-[15px] leading-relaxed text-slate-500 dark:text-slate-400" />
<p className="text-sm leading-relaxed text-slate-500 dark:text-slate-400" />
```

不要混用 `text-base`（项目里 90% 段落都是 15px 或 14px）。

### 6.3 菜单（侧边栏 / 顶部导航）

- 容器：`bg-sidebar text-sidebar-foreground`。
- 项：高度 `h-9` ~ `h-10`，`px-3`，`rounded-md`，间距 `gap-0.5`。
- 激活态：左侧 2px 色条 `border-l-2 border-primary`，背景 `bg-sidebar-accent`。
- 悬停：背景 `bg-sidebar-accent/60`，前景 `text-sidebar-primary`。
- 图标：`<IconBadge tone="vibrant-blue" size="sm">` 或裸 `lucide` 图标 16px，**统一 16px**。

### 6.4 按钮

复用 `src/components/ui/button.tsx` 的 CVA，**不要重写 Button 组件**。

| 场景 | variant | size | 备注 |
| --- | --- | --- | --- |
| 主操作 / 注册 / 支付 | `default` | `default` 或 `lg` | 主蓝底白字 |
| 次要操作 / 取消 | `outline` | `default` | 白底深字，hover `bg-accent` |
| 危险操作（删除 / 重置） | `destructive` | `default` | 红底白字 |
| 链接式按钮 | `link` | — | 仅文字 + 下划线 |
| Hero / 首屏 CTA | 自定义渐变 `bg-gradient-to-r from-blue-600 to-blue-500` | `h-11` | 阴影 `0 8px 24px -6px rgba(37,99,235,0.55)` |
| 卡片内 inline 行动 | `ghost` | `sm` | `text-sm`，hover `bg-accent` |

规则：
- 按钮内图标 `size-4`，文字 `text-sm font-medium`。
- 主按钮 hover：`hover:-translate-y-0.5` + 阴影增强。
- 不要在按钮内塞多于 2 个图标。

### 6.5 IconBadge（**强约束**）

- **必须**指定 `tone`；不允许 `neutral` 默认值出现在生产组件（除 placeholder）。
- **尺寸对照**：

| 用途 | size | 容器 | svg |
| --- | --- | --- | --- |
| 表格行 / 表单标签前缀 | `xs` | `size-5 rounded-md` | `size-3` |
| 菜单项前缀 / 紧凑行 | `sm` | `size-7 rounded-md` | `size-3.5` |
| 卡片小图标（标准） | `md` | `size-8 rounded-lg` | `size-4` |
| 区块标题左侧 | `title` | `size-8 sm:size-9 rounded-lg` | `size-4` |
| 概览卡片大图标 | `lg` | `size-10 rounded-xl` | `size-5` |
| 指标卡数字行前缀 | `stat` | `size-5 sm:size-7 rounded-md` | `size-3 sm:size-3.5` |

- 区块大图标（Features / HowItWorks 卡片内 `size-12`）通过 `<span className="inline-flex size-12 items-center justify-center rounded-[14px] bg-xxx-500/15 text-xxx-600">` 自实现，**复用 §2.3 的色板与色阶**，不要发明新的色相。

### 6.6 卡片

- 表面：`bg-card border border-slate-200/70 dark:border-white/10 rounded-xl`。
- 阴影：`shadow-[0_18px_40px_-30px_rgba(15,23,42,0.35)]`；hover 加强。
- 玻璃（`backdrop-blur-sm`）：仅用在浮于背景图 / Hero 卡片上，**不要全文铺玻璃**。
- 悬浮位移：`motion-safe:hover:-translate-y-1`，位移幅度 ≤ 4px。

### 6.7 标签 / Tag

```tsx
<span className="rounded-full border border-blue-100 bg-blue-50 px-2.5 py-1 text-[11px] font-medium text-blue-700 dark:border-blue-400/25 dark:bg-blue-500/10 dark:text-blue-300" />
```

颜色与 §2.3 IconBadge 对齐（如权限类用 violet，成本类用 amber）。

### 6.8 数据展示

- 数字统一 `tabular-nums`，加粗 `font-bold`。
- 价格使用 `font-mono text-[18px] font-bold tabular-nums text-red-500 dark:text-red-400`（中文 A 股惯例：红涨绿跌）。
- 增长徽章：`text-emerald-600 dark:text-emerald-400`；下降徽章：`text-rose-600 dark:text-rose-400`。

---

## 7. 可访问性

- 颜色对比度：`text-foreground` vs `bg-background` ≥ 7:1；`text-muted-foreground` vs `bg-background` ≥ 4.5:1。
- 焦点环 `ring-3 ring-ring/50` 必须在所有可聚焦元素保留（已在 `button.tsx` base 里强制）。
- IconBadge 默认 `aria-hidden`，仅在带语义时改为 `decorative={false}` 并加 `aria-label`.
- 动效：`motion-safe:` 修饰符包裹位移 / 渐入动画，并遵守 `prefers-reduced-motion`。
- 字号响应式：正文最小 `14px`，不允许 `<12px` 的正文；`text-[11px]` 仅用于 chip / tag。

---

## 8. 深色模式

- 默认随系统 `prefers-color-scheme`；手动开关由 `ThemeSwitch` 控制，写入 `localStorage` 并触发 `class="dark"`。
- 渐变 / 透明背景必须给 dark 端一个等价的更深版本，例如 `bg-blue-500/15 dark:bg-blue-400/15`。
- 背景图（Features / HowItWorks 整面板）需要确保 dark 模式下文字仍可读——目前通过 `bg-slate-900/55` 兜底；后续若更换素材要重新验证对比度。

---

## 9. 多端适配

- **断点**：`< 640 sm` / `≥ 640 md` / `≥ 1024 lg` / `≥ 1280 xl` / `≥ 1536 2xl`。
- 手机端：
  - 主导航折叠为汉堡按钮（已实现于 PublicHeader）。
  - 卡片网格降为 1 列；区块标题字号自动 `clamp` 收缩。
  - 表单输入 `font-size: 16px`（已在 base 兜底，避免 iOS 自动放大）。
- 桌面端：
  - 区块最大宽度 `max-w-[1400px]`。
  - 侧边栏常驻；顶部 PublicHeader 在滚动后收敛为胶囊（已在 PublicHeader 实现）。

---

## 10. 落地与维护

- **已落地的组件**：`IconBadge`（src/components/ui/icon-badge.tsx）、`Button`（button.tsx）、`ButtonGroup`、`StatCard`（dashboard/ui/stat-card.tsx）、`AnimateInView`、`PublicHeader`、`Card`、`Tag`、`Dialog`。
- **新增组件三步走**：
  1. 复用 §2 / §3 / §4 令牌，禁止新增临时色 / 字号。
  2. 同时提交 Light / Dark 预览。
  3. PR 描述里附「涉及到的设计规范段落」。
- **变更本规范**：在 PR 里同步修改本文件，并在 `dev-docs/dev进度.md` 标注；不得直接在组件里绕过规范。

---

## 附录 A：IconBadge 速查

```tsx
import { IconBadge, type IconBadgeTone } from '@/components/ui/icon-badge'

<IconBadge tone="vibrant-blue" size="md"><KeyRound /></IconBadge>
<IconBadge tone="vibrant-amber" size="md"><Wallet /></IconBadge>
<IconBadge tone="vibrant-violet" size="md"><ShieldCheck /></IconBadge>
<IconBadge tone="vibrant-emerald" size="md"><Route /></IconBadge>
<IconBadge tone="vibrant-pink" size="md"><PenTool /></IconBadge>
<IconBadge tone="vibrant-indigo" size="md"><Building2 /></IconBadge>
```

## 附录 B：渐变标题与卡片背景样板

```tsx
// 渐变标题（用于 H1 / H2）
<span className="bg-gradient-to-r from-blue-600 via-indigo-500 to-violet-500 bg-clip-text text-transparent dark:from-blue-400 dark:via-indigo-400 dark:to-violet-400">
  {title}
</span>

// 区块容器（背景图整面板左半）
<div className="relative isolate mx-auto max-w-[1400px] overflow-hidden rounded-[24px] border border-indigo-100/80 shadow-[0_30px_70px_-52px_rgba(79,70,229,0.5)] md:rounded-[32px] dark:border-white/10">
  <img aria-hidden src="/home/hero-slide-X.jpg" alt=""
       className="absolute inset-0 -z-10 h-full w-full object-cover object-left" />
  {/* 可选深色兜底：<div aria-hidden className="absolute inset-0 -z-10 bg-slate-900/55" /> */}
  <div className="relative ...">{children}</div>
</div>
```
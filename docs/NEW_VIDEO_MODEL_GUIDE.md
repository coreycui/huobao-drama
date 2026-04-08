# 新增视频生成模型接入指南

## 概述

本文档介绍在 huobao-drama 中新增一个视频生成模型需要修改的所有位置。

## 需要修改的文件

### 1. 后端 - VolcesArkClient（如果使用火山引擎）

文件：`pkg/video/volces_ark_client.go`

如果新模型的 API 格式与现有模型不同，需要在 `isSeedance2` 检测逻辑中新增判断。

### 2. 前端 - AI 配置下拉

文件：`web/src/components/common/AIConfigDialog.vue`

在 `providerConfigs.video` 中找到对应厂商的配置，在 `models` 数组中添加新的模型 ID。

### 3. 前端 - 专业编辑器模型能力配置

文件：`web/src/views/drama/ProfessionalEditor.vue`

在 `defaultModelCapabilities` 对象中添加新模型的能力配置：

```typescript
"<model-id>": {
  supportSingleImage: boolean,    // 是否支持单图
  supportMultipleImages: boolean, // 是否支持多图
  supportFirstLastFrame: boolean, // 是否支持首尾帧
  supportTextOnly: boolean,       // 是否支持纯文本
  maxImages: number,              // 最多支持几张图
},
```

### 4. 后端 - 模型路由（如需）

文件：`application/services/video_generation_service.go`

在 `getVideoClient()` 方法中添加新的 provider 分支。

## 示例：添加 Seedance 2.0（2026-04-08）

### 模型信息

- 模型ID：`doubao-seedance-2-0-260128` / `doubao-seedance-2-0-fast-260128`
- 厂商：火山引擎
- 支持：单图 + 多图（最多9张）+ 纯文本
- API格式：v3（与 Seedance 1.x 不同）

### 修改记录

- ✅ `AIConfigDialog.vue`：添加模型到下拉列表（volces 和 chatfire 两个 provider）
- ✅ `ProfessionalEditor.vue`：添加模型能力配置（supportMultipleImages: true, maxImages: 9）
- ✅ `volces_ark_client.go`：已有 isSeedance2 检测逻辑，自动适配

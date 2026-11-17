<img src="assets/icon.png" width="64" height="64" alt="InlineExplain 应用图标">

# InlineExplain

### 在任何应用里，连同上下文一起读懂英文单词。

鼠标指向一个单词，按 <kbd>⌥</kbd> <kbd>S</kbd>。InlineExplain 从屏幕上识别**这个词和附近的句子**，在原文旁解释它在当前语境中的意思。

**[下载 macOS 版 ↗](https://github.com/Wake-C/inlineExplain/releases/latest)** · [English](README.en.md) · [快速开始](#快速开始) · [隐私](#哪些内容会离开你的-mac)

<sub>macOS 15+ · Apple Silicon · 自备模型 API Key</sub>

![在阅读应用中指向 bank，捕获包含 of the river 的附近句子，旁边的 InlineExplain 卡片据此解释为河岸。](docs/assets/popup-preview.png)

<sub>场景示意，使用实际卡片渲染和示例解释。绿色标记用于说明取词范围，并非应用叠层。[打开交互预览源码](docs/popup-preview.html)</sub>

## 阅读留在原来的地方

网页、PDF、文档，或视频里当前可见的英文字幕，都可以作为取词来源。InlineExplain 通过屏幕截图和本地 OCR 识别文字，不需要浏览器插件，也不用先选中或复制。能否识别取决于文字是否可见、清晰且允许截图。

| 你做的事 | InlineExplain 做的事 |
| :--- | :--- |
| 指向一个英文单词，按 **Option+S** | 识别鼠标下的词，以及它周围可见的文字 |
| 看原文旁的解释 | 结合附近的句子说明词义，附上音标、例句和近反义词 |
| 按 **Escape** 或点击外部 | 收起卡片，继续阅读 |

浮窗的设计目标是让原应用保持焦点，全屏时也不切走阅读空间。它读取屏幕上可见的语境，不会读取整篇文档或屏幕外的句子。

## 想听发音，点一下

- 词头旁的喇叭播放有道词典美音，服务不可用时退回 macOS 语音。
- 设置中的**自动朗读单词**可在取词时播放发音，默认关闭。
- 例句也能朗读，使用 macOS 语音。
- 解释默认中文，可切换英文。两种语言分别保存自定义 Prompt。

## 快速开始

1. **安装。** [下载 DMG](https://github.com/Wake-C/inlineExplain/releases/latest)，将 InlineExplain 拖到 **Applications（应用程序）**，再从那里打开。
2. **连接模型。** 点击菜单栏的书本图标，打开**设置**，填写服务地址、模型和 API Key。默认是 `https://api.deepseek.com` 和 `deepseek-chat`，也支持其他兼容 Chat Completions 的服务。模型服务可能收费。
3. **允许取词。** 授予**屏幕录制**权限；如果 macOS 要求，请退出并重新打开。
4. **开始阅读。** 鼠标指向英文单词，按 **Option+S**。快捷键冲突时，可在设置中改为 **Option+Shift+S**。

> **macOS 首次打开提示**
>
> 发布包使用自签名证书，尚未通过 Apple 公证。仅在信任项目及下载来源时，前往**系统设置 → 隐私与安全性 → 仍要打开**。
>
> 从 0.1.6 及以前的 ad-hoc 签名版本升级，可能需要重新授予屏幕录制权限。新版使用固定签名身份，但跨版本权限保留仍需实机验证。更新不会清除已保存的模型服务设置。

## 哪些内容会离开你的 Mac？

| 内容 | 处理方式 |
| :--- | :--- |
| 屏幕截图 | 仅按快捷键时采集，在本机 OCR，只保留在内存中 |
| 单词和一小段上下文 | 发送给你配置的模型服务，用于生成解释 |
| 单词发音 | 手动播放或开启自动朗读后，将单词发送给有道词典的免 Key 非官方接口 |
| API Key | 仅存于 macOS Keychain，不写入配置文件 |

应用不保存或记录截图、阅读文字和解释结果。有道接口的可用性不受本项目控制。

**使用边界：** 受保护的视频、过小的文字和复杂排版可能无法识别，模型解释也可能出错。目前支持单词取词，暂不支持短语选择、阅读历史、单词卡片或自动更新。

## 开发与来源

[开发指南](docs/development.md) · [架构](docs/architecture.md) · [打包与发布](docs/releasing.md) · [第三方声明](NOTICE)

InlineExplain 是独立的 macOS 阅读工具。[Peekturn](https://github.com/Wake-C/peekturn) 是面向 Android 的英文阅读应用，部分 Prompt 和代码沿用其设计，来源见第三方声明。

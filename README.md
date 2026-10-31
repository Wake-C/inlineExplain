# InlineExplain

<p align="center"><img src="assets/icon.png" width="96" height="96" alt="InlineExplain 应用图标"></p>

在正在阅读的应用里，读懂一个英文单词在句子中的意思。

**鼠标指向单词 → 按 Option+S → 就地查看语境解释。**

**[下载 macOS 版](https://github.com/Wake-C/inlineExplain/releases/latest)** · [English](README.en.md) · macOS 15+ · Apple Silicon

![英文文章旁的 InlineExplain 浮窗，展示 grasp 的中文释义、音标、语境说明和例句。](docs/assets/popup-preview.png)

<sub>HTML 示意预览，使用应用实际的浮窗渲染代码。[交互原型源码](docs/popup-preview.html)。</sub>

## 读懂单词，接着读下去

InlineExplain 识别鼠标下的单词和附近的句子，在一个白色小浮窗里显示对应的释义、音标、例句和近反义词。需要时可以展开原句、朗读例句或复制解释。

浮窗的设计目标是让原应用保持焦点，全屏阅读时也能就地查看。按 **Escape** 或点击外部即可关闭。界面和解释默认使用中文，可在设置中切换英文；两种语言分别保存自己的自定义 Prompt。

## 开始使用

1. 下载并打开 DMG，将 **InlineExplain** 拖到 **Applications（应用程序）**，再从那里打开应用。
2. 点击菜单栏的 InlineExplain 书本图标，打开**设置**，填写服务地址、模型和 API Key。默认使用 `https://api.deepseek.com` 和 `deepseek-chat`，也支持其他兼容 Chat Completions 的服务。服务商可能按用量收费。
3. 按提示允许**屏幕录制**。如果 macOS 要求，请退出并重新打开 InlineExplain。
4. 将鼠标直接放在可见的英文单词上，按 **Option+S**，无需先选中或复制。

更新会保留已保存的服务设置；DeepSeek 默认地址仅在首次启动时使用。

当前发布包采用 ad-hoc 签名，尚未通过 Apple 公证。如果 macOS 阻止打开，仅在你信任项目及下载来源时，前往**系统设置 → 隐私与安全性 → 仍要打开**。

如果其他应用也使用 Option+S，请暂停其快捷键，或在 InlineExplain 设置中改用 **Option+Shift+S**。

## 隐私与限制

仅在按下快捷键时截图，并在 Mac 上通过 Apple Vision 识别文字。截图只保留在内存中；单词和一小段上下文会发送给你选择的模型服务。API Key 保存在 macOS 钥匙串中，应用不保存或记录阅读文字、截图和解释结果。

OCR 只能识别屏幕上可见的文字。屏幕外的句子、受保护的视频、过小的文字和复杂排版可能导致上下文不足。模型解释也可能出错。当前版本支持单词取词，暂不支持短语选择、阅读历史、单词卡片或自动更新。

## 开发

构建和验证命令见[开发指南](docs/development.md)。本机执行 `./scripts/package.sh` 即可打包发布 DMG；[发布指南](docs/releasing.md)说明了如何在手动触发的 GitHub Actions 中使用同一条命令。推送代码不会触发打包或测试任务。

InlineExplain 是独立的 macOS 阅读工具，[Peekturn](https://github.com/Wake-C/peekturn) 则是面向 Android 的英文阅读应用。Prompt 和代码来源见[第三方声明](NOTICE)。

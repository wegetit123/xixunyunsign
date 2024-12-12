本文档详细说明了如何在 GitHub Actions 中添加和使用 Secrets 来安全执行 习讯云自动签到 的相关命令。

### 在 GitHub Actions 中添加 Secrets 的步骤

1. 进入仓库设置
   打开你的 GitHub 仓库。
   点击 Settings（设置）选项卡。
   向下滚动到 Secrets and variables 部分，点击 Actions。

2. 添加 Secrets
   点击 New repository secret（新建仓库密钥）按钮。
   按以下名称和对应的值添加 Secrets：
   USERNAME：习讯云的账号。
   PASSWORD：习讯云的密码。
   ADDRESS：签到所需的地址值。
   ADDRESS_NAME：签到所需的地址名称。
   API_KEY_FANGTANG:微信公众号方糖提供的密钥apikey
### 例如

Secret 名称：USERNAME

Secret 值：your-username

按此方式分别添加所有所需的 Secrets。
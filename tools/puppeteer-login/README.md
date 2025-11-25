快速使用说明

先确保 Admin-FE 已由静态服务器提供，例如在 `Admin-FE` 目录下运行：

Windows PowerShell:

```powershell
cd d:\developTool\work\go\tea\Admin-FE
python -m http.server 8000
```

然后在本项目中运行 Puppeteer 脚本：

```powershell
cd d:\developTool\work\go\tea\tools\puppeteer-login
npm install
# 可选环境变量：ADMIN_FE_URL, ADMIN_USERNAME, ADMIN_PASSWORD
npm start
```

脚本会打开页面、触发登录并在控制台打印 `tea_admin_token` 的前 40 字符（以及 role/name）。如果你想在可视模式下查看，可以把 `login.js` 中 `puppeteer.launch({ headless: true })` 改为 `headless: false`。

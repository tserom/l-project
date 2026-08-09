销售开票（sales-front + sales-manage）— Windows 运行包
====================================================

对方机器需已安装并启动 MySQL。无需安装 Node / Go。

首次准备（只需一次）
--------------------
1. 建库（在 MySQL 里执行）：

     CREATE DATABASE IF NOT EXISTS sales_manage DEFAULT CHARSET utf8mb4;

2. 解压整个 zip，打开文件夹 sales-front-windows
3. 用记事本编辑同目录下的 .env，改成你的 MySQL 账号密码，例如：

     DB_USER=root
     DB_PASSWORD=你的密码
     DB_HOST=127.0.0.1
     DB_PORT=3306
     DB_NAME=sales_manage
     SERVER_PORT=8083

日常使用
--------
1. 确认 MySQL 已启动
2. 双击 start-sales.bat
3. 浏览器打开 http://127.0.0.1:5175
4. 关闭黑色窗口即停止前端，并会顺带关掉 sales-manage

包内应有
--------
  start-sales.bat
  serve.ps1
  sales-manage.exe
  .env
  index.html
  assets\

常见问题
--------
- sales-manage did not become healthy：检查 MySQL 是否启动、库是否建好、.env 密码是否正确、8083 是否被占用
- Missing index.html：解压不完整；请重新解压整个 zip
- 端口 5175 被占用：关掉占用程序后重试

说明
----
- 数据在 MySQL 库 sales_manage（多机不共享，除非连同一库）
- 与库存系统 stock-* 无关；不是 stock-manage 的 /sales-orders
- 打印时建议关闭页眉页脚，边距选默认或最小

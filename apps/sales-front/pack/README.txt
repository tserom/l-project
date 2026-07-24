销售单（sales-front）— Windows 运行包
================================

无需安装 Node / Go / MySQL。

使用步骤
--------
1. 解压整个 zip（不要只拷贝其中一个文件）
2. 打开解压后的文件夹 sales-front-windows
3. 确认里面同时有：
     start-sales.bat
     serve.ps1
     index.html
     assets\   （文件夹）
4. 双击 start-sales.bat
5. 浏览器打开 http://127.0.0.1:5175
6. 关闭黑色窗口即停止

常见问题
--------
- Missing index.html：解压不完整，或只复制了 bat。请重新解压完整 zip。
- 乱码 / 不是内部命令：请使用最新 zip（bat 为英文脚本）。
- 端口被占用：关掉占用 5175 的程序后重试。

说明
----
- 数据在本机浏览器 IndexedDB；清站点数据会丢
- 不同电脑数据不共享
- 打印时建议关闭页眉页脚，边距选默认或最小

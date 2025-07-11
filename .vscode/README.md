# VSCode 调试配置说明

本项目已配置好VSCode调试环境，可以直接在VSCode中调试BLE代理服务。

## 调试配置

### 1. Debug BLE Agent Service (推荐)
- **用途**: 主要调试配置，会先构建项目再启动调试
- **参数**: `-cp -r -rsh=192.168.8.216,192.168.8.196,0.0.0.0`
- **特点**: 
  - 自动构建项目
  - 使用集成终端
  - 详细日志输出

### 2. Debug BLE Agent Service (No Build)
- **用途**: 快速调试，不重新构建项目
- **参数**: 同上
- **特点**: 跳过构建步骤，适合频繁调试

### 3. Debug BLE Agent Service (Custom Args)
- **用途**: 自定义参数调试
- **参数**: 空（可在launch.json中修改）
- **特点**: 可以自定义启动参数

### 4. Debug with Breakpoint at main
- **用途**: 从main函数开始调试
- **参数**: 同默认配置
- **特点**: 在程序入口处自动停止

### 5. Attach to Process
- **用途**: 附加到正在运行的进程
- **特点**: 可以调试已经启动的服务

## 使用方法

### 方法1: 使用调试面板
1. 打开VSCode
2. 按 `Ctrl+Shift+D` 打开调试面板
3. 在顶部下拉菜单中选择调试配置
4. 点击绿色播放按钮开始调试

### 方法2: 使用快捷键
1. 按 `F5` 启动调试（使用默认配置）
2. 按 `Ctrl+F5` 运行而不调试

### 方法3: 使用命令面板
1. 按 `Ctrl+Shift+P` 打开命令面板
2. 输入 "Debug: Start Debugging"
3. 选择调试配置

## 调试技巧

### 设置断点
- 在代码行号左侧点击设置断点
- 按 `F9` 在当前行设置/取消断点
- 条件断点：右键点击断点，选择"编辑断点"

### 调试控制
- `F5`: 继续执行
- `F10`: 单步跳过
- `F11`: 单步进入
- `Shift+F11`: 单步跳出
- `Shift+F5`: 停止调试

### 查看变量
- 在调试面板的"变量"区域查看局部变量
- 在"监视"区域添加表达式监视
- 鼠标悬停在变量上查看值

## 构建任务

项目配置了以下构建任务：

- **build**: 构建项目 (`make build`)
- **clean**: 清理构建文件 (`make clean`)
- **test**: 运行测试 (`make test`)
- **go mod tidy**: 整理Go模块依赖

使用方法：
1. 按 `Ctrl+Shift+P` 打开命令面板
2. 输入 "Tasks: Run Task"
3. 选择要执行的任务

## 环境变量

调试配置中设置了以下环境变量：
- `EDGEX_SECURITY_SECRET_STORE=false`: 禁用EdgeX安全存储

如需添加其他环境变量，请编辑 `.vscode/launch.json` 文件中的 `env` 部分。

## 故障排除

### 1. 调试器无法启动
- 确保已安装Go扩展
- 检查Go环境是否正确配置
- 尝试重新构建项目：`make clean && make build`

### 2. 断点不生效
- 确保代码已保存
- 检查是否在正确的文件中设置断点
- 尝试重新启动调试会话

### 3. 找不到可执行文件
- 运行 `make build` 确保项目已构建
- 检查 `cmd/device-uart` 文件是否存在

### 4. 端口或设备访问问题
- 确保串口设备 `/dev/ttyS3` 可访问
- 检查网络地址是否正确
- 确保有足够的权限访问硬件设备

## 配置文件说明

- `.vscode/launch.json`: 调试配置
- `.vscode/tasks.json`: 构建任务配置  
- `.vscode/settings.json`: Go开发环境设置

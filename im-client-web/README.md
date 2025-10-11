# 志航密信 - 用户端IM客户端

**这是真正的IM聊天客户端，不是管理后台**

## 功能特性

### ✅ 已实现
- ✅ 用户登录/注册
- ✅ 实时聊天（WebSocket）
- ✅ 消息收发
- ✅ 联系人列表
- ✅ 用户设置
- ✅ 在线状态
- ✅ 响应式设计

### ⏳ 待实现（Devin完善）
- ⏳ 文件上传/下载
- ⏳ 图片预览
- ⏳ 语音视频通话
- ⏳ 群组聊天
- ⏳ 消息搜索
- ⏳ 消息已读状态
- ⏳ 表情包
- ⏳ @提及功能

## 技术栈

- **框架**: Vue 3
- **UI库**: Element Plus
- **状态管理**: Pinia
- **HTTP客户端**: Axios
- **WebSocket**: 原生WebSocket
- **构建工具**: Vite

## 开发运行

```bash
cd im-client-web

# 安装依赖
npm install

# 开发模式
npm run dev
# 访问 http://localhost:5173

# 构建生产版本
npm run build
# 产物在 dist/ 目录
```

## 环境配置

创建 `.env.production` 文件：

```env
VITE_API_BASE_URL=http://154.37.214.191:8080
VITE_WS_URL=ws://154.37.214.191:8080/ws
```

## 部署

### Docker部署

```bash
# 构建
npm run build

# 使用Nginx Docker镜像
docker run -d \
  -p 8000:80 \
  -v $(pwd)/dist:/usr/share/nginx/html \
  --name im-client-web \
  nginx:alpine
```

### 直接部署

```bash
npm run build
cd dist
zip -r ../im-client-web.zip .

# 解压到Web服务器
unzip im-client-web.zip -d /var/www/im-client
```

## 与im-admin的区别

| 项目 | im-admin | im-client-web |
|------|----------|---------------|
| 用途 | 管理后台 | 用户端IM客户端 |
| 用户 | 管理员 | 普通用户 |
| 功能 | 系统管理、用户管理、内容审核 | 聊天、消息、联系人 |
| 界面 | 管理界面 | 聊天界面 |

## API对接

所有API都对接到后端：
- ✅ POST `/api/auth/login` - 登录
- ✅ POST `/api/auth/register` - 注册
- ✅ GET `/api/users/me` - 获取用户信息
- ✅ GET `/api/users/friends` - 获取联系人
- ✅ POST `/api/messages/send` - 发送消息
- ✅ GET `/api/messages` - 获取消息
- ✅ WS `/ws` - WebSocket连接

## 截图

（待添加实际运行截图）

## 后续完善

1. 文件上传组件
2. 图片消息组件
3. 语音视频组件
4. 群组聊天页面
5. 更多功能...

---

**这是真正的IM客户端，可以让普通用户聊天使用！**


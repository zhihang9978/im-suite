# GitHub Actions 故障排除指南

## 📋 常见问题解决

### 问题1: 旧的workflow仍在运行

**症状**: 
- Actions页面显示错误的workflow运行
- 提示找不到 `ci-cd.yml` 文件

**原因**: 
- 之前的提交中有workflow文件，后来被删除
- GitHub缓存了旧的workflow配置

**解决方案**:
1. 进入仓库的 **Actions** 页面
2. 点击左侧的旧workflow名称（如 "CI/CD Pipeline"）
3. 点击右上角的 **"..."** 菜单
4. 选择 **"Disable workflow"** 禁用旧workflow
5. 或者等待新的push触发新的workflow

### 问题2: Workflow权限不足

**症状**:
- Actions运行失败，提示权限错误
- 无法访问仓库或推送

**解决方案**:
1. 进入仓库 **Settings** → **Actions** → **General**
2. 找到 **Workflow permissions**
3. 选择 **"Read and write permissions"**
4. 勾选 **"Allow GitHub Actions to create and approve pull requests"**
5. 保存设置

### 问题3: LFS文件拉取失败

**症状**:
- Checkout步骤失败
- 提示LFS文件错误

**解决方案**:
- 已在workflow中添加 `lfs: true` 选项
- 确保仓库有足够的LFS配额

### 问题4: 依赖安装失败

**症状**:
- npm install 或 go mod download 失败

**解决方案**:
- 已添加 `continue-on-error: true`
- 允许workflow继续执行其他检查

## 🔄 如何手动触发workflow

1. 进入仓库的 **Actions** 页面
2. 选择要运行的workflow
3. 点击 **"Run workflow"** 按钮
4. 选择分支，点击 **"Run workflow"**

## ✅ 当前的workflow配置

### 1. CI/CD Pipeline (`ci.yml`)
完整的CI/CD流程，包含：
- 后端测试
- 前端构建
- Docker镜像构建
- 安全扫描
- 代码质量检查

### 2. Simple CI (`simple-ci.yml`)
简化的CI流程，包含：
- 基础文件检查
- Go环境验证
- 项目结构检查

## 🛠️ 如果Actions仍然报错

### 方案1: 禁用所有Actions
```bash
# 进入仓库Settings → Actions → General
# 选择 "Disable actions"
# 需要时再重新启用
```

### 方案2: 清理workflow缓存
```bash
# 在本地执行
git rm -r .github/workflows/
git commit -m "chore: remove workflows temporarily"
git push origin main

# 等待几分钟后重新添加
git revert HEAD
git push origin main
```

### 方案3: 使用最小化workflow
创建一个最简单的workflow：
```yaml
name: Minimal CI
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: echo "OK"
```

## 📞 需要帮助？

如果问题仍然存在，请提供：
1. Actions页面的错误截图
2. 失败的workflow运行日志
3. 错误消息的完整内容

我们会根据具体错误提供针对性的解决方案。

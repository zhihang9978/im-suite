#!/bin/bash

# 志航密信测试执行脚本
# 针对中国用户手机进行综合测试

echo "=========================================="
echo "志航密信 - 中国用户手机综合测试"
echo "=========================================="

# 设置测试环境
export PYTHONPATH="${PYTHONPATH}:$(pwd)"
export TEST_ENV="chinese_phones"

# 创建测试目录
mkdir -p test_results
cd test_results

# 检查Python环境
echo "检查Python环境..."
python3 --version
if [ $? -ne 0 ]; then
    echo "错误: Python3 未安装或不在PATH中"
    exit 1
fi

# 检查依赖包
echo "检查依赖包..."
python3 -c "import requests, psutil, websocket" 2>/dev/null
if [ $? -ne 0 ]; then
    echo "安装依赖包..."
    pip3 install requests psutil websocket-client
fi

# 启动后端服务（如果未运行）
echo "检查后端服务..."
curl -s http://localhost:8080/api/ping > /dev/null
if [ $? -ne 0 ]; then
    echo "警告: 后端服务未运行，部分测试可能失败"
    echo "请确保后端服务在 http://localhost:8080 运行"
fi

# 运行综合测试
echo "开始运行综合测试..."
python3 ../run_all_tests.py

# 检查测试结果
if [ -f "comprehensive_test_report.json" ]; then
    echo "=========================================="
    echo "测试完成！"
    echo "=========================================="
    echo "测试报告文件:"
    echo "  - comprehensive_test_report.json (综合报告)"
    echo "  - test_report.json (功能测试)"
    echo "  - performance_report.json (性能测试)"
    echo "  - compatibility_report.json (兼容性测试)"
    echo "  - ux_report.json (用户体验测试)"
    echo ""
    echo "日志文件:"
    echo "  - comprehensive_test.log"
    echo "  - functional_test.log"
    echo "  - performance_test.log"
    echo "  - compatibility_test.log"
    echo "  - ux_test.log"
else
    echo "错误: 测试执行失败"
    exit 1
fi

echo "测试执行完成！"

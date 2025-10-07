#!/usr/bin/env python3
"""
志航密信性能测试脚本
用于测试系统的性能指标和负载能力
"""

import asyncio
import aiohttp
import time
import statistics
import json
import argparse
import sys
from typing import List, Dict, Any
from dataclasses import dataclass
from concurrent.futures import ThreadPoolExecutor
import logging

# 配置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

@dataclass
class TestResult:
    """测试结果数据类"""
    name: str
    success_count: int
    failure_count: int
    total_time: float
    avg_response_time: float
    min_response_time: float
    max_response_time: float
    p95_response_time: float
    p99_response_time: float
    requests_per_second: float

class PerformanceTest:
    """性能测试类"""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session = None
        self.results: List[TestResult] = []
    
    async def __aenter__(self):
        """异步上下文管理器入口"""
        self.session = aiohttp.ClientSession()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """异步上下文管理器出口"""
        if self.session:
            await self.session.close()
    
    async def make_request(self, method: str, endpoint: str, data: Dict[str, Any] = None, headers: Dict[str, str] = None) -> tuple:
        """
        发送 HTTP 请求
        
        Args:
            method: HTTP 方法
            endpoint: API 端点
            data: 请求数据
            headers: 请求头
            
        Returns:
            (success: bool, response_time: float, status_code: int)
        """
        url = f"{self.base_url}{endpoint}"
        start_time = time.time()
        
        try:
            async with self.session.request(
                method=method,
                url=url,
                json=data,
                headers=headers,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                response_time = time.time() - start_time
                success = response.status < 400
                return success, response_time, response.status
        except Exception as e:
            response_time = time.time() - start_time
            logger.error(f"请求失败: {e}")
            return False, response_time, 0
    
    async def run_concurrent_test(self, name: str, method: str, endpoint: str, 
                                concurrent_users: int, requests_per_user: int,
                                data: Dict[str, Any] = None, headers: Dict[str, str] = None) -> TestResult:
        """
        运行并发测试
        
        Args:
            name: 测试名称
            method: HTTP 方法
            endpoint: API 端点
            concurrent_users: 并发用户数
            requests_per_user: 每个用户的请求数
            data: 请求数据
            headers: 请求头
            
        Returns:
            测试结果
        """
        logger.info(f"开始运行测试: {name}")
        logger.info(f"并发用户数: {concurrent_users}, 每用户请求数: {requests_per_user}")
        
        start_time = time.time()
        response_times = []
        success_count = 0
        failure_count = 0
        
        # 创建并发任务
        async def user_session():
            user_response_times = []
            user_success = 0
            user_failure = 0
            
            for _ in range(requests_per_user):
                success, resp_time, status_code = await self.make_request(method, endpoint, data, headers)
                user_response_times.append(resp_time)
                
                if success:
                    user_success += 1
                else:
                    user_failure += 1
                    logger.warning(f"请求失败: {status_code}")
            
            return user_response_times, user_success, user_failure
        
        # 运行并发测试
        tasks = [user_session() for _ in range(concurrent_users)]
        results = await asyncio.gather(*tasks)
        
        # 汇总结果
        for user_response_times, user_success, user_failure in results:
            response_times.extend(user_response_times)
            success_count += user_success
            failure_count += user_failure
        
        total_time = time.time() - start_time
        
        # 计算统计指标
        if response_times:
            avg_response_time = statistics.mean(response_times)
            min_response_time = min(response_times)
            max_response_time = max(response_times)
            p95_response_time = self._percentile(response_times, 95)
            p99_response_time = self._percentile(response_times, 99)
        else:
            avg_response_time = min_response_time = max_response_time = 0
            p95_response_time = p99_response_time = 0
        
        requests_per_second = (success_count + failure_count) / total_time if total_time > 0 else 0
        
        result = TestResult(
            name=name,
            success_count=success_count,
            failure_count=failure_count,
            total_time=total_time,
            avg_response_time=avg_response_time,
            min_response_time=min_response_time,
            max_response_time=max_response_time,
            p95_response_time=p95_response_time,
            p99_response_time=p99_response_time,
            requests_per_second=requests_per_second
        )
        
        self.results.append(result)
        logger.info(f"测试完成: {name}")
        logger.info(f"成功率: {success_count/(success_count+failure_count)*100:.2f}%")
        logger.info(f"平均响应时间: {avg_response_time:.3f}s")
        logger.info(f"请求/秒: {requests_per_second:.2f}")
        
        return result
    
    def _percentile(self, data: List[float], percentile: int) -> float:
        """计算百分位数"""
        sorted_data = sorted(data)
        index = int(len(sorted_data) * percentile / 100)
        return sorted_data[min(index, len(sorted_data) - 1)]
    
    def generate_report(self) -> str:
        """生成测试报告"""
        if not self.results:
            return "没有测试结果"
        
        report = []
        report.append("=" * 80)
        report.append("志航密信性能测试报告")
        report.append("=" * 80)
        report.append(f"测试时间: {time.strftime('%Y-%m-%d %H:%M:%S')}")
        report.append(f"测试总数: {len(self.results)}")
        report.append("")
        
        for result in self.results:
            report.append(f"测试名称: {result.name}")
            report.append("-" * 40)
            report.append(f"总请求数: {result.success_count + result.failure_count}")
            report.append(f"成功请求: {result.success_count}")
            report.append(f"失败请求: {result.failure_count}")
            report.append(f"成功率: {result.success_count/(result.success_count+result.failure_count)*100:.2f}%")
            report.append(f"总测试时间: {result.total_time:.3f}s")
            report.append(f"平均响应时间: {result.avg_response_time:.3f}s")
            report.append(f"最小响应时间: {result.min_response_time:.3f}s")
            report.append(f"最大响应时间: {result.max_response_time:.3f}s")
            report.append(f"95%响应时间: {result.p95_response_time:.3f}s")
            report.append(f"99%响应时间: {result.p99_response_time:.3f}s")
            report.append(f"请求/秒: {result.requests_per_second:.2f}")
            report.append("")
        
        return "\n".join(report)

async def main():
    """主函数"""
    parser = argparse.ArgumentParser(description="志航密信性能测试")
    parser.add_argument("--base-url", default="http://localhost:8080", help="API 基础 URL")
    parser.add_argument("--concurrent-users", type=int, default=10, help="并发用户数")
    parser.add_argument("--requests-per-user", type=int, default=10, help="每个用户的请求数")
    parser.add_argument("--output", help="输出文件路径")
    parser.add_argument("--test-type", choices=["all", "api", "websocket", "database"], 
                       default="all", help="测试类型")
    
    args = parser.parse_args()
    
    logger.info("开始志航密信性能测试")
    logger.info(f"API 基础 URL: {args.base_url}")
    logger.info(f"并发用户数: {args.concurrent_users}")
    logger.info(f"每用户请求数: {args.requests_per_user}")
    
    async with PerformanceTest(args.base_url) as tester:
        # API 健康检查测试
        if args.test_type in ["all", "api"]:
            await tester.run_concurrent_test(
                name="API 健康检查",
                method="GET",
                endpoint="/api/ping",
                concurrent_users=args.concurrent_users,
                requests_per_user=args.requests_per_user
            )
        
        # 用户认证测试
        if args.test_type in ["all", "api"]:
            await tester.run_concurrent_test(
                name="用户登录测试",
                method="POST",
                endpoint="/api/auth/login",
                concurrent_users=args.concurrent_users,
                requests_per_user=args.requests_per_user,
                data={"phone": "13800138000", "code": "123456"}
            )
        
        # 消息发送测试
        if args.test_type in ["all", "api"]:
            await tester.run_concurrent_test(
                name="消息发送测试",
                method="POST",
                endpoint="/api/messages/send",
                concurrent_users=args.concurrent_users,
                requests_per_user=args.requests_per_user,
                data={
                    "chat_id": 1,
                    "content": "性能测试消息",
                    "type": "text"
                },
                headers={"Authorization": "Bearer test-token"}
            )
        
        # 获取消息列表测试
        if args.test_type in ["all", "api"]:
            await tester.run_concurrent_test(
                name="获取消息列表测试",
                method="GET",
                endpoint="/api/messages?chat_id=1&limit=50",
                concurrent_users=args.concurrent_users,
                requests_per_user=args.requests_per_user,
                headers={"Authorization": "Bearer test-token"}
            )
    
    # 生成报告
    report = tester.generate_report()
    print(report)
    
    # 保存报告
    if args.output:
        with open(args.output, 'w', encoding='utf-8') as f:
            f.write(report)
        logger.info(f"测试报告已保存到: {args.output}")
    
    logger.info("性能测试完成")

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("测试被用户中断")
        sys.exit(1)
    except Exception as e:
        logger.error(f"测试执行失败: {e}")
        sys.exit(1)
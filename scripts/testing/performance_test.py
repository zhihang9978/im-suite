#!/usr/bin/env python3
"""
志航密信性能测试脚本
针对中国用户手机进行性能测试
"""

import requests
import json
import time
import psutil
import threading
import concurrent.futures
from typing import Dict, List, Any
from dataclasses import dataclass
from datetime import datetime
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('performance_test.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

@dataclass
class PerformanceMetrics:
    """性能指标数据类"""
    test_name: str
    response_time: float
    memory_usage: float
    cpu_usage: float
    throughput: float
    error_rate: float
    timestamp: datetime

class ChinesePhonePerformanceTester:
    """中国手机品牌性能测试器"""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()
        self.metrics: List[PerformanceMetrics] = []
        
        # 中国手机品牌性能基准
        self.brand_benchmarks = {
            "xiaomi": {
                "name": "小米",
                "expected_response_time": 0.5,  # 秒
                "expected_memory_usage": 100,   # MB
                "expected_cpu_usage": 15,       # %
                "expected_throughput": 100       # 请求/秒
            },
            "huawei": {
                "name": "华为",
                "expected_response_time": 0.6,
                "expected_memory_usage": 120,
                "expected_cpu_usage": 18,
                "expected_throughput": 90
            },
            "oppo": {
                "name": "OPPO",
                "expected_response_time": 0.7,
                "expected_memory_usage": 110,
                "expected_cpu_usage": 16,
                "expected_throughput": 85
            },
            "vivo": {
                "name": "vivo",
                "expected_response_time": 0.8,
                "expected_memory_usage": 115,
                "expected_cpu_usage": 17,
                "expected_throughput": 80
            },
            "oneplus": {
                "name": "一加",
                "expected_response_time": 0.4,
                "expected_memory_usage": 95,
                "expected_cpu_usage": 12,
                "expected_throughput": 120
            },
            "realme": {
                "name": "realme",
                "expected_response_time": 0.6,
                "expected_memory_usage": 105,
                "expected_cpu_usage": 14,
                "expected_throughput": 95
            },
            "meizu": {
                "name": "魅族",
                "expected_response_time": 0.7,
                "expected_memory_usage": 108,
                "expected_cpu_usage": 16,
                "expected_throughput": 88
            }
        }
    
    def get_system_metrics(self) -> Dict[str, float]:
        """获取系统性能指标"""
        try:
            # CPU使用率
            cpu_percent = psutil.cpu_percent(interval=1)
            
            # 内存使用情况
            memory = psutil.virtual_memory()
            memory_usage = memory.used / (1024 * 1024)  # 转换为MB
            
            # 磁盘使用情况
            disk = psutil.disk_usage('/')
            disk_usage = disk.used / (1024 * 1024 * 1024)  # 转换为GB
            
            return {
                "cpu_usage": cpu_percent,
                "memory_usage": memory_usage,
                "disk_usage": disk_usage,
                "memory_percent": memory.percent
            }
        except Exception as e:
            logger.error(f"获取系统指标失败: {e}")
            return {}
    
    def test_api_response_time(self, endpoint: str, method: str = "GET", data: Dict = None) -> float:
        """测试API响应时间"""
        try:
            start_time = time.time()
            
            if method.upper() == "GET":
                response = self.session.get(f"{self.base_url}{endpoint}", timeout=10)
            elif method.upper() == "POST":
                response = self.session.post(f"{self.base_url}{endpoint}", json=data, timeout=10)
            elif method.upper() == "PUT":
                response = self.session.put(f"{self.base_url}{endpoint}", json=data, timeout=10)
            elif method.upper() == "DELETE":
                response = self.session.delete(f"{self.base_url}{endpoint}", timeout=10)
            
            response_time = time.time() - start_time
            
            if response.status_code >= 400:
                logger.warning(f"API请求失败: {endpoint} - {response.status_code}")
            
            return response_time
        except Exception as e:
            logger.error(f"API响应时间测试失败: {e}")
            return -1
    
    def test_concurrent_requests(self, endpoint: str, num_requests: int = 100) -> Dict[str, Any]:
        """测试并发请求性能"""
        try:
            start_time = time.time()
            success_count = 0
            error_count = 0
            response_times = []
            
            def make_request():
                try:
                    req_start = time.time()
                    response = self.session.get(f"{self.base_url}{endpoint}", timeout=5)
                    req_time = time.time() - req_start
                    
                    if response.status_code == 200:
                        return True, req_time
                    else:
                        return False, req_time
                except Exception:
                    return False, 0
            
            # 使用线程池执行并发请求
            with concurrent.futures.ThreadPoolExecutor(max_workers=20) as executor:
                futures = [executor.submit(make_request) for _ in range(num_requests)]
                results = [future.result() for future in futures]
            
            total_time = time.time() - start_time
            
            for success, req_time in results:
                if success:
                    success_count += 1
                    response_times.append(req_time)
                else:
                    error_count += 1
            
            throughput = num_requests / total_time if total_time > 0 else 0
            error_rate = (error_count / num_requests) * 100 if num_requests > 0 else 0
            avg_response_time = sum(response_times) / len(response_times) if response_times else 0
            
            return {
                "total_requests": num_requests,
                "success_count": success_count,
                "error_count": error_count,
                "throughput": throughput,
                "error_rate": error_rate,
                "avg_response_time": avg_response_time,
                "total_time": total_time
            }
        except Exception as e:
            logger.error(f"并发请求测试失败: {e}")
            return {}
    
    def test_memory_usage(self, duration: int = 60) -> Dict[str, Any]:
        """测试内存使用情况"""
        try:
            memory_samples = []
            start_time = time.time()
            
            while time.time() - start_time < duration:
                metrics = self.get_system_metrics()
                if metrics:
                    memory_samples.append(metrics["memory_usage"])
                time.sleep(1)
            
            if memory_samples:
                avg_memory = sum(memory_samples) / len(memory_samples)
                max_memory = max(memory_samples)
                min_memory = min(memory_samples)
                
                return {
                    "duration": duration,
                    "samples": len(memory_samples),
                    "avg_memory": avg_memory,
                    "max_memory": max_memory,
                    "min_memory": min_memory,
                    "memory_trend": memory_samples
                }
            else:
                return {}
        except Exception as e:
            logger.error(f"内存使用测试失败: {e}")
            return {}
    
    def test_cpu_usage(self, duration: int = 60) -> Dict[str, Any]:
        """测试CPU使用情况"""
        try:
            cpu_samples = []
            start_time = time.time()
            
            while time.time() - start_time < duration:
                metrics = self.get_system_metrics()
                if metrics:
                    cpu_samples.append(metrics["cpu_usage"])
                time.sleep(1)
            
            if cpu_samples:
                avg_cpu = sum(cpu_samples) / len(cpu_samples)
                max_cpu = max(cpu_samples)
                min_cpu = min(cpu_samples)
                
                return {
                    "duration": duration,
                    "samples": len(cpu_samples),
                    "avg_cpu": avg_cpu,
                    "max_cpu": max_cpu,
                    "min_cpu": min_cpu,
                    "cpu_trend": cpu_samples
                }
            else:
                return {}
        except Exception as e:
            logger.error(f"CPU使用测试失败: {e}")
            return {}
    
    def test_brand_optimization_performance(self, brand: str) -> Dict[str, Any]:
        """测试品牌优化性能"""
        try:
            brand_config = self.brand_benchmarks.get(brand)
            if not brand_config:
                return {}
            
            # 测试品牌特定优化
            optimization_endpoints = [
                "/api/optimization/xiaomi" if brand == "xiaomi" else f"/api/optimization/{brand}",
                "/api/permissions/test",
                "/api/theme/apply",
                "/api/notification/test"
            ]
            
            results = {}
            total_response_time = 0
            successful_requests = 0
            
            for endpoint in optimization_endpoints:
                response_time = self.test_api_response_time(endpoint)
                if response_time > 0:
                    total_response_time += response_time
                    successful_requests += 1
                    results[endpoint] = response_time
            
            avg_response_time = total_response_time / successful_requests if successful_requests > 0 else 0
            
            # 与基准对比
            benchmark = brand_config
            performance_score = 0
            
            if avg_response_time <= benchmark["expected_response_time"]:
                performance_score += 25
            if avg_response_time <= benchmark["expected_response_time"] * 0.8:
                performance_score += 25
            
            return {
                "brand": brand,
                "brand_name": brand_config["name"],
                "avg_response_time": avg_response_time,
                "expected_response_time": benchmark["expected_response_time"],
                "performance_score": performance_score,
                "successful_requests": successful_requests,
                "total_requests": len(optimization_endpoints),
                "details": results
            }
        except Exception as e:
            logger.error(f"品牌优化性能测试失败 ({brand}): {e}")
            return {}
    
    def test_network_performance(self) -> Dict[str, Any]:
        """测试网络性能"""
        try:
            # 测试不同网络环境下的性能
            network_tests = [
                {"name": "本地网络", "endpoint": "/api/ping"},
                {"name": "用户信息", "endpoint": "/api/users/me"},
                {"name": "消息列表", "endpoint": "/api/messages"},
                {"name": "文件上传", "endpoint": "/api/files/upload"}
            ]
            
            results = {}
            
            for test in network_tests:
                response_time = self.test_api_response_time(test["endpoint"])
                if response_time > 0:
                    results[test["name"]] = {
                        "endpoint": test["endpoint"],
                        "response_time": response_time,
                        "status": "success"
                    }
                else:
                    results[test["name"]] = {
                        "endpoint": test["endpoint"],
                        "response_time": -1,
                        "status": "failed"
                    }
            
            return results
        except Exception as e:
            logger.error(f"网络性能测试失败: {e}")
            return {}
    
    def run_performance_tests(self):
        """运行所有性能测试"""
        logger.info("开始志航密信性能测试")
        logger.info("=" * 50)
        
        # 基础性能测试
        logger.info("1. 基础性能测试")
        ping_time = self.test_api_response_time("/api/ping")
        logger.info(f"   Ping响应时间: {ping_time:.3f}秒")
        
        # 并发性能测试
        logger.info("2. 并发性能测试")
        concurrent_results = self.test_concurrent_requests("/api/ping", 100)
        if concurrent_results:
            logger.info(f"   吞吐量: {concurrent_results['throughput']:.2f} 请求/秒")
            logger.info(f"   错误率: {concurrent_results['error_rate']:.2f}%")
            logger.info(f"   平均响应时间: {concurrent_results['avg_response_time']:.3f}秒")
        
        # 内存使用测试
        logger.info("3. 内存使用测试")
        memory_results = self.test_memory_usage(30)
        if memory_results:
            logger.info(f"   平均内存使用: {memory_results['avg_memory']:.2f}MB")
            logger.info(f"   最大内存使用: {memory_results['max_memory']:.2f}MB")
        
        # CPU使用测试
        logger.info("4. CPU使用测试")
        cpu_results = self.test_cpu_usage(30)
        if cpu_results:
            logger.info(f"   平均CPU使用: {cpu_results['avg_cpu']:.2f}%")
            logger.info(f"   最大CPU使用: {cpu_results['max_cpu']:.2f}%")
        
        # 中国品牌优化性能测试
        logger.info("5. 中国品牌优化性能测试")
        for brand in self.brand_benchmarks.keys():
            brand_results = self.test_brand_optimization_performance(brand)
            if brand_results:
                brand_name = brand_results.get("brand_name", brand)
                performance_score = brand_results.get("performance_score", 0)
                avg_response_time = brand_results.get("avg_response_time", 0)
                
                logger.info(f"   {brand_name}: 性能得分 {performance_score}/100, 响应时间 {avg_response_time:.3f}秒")
        
        # 网络性能测试
        logger.info("6. 网络性能测试")
        network_results = self.test_network_performance()
        for test_name, result in network_results.items():
            if result["status"] == "success":
                logger.info(f"   {test_name}: {result['response_time']:.3f}秒")
            else:
                logger.info(f"   {test_name}: 测试失败")
        
        # 生成性能报告
        self.generate_performance_report()
    
    def generate_performance_report(self):
        """生成性能测试报告"""
        report_data = {
            "test_time": datetime.now().isoformat(),
            "system_info": self.get_system_metrics(),
            "brand_benchmarks": self.brand_benchmarks,
            "summary": {
                "total_tests": len(self.metrics),
                "avg_response_time": 0,
                "avg_memory_usage": 0,
                "avg_cpu_usage": 0
            }
        }
        
        # 保存报告
        with open('performance_report.json', 'w', encoding='utf-8') as f:
            json.dump(report_data, f, ensure_ascii=False, indent=2)
        
        logger.info(f"\n性能测试报告已保存到: performance_report.json")

def main():
    """主函数"""
    tester = ChinesePhonePerformanceTester()
    tester.run_performance_tests()

if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""
志航密信功能测试脚本
针对中国用户手机进行功能测试
"""

import requests
import json
import time
import logging
from typing import Dict, List, Any
from dataclasses import dataclass
from enum import Enum

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('functional_test.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class TestStatus(Enum):
    PASS = "PASS"
    FAIL = "FAIL"
    SKIP = "SKIP"
    ERROR = "ERROR"

@dataclass
class TestResult:
    test_name: str
    status: TestStatus
    message: str
    duration: float
    details: Dict[str, Any] = None

class ChinesePhoneTester:
    """中国手机品牌测试器"""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()
        self.test_results: List[TestResult] = []
        
        # 中国手机品牌配置
        self.chinese_brands = {
            "xiaomi": {
                "name": "小米",
                "systems": ["MIUI 12", "MIUI 13", "MIUI 14"],
                "android_versions": ["Android 10", "Android 11", "Android 12", "Android 13"]
            },
            "huawei": {
                "name": "华为",
                "systems": ["EMUI 11", "EMUI 12", "HarmonyOS 2", "HarmonyOS 3"],
                "android_versions": ["Android 10", "Android 11", "Android 12"]
            },
            "oppo": {
                "name": "OPPO",
                "systems": ["ColorOS 11", "ColorOS 12", "ColorOS 13"],
                "android_versions": ["Android 10", "Android 11", "Android 12", "Android 13"]
            },
            "vivo": {
                "name": "vivo",
                "systems": ["OriginOS 1", "OriginOS 2", "OriginOS 3", "FuntouchOS 11", "FuntouchOS 12"],
                "android_versions": ["Android 10", "Android 11", "Android 12", "Android 13"]
            },
            "oneplus": {
                "name": "一加",
                "systems": ["OxygenOS 11", "OxygenOS 12", "OxygenOS 13"],
                "android_versions": ["Android 10", "Android 11", "Android 12", "Android 13"]
            },
            "realme": {
                "name": "realme",
                "systems": ["realme UI 2", "realme UI 3", "realme UI 4"],
                "android_versions": ["Android 10", "Android 11", "Android 12", "Android 13"]
            },
            "meizu": {
                "name": "魅族",
                "systems": ["Flyme 8", "Flyme 9"],
                "android_versions": ["Android 10", "Android 11", "Android 12"]
            }
        }
    
    def run_test(self, test_func, test_name: str, *args, **kwargs) -> TestResult:
        """运行单个测试"""
        start_time = time.time()
        try:
            result = test_func(*args, **kwargs)
            duration = time.time() - start_time
            
            if result:
                status = TestStatus.PASS
                message = f"{test_name} 测试通过"
            else:
                status = TestStatus.FAIL
                message = f"{test_name} 测试失败"
                
        except Exception as e:
            duration = time.time() - start_time
            status = TestStatus.ERROR
            message = f"{test_name} 测试出错: {str(e)}"
            result = None
            
        test_result = TestResult(
            test_name=test_name,
            status=status,
            message=message,
            duration=duration,
            details={"result": result}
        )
        
        self.test_results.append(test_result)
        logger.info(f"{test_name}: {status.value} - {message} ({duration:.2f}s)")
        
        return test_result
    
    def test_api_connectivity(self) -> bool:
        """测试API连接性"""
        try:
            response = self.session.get(f"{self.base_url}/api/ping", timeout=10)
            return response.status_code == 200
        except Exception as e:
            logger.error(f"API连接测试失败: {e}")
            return False
    
    def test_user_registration(self) -> bool:
        """测试用户注册功能"""
        try:
            # 测试数据
            test_user = {
                "phone": "13800138000",
                "password": "test123456",
                "verification_code": "123456"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/auth/register",
                json=test_user,
                timeout=10
            )
            
            return response.status_code in [200, 201, 409]  # 409表示用户已存在
        except Exception as e:
            logger.error(f"用户注册测试失败: {e}")
            return False
    
    def test_user_login(self) -> bool:
        """测试用户登录功能"""
        try:
            login_data = {
                "phone": "13800138000",
                "password": "test123456"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/auth/login",
                json=login_data,
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                # 保存token用于后续测试
                if "token" in data:
                    self.session.headers.update({"Authorization": f"Bearer {data['token']}"})
                return True
            return False
        except Exception as e:
            logger.error(f"用户登录测试失败: {e}")
            return False
    
    def test_message_sending(self) -> bool:
        """测试消息发送功能"""
        try:
            message_data = {
                "chat_id": "test_chat_001",
                "content": "测试消息",
                "type": "text"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/messages/send",
                json=message_data,
                timeout=10
            )
            
            return response.status_code in [200, 201]
        except Exception as e:
            logger.error(f"消息发送测试失败: {e}")
            return False
    
    def test_file_upload(self) -> bool:
        """测试文件上传功能"""
        try:
            # 创建测试文件
            test_file = ("test.txt", "这是一个测试文件", "text/plain")
            
            files = {"file": test_file}
            response = self.session.post(
                f"{self.base_url}/api/files/upload",
                files=files,
                timeout=30
            )
            
            return response.status_code in [200, 201]
        except Exception as e:
            logger.error(f"文件上传测试失败: {e}")
            return False
    
    def test_websocket_connection(self) -> bool:
        """测试WebSocket连接"""
        try:
            import websocket
            
            def on_message(ws, message):
                logger.info(f"WebSocket消息: {message}")
            
            def on_error(ws, error):
                logger.error(f"WebSocket错误: {error}")
            
            def on_close(ws, close_status_code, close_msg):
                logger.info("WebSocket连接关闭")
            
            def on_open(ws):
                logger.info("WebSocket连接成功")
                ws.close()
            
            ws_url = self.base_url.replace("http", "ws") + "/ws"
            ws = websocket.WebSocketApp(
                ws_url,
                on_message=on_message,
                on_error=on_error,
                on_close=on_close,
                on_open=on_open
            )
            
            ws.run_forever(timeout=10)
            return True
        except Exception as e:
            logger.error(f"WebSocket连接测试失败: {e}")
            return False
    
    def test_chinese_brand_optimization(self, brand: str) -> bool:
        """测试中国品牌优化"""
        try:
            brand_config = self.chinese_brands.get(brand)
            if not brand_config:
                return False
            
            # 测试品牌特定功能
            optimization_data = {
                "brand": brand,
                "system": brand_config["systems"][0],
                "android_version": brand_config["android_versions"][0]
            }
            
            response = self.session.post(
                f"{self.base_url}/api/optimization/test",
                json=optimization_data,
                timeout=10
            )
            
            return response.status_code == 200
        except Exception as e:
            logger.error(f"中国品牌优化测试失败 ({brand}): {e}")
            return False
    
    def test_permission_management(self, brand: str) -> bool:
        """测试权限管理"""
        try:
            permission_data = {
                "brand": brand,
                "permissions": ["INTERNET", "ACCESS_NETWORK_STATE", "WAKE_LOCK"]
            }
            
            response = self.session.post(
                f"{self.base_url}/api/permissions/test",
                json=permission_data,
                timeout=10
            )
            
            return response.status_code == 200
        except Exception as e:
            logger.error(f"权限管理测试失败 ({brand}): {e}")
            return False
    
    def test_performance_metrics(self) -> Dict[str, Any]:
        """测试性能指标"""
        try:
            start_time = time.time()
            
            # 测试API响应时间
            response = self.session.get(f"{self.base_url}/api/ping", timeout=5)
            api_response_time = time.time() - start_time
            
            # 测试并发请求
            import concurrent.futures
            with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
                futures = [
                    executor.submit(self.session.get, f"{self.base_url}/api/ping")
                    for _ in range(10)
                ]
                concurrent_results = [f.result() for f in futures]
            
            concurrent_success_rate = sum(1 for r in concurrent_results if r.status_code == 200) / len(concurrent_results)
            
            return {
                "api_response_time": api_response_time,
                "concurrent_success_rate": concurrent_success_rate,
                "total_requests": len(concurrent_results)
            }
        except Exception as e:
            logger.error(f"性能测试失败: {e}")
            return {}
    
    def run_all_tests(self):
        """运行所有测试"""
        logger.info("开始志航密信功能测试")
        logger.info("=" * 50)
        
        # 基础功能测试
        self.run_test(self.test_api_connectivity, "API连接性测试")
        self.run_test(self.test_user_registration, "用户注册测试")
        self.run_test(self.test_user_login, "用户登录测试")
        self.run_test(self.test_message_sending, "消息发送测试")
        self.run_test(self.test_file_upload, "文件上传测试")
        self.run_test(self.test_websocket_connection, "WebSocket连接测试")
        
        # 中国品牌优化测试
        for brand in self.chinese_brands.keys():
            self.run_test(
                self.test_chinese_brand_optimization, 
                f"{self.chinese_brands[brand]['name']}品牌优化测试",
                brand
            )
            self.run_test(
                self.test_permission_management,
                f"{self.chinese_brands[brand]['name']}权限管理测试",
                brand
            )
        
        # 性能测试
        performance_result = self.run_test(
            self.test_performance_metrics,
            "性能指标测试"
        )
        
        # 生成测试报告
        self.generate_test_report()
    
    def generate_test_report(self):
        """生成测试报告"""
        total_tests = len(self.test_results)
        passed_tests = sum(1 for r in self.test_results if r.status == TestStatus.PASS)
        failed_tests = sum(1 for r in self.test_results if r.status == TestStatus.FAIL)
        error_tests = sum(1 for r in self.test_results if r.status == TestStatus.ERROR)
        
        pass_rate = (passed_tests / total_tests) * 100 if total_tests > 0 else 0
        
        logger.info("=" * 50)
        logger.info("测试报告")
        logger.info("=" * 50)
        logger.info(f"总测试数: {total_tests}")
        logger.info(f"通过测试: {passed_tests}")
        logger.info(f"失败测试: {failed_tests}")
        logger.info(f"错误测试: {error_tests}")
        logger.info(f"通过率: {pass_rate:.2f}%")
        
        # 详细结果
        logger.info("\n详细测试结果:")
        for result in self.test_results:
            logger.info(f"  {result.test_name}: {result.status.value} ({result.duration:.2f}s)")
            if result.status != TestStatus.PASS:
                logger.info(f"    错误信息: {result.message}")
        
        # 保存测试报告到文件
        report_data = {
            "summary": {
                "total_tests": total_tests,
                "passed_tests": passed_tests,
                "failed_tests": failed_tests,
                "error_tests": error_tests,
                "pass_rate": pass_rate
            },
            "results": [
                {
                    "test_name": r.test_name,
                    "status": r.status.value,
                    "message": r.message,
                    "duration": r.duration,
                    "details": r.details
                }
                for r in self.test_results
            ]
        }
        
        with open('test_report.json', 'w', encoding='utf-8') as f:
            json.dump(report_data, f, ensure_ascii=False, indent=2)
        
        logger.info(f"\n测试报告已保存到: test_report.json")

def main():
    """主函数"""
    tester = ChinesePhoneTester()
    tester.run_all_tests()

if __name__ == "__main__":
    main()

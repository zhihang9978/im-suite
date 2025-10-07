#!/usr/bin/env python3
"""
志航密信兼容性测试脚本
针对中国用户手机进行兼容性测试
"""

import requests
import json
import time
import platform
import sys
from typing import Dict, List, Any, Tuple
from dataclasses import dataclass
from datetime import datetime
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('compatibility_test.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

@dataclass
class CompatibilityResult:
    """兼容性测试结果"""
    test_name: str
    brand: str
    system: str
    android_version: str
    status: str
    details: Dict[str, Any]
    timestamp: datetime

class ChinesePhoneCompatibilityTester:
    """中国手机品牌兼容性测试器"""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()
        self.compatibility_results: List[CompatibilityResult] = []
        
        # 中国手机品牌兼容性配置
        self.compatibility_matrix = {
            "xiaomi": {
                "name": "小米",
                "systems": {
                    "MIUI 12": {"android": ["Android 10", "Android 11"], "compatibility": "high"},
                    "MIUI 13": {"android": ["Android 11", "Android 12"], "compatibility": "high"},
                    "MIUI 14": {"android": ["Android 12", "Android 13"], "compatibility": "high"}
                },
                "features": ["权限管理", "电池优化", "通知管理", "悬浮窗"]
            },
            "huawei": {
                "name": "华为",
                "systems": {
                    "EMUI 11": {"android": ["Android 10", "Android 11"], "compatibility": "high"},
                    "EMUI 12": {"android": ["Android 11", "Android 12"], "compatibility": "high"},
                    "HarmonyOS 2": {"android": ["Android 10", "Android 11"], "compatibility": "medium"},
                    "HarmonyOS 3": {"android": ["Android 11", "Android 12"], "compatibility": "medium"}
                },
                "features": ["权限管理", "电池优化", "推送服务", "系统集成"]
            },
            "oppo": {
                "name": "OPPO",
                "systems": {
                    "ColorOS 11": {"android": ["Android 10", "Android 11"], "compatibility": "high"},
                    "ColorOS 12": {"android": ["Android 11", "Android 12"], "compatibility": "high"},
                    "ColorOS 13": {"android": ["Android 12", "Android 13"], "compatibility": "high"}
                },
                "features": ["权限管理", "电池优化", "通知管理", "主题系统"]
            },
            "vivo": {
                "name": "vivo",
                "systems": {
                    "OriginOS 1": {"android": ["Android 10", "Android 11"], "compatibility": "high"},
                    "OriginOS 2": {"android": ["Android 11", "Android 12"], "compatibility": "high"},
                    "OriginOS 3": {"android": ["Android 12", "Android 13"], "compatibility": "high"},
                    "FuntouchOS 11": {"android": ["Android 10", "Android 11"], "compatibility": "medium"},
                    "FuntouchOS 12": {"android": ["Android 11", "Android 12"], "compatibility": "medium"}
                },
                "features": ["权限管理", "电池优化", "通知管理", "系统优化"]
            },
            "oneplus": {
                "name": "一加",
                "systems": {
                    "OxygenOS 11": {"android": ["Android 10", "Android 11"], "compatibility": "high"},
                    "OxygenOS 12": {"android": ["Android 11", "Android 12"], "compatibility": "high"},
                    "OxygenOS 13": {"android": ["Android 12", "Android 13"], "compatibility": "high"}
                },
                "features": ["权限管理", "电池优化", "通知管理", "系统特性"]
            },
            "realme": {
                "name": "realme",
                "systems": {
                    "realme UI 2": {"android": ["Android 10", "Android 11"], "compatibility": "high"},
                    "realme UI 3": {"android": ["Android 11", "Android 12"], "compatibility": "high"},
                    "realme UI 4": {"android": ["Android 12", "Android 13"], "compatibility": "high"}
                },
                "features": ["权限管理", "电池优化", "通知管理", "系统优化"]
            },
            "meizu": {
                "name": "魅族",
                "systems": {
                    "Flyme 8": {"android": ["Android 10", "Android 11"], "compatibility": "medium"},
                    "Flyme 9": {"android": ["Android 11", "Android 12"], "compatibility": "medium"}
                },
                "features": ["权限管理", "电池优化", "通知管理", "系统特性"]
            }
        }
    
    def test_basic_compatibility(self, brand: str, system: str, android_version: str) -> Dict[str, Any]:
        """测试基础兼容性"""
        try:
            # 测试API连接
            response = self.session.get(f"{self.base_url}/api/ping", timeout=10)
            api_connectivity = response.status_code == 200
            
            # 测试用户认证
            auth_response = self.session.post(
                f"{self.base_url}/api/auth/login",
                json={"phone": "13800138000", "password": "test123456"},
                timeout=10
            )
            auth_compatibility = auth_response.status_code in [200, 401]  # 401表示认证失败但接口正常
            
            # 测试消息功能
            message_response = self.session.get(f"{self.base_url}/api/messages", timeout=10)
            message_compatibility = message_response.status_code in [200, 401, 403]
            
            return {
                "api_connectivity": api_connectivity,
                "auth_compatibility": auth_compatibility,
                "message_compatibility": message_compatibility,
                "overall_compatibility": api_connectivity and auth_compatibility and message_compatibility
            }
        except Exception as e:
            logger.error(f"基础兼容性测试失败 ({brand}): {e}")
            return {"overall_compatibility": False, "error": str(e)}
    
    def test_brand_specific_features(self, brand: str, system: str) -> Dict[str, Any]:
        """测试品牌特定功能"""
        try:
            brand_config = self.compatibility_matrix.get(brand)
            if not brand_config:
                return {"compatibility": False, "error": "品牌配置不存在"}
            
            features = brand_config.get("features", [])
            feature_results = {}
            
            for feature in features:
                if feature == "权限管理":
                    result = self.test_permission_compatibility(brand)
                elif feature == "电池优化":
                    result = self.test_battery_compatibility(brand)
                elif feature == "通知管理":
                    result = self.test_notification_compatibility(brand)
                elif feature == "悬浮窗":
                    result = self.test_floating_window_compatibility(brand)
                elif feature == "推送服务":
                    result = self.test_push_service_compatibility(brand)
                elif feature == "系统集成":
                    result = self.test_system_integration_compatibility(brand)
                elif feature == "主题系统":
                    result = self.test_theme_compatibility(brand)
                elif feature == "系统优化":
                    result = self.test_system_optimization_compatibility(brand)
                elif feature == "系统特性":
                    result = self.test_system_features_compatibility(brand)
                else:
                    result = {"compatibility": True, "message": f"{feature}功能正常"}
                
                feature_results[feature] = result
            
            # 计算整体兼容性
            compatible_features = sum(1 for result in feature_results.values() 
                                    if result.get("compatibility", False))
            total_features = len(feature_results)
            compatibility_score = (compatible_features / total_features) * 100 if total_features > 0 else 0
            
            return {
                "feature_results": feature_results,
                "compatibility_score": compatibility_score,
                "compatible_features": compatible_features,
                "total_features": total_features
            }
        except Exception as e:
            logger.error(f"品牌特定功能测试失败 ({brand}): {e}")
            return {"compatibility": False, "error": str(e)}
    
    def test_permission_compatibility(self, brand: str) -> Dict[str, Any]:
        """测试权限管理兼容性"""
        try:
            permission_data = {
                "brand": brand,
                "permissions": ["INTERNET", "ACCESS_NETWORK_STATE", "WAKE_LOCK", "VIBRATE"]
            }
            
            response = self.session.post(
                f"{self.base_url}/api/permissions/test",
                json=permission_data,
                timeout=10
            )
            
            return {
                "compatibility": response.status_code == 200,
                "status_code": response.status_code,
                "message": "权限管理兼容性测试完成"
            }
        except Exception as e:
            return {"compatibility": False, "error": str(e)}
    
    def test_battery_compatibility(self, brand: str) -> Dict[str, Any]:
        """测试电池优化兼容性"""
        try:
            battery_data = {
                "brand": brand,
                "optimization": "battery"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/optimization/battery",
                json=battery_data,
                timeout=10
            )
            
            return {
                "compatibility": response.status_code == 200,
                "status_code": response.status_code,
                "message": "电池优化兼容性测试完成"
            }
        except Exception as e:
            return {"compatibility": False, "error": str(e)}
    
    def test_notification_compatibility(self, brand: str) -> Dict[str, Any]:
        """测试通知管理兼容性"""
        try:
            notification_data = {
                "brand": brand,
                "type": "test"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/notifications/test",
                json=notification_data,
                timeout=10
            )
            
            return {
                "compatibility": response.status_code == 200,
                "status_code": response.status_code,
                "message": "通知管理兼容性测试完成"
            }
        except Exception as e:
            return {"compatibility": False, "error": str(e)}
    
    def test_floating_window_compatibility(self, brand: str) -> Dict[str, Any]:
        """测试悬浮窗兼容性"""
        try:
            floating_data = {
                "brand": brand,
                "feature": "floating_window"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/features/floating_window",
                json=floating_data,
                timeout=10
            )
            
            return {
                "compatibility": response.status_code == 200,
                "status_code": response.status_code,
                "message": "悬浮窗兼容性测试完成"
            }
        except Exception as e:
            return {"compatibility": False, "error": str(e)}
    
    def test_push_service_compatibility(self, brand: str) -> Dict[str, Any]:
        """测试推送服务兼容性"""
        try:
            push_data = {
                "brand": brand,
                "service": "push"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/push/test",
                json=push_data,
                timeout=10
            )
            
            return {
                "compatibility": response.status_code == 200,
                "status_code": response.status_code,
                "message": "推送服务兼容性测试完成"
            }
        except Exception as e:
            return {"compatibility": False, "error": str(e)}
    
    def test_system_integration_compatibility(self, brand: str) -> Dict[str, Any]:
        """测试系统集成兼容性"""
        try:
            integration_data = {
                "brand": brand,
                "integration": "system"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/integration/test",
                json=integration_data,
                timeout=10
            )
            
            return {
                "compatibility": response.status_code == 200,
                "status_code": response.status_code,
                "message": "系统集成兼容性测试完成"
            }
        except Exception as e:
            return {"compatibility": False, "error": str(e)}
    
    def test_theme_compatibility(self, brand: str) -> Dict[str, Any]:
        """测试主题系统兼容性"""
        try:
            theme_data = {
                "brand": brand,
                "theme": "default"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/theme/apply",
                json=theme_data,
                timeout=10
            )
            
            return {
                "compatibility": response.status_code == 200,
                "status_code": response.status_code,
                "message": "主题系统兼容性测试完成"
            }
        except Exception as e:
            return {"compatibility": False, "error": str(e)}
    
    def test_system_optimization_compatibility(self, brand: str) -> Dict[str, Any]:
        """测试系统优化兼容性"""
        try:
            optimization_data = {
                "brand": brand,
                "optimization": "system"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/optimization/system",
                json=optimization_data,
                timeout=10
            )
            
            return {
                "compatibility": response.status_code == 200,
                "status_code": response.status_code,
                "message": "系统优化兼容性测试完成"
            }
        except Exception as e:
            return {"compatibility": False, "error": str(e)}
    
    def test_system_features_compatibility(self, brand: str) -> Dict[str, Any]:
        """测试系统特性兼容性"""
        try:
            features_data = {
                "brand": brand,
                "features": ["gesture", "multitasking", "security"]
            }
            
            response = self.session.post(
                f"{self.base_url}/api/features/test",
                json=features_data,
                timeout=10
            )
            
            return {
                "compatibility": response.status_code == 200,
                "status_code": response.status_code,
                "message": "系统特性兼容性测试完成"
            }
        except Exception as e:
            return {"compatibility": False, "error": str(e)}
    
    def test_android_version_compatibility(self, android_version: str) -> Dict[str, Any]:
        """测试Android版本兼容性"""
        try:
            version_data = {
                "android_version": android_version,
                "api_level": self.get_api_level(android_version)
            }
            
            response = self.session.post(
                f"{self.base_url}/api/compatibility/android",
                json=version_data,
                timeout=10
            )
            
            return {
                "compatibility": response.status_code == 200,
                "android_version": android_version,
                "api_level": version_data["api_level"],
                "status_code": response.status_code
            }
        except Exception as e:
            return {"compatibility": False, "error": str(e)}
    
    def get_api_level(self, android_version: str) -> int:
        """获取Android API级别"""
        version_mapping = {
            "Android 10": 29,
            "Android 11": 30,
            "Android 12": 31,
            "Android 13": 33
        }
        return version_mapping.get(android_version, 29)
    
    def run_compatibility_tests(self):
        """运行所有兼容性测试"""
        logger.info("开始志航密信兼容性测试")
        logger.info("=" * 50)
        
        total_tests = 0
        passed_tests = 0
        
        for brand, brand_config in self.compatibility_matrix.items():
            brand_name = brand_config["name"]
            logger.info(f"\n测试品牌: {brand_name}")
            logger.info("-" * 30)
            
            for system, system_config in brand_config["systems"].items():
                for android_version in system_config["android"]:
                    total_tests += 1
                    
                    logger.info(f"  测试配置: {system} + {android_version}")
                    
                    # 基础兼容性测试
                    basic_result = self.test_basic_compatibility(brand, system, android_version)
                    basic_compatible = basic_result.get("overall_compatibility", False)
                    
                    # 品牌特定功能测试
                    feature_result = self.test_brand_specific_features(brand, system)
                    feature_compatible = feature_result.get("compatibility_score", 0) >= 80
                    
                    # Android版本兼容性测试
                    android_result = self.test_android_version_compatibility(android_version)
                    android_compatible = android_result.get("compatibility", False)
                    
                    # 整体兼容性
                    overall_compatible = basic_compatible and feature_compatible and android_compatible
                    
                    if overall_compatible:
                        passed_tests += 1
                        status = "✅ 兼容"
                    else:
                        status = "❌ 不兼容"
                    
                    logger.info(f"    基础兼容性: {'✅' if basic_compatible else '❌'}")
                    logger.info(f"    功能兼容性: {'✅' if feature_compatible else '❌'}")
                    logger.info(f"    系统兼容性: {'✅' if android_compatible else '❌'}")
                    logger.info(f"    整体结果: {status}")
                    
                    # 保存测试结果
                    result = CompatibilityResult(
                        test_name=f"{brand_name}_{system}_{android_version}",
                        brand=brand_name,
                        system=system,
                        android_version=android_version,
                        status="compatible" if overall_compatible else "incompatible",
                        details={
                            "basic_compatibility": basic_result,
                            "feature_compatibility": feature_result,
                            "android_compatibility": android_result
                        },
                        timestamp=datetime.now()
                    )
                    self.compatibility_results.append(result)
        
        # 生成兼容性报告
        self.generate_compatibility_report(total_tests, passed_tests)
    
    def generate_compatibility_report(self, total_tests: int, passed_tests: int):
        """生成兼容性测试报告"""
        pass_rate = (passed_tests / total_tests) * 100 if total_tests > 0 else 0
        
        logger.info("\n" + "=" * 50)
        logger.info("兼容性测试报告")
        logger.info("=" * 50)
        logger.info(f"总测试数: {total_tests}")
        logger.info(f"通过测试: {passed_tests}")
        logger.info(f"失败测试: {total_tests - passed_tests}")
        logger.info(f"通过率: {pass_rate:.2f}%")
        
        # 按品牌统计
        brand_stats = {}
        for result in self.compatibility_results:
            brand = result.brand
            if brand not in brand_stats:
                brand_stats[brand] = {"total": 0, "passed": 0}
            brand_stats[brand]["total"] += 1
            if result.status == "compatible":
                brand_stats[brand]["passed"] += 1
        
        logger.info("\n品牌兼容性统计:")
        for brand, stats in brand_stats.items():
            brand_pass_rate = (stats["passed"] / stats["total"]) * 100 if stats["total"] > 0 else 0
            logger.info(f"  {brand}: {stats['passed']}/{stats['total']} ({brand_pass_rate:.1f}%)")
        
        # 保存详细报告
        report_data = {
            "summary": {
                "total_tests": total_tests,
                "passed_tests": passed_tests,
                "failed_tests": total_tests - passed_tests,
                "pass_rate": pass_rate
            },
            "brand_stats": brand_stats,
            "detailed_results": [
                {
                    "test_name": r.test_name,
                    "brand": r.brand,
                    "system": r.system,
                    "android_version": r.android_version,
                    "status": r.status,
                    "details": r.details,
                    "timestamp": r.timestamp.isoformat()
                }
                for r in self.compatibility_results
            ]
        }
        
        with open('compatibility_report.json', 'w', encoding='utf-8') as f:
            json.dump(report_data, f, ensure_ascii=False, indent=2)
        
        logger.info(f"\n兼容性测试报告已保存到: compatibility_report.json")

def main():
    """主函数"""
    tester = ChinesePhoneCompatibilityTester()
    tester.run_compatibility_tests()

if __name__ == "__main__":
    main()

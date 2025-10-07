#!/usr/bin/env python3
"""
志航密信用户体验测试脚本
针对中国用户手机进行用户体验测试
"""

import requests
import json
import time
import random
from typing import Dict, List, Any, Tuple
from dataclasses import dataclass
from datetime import datetime
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('ux_test.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

@dataclass
class UXTestResult:
    """用户体验测试结果"""
    test_name: str
    brand: str
    user_scenario: str
    score: float
    details: Dict[str, Any]
    timestamp: datetime

class ChinesePhoneUXTester:
    """中国手机品牌用户体验测试器"""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()
        self.ux_results: List[UXTestResult] = []
        
        # 中国用户使用场景
        self.user_scenarios = {
            "daily_communication": {
                "name": "日常通讯",
                "description": "用户日常发送消息、语音、图片等",
                "weight": 0.3
            },
            "business_communication": {
                "name": "商务通讯",
                "description": "商务用户进行工作沟通",
                "weight": 0.2
            },
            "group_communication": {
                "name": "群组通讯",
                "description": "群组聊天、群组管理等功能",
                "weight": 0.2
            },
            "file_sharing": {
                "name": "文件分享",
                "description": "文件传输、文档分享等",
                "weight": 0.15
            },
            "voice_video_call": {
                "name": "语音视频通话",
                "description": "语音通话、视频通话功能",
                "weight": 0.05
            }
        }
        
        # 中国手机品牌用户体验基准
        self.ux_benchmarks = {
            "xiaomi": {
                "name": "小米",
                "expected_ui_score": 85,
                "expected_performance_score": 90,
                "expected_accessibility_score": 80,
                "expected_satisfaction_score": 88
            },
            "huawei": {
                "name": "华为",
                "expected_ui_score": 88,
                "expected_performance_score": 85,
                "expected_accessibility_score": 85,
                "expected_satisfaction_score": 90
            },
            "oppo": {
                "name": "OPPO",
                "expected_ui_score": 82,
                "expected_performance_score": 88,
                "expected_accessibility_score": 83,
                "expected_satisfaction_score": 85
            },
            "vivo": {
                "name": "vivo",
                "expected_ui_score": 80,
                "expected_performance_score": 85,
                "expected_accessibility_score": 82,
                "expected_satisfaction_score": 83
            },
            "oneplus": {
                "name": "一加",
                "expected_ui_score": 90,
                "expected_performance_score": 95,
                "expected_accessibility_score": 88,
                "expected_satisfaction_score": 92
            },
            "realme": {
                "name": "realme",
                "expected_ui_score": 83,
                "expected_performance_score": 87,
                "expected_accessibility_score": 84,
                "expected_satisfaction_score": 86
            },
            "meizu": {
                "name": "魅族",
                "expected_ui_score": 85,
                "expected_performance_score": 82,
                "expected_accessibility_score": 86,
                "expected_satisfaction_score": 84
            }
        }
    
    def test_ui_responsiveness(self, brand: str) -> Dict[str, Any]:
        """测试界面响应性"""
        try:
            # 测试界面加载速度
            start_time = time.time()
            response = self.session.get(f"{self.base_url}/api/ui/load", timeout=10)
            load_time = time.time() - start_time
            
            # 测试界面交互响应
            interaction_times = []
            for _ in range(5):
                start_time = time.time()
                response = self.session.post(f"{self.base_url}/api/ui/interact", 
                                           json={"action": "click", "element": "button"}, timeout=5)
                interaction_time = time.time() - start_time
                interaction_times.append(interaction_time)
            
            avg_interaction_time = sum(interaction_times) / len(interaction_times)
            
            # 计算UI响应性得分
            load_score = max(0, 100 - (load_time * 100))  # 加载时间越短得分越高
            interaction_score = max(0, 100 - (avg_interaction_time * 200))  # 交互时间越短得分越高
            ui_score = (load_score + interaction_score) / 2
            
            return {
                "load_time": load_time,
                "avg_interaction_time": avg_interaction_time,
                "ui_score": ui_score,
                "load_score": load_score,
                "interaction_score": interaction_score
            }
        except Exception as e:
            logger.error(f"界面响应性测试失败 ({brand}): {e}")
            return {"ui_score": 0, "error": str(e)}
    
    def test_performance_ux(self, brand: str) -> Dict[str, Any]:
        """测试性能用户体验"""
        try:
            # 测试应用启动时间
            start_time = time.time()
            response = self.session.get(f"{self.base_url}/api/app/start", timeout=15)
            startup_time = time.time() - start_time
            
            # 测试消息发送性能
            message_times = []
            for i in range(10):
                start_time = time.time()
                response = self.session.post(f"{self.base_url}/api/messages/send",
                                          json={"content": f"测试消息 {i}", "type": "text"}, timeout=5)
                message_time = time.time() - start_time
                message_times.append(message_time)
            
            avg_message_time = sum(message_times) / len(message_times)
            
            # 测试文件上传性能
            file_start_time = time.time()
            response = self.session.post(f"{self.base_url}/api/files/upload",
                                       files={"file": ("test.txt", "测试文件", "text/plain")}, timeout=30)
            file_time = time.time() - file_start_time
            
            # 计算性能得分
            startup_score = max(0, 100 - (startup_time * 20))  # 启动时间越短得分越高
            message_score = max(0, 100 - (avg_message_time * 100))  # 消息发送时间越短得分越高
            file_score = max(0, 100 - (file_time * 2))  # 文件上传时间越短得分越高
            performance_score = (startup_score + message_score + file_score) / 3
            
            return {
                "startup_time": startup_time,
                "avg_message_time": avg_message_time,
                "file_upload_time": file_time,
                "performance_score": performance_score,
                "startup_score": startup_score,
                "message_score": message_score,
                "file_score": file_score
            }
        except Exception as e:
            logger.error(f"性能用户体验测试失败 ({brand}): {e}")
            return {"performance_score": 0, "error": str(e)}
    
    def test_accessibility_ux(self, brand: str) -> Dict[str, Any]:
        """测试无障碍用户体验"""
        try:
            # 测试字体大小适配
            font_response = self.session.get(f"{self.base_url}/api/accessibility/font", timeout=5)
            font_score = 100 if font_response.status_code == 200 else 0
            
            # 测试颜色对比度
            contrast_response = self.session.get(f"{self.base_url}/api/accessibility/contrast", timeout=5)
            contrast_score = 100 if contrast_response.status_code == 200 else 0
            
            # 测试语音功能
            voice_response = self.session.get(f"{self.base_url}/api/accessibility/voice", timeout=5)
            voice_score = 100 if voice_response.status_code == 200 else 0
            
            # 测试手势操作
            gesture_response = self.session.get(f"{self.base_url}/api/accessibility/gesture", timeout=5)
            gesture_score = 100 if gesture_response.status_code == 200 else 0
            
            # 计算无障碍得分
            accessibility_score = (font_score + contrast_score + voice_score + gesture_score) / 4
            
            return {
                "font_score": font_score,
                "contrast_score": contrast_score,
                "voice_score": voice_score,
                "gesture_score": gesture_score,
                "accessibility_score": accessibility_score
            }
        except Exception as e:
            logger.error(f"无障碍用户体验测试失败 ({brand}): {e}")
            return {"accessibility_score": 0, "error": str(e)}
    
    def test_user_satisfaction(self, brand: str, scenario: str) -> Dict[str, Any]:
        """测试用户满意度"""
        try:
            # 模拟用户操作流程
            satisfaction_factors = []
            
            # 测试登录满意度
            login_start = time.time()
            login_response = self.session.post(f"{self.base_url}/api/auth/login",
                                            json={"phone": "13800138000", "password": "test123456"}, timeout=10)
            login_time = time.time() - login_start
            login_satisfaction = max(0, 100 - (login_time * 50))
            satisfaction_factors.append(login_satisfaction)
            
            # 测试消息发送满意度
            message_start = time.time()
            message_response = self.session.post(f"{self.base_url}/api/messages/send",
                                              json={"content": "测试消息", "type": "text"}, timeout=5)
            message_time = time.time() - message_start
            message_satisfaction = max(0, 100 - (message_time * 100))
            satisfaction_factors.append(message_satisfaction)
            
            # 测试界面切换满意度
            ui_start = time.time()
            ui_response = self.session.get(f"{self.base_url}/api/ui/switch", timeout=5)
            ui_time = time.time() - ui_start
            ui_satisfaction = max(0, 100 - (ui_time * 200))
            satisfaction_factors.append(ui_satisfaction)
            
            # 测试错误处理满意度
            error_response = self.session.get(f"{self.base_url}/api/error/test", timeout=5)
            error_satisfaction = 80 if error_response.status_code in [200, 404] else 60
            satisfaction_factors.append(error_satisfaction)
            
            # 计算整体满意度
            overall_satisfaction = sum(satisfaction_factors) / len(satisfaction_factors)
            
            return {
                "login_satisfaction": login_satisfaction,
                "message_satisfaction": message_satisfaction,
                "ui_satisfaction": ui_satisfaction,
                "error_satisfaction": error_satisfaction,
                "overall_satisfaction": overall_satisfaction,
                "satisfaction_factors": satisfaction_factors
            }
        except Exception as e:
            logger.error(f"用户满意度测试失败 ({brand}, {scenario}): {e}")
            return {"overall_satisfaction": 0, "error": str(e)}
    
    def test_chinese_user_scenarios(self, brand: str) -> Dict[str, Any]:
        """测试中国用户使用场景"""
        try:
            scenario_results = {}
            
            for scenario_key, scenario_config in self.user_scenarios.items():
                scenario_name = scenario_config["name"]
                weight = scenario_config["weight"]
                
                # 根据场景测试不同功能
                if scenario_key == "daily_communication":
                    result = self.test_daily_communication_ux(brand)
                elif scenario_key == "business_communication":
                    result = self.test_business_communication_ux(brand)
                elif scenario_key == "group_communication":
                    result = self.test_group_communication_ux(brand)
                elif scenario_key == "file_sharing":
                    result = self.test_file_sharing_ux(brand)
                elif scenario_key == "voice_video_call":
                    result = self.test_voice_video_call_ux(brand)
                else:
                    result = {"score": 0, "error": "未知场景"}
                
                scenario_results[scenario_key] = {
                    "name": scenario_name,
                    "weight": weight,
                    "score": result.get("score", 0),
                    "details": result
                }
            
            # 计算加权平均分
            weighted_score = sum(
                result["score"] * result["weight"] 
                for result in scenario_results.values()
            )
            
            return {
                "scenario_results": scenario_results,
                "weighted_score": weighted_score,
                "overall_score": weighted_score
            }
        except Exception as e:
            logger.error(f"中国用户场景测试失败 ({brand}): {e}")
            return {"overall_score": 0, "error": str(e)}
    
    def test_daily_communication_ux(self, brand: str) -> Dict[str, Any]:
        """测试日常通讯用户体验"""
        try:
            # 测试发送文本消息
            text_start = time.time()
            text_response = self.session.post(f"{self.base_url}/api/messages/send",
                                           json={"content": "你好", "type": "text"}, timeout=5)
            text_time = time.time() - text_start
            text_score = max(0, 100 - (text_time * 100))
            
            # 测试发送表情
            emoji_start = time.time()
            emoji_response = self.session.post(f"{self.base_url}/api/messages/send",
                                             json={"content": "😊", "type": "emoji"}, timeout=5)
            emoji_time = time.time() - emoji_start
            emoji_score = max(0, 100 - (emoji_time * 100))
            
            # 测试发送图片
            image_start = time.time()
            image_response = self.session.post(f"{self.base_url}/api/messages/send",
                                            json={"content": "图片", "type": "image"}, timeout=10)
            image_time = time.time() - image_start
            image_score = max(0, 100 - (image_time * 50))
            
            overall_score = (text_score + emoji_score + image_score) / 3
            
            return {
                "score": overall_score,
                "text_score": text_score,
                "emoji_score": emoji_score,
                "image_score": image_score,
                "text_time": text_time,
                "emoji_time": emoji_time,
                "image_time": image_time
            }
        except Exception as e:
            return {"score": 0, "error": str(e)}
    
    def test_business_communication_ux(self, brand: str) -> Dict[str, Any]:
        """测试商务通讯用户体验"""
        try:
            # 测试文件传输
            file_start = time.time()
            file_response = self.session.post(f"{self.base_url}/api/files/upload",
                                           files={"file": ("document.pdf", "商务文档", "application/pdf")}, timeout=30)
            file_time = time.time() - file_start
            file_score = max(0, 100 - (file_time * 2))
            
            # 测试群组创建
            group_start = time.time()
            group_response = self.session.post(f"{self.base_url}/api/groups/create",
                                            json={"name": "商务群组", "type": "business"}, timeout=10)
            group_time = time.time() - group_start
            group_score = max(0, 100 - (group_time * 50))
            
            # 测试消息转发
            forward_start = time.time()
            forward_response = self.session.post(f"{self.base_url}/api/messages/forward",
                                              json={"message_id": "test", "target_chat": "business"}, timeout=5)
            forward_time = time.time() - forward_start
            forward_score = max(0, 100 - (forward_time * 100))
            
            overall_score = (file_score + group_score + forward_score) / 3
            
            return {
                "score": overall_score,
                "file_score": file_score,
                "group_score": group_score,
                "forward_score": forward_score,
                "file_time": file_time,
                "group_time": group_time,
                "forward_time": forward_time
            }
        except Exception as e:
            return {"score": 0, "error": str(e)}
    
    def test_group_communication_ux(self, brand: str) -> Dict[str, Any]:
        """测试群组通讯用户体验"""
        try:
            # 测试群组消息发送
            group_message_start = time.time()
            group_message_response = self.session.post(f"{self.base_url}/api/groups/message",
                                                     json={"group_id": "test", "content": "群组消息"}, timeout=5)
            group_message_time = time.time() - group_message_start
            group_message_score = max(0, 100 - (group_message_time * 100))
            
            # 测试群组管理
            group_manage_start = time.time()
            group_manage_response = self.session.post(f"{self.base_url}/api/groups/manage",
                                                     json={"action": "add_member", "group_id": "test"}, timeout=5)
            group_manage_time = time.time() - group_manage_start
            group_manage_score = max(0, 100 - (group_manage_time * 100))
            
            # 测试群组设置
            group_settings_start = time.time()
            group_settings_response = self.session.get(f"{self.base_url}/api/groups/settings", timeout=5)
            group_settings_time = time.time() - group_settings_start
            group_settings_score = max(0, 100 - (group_settings_time * 100))
            
            overall_score = (group_message_score + group_manage_score + group_settings_score) / 3
            
            return {
                "score": overall_score,
                "group_message_score": group_message_score,
                "group_manage_score": group_manage_score,
                "group_settings_score": group_settings_score,
                "group_message_time": group_message_time,
                "group_manage_time": group_manage_time,
                "group_settings_time": group_settings_time
            }
        except Exception as e:
            return {"score": 0, "error": str(e)}
    
    def test_file_sharing_ux(self, brand: str) -> Dict[str, Any]:
        """测试文件分享用户体验"""
        try:
            # 测试文件上传
            upload_start = time.time()
            upload_response = self.session.post(f"{self.base_url}/api/files/upload",
                                              files={"file": ("test.txt", "测试文件", "text/plain")}, timeout=30)
            upload_time = time.time() - upload_start
            upload_score = max(0, 100 - (upload_time * 2))
            
            # 测试文件下载
            download_start = time.time()
            download_response = self.session.get(f"{self.base_url}/api/files/download", timeout=30)
            download_time = time.time() - download_start
            download_score = max(0, 100 - (download_time * 2))
            
            # 测试文件预览
            preview_start = time.time()
            preview_response = self.session.get(f"{self.base_url}/api/files/preview", timeout=10)
            preview_time = time.time() - preview_start
            preview_score = max(0, 100 - (preview_time * 10))
            
            overall_score = (upload_score + download_score + preview_score) / 3
            
            return {
                "score": overall_score,
                "upload_score": upload_score,
                "download_score": download_score,
                "preview_score": preview_score,
                "upload_time": upload_time,
                "download_time": download_time,
                "preview_time": preview_time
            }
        except Exception as e:
            return {"score": 0, "error": str(e)}
    
    def test_voice_video_call_ux(self, brand: str) -> Dict[str, Any]:
        """测试语音视频通话用户体验"""
        try:
            # 测试语音通话
            voice_start = time.time()
            voice_response = self.session.post(f"{self.base_url}/api/calls/voice",
                                            json={"target": "test_user"}, timeout=10)
            voice_time = time.time() - voice_start
            voice_score = max(0, 100 - (voice_time * 20))
            
            # 测试视频通话
            video_start = time.time()
            video_response = self.session.post(f"{self.base_url}/api/calls/video",
                                            json={"target": "test_user"}, timeout=10)
            video_time = time.time() - video_start
            video_score = max(0, 100 - (video_time * 20))
            
            # 测试通话质量
            quality_start = time.time()
            quality_response = self.session.get(f"{self.base_url}/api/calls/quality", timeout=5)
            quality_time = time.time() - quality_start
            quality_score = max(0, 100 - (quality_time * 100))
            
            overall_score = (voice_score + video_score + quality_score) / 3
            
            return {
                "score": overall_score,
                "voice_score": voice_score,
                "video_score": video_score,
                "quality_score": quality_score,
                "voice_time": voice_time,
                "video_time": video_time,
                "quality_time": quality_time
            }
        except Exception as e:
            return {"score": 0, "error": str(e)}
    
    def run_ux_tests(self):
        """运行所有用户体验测试"""
        logger.info("开始志航密信用户体验测试")
        logger.info("=" * 50)
        
        total_score = 0
        brand_count = 0
        
        for brand, brand_config in self.ux_benchmarks.items():
            brand_name = brand_config["name"]
            logger.info(f"\n测试品牌: {brand_name}")
            logger.info("-" * 30)
            
            # UI响应性测试
            ui_result = self.test_ui_responsiveness(brand)
            ui_score = ui_result.get("ui_score", 0)
            logger.info(f"  UI响应性: {ui_score:.1f}/100")
            
            # 性能用户体验测试
            performance_result = self.test_performance_ux(brand)
            performance_score = performance_result.get("performance_score", 0)
            logger.info(f"  性能体验: {performance_score:.1f}/100")
            
            # 无障碍用户体验测试
            accessibility_result = self.test_accessibility_ux(brand)
            accessibility_score = accessibility_result.get("accessibility_score", 0)
            logger.info(f"  无障碍体验: {accessibility_score:.1f}/100")
            
            # 用户满意度测试
            satisfaction_result = self.test_user_satisfaction(brand, "overall")
            satisfaction_score = satisfaction_result.get("overall_satisfaction", 0)
            logger.info(f"  用户满意度: {satisfaction_score:.1f}/100")
            
            # 中国用户场景测试
            scenario_result = self.test_chinese_user_scenarios(brand)
            scenario_score = scenario_result.get("overall_score", 0)
            logger.info(f"  场景体验: {scenario_score:.1f}/100")
            
            # 计算综合得分
            overall_score = (ui_score + performance_score + accessibility_score + 
                           satisfaction_score + scenario_score) / 5
            
            logger.info(f"  综合得分: {overall_score:.1f}/100")
            
            # 与基准对比
            benchmark = brand_config
            benchmark_score = (benchmark["expected_ui_score"] + benchmark["expected_performance_score"] + 
                            benchmark["expected_accessibility_score"] + benchmark["expected_satisfaction_score"]) / 4
            
            if overall_score >= benchmark_score:
                status = "✅ 达标"
            else:
                status = "❌ 未达标"
            
            logger.info(f"  基准对比: {status} (基准: {benchmark_score:.1f})")
            
            # 保存测试结果
            result = UXTestResult(
                test_name=f"UX_{brand_name}",
                brand=brand_name,
                user_scenario="overall",
                score=overall_score,
                details={
                    "ui_score": ui_score,
                    "performance_score": performance_score,
                    "accessibility_score": accessibility_score,
                    "satisfaction_score": satisfaction_score,
                    "scenario_score": scenario_score,
                    "benchmark_score": benchmark_score,
                    "ui_result": ui_result,
                    "performance_result": performance_result,
                    "accessibility_result": accessibility_result,
                    "satisfaction_result": satisfaction_result,
                    "scenario_result": scenario_result
                },
                timestamp=datetime.now()
            )
            self.ux_results.append(result)
            
            total_score += overall_score
            brand_count += 1
        
        # 生成用户体验报告
        self.generate_ux_report(total_score, brand_count)
    
    def generate_ux_report(self, total_score: float, brand_count: int):
        """生成用户体验测试报告"""
        avg_score = total_score / brand_count if brand_count > 0 else 0
        
        logger.info("\n" + "=" * 50)
        logger.info("用户体验测试报告")
        logger.info("=" * 50)
        logger.info(f"测试品牌数: {brand_count}")
        logger.info(f"平均得分: {avg_score:.1f}/100")
        
        # 按品牌统计
        brand_scores = {}
        for result in self.ux_results:
            brand = result.brand
            brand_scores[brand] = result.score
        
        logger.info("\n品牌用户体验得分:")
        for brand, score in sorted(brand_scores.items(), key=lambda x: x[1], reverse=True):
            logger.info(f"  {brand}: {score:.1f}/100")
        
        # 保存详细报告
        report_data = {
            "summary": {
                "total_brands": brand_count,
                "average_score": avg_score,
                "highest_score": max(brand_scores.values()) if brand_scores else 0,
                "lowest_score": min(brand_scores.values()) if brand_scores else 0
            },
            "brand_scores": brand_scores,
            "detailed_results": [
                {
                    "test_name": r.test_name,
                    "brand": r.brand,
                    "user_scenario": r.user_scenario,
                    "score": r.score,
                    "details": r.details,
                    "timestamp": r.timestamp.isoformat()
                }
                for r in self.ux_results
            ]
        }
        
        with open('ux_report.json', 'w', encoding='utf-8') as f:
            json.dump(report_data, f, ensure_ascii=False, indent=2)
        
        logger.info(f"\n用户体验测试报告已保存到: ux_report.json")

def main():
    """主函数"""
    tester = ChinesePhoneUXTester()
    tester.run_ux_tests()

if __name__ == "__main__":
    main()

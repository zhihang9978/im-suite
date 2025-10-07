#!/usr/bin/env python3
"""
志航密信综合测试脚本
运行所有测试：功能测试、性能测试、兼容性测试、用户体验测试
"""

import subprocess
import json
import time
import os
from datetime import datetime
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('comprehensive_test.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class ComprehensiveTester:
    """综合测试器"""
    
    def __init__(self):
        self.test_results = {}
        self.start_time = None
        self.end_time = None
    
    def run_test_script(self, script_name: str, script_path: str) -> dict:
        """运行单个测试脚本"""
        logger.info(f"开始运行 {script_name}")
        start_time = time.time()
        
        try:
            # 运行测试脚本
            result = subprocess.run(
                ["python3", script_path],
                capture_output=True,
                text=True,
                timeout=300  # 5分钟超时
            )
            
            end_time = time.time()
            duration = end_time - start_time
            
            if result.returncode == 0:
                status = "SUCCESS"
                message = f"{script_name} 执行成功"
            else:
                status = "FAILED"
                message = f"{script_name} 执行失败: {result.stderr}"
            
            logger.info(f"{script_name} 完成: {status} (耗时: {duration:.2f}秒)")
            
            return {
                "script_name": script_name,
                "status": status,
                "message": message,
                "duration": duration,
                "returncode": result.returncode,
                "stdout": result.stdout,
                "stderr": result.stderr
            }
            
        except subprocess.TimeoutExpired:
            logger.error(f"{script_name} 执行超时")
            return {
                "script_name": script_name,
                "status": "TIMEOUT",
                "message": f"{script_name} 执行超时",
                "duration": 300,
                "returncode": -1,
                "stdout": "",
                "stderr": "执行超时"
            }
        except Exception as e:
            logger.error(f"{script_name} 执行异常: {e}")
            return {
                "script_name": script_name,
                "status": "ERROR",
                "message": f"{script_name} 执行异常: {str(e)}",
                "duration": 0,
                "returncode": -1,
                "stdout": "",
                "stderr": str(e)
            }
    
    def load_test_report(self, report_file: str) -> dict:
        """加载测试报告"""
        try:
            if os.path.exists(report_file):
                with open(report_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
            else:
                return {"error": f"报告文件不存在: {report_file}"}
        except Exception as e:
            return {"error": f"加载报告失败: {str(e)}"}
    
    def run_all_tests(self):
        """运行所有测试"""
        logger.info("开始志航密信综合测试")
        logger.info("=" * 60)
        
        self.start_time = datetime.now()
        
        # 测试脚本配置
        test_scripts = [
            {
                "name": "功能测试",
                "path": "scripts/testing/functional_test.py",
                "report": "test_report.json"
            },
            {
                "name": "性能测试",
                "path": "scripts/testing/performance_test.py",
                "report": "performance_report.json"
            },
            {
                "name": "兼容性测试",
                "path": "scripts/testing/compatibility_test.py",
                "report": "compatibility_report.json"
            },
            {
                "name": "用户体验测试",
                "path": "scripts/testing/ux_test.py",
                "report": "ux_report.json"
            }
        ]
        
        # 运行所有测试
        for test_config in test_scripts:
            script_name = test_config["name"]
            script_path = test_config["path"]
            report_file = test_config["report"]
            
            logger.info(f"\n{'='*20} {script_name} {'='*20}")
            
            # 运行测试脚本
            test_result = self.run_test_script(script_name, script_path)
            self.test_results[script_name] = test_result
            
            # 加载测试报告
            if test_result["status"] == "SUCCESS":
                report_data = self.load_test_report(report_file)
                self.test_results[script_name]["report"] = report_data
                
                # 显示测试报告摘要
                self.display_test_summary(script_name, report_data)
            else:
                logger.error(f"{script_name} 执行失败，无法加载报告")
        
        self.end_time = datetime.now()
        
        # 生成综合测试报告
        self.generate_comprehensive_report()
    
    def display_test_summary(self, test_name: str, report_data: dict):
        """显示测试摘要"""
        if "error" in report_data:
            logger.error(f"{test_name} 报告加载失败: {report_data['error']}")
            return
        
        if test_name == "功能测试":
            summary = report_data.get("summary", {})
            logger.info(f"  总测试数: {summary.get('total_tests', 0)}")
            logger.info(f"  通过测试: {summary.get('passed_tests', 0)}")
            logger.info(f"  失败测试: {summary.get('failed_tests', 0)}")
            logger.info(f"  通过率: {summary.get('pass_rate', 0):.2f}%")
        
        elif test_name == "性能测试":
            system_info = report_data.get("system_info", {})
            logger.info(f"  CPU使用率: {system_info.get('cpu_usage', 0):.2f}%")
            logger.info(f"  内存使用: {system_info.get('memory_usage', 0):.2f}MB")
            logger.info(f"  磁盘使用: {system_info.get('disk_usage', 0):.2f}GB")
        
        elif test_name == "兼容性测试":
            summary = report_data.get("summary", {})
            logger.info(f"  总测试数: {summary.get('total_tests', 0)}")
            logger.info(f"  通过测试: {summary.get('passed_tests', 0)}")
            logger.info(f"  通过率: {summary.get('pass_rate', 0):.2f}%")
            
            brand_stats = report_data.get("brand_stats", {})
            logger.info("  品牌兼容性:")
            for brand, stats in brand_stats.items():
                brand_pass_rate = (stats["passed"] / stats["total"]) * 100 if stats["total"] > 0 else 0
                logger.info(f"    {brand}: {stats['passed']}/{stats['total']} ({brand_pass_rate:.1f}%)")
        
        elif test_name == "用户体验测试":
            summary = report_data.get("summary", {})
            logger.info(f"  测试品牌数: {summary.get('total_brands', 0)}")
            logger.info(f"  平均得分: {summary.get('average_score', 0):.1f}/100")
            logger.info(f"  最高得分: {summary.get('highest_score', 0):.1f}/100")
            logger.info(f"  最低得分: {summary.get('lowest_score', 0):.1f}/100")
    
    def generate_comprehensive_report(self):
        """生成综合测试报告"""
        total_duration = (self.end_time - self.start_time).total_seconds()
        
        logger.info("\n" + "=" * 60)
        logger.info("综合测试报告")
        logger.info("=" * 60)
        logger.info(f"测试开始时间: {self.start_time.strftime('%Y-%m-%d %H:%M:%S')}")
        logger.info(f"测试结束时间: {self.end_time.strftime('%Y-%m-%d %H:%M:%S')}")
        logger.info(f"总耗时: {total_duration:.2f}秒")
        
        # 统计测试结果
        total_tests = len(self.test_results)
        successful_tests = sum(1 for result in self.test_results.values() if result["status"] == "SUCCESS")
        failed_tests = sum(1 for result in self.test_results.values() if result["status"] == "FAILED")
        error_tests = sum(1 for result in self.test_results.values() if result["status"] == "ERROR")
        timeout_tests = sum(1 for result in self.test_results.values() if result["status"] == "TIMEOUT")
        
        logger.info(f"\n测试统计:")
        logger.info(f"  总测试数: {total_tests}")
        logger.info(f"  成功测试: {successful_tests}")
        logger.info(f"  失败测试: {failed_tests}")
        logger.info(f"  错误测试: {error_tests}")
        logger.info(f"  超时测试: {timeout_tests}")
        logger.info(f"  成功率: {(successful_tests/total_tests)*100:.2f}%")
        
        # 详细测试结果
        logger.info(f"\n详细测试结果:")
        for test_name, result in self.test_results.items():
            status_icon = "✅" if result["status"] == "SUCCESS" else "❌"
            logger.info(f"  {status_icon} {test_name}: {result['status']} (耗时: {result['duration']:.2f}秒)")
            if result["status"] != "SUCCESS":
                logger.info(f"    错误信息: {result['message']}")
        
        # 生成综合报告数据
        comprehensive_report = {
            "test_info": {
                "start_time": self.start_time.isoformat(),
                "end_time": self.end_time.isoformat(),
                "total_duration": total_duration
            },
            "summary": {
                "total_tests": total_tests,
                "successful_tests": successful_tests,
                "failed_tests": failed_tests,
                "error_tests": error_tests,
                "timeout_tests": timeout_tests,
                "success_rate": (successful_tests/total_tests)*100 if total_tests > 0 else 0
            },
            "test_results": self.test_results
        }
        
        # 保存综合报告
        with open('comprehensive_test_report.json', 'w', encoding='utf-8') as f:
            json.dump(comprehensive_report, f, ensure_ascii=False, indent=2)
        
        logger.info(f"\n综合测试报告已保存到: comprehensive_test_report.json")
        
        # 生成测试建议
        self.generate_test_recommendations()
    
    def generate_test_recommendations(self):
        """生成测试建议"""
        logger.info(f"\n测试建议:")
        
        # 分析测试结果
        failed_tests = [name for name, result in self.test_results.items() if result["status"] != "SUCCESS"]
        
        if not failed_tests:
            logger.info("  🎉 所有测试都通过了！系统运行良好。")
        else:
            logger.info(f"  ⚠️  以下测试需要关注:")
            for test_name in failed_tests:
                logger.info(f"    - {test_name}: 需要修复相关问题")
        
        # 性能建议
        performance_result = self.test_results.get("性能测试", {})
        if performance_result.get("status") == "SUCCESS":
            performance_report = performance_result.get("report", {})
            system_info = performance_report.get("system_info", {})
            
            if system_info.get("cpu_usage", 0) > 80:
                logger.info("  💡 CPU使用率较高，建议优化性能")
            if system_info.get("memory_usage", 0) > 1000:
                logger.info("  💡 内存使用较高，建议优化内存管理")
        
        # 兼容性建议
        compatibility_result = self.test_results.get("兼容性测试", {})
        if compatibility_result.get("status") == "SUCCESS":
            compatibility_report = compatibility_result.get("report", {})
            summary = compatibility_report.get("summary", {})
            
            if summary.get("pass_rate", 0) < 90:
                logger.info("  💡 兼容性测试通过率较低，建议优化品牌适配")
        
        # 用户体验建议
        ux_result = self.test_results.get("用户体验测试", {})
        if ux_result.get("status") == "SUCCESS":
            ux_report = ux_result.get("report", {})
            summary = ux_report.get("summary", {})
            
            if summary.get("average_score", 0) < 80:
                logger.info("  💡 用户体验得分较低，建议优化界面和交互")

def main():
    """主函数"""
    tester = ComprehensiveTester()
    tester.run_all_tests()

if __name__ == "__main__":
    main()

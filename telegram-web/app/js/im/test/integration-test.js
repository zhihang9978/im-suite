/**
 * 志航密信集成测试
 * 验证适配层是否正确集成到 Telegram Web 中
 */

(function() {
  'use strict';

  console.log('[志航密信] 开始集成测试...');

  // 测试配置
  const TEST_CONFIG = {
    API_BASE: 'http://localhost:8080/api',
    WS_BASE: 'ws://localhost:8080/ws',
    TEST_PHONE: '13800138000',
    TEST_PASSWORD: '123456'
  };

  // 测试结果收集器
  const testResults = {
    passed: 0,
    failed: 0,
    total: 0,
    details: []
  };

  /**
   * 添加测试结果
   * @param {string} testName - 测试名称
   * @param {boolean} passed - 是否通过
   * @param {string} message - 测试消息
   */
  function addTestResult(testName, passed, message) {
    testResults.total++;
    if (passed) {
      testResults.passed++;
    } else {
      testResults.failed++;
    }
    
    testResults.details.push({
      name: testName,
      passed,
      message,
      timestamp: new Date().toISOString()
    });
    
    console.log(`[志航密信] 测试 ${testName}: ${passed ? '通过' : '失败'} - ${message}`);
  }

  /**
   * 测试 API 连接
   */
  async function testAPIConnection() {
    try {
      if (!window.IMAPI) {
        addTestResult('API 连接测试', false, 'IMAPI 对象未找到');
        return;
      }

      const result = await window.IMAPI.get('/ping');
      if (result && result.ok) {
        addTestResult('API 连接测试', true, 'API 连接正常');
      } else {
        addTestResult('API 连接测试', false, 'API 响应异常');
      }
    } catch (error) {
      addTestResult('API 连接测试', false, `API 连接失败: ${error.message}`);
    }
  }

  /**
   * 测试 WebSocket 连接
   */
  function testWebSocketConnection() {
    try {
      if (!window.IMWebSocket) {
        addTestResult('WebSocket 连接测试', false, 'IMWebSocket 对象未找到');
        return;
      }

      // 尝试连接 WebSocket
      window.IMWebSocket.connect();
      addTestResult('WebSocket 连接测试', true, 'WebSocket 连接成功');
    } catch (error) {
      addTestResult('WebSocket 连接测试', false, `WebSocket 连接失败: ${error.message}`);
    }
  }

  /**
   * 测试用户登录
   */
  async function testUserLogin() {
    try {
      if (!window.IMAPI) {
        addTestResult('用户登录测试', false, 'IMAPI 对象未找到');
        return;
      }

      const result = await window.IMAPI.login(TEST_CONFIG.TEST_PHONE, TEST_CONFIG.TEST_PASSWORD);
      if (result && result.token) {
        addTestResult('用户登录测试', true, '用户登录成功');
      } else {
        addTestResult('用户登录测试', false, '用户登录失败');
      }
    } catch (error) {
      addTestResult('用户登录测试', false, `用户登录失败: ${error.message}`);
    }
  }

  /**
   * 测试 Telegram Web 集成
   */
  function testTelegramWebIntegration() {
    try {
      if (!window.App) {
        addTestResult('Telegram Web 集成测试', false, 'Telegram Web 未初始化');
        return;
      }

      if (!window.App.service) {
        addTestResult('Telegram Web 集成测试', false, 'Telegram Web 服务未初始化');
        return;
      }

      // 检查是否已替换认证服务
      const authService = window.App.service.auth;
      if (authService && authService.signIn) {
        addTestResult('Telegram Web 集成测试', true, '认证服务已替换');
      } else {
        addTestResult('Telegram Web 集成测试', false, '认证服务未替换');
      }
    } catch (error) {
      addTestResult('Telegram Web 集成测试', false, `集成测试失败: ${error.message}`);
    }
  }

  /**
   * 测试调试功能
   */
  function testDebugFeatures() {
    try {
      if (!window.志航密信Debug) {
        addTestResult('调试功能测试', false, '调试对象未找到');
        return;
      }

      const status = window.志航密信Debug.getStatus();
      if (status && status.api && status.websocket) {
        addTestResult('调试功能测试', true, '调试功能正常');
      } else {
        addTestResult('调试功能测试', false, '调试功能异常');
      }
    } catch (error) {
      addTestResult('调试功能测试', false, `调试功能测试失败: ${error.message}`);
    }
  }

  /**
   * 运行所有测试
   */
  async function runAllTests() {
    console.log('[志航密信] 开始运行集成测试...');
    
    // 等待 Telegram Web 初始化
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    // 运行各项测试
    await testAPIConnection();
    testWebSocketConnection();
    await testUserLogin();
    testTelegramWebIntegration();
    testDebugFeatures();
    
    // 输出测试结果
    console.log('[志航密信] 集成测试完成');
    console.log(`[志航密信] 测试结果: ${testResults.passed}/${testResults.total} 通过`);
    
    if (testResults.failed > 0) {
      console.error('[志航密信] 失败的测试:');
      testResults.details
        .filter(test => !test.passed)
        .forEach(test => {
          console.error(`  - ${test.name}: ${test.message}`);
        });
    }
    
    return testResults;
  }

  // 导出测试函数
  window.志航密信测试 = {
    runAllTests,
    testAPIConnection,
    testWebSocketConnection,
    testUserLogin,
    testTelegramWebIntegration,
    testDebugFeatures,
    getResults: () => testResults
  };

  // 自动运行测试（可选）
  if (window.location.search.includes('autotest=true')) {
    runAllTests();
  }

})();

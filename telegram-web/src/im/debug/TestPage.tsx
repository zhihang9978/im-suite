/**
 * IM-Suite 调试页面
 * 用于测试和调试适配层功能
 */

import React, { useState, useEffect } from 'react';

interface TestResult {
    success: boolean;
    message: string;
    data?: any;
    error?: string;
}

const TestPage: React.FC = () => {
    const [results, setResults] = useState<TestResult[]>([]);
    const [isLoading, setIsLoading] = useState(false);

    // 添加测试结果
    const addResult = (result: TestResult) => {
        setResults(prev => [...prev, result]);
    };

    // 清除结果
    const clearResults = () => {
        setResults([]);
    };

    // 测试 API 连接
    const testApiConnection = async () => {
        setIsLoading(true);
        try {
            const response = await fetch('/api/ping');
            const data = await response.json();
            
            addResult({
                success: response.ok,
                message: 'API 连接测试',
                data: data,
                error: response.ok ? undefined : `HTTP ${response.status}`
            });
        } catch (error) {
            addResult({
                success: false,
                message: 'API 连接测试',
                error: error instanceof Error ? error.message : '未知错误'
            });
        } finally {
            setIsLoading(false);
        }
    };

    // 测试 WebSocket 连接
    const testWebSocketConnection = () => {
        try {
            if (window.IMWebSocket) {
                const ws = window.IMWebSocket;
                
                if (ws.isConnected()) {
                    addResult({
                        success: true,
                        message: 'WebSocket 连接测试',
                        data: { status: '已连接', readyState: ws.getReadyState() }
                    });
                } else {
                    addResult({
                        success: false,
                        message: 'WebSocket 连接测试',
                        error: '未连接到 WebSocket 服务器'
                    });
                }
            } else {
                addResult({
                    success: false,
                    message: 'WebSocket 连接测试',
                    error: 'WebSocket 实例未初始化'
                });
            }
        } catch (error) {
            addResult({
                success: false,
                message: 'WebSocket 连接测试',
                error: error instanceof Error ? error.message : '未知错误'
            });
        }
    };

    // 测试发送消息
    const testSendMessage = async () => {
        setIsLoading(true);
        try {
            if (window.IMAPI) {
                const result = await window.IMAPI.sendMessage(1, {
                    content: '这是一条测试消息',
                    type: 'text'
                });
                
                addResult({
                    success: true,
                    message: '发送消息测试',
                    data: result
                });
            } else {
                addResult({
                    success: false,
                    message: '发送消息测试',
                    error: 'API 实例未初始化'
                });
            }
        } catch (error) {
            addResult({
                success: false,
                message: '发送消息测试',
                error: error instanceof Error ? error.message : '未知错误'
            });
        } finally {
            setIsLoading(false);
        }
    };

    // 测试用户登录
    const testUserLogin = async () => {
        setIsLoading(true);
        try {
            if (window.IMAPI) {
                const result = await window.IMAPI.login('13800138000', '123456');
                
                addResult({
                    success: true,
                    message: '用户登录测试',
                    data: { user: result.user, hasToken: !!result.token }
                });
            } else {
                addResult({
                    success: false,
                    message: '用户登录测试',
                    error: 'API 实例未初始化'
                });
            }
        } catch (error) {
            addResult({
                success: false,
                message: '用户登录测试',
                error: error instanceof Error ? error.message : '未知错误'
            });
        } finally {
            setIsLoading(false);
        }
    };

    // 测试获取聊天列表
    const testGetChats = async () => {
        setIsLoading(true);
        try {
            if (window.IMAPI) {
                const result = await window.IMAPI.getChats();
                
                addResult({
                    success: true,
                    message: '获取聊天列表测试',
                    data: { chats: result, count: result.length }
                });
            } else {
                addResult({
                    success: false,
                    message: '获取聊天列表测试',
                    error: 'API 实例未初始化'
                });
            }
        } catch (error) {
            addResult({
                success: false,
                message: '获取聊天列表测试',
                error: error instanceof Error ? error.message : '未知错误'
            });
        } finally {
            setIsLoading(false);
        }
    };

    // 测试 WebSocket 消息发送
    const testWebSocketMessage = () => {
        try {
            if (window.IMWebSocket && window.IMWebSocket.isConnected()) {
                window.IMWebSocket.send({
                    type: 'test',
                    data: { message: '这是一条 WebSocket 测试消息' }
                });
                
                addResult({
                    success: true,
                    message: 'WebSocket 消息发送测试',
                    data: { message: '消息已发送' }
                });
            } else {
                addResult({
                    success: false,
                    message: 'WebSocket 消息发送测试',
                    error: 'WebSocket 未连接'
                });
            }
        } catch (error) {
            addResult({
                success: false,
                message: 'WebSocket 消息发送测试',
                error: error instanceof Error ? error.message : '未知错误'
            });
        }
    };

    // 运行所有测试
    const runAllTests = async () => {
        clearResults();
        setIsLoading(true);
        
        try {
            await testApiConnection();
            testWebSocketConnection();
            await testUserLogin();
            await testGetChats();
            testWebSocketMessage();
            await testSendMessage();
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
            <h1>志航密信调试页面</h1>
            <p>此页面用于测试和调试适配层功能</p>
            
            <div style={{ marginBottom: '20px' }}>
                <button 
                    onClick={runAllTests} 
                    disabled={isLoading}
                    style={{ 
                        padding: '10px 20px', 
                        marginRight: '10px',
                        backgroundColor: '#007bff',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: isLoading ? 'not-allowed' : 'pointer'
                    }}
                >
                    {isLoading ? '测试中...' : '运行所有测试'}
                </button>
                
                <button 
                    onClick={clearResults}
                    style={{ 
                        padding: '10px 20px',
                        backgroundColor: '#6c757d',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: 'pointer'
                    }}
                >
                    清除结果
                </button>
            </div>

            <div style={{ marginBottom: '20px' }}>
                <h3>单项测试</h3>
                <button onClick={testApiConnection} disabled={isLoading} style={{ marginRight: '10px', padding: '8px 16px' }}>
                    API 连接测试
                </button>
                <button onClick={testWebSocketConnection} disabled={isLoading} style={{ marginRight: '10px', padding: '8px 16px' }}>
                    WebSocket 连接测试
                </button>
                <button onClick={testUserLogin} disabled={isLoading} style={{ marginRight: '10px', padding: '8px 16px' }}>
                    用户登录测试
                </button>
                <button onClick={testGetChats} disabled={isLoading} style={{ marginRight: '10px', padding: '8px 16px' }}>
                    获取聊天列表测试
                </button>
                <button onClick={testSendMessage} disabled={isLoading} style={{ marginRight: '10px', padding: '8px 16px' }}>
                    发送消息测试
                </button>
                <button onClick={testWebSocketMessage} disabled={isLoading} style={{ padding: '8px 16px' }}>
                    WebSocket 消息测试
                </button>
            </div>

            <div>
                <h3>测试结果</h3>
                {results.length === 0 ? (
                    <p style={{ color: '#6c757d' }}>暂无测试结果</p>
                ) : (
                    <div style={{ maxHeight: '400px', overflowY: 'auto' }}>
                        {results.map((result, index) => (
                            <div 
                                key={index}
                                style={{ 
                                    marginBottom: '10px',
                                    padding: '10px',
                                    border: '1px solid #dee2e6',
                                    borderRadius: '4px',
                                    backgroundColor: result.success ? '#d4edda' : '#f8d7da'
                                }}
                            >
                                <div style={{ fontWeight: 'bold', marginBottom: '5px' }}>
                                    {result.success ? '✅' : '❌'} {result.message}
                                </div>
                                {result.data && (
                                    <div style={{ marginBottom: '5px' }}>
                                        <strong>数据:</strong>
                                        <pre style={{ margin: '5px 0', fontSize: '12px' }}>
                                            {JSON.stringify(result.data, null, 2)}
                                        </pre>
                                    </div>
                                )}
                                {result.error && (
                                    <div style={{ color: '#721c24' }}>
                                        <strong>错误:</strong> {result.error}
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                )}
            </div>

            <div style={{ marginTop: '20px', padding: '10px', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                <h4>调试信息</h4>
                <p><strong>API 基础地址:</strong> {window.VITE_API_BASE_URL || 'http://localhost:8080/api'}</p>
                <p><strong>WebSocket 地址:</strong> {window.VITE_WS_BASE_URL || 'ws://localhost:8080/ws'}</p>
                <p><strong>API 实例:</strong> {window.IMAPI ? '已初始化' : '未初始化'}</p>
                <p><strong>WebSocket 实例:</strong> {window.IMWebSocket ? '已初始化' : '未初始化'}</p>
                {window.IMWebSocket && (
                    <p><strong>WebSocket 状态:</strong> {window.IMWebSocket.isConnected() ? '已连接' : '未连接'}</p>
                )}
            </div>
        </div>
    );
};

export default TestPage;



/**
 * 志航密信 Web 端集成脚本
 * 将适配层集成到 Telegram Web 中，替换原有的网络调用
 */

(function () {
  'use strict';

  console.log('[志航密信] 开始集成适配层...');

  // 等待 Telegram Web 初始化完成
  function waitForTelegramWeb() {
    return new Promise((resolve) => {
      const checkInterval = setInterval(() => {
        if (window.App && window.App.service) {
          clearInterval(checkInterval);
          resolve();
        }
      }, 100);
    });
  }

  // 集成 API 适配层
  function integrateAPIAdapter() {
    console.log('[志航密信] 集成 API 适配层...');
    
    // 替换认证服务
    if (window.App && window.App.service && window.App.service.auth) {
      const originalAuth = window.App.service.auth;
      
      // 替换登录方法
      if (originalAuth.signIn) {
        originalAuth.signIn = async function(phone, code) {
          console.log('[志航密信] 使用自定义登录方法');
          try {
            const result = await window.IMAPI.login(phone, code);
            return {
              user: result.user,
              session: {
                id: result.token,
                expires_at: result.expires_at
              }
            };
          } catch (error) {
            console.error('[志航密信] 登录失败:', error);
            throw error;
          }
        };
      }

      // 替换登出方法
      if (originalAuth.logOut) {
        originalAuth.logOut = async function() {
          console.log('[志航密信] 使用自定义登出方法');
          try {
            await window.IMAPI.logout();
            return true;
          } catch (error) {
            console.error('[志航密信] 登出失败:', error);
            throw error;
          }
        };
      }
    }
  }

  // 集成 WebSocket 适配层
  function integrateWebSocketAdapter() {
    console.log('[志航密信] 集成 WebSocket 适配层...');
    
    // 替换 WebSocket 连接
    if (window.App && window.App.service && window.App.service.connection) {
      const originalConnection = window.App.service.connection;
      
      // 替换连接方法
      if (originalConnection.connect) {
        originalConnection.connect = function() {
          console.log('[志航密信] 使用自定义 WebSocket 连接');
          return window.IMWebSocket.connect();
        };
      }

      // 替换发送方法
      if (originalConnection.send) {
        originalConnection.send = function(data) {
          console.log('[志航密信] 使用自定义 WebSocket 发送');
          return window.IMWebSocket.send(data);
        };
      }
    }
  }

  // 集成消息服务适配层
  function integrateMessageAdapter() {
    console.log('[志航密信] 集成消息服务适配层...');
    
    // 替换消息服务
    if (window.App && window.App.service && window.App.service.messages) {
      const originalMessages = window.App.service.messages;
      
      // 替换发送消息方法
      if (originalMessages.sendMessage) {
        originalMessages.sendMessage = async function(chatId, messageData) {
          console.log('[志航密信] 使用自定义发送消息方法');
          try {
            const result = await window.IMAPI.sendMessage(chatId, messageData);
            return result;
          } catch (error) {
            console.error('[志航密信] 发送消息失败:', error);
            throw error;
          }
        };
      }

      // 替换获取消息方法
      if (originalMessages.getHistory) {
        originalMessages.getHistory = async function(chatId, params) {
          console.log('[志航密信] 使用自定义获取消息方法');
          try {
            const result = await window.IMAPI.getMessages(chatId, params);
            return result;
          } catch (error) {
            console.error('[志航密信] 获取消息失败:', error);
            throw error;
          }
        };
      }
    }
  }

  // 集成用户服务适配层
  function integrateUserAdapter() {
    console.log('[志航密信] 集成用户服务适配层...');
    
    // 替换用户服务
    if (window.App && window.App.service && window.App.service.users) {
      const originalUsers = window.App.service.users;
      
      // 替换获取用户信息方法
      if (originalUsers.getFullUser) {
        originalUsers.getFullUser = async function(userId) {
          console.log('[志航密信] 使用自定义获取用户信息方法');
          try {
            if (userId === 'me') {
              return await window.IMAPI.getCurrentUser();
            }
            // 这里可以添加获取其他用户信息的逻辑
            return null;
          } catch (error) {
            console.error('[志航密信] 获取用户信息失败:', error);
            throw error;
          }
        };
      }
    }
  }

  // 集成联系人服务适配层
  function integrateContactAdapter() {
    console.log('[志航密信] 集成联系人服务适配层...');
    
    // 替换联系人服务
    if (window.App && window.App.service && window.App.service.contacts) {
      const originalContacts = window.App.service.contacts;
      
      // 替换获取联系人列表方法
      if (originalContacts.getContacts) {
        originalContacts.getContacts = async function() {
          console.log('[志航密信] 使用自定义获取联系人列表方法');
          try {
            const result = await window.IMAPI.getContacts();
            return result;
          } catch (error) {
            console.error('[志航密信] 获取联系人列表失败:', error);
            throw error;
          }
        };
      }

      // 替换添加联系人方法
      if (originalContacts.addContact) {
        originalContacts.addContact = async function(phone, nickname) {
          console.log('[志航密信] 使用自定义添加联系人方法');
          try {
            const result = await window.IMAPI.addContact(phone, nickname);
            return result;
          } catch (error) {
            console.error('[志航密信] 添加联系人失败:', error);
            throw error;
          }
        };
      }
    }
  }

  // 添加调试功能
  function addDebugFeatures() {
    console.log('[志航密信] 添加调试功能...');
    
    // 添加全局调试对象
    window.志航密信Debug = {
      // 测试 API 连接
      testAPI: async function() {
        try {
          const result = await window.IMAPI.get('/ping');
          console.log('[志航密信] API 连接测试成功:', result);
          return result;
        } catch (error) {
          console.error('[志航密信] API 连接测试失败:', error);
          throw error;
        }
      },
      
      // 测试 WebSocket 连接
      testWebSocket: function() {
        try {
          window.IMWebSocket.connect();
          console.log('[志航密信] WebSocket 连接测试成功');
          return true;
        } catch (error) {
          console.error('[志航密信] WebSocket 连接测试失败:', error);
          throw error;
        }
      },
      
      // 获取当前状态
      getStatus: function() {
        return {
          api: !!window.IMAPI,
          websocket: !!window.IMWebSocket,
          telegram: !!window.App,
          token: window.IMAPI ? window.IMAPI.token : null
        };
      }
    };

    // 添加键盘快捷键
    document.addEventListener('keydown', function(event) {
      // Ctrl+Shift+D 打开调试面板
      if (event.ctrlKey && event.shiftKey && event.key === 'D') {
        event.preventDefault();
        console.log('[志航密信] 调试面板快捷键触发');
        if (window.志航密信Debug) {
          window.志航密信Debug.testAPI();
          window.志航密信Debug.testWebSocket();
        }
      }
    });
  }

  // 主集成函数
  async function integrate() {
    try {
      console.log('[志航密信] 等待 Telegram Web 初始化...');
      await waitForTelegramWeb();
      
      console.log('[志航密信] 开始集成适配层...');
      
      // 集成各个服务
      integrateAPIAdapter();
      integrateWebSocketAdapter();
      integrateMessageAdapter();
      integrateUserAdapter();
      integrateContactAdapter();
      
      // 添加调试功能
      addDebugFeatures();
      
      console.log('[志航密信] 适配层集成完成！');
      console.log('[志航密信] 使用 Ctrl+Shift+D 打开调试面板');
      
    } catch (error) {
      console.error('[志航密信] 集成失败:', error);
    }
  }

  // 页面加载完成后开始集成
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', integrate);
  } else {
    integrate();
  }

})();

/**
 * IM-Suite Web 端到端加密管理器
 * 实现端到端加密功能，包括密钥管理、消息加密、数字签名等
 */

(function () {
  'use strict';

  var EncryptionManager = {
    // 配置
    config: {
      algorithm: 'XChaCha20-Poly1305',
      keySize: 256,
      nonceSize: 192,
      tagSize: 128,
      keyDerivation: 'HKDF-SHA256',
      signatureAlgorithm: 'Ed25519'
    },

    // 状态
    isInitialized: false,
    masterKey: null,
    identityKey: null,
    signingKey: null,
    sessionKeys: new Map(),
    keyStorage: new Map(),

    // 事件监听器
    eventListeners: {},

    /**
     * 初始化加密管理器
     */
    init: function () {
      if (this.isInitialized) {
        console.log('[Encryption] 已经初始化');
        return;
      }

      console.log('[Encryption] 初始化端到端加密管理器');
      this.isInitialized = true;

      // 检查浏览器支持
      if (!this.checkBrowserSupport()) {
        console.error('[Encryption] 浏览器不支持端到端加密');
        this.emit('error', { code: 'BROWSER_NOT_SUPPORTED', message: '浏览器不支持端到端加密' });
        return;
      }

      // 初始化密钥
      this.initializeKeys();
    },

    /**
     * 检查浏览器支持
     */
    checkBrowserSupport: function () {
      return !!(window.crypto && window.crypto.subtle && window.crypto.getRandomValues);
    },

    /**
     * 初始化密钥
     */
    initializeKeys: function () {
      var self = this;

      // 生成主密钥
      this.generateMasterKey()
        .then(function (masterKey) {
          self.masterKey = masterKey;
          return self.generateIdentityKey();
        })
        .then(function (identityKey) {
          self.identityKey = identityKey;
          return self.generateSigningKey();
        })
        .then(function (signingKey) {
          self.signingKey = signingKey;
          console.log('[Encryption] 密钥初始化完成');
          self.emit('initialized');
        })
        .catch(function (error) {
          console.error('[Encryption] 密钥初始化失败:', error);
          self.emit('error', { code: 'KEY_INIT_FAILED', message: error.message });
        });
    },

    /**
     * 生成主密钥
     */
    generateMasterKey: function () {
      return new Promise(function (resolve, reject) {
        try {
          // 生成随机主密钥
          var masterKey = crypto.getRandomValues(new Uint8Array(32));
          
          // 使用 PBKDF2 进行密钥强化
          crypto.subtle.importKey(
            'raw',
            masterKey,
            { name: 'PBKDF2' },
            false,
            ['deriveKey']
          ).then(function (keyMaterial) {
            return crypto.subtle.deriveKey(
              {
                name: 'PBKDF2',
                salt: crypto.getRandomValues(new Uint8Array(16)),
                iterations: 100000,
                hash: 'SHA-256'
              },
              keyMaterial,
              { name: 'HKDF', hash: 'SHA-256' },
              false,
              ['deriveKey']
            );
          }).then(resolve).catch(reject);
        } catch (error) {
          reject(error);
        }
      });
    },

    /**
     * 生成身份密钥
     */
    generateIdentityKey: function () {
      return crypto.subtle.generateKey(
        {
          name: 'X25519',
          namedCurve: 'X25519'
        },
        true,
        ['deriveKey']
      );
    },

    /**
     * 生成签名密钥
     */
    generateSigningKey: function () {
      return crypto.subtle.generateKey(
        {
          name: 'Ed25519',
          namedCurve: 'Ed25519'
        },
        true,
        ['sign', 'verify']
      );
    },

    /**
     * 生成会话密钥
     * @param {string} contactId 联系人ID
     * @param {CryptoKey} ephemeralKey 临时密钥
     */
    generateSessionKey: function (contactId, ephemeralKey) {
      var self = this;

      return crypto.subtle.deriveKey(
        {
          name: 'X25519',
          public: ephemeralKey
        },
        this.identityKey.privateKey,
        {
          name: 'HKDF',
          hash: 'SHA-256',
          salt: crypto.getRandomValues(new Uint8Array(32)),
          info: new TextEncoder().encode('session-key-' + contactId)
        },
        false,
        ['encrypt', 'decrypt']
      ).then(function (sessionKey) {
        self.sessionKeys.set(contactId, sessionKey);
        return sessionKey;
      });
    },

    /**
     * 加密消息
     * @param {string} plaintext 明文
     * @param {string} contactId 联系人ID
     * @param {object} options 加密选项
     */
    encryptMessage: function (plaintext, contactId, options) {
      var self = this;

      return new Promise(function (resolve, reject) {
        try {
          // 获取会话密钥
          var sessionKey = self.sessionKeys.get(contactId);
          if (!sessionKey) {
            reject(new Error('会话密钥不存在'));
            return;
          }

          // 生成随机 nonce
          var nonce = crypto.getRandomValues(new Uint8Array(24));
          
          // 加密消息
          crypto.subtle.encrypt(
            {
              name: 'AES-GCM',
              iv: nonce,
              tagLength: 128
            },
            sessionKey,
            new TextEncoder().encode(plaintext)
          ).then(function (ciphertext) {
            // 创建加密消息结构
            var encryptedMessage = {
              version: 1,
              algorithm: self.config.algorithm,
              contactId: contactId,
              nonce: Array.from(nonce),
              ciphertext: Array.from(new Uint8Array(ciphertext)),
              timestamp: Date.now(),
              ttl: options.ttl || 0
            };

            // 数字签名
            return self.signMessage(encryptedMessage);
          }).then(function (signature) {
            encryptedMessage.signature = Array.from(signature);
            resolve(encryptedMessage);
          }).catch(reject);
        } catch (error) {
          reject(error);
        }
      });
    },

    /**
     * 解密消息
     * @param {object} encryptedMessage 加密消息
     * @param {string} contactId 联系人ID
     */
    decryptMessage: function (encryptedMessage, contactId) {
      var self = this;

      return new Promise(function (resolve, reject) {
        try {
          // 验证签名
          self.verifySignature(encryptedMessage, contactId)
            .then(function (isValid) {
              if (!isValid) {
                reject(new Error('消息签名验证失败'));
                return;
              }

              // 获取会话密钥
              var sessionKey = self.sessionKeys.get(contactId);
              if (!sessionKey) {
                reject(new Error('会话密钥不存在'));
                return;
              }

              // 解密消息
              var nonce = new Uint8Array(encryptedMessage.nonce);
              var ciphertext = new Uint8Array(encryptedMessage.ciphertext);
              
              return crypto.subtle.decrypt(
                {
                  name: 'AES-GCM',
                  iv: nonce,
                  tagLength: 128
                },
                sessionKey,
                ciphertext
              );
            }).then(function (plaintext) {
              var decodedText = new TextDecoder().decode(plaintext);
              resolve(decodedText);
            }).catch(reject);
        } catch (error) {
          reject(error);
        }
      });
    },

    /**
     * 签名消息
     * @param {object} message 消息对象
     */
    signMessage: function (message) {
      var self = this;

      return new Promise(function (resolve, reject) {
        try {
          // 创建消息摘要
          var messageData = JSON.stringify(message);
          var messageBuffer = new TextEncoder().encode(messageData);
          
          // 使用私钥签名
          crypto.subtle.sign(
            {
              name: 'Ed25519'
            },
            self.signingKey.privateKey,
            messageBuffer
          ).then(resolve).catch(reject);
        } catch (error) {
          reject(error);
        }
      });
    },

    /**
     * 验证签名
     * @param {object} message 消息对象
     * @param {string} contactId 联系人ID
     */
    verifySignature: function (message, contactId) {
      var self = this;

      return new Promise(function (resolve, reject) {
        try {
          // 获取发送者的公钥
          var senderPublicKey = self.getContactPublicKey(contactId);
          if (!senderPublicKey) {
            resolve(false);
            return;
          }

          // 创建消息摘要
          var messageCopy = Object.assign({}, message);
          delete messageCopy.signature;
          var messageData = JSON.stringify(messageCopy);
          var messageBuffer = new TextEncoder().encode(messageData);
          
          // 验证签名
          crypto.subtle.verify(
            {
              name: 'Ed25519'
            },
            senderPublicKey,
            new Uint8Array(message.signature),
            messageBuffer
          ).then(resolve).catch(function (error) {
            console.error('[Encryption] 签名验证失败:', error);
            resolve(false);
          });
        } catch (error) {
          console.error('[Encryption] 签名验证错误:', error);
          resolve(false);
        }
      });
    },

    /**
     * 创建阅后即焚消息
     * @param {string} plaintext 明文
     * @param {string} contactId 联系人ID
     * @param {number} ttlSeconds 生存时间（秒）
     */
    createSelfDestructingMessage: function (plaintext, contactId, ttlSeconds) {
      var self = this;

      return new Promise(function (resolve, reject) {
        try {
          // 生成临时密钥
          var messageKey = crypto.getRandomValues(new Uint8Array(32));
          
          // 加密消息
          var nonce = crypto.getRandomValues(new Uint8Array(24));
          crypto.subtle.importKey(
            'raw',
            messageKey,
            { name: 'AES-GCM' },
            false,
            ['encrypt', 'decrypt']
          ).then(function (key) {
            return crypto.subtle.encrypt(
              {
                name: 'AES-GCM',
                iv: nonce,
                tagLength: 128
              },
              key,
              new TextEncoder().encode(plaintext)
            );
          }).then(function (ciphertext) {
            // 创建自毁消息
            var selfDestructingMessage = {
              version: 1,
              algorithm: 'AES-GCM',
              contactId: contactId,
              nonce: Array.from(nonce),
              ciphertext: Array.from(new Uint8Array(ciphertext)),
              ttl: ttlSeconds,
              timestamp: Date.now(),
              type: 'self-destructing'
            };

            // 设置自毁定时器
            setTimeout(function () {
              self.destroyMessageKey(messageKey);
            }, ttlSeconds * 1000);

            resolve({
              message: selfDestructingMessage,
              messageKey: messageKey
            });
          }).catch(reject);
        } catch (error) {
          reject(error);
        }
      });
    },

    /**
     * 解密阅后即焚消息
     * @param {object} encryptedMessage 加密消息
     * @param {Uint8Array} messageKey 消息密钥
     */
    decryptSelfDestructingMessage: function (encryptedMessage, messageKey) {
      return new Promise(function (resolve, reject) {
        try {
          // 导入密钥
          crypto.subtle.importKey(
            'raw',
            messageKey,
            { name: 'AES-GCM' },
            false,
            ['encrypt', 'decrypt']
          ).then(function (key) {
            // 解密消息
            var nonce = new Uint8Array(encryptedMessage.nonce);
            var ciphertext = new Uint8Array(encryptedMessage.ciphertext);
            
            return crypto.subtle.decrypt(
              {
                name: 'AES-GCM',
                iv: nonce,
                tagLength: 128
              },
              key,
              ciphertext
            );
          }).then(function (plaintext) {
            // 解密后立即销毁密钥
            messageKey.fill(0);
            messageKey = null;
            
            var decodedText = new TextDecoder().decode(plaintext);
            resolve(decodedText);
          }).catch(reject);
        } catch (error) {
          reject(error);
        }
      });
    },

    /**
     * 获取联系人公钥
     * @param {string} contactId 联系人ID
     */
    getContactPublicKey: function (contactId) {
      // 这里应该从服务器或本地存储获取联系人的公钥
      // 简化实现，返回 null
      return null;
    },

    /**
     * 销毁消息密钥
     * @param {Uint8Array} messageKey 消息密钥
     */
    destroyMessageKey: function (messageKey) {
      if (messageKey) {
        messageKey.fill(0);
        messageKey = null;
      }
    },

    /**
     * 导出公钥
     * @param {string} format 导出格式
     */
    exportPublicKey: function (format) {
      var self = this;

      return new Promise(function (resolve, reject) {
        try {
          if (format === 'raw') {
            crypto.subtle.exportKey('raw', self.identityKey.publicKey)
              .then(resolve).catch(reject);
          } else if (format === 'spki') {
            crypto.subtle.exportKey('spki', self.identityKey.publicKey)
              .then(resolve).catch(reject);
          } else {
            reject(new Error('不支持的导出格式'));
          }
        } catch (error) {
          reject(error);
        }
      });
    },

    /**
     * 导入公钥
     * @param {ArrayBuffer} keyData 密钥数据
     * @param {string} format 导入格式
     */
    importPublicKey: function (keyData, format) {
      return new Promise(function (resolve, reject) {
        try {
          if (format === 'raw') {
            crypto.subtle.importKey(
              'raw',
              keyData,
              { name: 'X25519', namedCurve: 'X25519' },
              false,
              ['deriveKey']
            ).then(resolve).catch(reject);
          } else if (format === 'spki') {
            crypto.subtle.importKey(
              'spki',
              keyData,
              { name: 'X25519', namedCurve: 'X25519' },
              false,
              ['deriveKey']
            ).then(resolve).catch(reject);
          } else {
            reject(new Error('不支持的导入格式'));
          }
        } catch (error) {
          reject(error);
        }
      });
    },

    /**
     * 订阅事件
     * @param {string} event 事件名称
     * @param {function} callback 回调函数
     */
    subscribe: function (event, callback) {
      if (!this.eventListeners[event]) {
        this.eventListeners[event] = [];
      }
      this.eventListeners[event].push(callback);
    },

    /**
     * 取消订阅事件
     * @param {string} event 事件名称
     * @param {function} callback 回调函数
     */
    unsubscribe: function (event, callback) {
      if (this.eventListeners[event]) {
        this.eventListeners[event] = this.eventListeners[event].filter(function (listener) {
          return listener !== callback;
        });
      }
    },

    /**
     * 触发事件
     * @param {string} event 事件名称
     * @param {*} data 事件数据
     */
    emit: function (event, data) {
      if (this.eventListeners[event]) {
        this.eventListeners[event].forEach(function (listener) {
          listener(data);
        });
      }
    },

    /**
     * 获取加密状态
     */
    getEncryptionStatus: function () {
      return {
        isInitialized: this.isInitialized,
        hasMasterKey: !!this.masterKey,
        hasIdentityKey: !!this.identityKey,
        hasSigningKey: !!this.signingKey,
        sessionKeysCount: this.sessionKeys.size,
        algorithm: this.config.algorithm
      };
    }
  };

  // 暴露到全局
  window.EncryptionManager = EncryptionManager;

  // 自动初始化
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () {
      EncryptionManager.init();
    });
  } else {
    EncryptionManager.init();
  }
})();

/**
 * IM-Suite Web WebRTC 管理器
 * 负责处理语音和视频通话功能
 */

(function () {
  'use strict';

  var WebRTCManager = {
    // 配置
    config: {
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' },
        { urls: 'stun:stun1.l.google.com:19302' },
        { urls: 'stun:stun2.l.google.com:19302' }
      ],
      iceCandidatePoolSize: 10,
      bundlePolicy: 'max-bundle',
      rtcpMuxPolicy: 'require'
    },

    // 状态
    isInitialized: false,
    currentCall: null,
    peerConnection: null,
    localStream: null,
    remoteStream: null,

    // 事件监听器
    eventListeners: {},

    /**
     * 初始化 WebRTC 管理器
     */
    init: function () {
      if (this.isInitialized) {
        console.log('[WebRTC] 已经初始化');
        return;
      }

      console.log('[WebRTC] 初始化 WebRTC 管理器');
      this.isInitialized = true;

      // 检查浏览器支持
      if (!this.checkBrowserSupport()) {
        console.error('[WebRTC] 浏览器不支持 WebRTC');
        this.emit('error', { code: 'BROWSER_NOT_SUPPORTED', message: '浏览器不支持 WebRTC' });
        return;
      }

      // 设置 WebSocket 事件监听
      this.setupWebSocketListeners();
    },

    /**
     * 检查浏览器支持
     */
    checkBrowserSupport: function () {
      return !!(navigator.mediaDevices && navigator.mediaDevices.getUserMedia && window.RTCPeerConnection);
    },

    /**
     * 设置 WebSocket 事件监听
     */
    setupWebSocketListeners: function () {
      var self = this;

      // 监听信令事件
      if (window.IMWebSocket) {
        window.IMWebSocket.subscribe('call.offer', function (data) {
          self.handleCallOffer(data);
        });

        window.IMWebSocket.subscribe('call.answer', function (data) {
          self.handleCallAnswer(data);
        });

        window.IMWebSocket.subscribe('call.ice', function (data) {
          self.handleIceCandidate(data);
        });

        window.IMWebSocket.subscribe('call.end', function (data) {
          self.handleCallEnd(data);
        });

        window.IMWebSocket.subscribe('call.error', function (data) {
          self.handleCallError(data);
        });
      }
    },

    /**
     * 发起通话
     * @param {number} toUserId 目标用户ID
     * @param {string} callType 通话类型 ('audio' 或 'video')
     * @param {object} constraints 媒体约束
     */
    startCall: function (toUserId, callType, constraints) {
      var self = this;

      if (this.currentCall) {
        console.warn('[WebRTC] 已有通话进行中');
        this.emit('error', { code: 'CALL_IN_PROGRESS', message: '已有通话进行中' });
        return;
      }

      console.log('[WebRTC] 发起通话:', toUserId, callType);

      // 生成通话ID
      var callId = 'call_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
      this.currentCall = {
        id: callId,
        toUser: toUserId,
        type: callType,
        status: 'initiating'
      };

      // 获取媒体流
      this.getUserMedia(constraints || this.getDefaultConstraints(callType))
        .then(function (stream) {
          self.localStream = stream;
          self.emit('localStream', stream);

          // 创建 PeerConnection
          return self.createPeerConnection();
        })
        .then(function () {
          // 添加本地流到 PeerConnection
          self.localStream.getTracks().forEach(function (track) {
            self.peerConnection.addTrack(track, self.localStream);
          });

          // 创建 Offer
          return self.peerConnection.createOffer();
        })
        .then(function (offer) {
          // 设置本地描述
          return self.peerConnection.setLocalDescription(offer);
        })
        .then(function () {
          // 发送信令消息
          self.sendCallOffer(callId, toUserId, callType);
        })
        .catch(function (error) {
          console.error('[WebRTC] 发起通话失败:', error);
          self.emit('error', { code: 'START_CALL_FAILED', message: error.message });
          self.endCall();
        });
    },

    /**
     * 应答通话
     * @param {string} callId 通话ID
     */
    answerCall: function (callId) {
      var self = this;

      if (!this.currentCall || this.currentCall.id !== callId) {
        console.warn('[WebRTC] 无效的通话ID');
        return;
      }

      console.log('[WebRTC] 应答通话:', callId);

      // 获取媒体流
      this.getUserMedia(this.getDefaultConstraints(this.currentCall.type))
        .then(function (stream) {
          self.localStream = stream;
          self.emit('localStream', stream);

          // 创建 PeerConnection
          return self.createPeerConnection();
        })
        .then(function () {
          // 添加本地流到 PeerConnection
          self.localStream.getTracks().forEach(function (track) {
            self.peerConnection.addTrack(track, self.localStream);
          });

          // 创建 Answer
          return self.peerConnection.createAnswer();
        })
        .then(function (answer) {
          // 设置本地描述
          return self.peerConnection.setLocalDescription(answer);
        })
        .then(function () {
          // 发送信令消息
          self.sendCallAnswer(callId);
        })
        .catch(function (error) {
          console.error('[WebRTC] 应答通话失败:', error);
          self.emit('error', { code: 'ANSWER_CALL_FAILED', message: error.message });
          self.endCall();
        });
    },

    /**
     * 结束通话
     */
    endCall: function () {
      if (!this.currentCall) {
        return;
      }

      console.log('[WebRTC] 结束通话:', this.currentCall.id);

      // 发送结束信令
      this.sendCallEnd(this.currentCall.id, 'user_hangup');

      // 清理资源
      this.cleanup();

      // 触发事件
      this.emit('callEnded', { callId: this.currentCall.id });
    },

    /**
     * 拒绝通话
     * @param {string} callId 通话ID
     */
    rejectCall: function (callId) {
      console.log('[WebRTC] 拒绝通话:', callId);
      this.sendCallEnd(callId, 'user_rejected');
      this.cleanup();
    },

    /**
     * 获取用户媒体
     * @param {object} constraints 媒体约束
     */
    getUserMedia: function (constraints) {
      return navigator.mediaDevices.getUserMedia(constraints)
        .catch(function (error) {
          console.error('[WebRTC] 获取媒体失败:', error);
          throw new Error('无法访问摄像头和麦克风: ' + error.message);
        });
    },

    /**
     * 获取默认媒体约束
     * @param {string} callType 通话类型
     */
    getDefaultConstraints: function (callType) {
      var constraints = {
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true
        }
      };

      if (callType === 'video') {
        constraints.video = {
          width: { ideal: 1280 },
          height: { ideal: 720 },
          frameRate: { ideal: 30 }
        };
      }

      return constraints;
    },

    /**
     * 创建 PeerConnection
     */
    createPeerConnection: function () {
      var self = this;

      this.peerConnection = new RTCPeerConnection(this.config);

      // 监听连接状态变化
      this.peerConnection.onconnectionstatechange = function () {
        console.log('[WebRTC] 连接状态:', self.peerConnection.connectionState);
        self.emit('connectionStateChange', self.peerConnection.connectionState);
      };

      // 监听 ICE 连接状态变化
      this.peerConnection.oniceconnectionstatechange = function () {
        console.log('[WebRTC] ICE 连接状态:', self.peerConnection.iceConnectionState);
        self.emit('iceConnectionStateChange', self.peerConnection.iceConnectionState);
      };

      // 监听远程流
      this.peerConnection.ontrack = function (event) {
        console.log('[WebRTC] 收到远程流');
        self.remoteStream = event.streams[0];
        self.emit('remoteStream', self.remoteStream);
      };

      // 监听 ICE 候选
      this.peerConnection.onicecandidate = function (event) {
        if (event.candidate) {
          self.sendIceCandidate(event.candidate);
        }
      };

      return Promise.resolve();
    },

    /**
     * 处理通话邀请
     * @param {object} data 信令数据
     */
    handleCallOffer: function (data) {
      var self = this;

      console.log('[WebRTC] 收到通话邀请:', data);

      // 设置当前通话
      this.currentCall = {
        id: data.call_id,
        fromUser: data.from_user,
        type: data.call_type,
        status: 'ringing'
      };

      // 触发事件
      this.emit('incomingCall', {
        callId: data.call_id,
        fromUser: data.from_user,
        callType: data.call_type
      });

      // 创建 PeerConnection
      this.createPeerConnection()
        .then(function () {
          // 设置远程描述
          return self.peerConnection.setRemoteDescription(data.sdp);
        })
        .then(function () {
          // 添加 ICE 候选
          if (data.ice_candidates) {
            data.ice_candidates.forEach(function (candidate) {
              self.peerConnection.addIceCandidate(candidate);
            });
          }
        })
        .catch(function (error) {
          console.error('[WebRTC] 处理通话邀请失败:', error);
          self.emit('error', { code: 'HANDLE_OFFER_FAILED', message: error.message });
        });
    },

    /**
     * 处理通话应答
     * @param {object} data 信令数据
     */
    handleCallAnswer: function (data) {
      var self = this;

      console.log('[WebRTC] 收到通话应答:', data);

      if (!this.currentCall || this.currentCall.id !== data.call_id) {
        console.warn('[WebRTC] 无效的通话应答');
        return;
      }

      // 设置远程描述
      this.peerConnection.setRemoteDescription(data.sdp)
        .then(function () {
          // 添加 ICE 候选
          if (data.ice_candidates) {
            data.ice_candidates.forEach(function (candidate) {
              self.peerConnection.addIceCandidate(candidate);
            });
          }
        })
        .catch(function (error) {
          console.error('[WebRTC] 处理通话应答失败:', error);
          self.emit('error', { code: 'HANDLE_ANSWER_FAILED', message: error.message });
        });
    },

    /**
     * 处理 ICE 候选
     * @param {object} data 信令数据
     */
    handleIceCandidate: function (data) {
      if (!this.peerConnection) {
        return;
      }

      this.peerConnection.addIceCandidate(data.ice_candidate)
        .catch(function (error) {
          console.error('[WebRTC] 添加 ICE 候选失败:', error);
        });
    },

    /**
     * 处理通话结束
     * @param {object} data 信令数据
     */
    handleCallEnd: function (data) {
      console.log('[WebRTC] 通话结束:', data);
      this.cleanup();
      this.emit('callEnded', data);
    },

    /**
     * 处理通话错误
     * @param {object} data 信令数据
     */
    handleCallError: function (data) {
      console.error('[WebRTC] 通话错误:', data);
      this.cleanup();
      this.emit('error', data);
    },

    /**
     * 发送通话邀请
     * @param {string} callId 通话ID
     * @param {number} toUserId 目标用户ID
     * @param {string} callType 通话类型
     */
    sendCallOffer: function (callId, toUserId, callType) {
      var offer = this.peerConnection.localDescription;
      var iceCandidates = this.collectIceCandidates();

      var message = {
        type: 'call.offer',
        payload: {
          call_id: callId,
          from_user: window.currentUserId || 1, // 假设有当前用户ID
          to_user: toUserId,
          call_type: callType,
          sdp: offer,
          ice_candidates: iceCandidates,
          constraints: this.getDefaultConstraints(callType),
          created_at: new Date().toISOString()
        }
      };

      window.IMWebSocket.send('call.offer', message.payload);
    },

    /**
     * 发送通话应答
     * @param {string} callId 通话ID
     */
    sendCallAnswer: function (callId) {
      var answer = this.peerConnection.localDescription;
      var iceCandidates = this.collectIceCandidates();

      var message = {
        type: 'call.answer',
        payload: {
          call_id: callId,
          from_user: window.currentUserId || 1,
          to_user: this.currentCall.fromUser,
          sdp: answer,
          ice_candidates: iceCandidates,
          answered_at: new Date().toISOString()
        }
      };

      window.IMWebSocket.send('call.answer', message.payload);
    },

    /**
     * 发送 ICE 候选
     * @param {object} candidate ICE 候选
     */
    sendIceCandidate: function (candidate) {
      if (!this.currentCall) {
        return;
      }

      var message = {
        type: 'call.ice',
        payload: {
          call_id: this.currentCall.id,
          from_user: window.currentUserId || 1,
          to_user: this.currentCall.toUser || this.currentCall.fromUser,
          ice_candidate: {
            candidate: candidate.candidate,
            sdpMLineIndex: candidate.sdpMLineIndex,
            sdpMid: candidate.sdpMid
          }
        }
      };

      window.IMWebSocket.send('call.ice', message.payload);
    },

    /**
     * 发送通话结束
     * @param {string} callId 通话ID
     * @param {string} reason 结束原因
     */
    sendCallEnd: function (callId, reason) {
      var message = {
        type: 'call.end',
        payload: {
          call_id: callId,
          from_user: window.currentUserId || 1,
          to_user: this.currentCall.toUser || this.currentCall.fromUser,
          reason: reason,
          duration: this.getCallDuration(),
          ended_at: new Date().toISOString()
        }
      };

      window.IMWebSocket.send('call.end', message.payload);
    },

    /**
     * 收集 ICE 候选
     */
    collectIceCandidates: function () {
      // 这里应该收集所有已收集的 ICE 候选
      // 简化实现，返回空数组
      return [];
    },

    /**
     * 获取通话时长
     */
    getCallDuration: function () {
      if (!this.currentCall || !this.currentCall.startTime) {
        return 0;
      }
      return Math.floor((Date.now() - this.currentCall.startTime) / 1000);
    },

    /**
     * 清理资源
     */
    cleanup: function () {
      // 停止本地流
      if (this.localStream) {
        this.localStream.getTracks().forEach(function (track) {
          track.stop();
        });
        this.localStream = null;
      }

      // 关闭 PeerConnection
      if (this.peerConnection) {
        this.peerConnection.close();
        this.peerConnection = null;
      }

      // 清理远程流
      this.remoteStream = null;

      // 清理当前通话
      this.currentCall = null;
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
     * 获取当前通话状态
     */
    getCallStatus: function () {
      return {
        isInCall: !!this.currentCall,
        callId: this.currentCall ? this.currentCall.id : null,
        callType: this.currentCall ? this.currentCall.type : null,
        connectionState: this.peerConnection ? this.peerConnection.connectionState : null,
        iceConnectionState: this.peerConnection ? this.peerConnection.iceConnectionState : null
      };
    }
  };

  // 暴露到全局
  window.WebRTCManager = WebRTCManager;

  // 自动初始化
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () {
      WebRTCManager.init();
    });
  } else {
    WebRTCManager.init();
  }
})();

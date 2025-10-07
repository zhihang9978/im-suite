# IM-Suite WebRTC 集成示例

## 概述

本文档提供了 IM-Suite 中 WebRTC 功能的完整集成示例，包括 Web 端和 Android 端的实现。

## Web 端集成示例

### 1. 基础通话功能

#### HTML 结构
```html
<!DOCTYPE html>
<html>
<head>
    <title>IM-Suite 通话</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
    <div id="call-container">
        <!-- 通话控制面板 -->
        <div id="call-controls" style="display: none;">
            <button id="start-audio-call">语音通话</button>
            <button id="start-video-call">视频通话</button>
            <button id="answer-call">接听</button>
            <button id="reject-call">拒绝</button>
            <button id="end-call">挂断</button>
            <button id="mute-audio">静音</button>
            <button id="mute-video">关闭视频</button>
        </div>
        
        <!-- 视频显示区域 -->
        <div id="video-container">
            <video id="local-video" autoplay muted></video>
            <video id="remote-video" autoplay></video>
        </div>
        
        <!-- 通话状态显示 -->
        <div id="call-status">
            <span id="connection-status">未连接</span>
            <span id="call-duration">00:00</span>
        </div>
    </div>
    
    <!-- 引入必要的脚本 -->
    <script src="js/im/adapter/api.js"></script>
    <script src="js/im/adapter/ws.js"></script>
    <script src="js/im/webrtc/WebRTCManager.js"></script>
    <script src="js/call-app.js"></script>
</body>
</html>
```

#### CSS 样式
```css
#call-container {
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
    font-family: Arial, sans-serif;
}

#call-controls {
    text-align: center;
    margin: 20px 0;
}

#call-controls button {
    margin: 5px;
    padding: 10px 20px;
    font-size: 16px;
    border: none;
    border-radius: 5px;
    cursor: pointer;
    background-color: #007bff;
    color: white;
}

#call-controls button:hover {
    background-color: #0056b3;
}

#call-controls button:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
}

#video-container {
    display: flex;
    gap: 10px;
    margin: 20px 0;
}

#local-video, #remote-video {
    width: 300px;
    height: 200px;
    border: 2px solid #ddd;
    border-radius: 10px;
    background-color: #000;
}

#remote-video {
    width: 100%;
    height: 400px;
}

#call-status {
    text-align: center;
    margin: 20px 0;
    font-size: 18px;
}

#connection-status {
    color: #28a745;
    font-weight: bold;
}

#call-duration {
    color: #6c757d;
    margin-left: 20px;
}
```

#### JavaScript 实现
```javascript
// call-app.js
(function () {
  'use strict';

  var CallApp = {
    // 状态
    isInCall: false,
    callDuration: 0,
    durationTimer: null,
    
    // 元素引用
    elements: {},
    
    // 初始化
    init: function () {
      this.initElements();
      this.bindEvents();
      this.setupWebRTCListeners();
      
      console.log('[CallApp] 通话应用已初始化');
    },
    
    // 初始化元素引用
    initElements: function () {
      this.elements = {
        startAudioCall: document.getElementById('start-audio-call'),
        startVideoCall: document.getElementById('start-video-call'),
        answerCall: document.getElementById('answer-call'),
        rejectCall: document.getElementById('reject-call'),
        endCall: document.getElementById('end-call'),
        muteAudio: document.getElementById('mute-audio'),
        muteVideo: document.getElementById('mute-video'),
        callControls: document.getElementById('call-controls'),
        localVideo: document.getElementById('local-video'),
        remoteVideo: document.getElementById('remote-video'),
        connectionStatus: document.getElementById('connection-status'),
        callDuration: document.getElementById('call-duration')
      };
    },
    
    // 绑定事件
    bindEvents: function () {
      var self = this;
      
      // 发起语音通话
      this.elements.startAudioCall.addEventListener('click', function () {
        var toUserId = prompt('请输入目标用户ID:');
        if (toUserId) {
          self.startCall(parseInt(toUserId), 'audio');
        }
      });
      
      // 发起视频通话
      this.elements.startVideoCall.addEventListener('click', function () {
        var toUserId = prompt('请输入目标用户ID:');
        if (toUserId) {
          self.startCall(parseInt(toUserId), 'video');
        }
      });
      
      // 接听通话
      this.elements.answerCall.addEventListener('click', function () {
        self.answerCall();
      });
      
      // 拒绝通话
      this.elements.rejectCall.addEventListener('click', function () {
        self.rejectCall();
      });
      
      // 结束通话
      this.elements.endCall.addEventListener('click', function () {
        self.endCall();
      });
      
      // 静音/取消静音
      this.elements.muteAudio.addEventListener('click', function () {
        self.toggleAudioMute();
      });
      
      // 关闭/开启视频
      this.elements.muteVideo.addEventListener('click', function () {
        self.toggleVideoMute();
      });
    },
    
    // 设置 WebRTC 事件监听
    setupWebRTCListeners: function () {
      var self = this;
      
      if (window.WebRTCManager) {
        // 本地流事件
        window.WebRTCManager.subscribe('localStream', function (stream) {
          self.elements.localVideo.srcObject = stream;
          console.log('[CallApp] 本地流已设置');
        });
        
        // 远程流事件
        window.WebRTCManager.subscribe('remoteStream', function (stream) {
          self.elements.remoteVideo.srcObject = stream;
          console.log('[CallApp] 远程流已设置');
        });
        
        // 来电事件
        window.WebRTCManager.subscribe('incomingCall', function (data) {
          self.showIncomingCall(data);
        });
        
        // 通话结束事件
        window.WebRTCManager.subscribe('callEnded', function (data) {
          self.handleCallEnded(data);
        });
        
        // 连接状态变化
        window.WebRTCManager.subscribe('connectionStateChange', function (state) {
          self.updateConnectionStatus(state);
        });
        
        // 错误事件
        window.WebRTCManager.subscribe('error', function (error) {
          self.handleError(error);
        });
      }
    },
    
    // 发起通话
    startCall: function (toUserId, callType) {
      console.log('[CallApp] 发起通话:', toUserId, callType);
      
      if (window.WebRTCManager) {
        window.WebRTCManager.startCall(toUserId, callType);
        this.showCallControls();
        this.startDurationTimer();
      }
    },
    
    // 应答通话
    answerCall: function () {
      console.log('[CallApp] 应答通话');
      
      if (window.WebRTCManager && window.WebRTCManager.currentCall) {
        window.WebRTCManager.answerCall(window.WebRTCManager.currentCall.id);
        this.showCallControls();
        this.startDurationTimer();
      }
    },
    
    // 拒绝通话
    rejectCall: function () {
      console.log('[CallApp] 拒绝通话');
      
      if (window.WebRTCManager && window.WebRTCManager.currentCall) {
        window.WebRTCManager.rejectCall(window.WebRTCManager.currentCall.id);
        this.hideCallControls();
      }
    },
    
    // 结束通话
    endCall: function () {
      console.log('[CallApp] 结束通话');
      
      if (window.WebRTCManager) {
        window.WebRTCManager.endCall();
        this.hideCallControls();
        this.stopDurationTimer();
      }
    },
    
    // 切换音频静音
    toggleAudioMute: function () {
      if (this.elements.localVideo.srcObject) {
        var audioTracks = this.elements.localVideo.srcObject.getAudioTracks();
        audioTracks.forEach(function (track) {
          track.enabled = !track.enabled;
        });
        
        var button = this.elements.muteAudio;
        button.textContent = track.enabled ? '静音' : '取消静音';
        button.style.backgroundColor = track.enabled ? '#007bff' : '#dc3545';
      }
    },
    
    // 切换视频静音
    toggleVideoMute: function () {
      if (this.elements.localVideo.srcObject) {
        var videoTracks = this.elements.localVideo.srcObject.getVideoTracks();
        videoTracks.forEach(function (track) {
          track.enabled = !track.enabled;
        });
        
        var button = this.elements.muteVideo;
        button.textContent = track.enabled ? '关闭视频' : '开启视频';
        button.style.backgroundColor = track.enabled ? '#007bff' : '#dc3545';
      }
    },
    
    // 显示来电界面
    showIncomingCall: function (data) {
      console.log('[CallApp] 收到来电:', data);
      
      // 显示接听/拒绝按钮
      this.elements.answerCall.style.display = 'inline-block';
      this.elements.rejectCall.style.display = 'inline-block';
      this.elements.startAudioCall.style.display = 'none';
      this.elements.startVideoCall.style.display = 'none';
      this.elements.endCall.style.display = 'none';
      
      // 更新状态
      this.updateConnectionStatus('ringing');
      
      // 播放铃声
      this.playRingtone();
    },
    
    // 显示通话控制界面
    showCallControls: function () {
      this.elements.callControls.style.display = 'block';
      this.elements.answerCall.style.display = 'none';
      this.elements.rejectCall.style.display = 'none';
      this.elements.startAudioCall.style.display = 'none';
      this.elements.startVideoCall.style.display = 'none';
      this.elements.endCall.style.display = 'inline-block';
      this.elements.muteAudio.style.display = 'inline-block';
      this.elements.muteVideo.style.display = 'inline-block';
    },
    
    // 隐藏通话控制界面
    hideCallControls: function () {
      this.elements.callControls.style.display = 'none';
      this.elements.localVideo.srcObject = null;
      this.elements.remoteVideo.srcObject = null;
      this.updateConnectionStatus('未连接');
    },
    
    // 处理通话结束
    handleCallEnded: function (data) {
      console.log('[CallApp] 通话结束:', data);
      this.hideCallControls();
      this.stopDurationTimer();
      this.stopRingtone();
    },
    
    // 更新连接状态
    updateConnectionStatus: function (state) {
      var statusText = '';
      var statusColor = '';
      
      switch (state) {
        case 'connecting':
          statusText = '连接中...';
          statusColor = '#ffc107';
          break;
        case 'connected':
          statusText = '已连接';
          statusColor = '#28a745';
          break;
        case 'disconnected':
          statusText = '已断开';
          statusColor = '#dc3545';
          break;
        case 'ringing':
          statusText = '响铃中...';
          statusColor = '#17a2b8';
          break;
        default:
          statusText = '未连接';
          statusColor = '#6c757d';
      }
      
      this.elements.connectionStatus.textContent = statusText;
      this.elements.connectionStatus.style.color = statusColor;
    },
    
    // 开始计时
    startDurationTimer: function () {
      this.callDuration = 0;
      this.durationTimer = setInterval(function () {
        this.callDuration++;
        this.updateCallDuration();
      }.bind(this), 1000);
    },
    
    // 停止计时
    stopDurationTimer: function () {
      if (this.durationTimer) {
        clearInterval(this.durationTimer);
        this.durationTimer = null;
      }
      this.callDuration = 0;
      this.updateCallDuration();
    },
    
    // 更新通话时长显示
    updateCallDuration: function () {
      var minutes = Math.floor(this.callDuration / 60);
      var seconds = this.callDuration % 60;
      var durationText = String(minutes).padStart(2, '0') + ':' + String(seconds).padStart(2, '0');
      this.elements.callDuration.textContent = durationText;
    },
    
    // 播放铃声
    playRingtone: function () {
      // 这里可以播放铃声
      console.log('[CallApp] 播放铃声');
    },
    
    // 停止铃声
    stopRingtone: function () {
      // 这里停止铃声
      console.log('[CallApp] 停止铃声');
    },
    
    // 处理错误
    handleError: function (error) {
      console.error('[CallApp] 通话错误:', error);
      alert('通话错误: ' + error.message);
    }
  };
  
  // 页面加载完成后初始化
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () {
      CallApp.init();
    });
  } else {
    CallApp.init();
  }
})();
```

### 2. 高级功能示例

#### 屏幕共享
```javascript
// 屏幕共享功能
function startScreenShare() {
  if (navigator.mediaDevices && navigator.mediaDevices.getDisplayMedia) {
    navigator.mediaDevices.getDisplayMedia({
      video: {
        cursor: 'always',
        displaySurface: 'monitor'
      },
      audio: false
    }).then(function (stream) {
      // 将屏幕共享流添加到 PeerConnection
      var screenTrack = stream.getVideoTracks()[0];
      var sender = peerConnection.getSenders().find(function (s) {
        return s.track && s.track.kind === 'video';
      });
      
      if (sender) {
        sender.replaceTrack(screenTrack);
      }
      
      // 显示屏幕共享流
      localVideo.srcObject = stream;
      
      // 监听屏幕共享结束
      screenTrack.onended = function () {
        console.log('屏幕共享已结束');
        // 恢复摄像头
        startCamera();
      };
      
    }).catch(function (error) {
      console.error('屏幕共享失败:', error);
    });
  } else {
    console.error('浏览器不支持屏幕共享');
  }
}

// 开始摄像头
function startCamera() {
  navigator.mediaDevices.getUserMedia({
    video: true,
    audio: true
  }).then(function (stream) {
    var videoTrack = stream.getVideoTracks()[0];
    var sender = peerConnection.getSenders().find(function (s) {
      return s.track && s.track.kind === 'video';
    });
    
    if (sender) {
      sender.replaceTrack(videoTrack);
    }
    
    localVideo.srcObject = stream;
  });
}
```

#### 录制功能
```javascript
// 录制通话
var mediaRecorder;
var recordedChunks = [];

function startRecording() {
  if (remoteVideo.srcObject) {
    mediaRecorder = new MediaRecorder(remoteVideo.srcObject);
    
    mediaRecorder.ondataavailable = function (event) {
      if (event.data.size > 0) {
        recordedChunks.push(event.data);
      }
    };
    
    mediaRecorder.onstop = function () {
      var blob = new Blob(recordedChunks, { type: 'video/webm' });
      var url = URL.createObjectURL(blob);
      
      // 下载录制文件
      var a = document.createElement('a');
      a.href = url;
      a.download = 'call-recording-' + Date.now() + '.webm';
      a.click();
      
      URL.revokeObjectURL(url);
      recordedChunks = [];
    };
    
    mediaRecorder.start();
    console.log('开始录制');
  }
}

function stopRecording() {
  if (mediaRecorder && mediaRecorder.state === 'recording') {
    mediaRecorder.stop();
    console.log('停止录制');
  }
}
```

## Android 端集成示例

### 1. 通话界面

#### 布局文件 (activity_call.xml)
```xml
<?xml version="1.0" encoding="utf-8"?>
<LinearLayout xmlns:android="http://schemas.android.com/apk/res/android"
    android:layout_width="match_parent"
    android:layout_height="match_parent"
    android:orientation="vertical"
    android:background="@color/black">

    <!-- 远程视频 -->
    <org.webrtc.SurfaceViewRenderer
        android:id="@+id/remote_video_view"
        android:layout_width="match_parent"
        android:layout_height="0dp"
        android:layout_weight="1" />

    <!-- 本地视频 -->
    <org.webrtc.SurfaceViewRenderer
        android:id="@+id/local_video_view"
        android:layout_width="120dp"
        android:layout_height="160dp"
        android:layout_gravity="top|end"
        android:layout_margin="16dp" />

    <!-- 通话控制 -->
    <LinearLayout
        android:layout_width="match_parent"
        android:layout_height="wrap_content"
        android:orientation="horizontal"
        android:gravity="center"
        android:padding="16dp"
        android:background="@color/black_transparent">

        <!-- 静音按钮 -->
        <ImageButton
            android:id="@+id/btn_mute_audio"
            android:layout_width="60dp"
            android:layout_height="60dp"
            android:layout_margin="8dp"
            android:src="@drawable/ic_mic_on"
            android:background="@drawable/circle_button"
            android:scaleType="centerInside" />

        <!-- 挂断按钮 -->
        <ImageButton
            android:id="@+id/btn_end_call"
            android:layout_width="60dp"
            android:layout_height="60dp"
            android:layout_margin="8dp"
            android:src="@drawable/ic_call_end"
            android:background="@drawable/circle_button_red"
            android:scaleType="centerInside" />

        <!-- 切换摄像头按钮 -->
        <ImageButton
            android:id="@+id/btn_switch_camera"
            android:layout_width="60dp"
            android:layout_height="60dp"
            android:layout_margin="8dp"
            android:src="@drawable/ic_switch_camera"
            android:background="@drawable/circle_button"
            android:scaleType="centerInside" />

        <!-- 关闭视频按钮 -->
        <ImageButton
            android:id="@+id/btn_mute_video"
            android:layout_width="60dp"
            android:layout_height="60dp"
            android:layout_margin="8dp"
            android:src="@drawable/ic_videocam_on"
            android:background="@drawable/circle_button"
            android:scaleType="centerInside" />

    </LinearLayout>

    <!-- 通话状态 -->
    <TextView
        android:id="@+id/tv_call_status"
        android:layout_width="match_parent"
        android:layout_height="wrap_content"
        android:text="连接中..."
        android:textColor="@color/white"
        android:textSize="16sp"
        android:gravity="center"
        android:padding="8dp" />

    <!-- 通话时长 -->
    <TextView
        android:id="@+id/tv_call_duration"
        android:layout_width="match_parent"
        android:layout_height="wrap_content"
        android:text="00:00"
        android:textColor="@color/white"
        android:textSize="14sp"
        android:gravity="center"
        android:padding="4dp" />

</LinearLayout>
```

#### Activity 实现 (CallActivity.kt)
```kotlin
package org.telegram.im.webrtc

import android.Manifest
import android.content.pm.PackageManager
import android.os.Bundle
import android.view.View
import android.widget.ImageButton
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import kotlinx.coroutines.*
import org.telegram.im.adapter.WebSocketService
import org.webrtc.*

class CallActivity : AppCompatActivity() {
    
    companion object {
        private const val TAG = "CallActivity"
        private const val PERMISSION_REQUEST_CODE = 1001
        
        const val EXTRA_CALL_ID = "call_id"
        const val EXTRA_CALL_TYPE = "call_type"
        const val EXTRA_IS_INCOMING = "is_incoming"
    }
    
    private lateinit var webRTCManager: WebRTCManager
    private lateinit var wsService: WebSocketService
    
    // UI 元素
    private lateinit var remoteVideoView: SurfaceViewRenderer
    private lateinit var localVideoView: SurfaceViewRenderer
    private lateinit var btnMuteAudio: ImageButton
    private lateinit var btnEndCall: ImageButton
    private lateinit var btnSwitchCamera: ImageButton
    private lateinit var btnMuteVideo: ImageButton
    private lateinit var tvCallStatus: TextView
    private lateinit var tvCallDuration: TextView
    
    // 状态
    private var callId: String? = null
    private var callType: String? = null
    private var isIncoming: Boolean = false
    private var isAudioMuted: Boolean = false
    private var isVideoMuted: Boolean = false
    private var callDuration: Long = 0
    private var durationTimer: Job? = null
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_call)
        
        // 获取参数
        callId = intent.getStringExtra(EXTRA_CALL_ID)
        callType = intent.getStringExtra(EXTRA_CALL_TYPE)
        isIncoming = intent.getBooleanExtra(EXTRA_IS_INCOMING, false)
        
        // 初始化 UI
        initViews()
        
        // 检查权限
        if (checkPermissions()) {
            initWebRTC()
        } else {
            requestPermissions()
        }
    }
    
    private fun initViews() {
        remoteVideoView = findViewById(R.id.remote_video_view)
        localVideoView = findViewById(R.id.local_video_view)
        btnMuteAudio = findViewById(R.id.btn_mute_audio)
        btnEndCall = findViewById(R.id.btn_end_call)
        btnSwitchCamera = findViewById(R.id.btn_switch_camera)
        btnMuteVideo = findViewById(R.id.btn_mute_video)
        tvCallStatus = findViewById(R.id.tv_call_status)
        tvCallDuration = findViewById(R.id.tv_call_duration)
        
        // 设置点击事件
        btnMuteAudio.setOnClickListener { toggleAudioMute() }
        btnEndCall.setOnClickListener { endCall() }
        btnSwitchCamera.setOnClickListener { switchCamera() }
        btnMuteVideo.setOnClickListener { toggleVideoMute() }
        
        // 初始化视频视图
        remoteVideoView.init(EglBase.create().eglBaseContext, null)
        localVideoView.init(EglBase.create().eglBaseContext, null)
    }
    
    private fun checkPermissions(): Boolean {
        val permissions = arrayOf(
            Manifest.permission.CAMERA,
            Manifest.permission.RECORD_AUDIO
        )
        
        return permissions.all { permission ->
            ContextCompat.checkSelfPermission(this, permission) == PackageManager.PERMISSION_GRANTED
        }
    }
    
    private fun requestPermissions() {
        val permissions = arrayOf(
            Manifest.permission.CAMERA,
            Manifest.permission.RECORD_AUDIO
        )
        
        ActivityCompat.requestPermissions(this, permissions, PERMISSION_REQUEST_CODE)
    }
    
    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<out String>,
        grantResults: IntArray
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        
        if (requestCode == PERMISSION_REQUEST_CODE) {
            if (grantResults.all { it == PackageManager.PERMISSION_GRANTED }) {
                initWebRTC()
            } else {
                finish()
            }
        }
    }
    
    private fun initWebRTC() {
        webRTCManager = WebRTCManager.getInstance()
        wsService = WebSocketService.getInstance()
        
        // 初始化 WebRTC 管理器
        webRTCManager.initialize(this)
        
        // 设置事件监听
        setupWebRTCListeners()
        
        // 根据通话类型处理
        if (isIncoming) {
            handleIncomingCall()
        } else {
            startOutgoingCall()
        }
    }
    
    private fun setupWebRTCListeners() {
        // 本地流事件
        webRTCManager.subscribe("localStream") { stream ->
            runOnUiThread {
                val mediaStream = stream as MediaStream
                mediaStream.videoTracks.forEach { track ->
                    track.addSink(localVideoView)
                }
            }
        }
        
        // 远程流事件
        webRTCManager.subscribe("remoteStream") { stream ->
            runOnUiThread {
                val mediaStream = stream as MediaStream
                mediaStream.videoTracks.forEach { track ->
                    track.addSink(remoteVideoView)
                }
            }
        }
        
        // 通话结束事件
        webRTCManager.subscribe("callEnded") { data ->
            runOnUiThread {
                finish()
            }
        }
        
        // 连接状态变化
        webRTCManager.subscribe("connectionStateChange") { state ->
            runOnUiThread {
                updateCallStatus(state as String)
            }
        }
        
        // 错误事件
        webRTCManager.subscribe("error") { error ->
            runOnUiThread {
                handleError(error)
            }
        }
    }
    
    private fun handleIncomingCall() {
        updateCallStatus("响铃中...")
        // 显示接听/拒绝按钮
        // 播放铃声
    }
    
    private fun startOutgoingCall() {
        callId?.let { id ->
            callType?.let { type ->
                CoroutineScope(Dispatchers.Main).launch {
                    webRTCManager.startCall(2L, type) // 假设目标用户ID为2
                    startDurationTimer()
                }
            }
        }
    }
    
    private fun answerCall() {
        callId?.let { id ->
            CoroutineScope(Dispatchers.Main).launch {
                webRTCManager.answerCall(id)
                startDurationTimer()
            }
        }
    }
    
    private fun endCall() {
        webRTCManager.endCall()
        finish()
    }
    
    private fun toggleAudioMute() {
        isAudioMuted = !isAudioMuted
        btnMuteAudio.setImageResource(
            if (isAudioMuted) R.drawable.ic_mic_off else R.drawable.ic_mic_on
        )
        // 实现音频静音逻辑
    }
    
    private fun toggleVideoMute() {
        isVideoMuted = !isVideoMuted
        btnMuteVideo.setImageResource(
            if (isVideoMuted) R.drawable.ic_videocam_off else R.drawable.ic_videocam_on
        )
        // 实现视频静音逻辑
    }
    
    private fun switchCamera() {
        // 实现摄像头切换逻辑
    }
    
    private fun updateCallStatus(status: String) {
        tvCallStatus.text = status
    }
    
    private fun startDurationTimer() {
        durationTimer = CoroutineScope(Dispatchers.Main).launch {
            while (isActive) {
                callDuration++
                updateCallDuration()
                delay(1000)
            }
        }
    }
    
    private fun updateCallDuration() {
        val minutes = callDuration / 60
        val seconds = callDuration % 60
        val durationText = String.format("%02d:%02d", minutes, seconds)
        tvCallDuration.text = durationText
    }
    
    private fun handleError(error: Any) {
        // 处理错误
        finish()
    }
    
    override fun onDestroy() {
        super.onDestroy()
        durationTimer?.cancel()
        webRTCManager.dispose()
    }
}
```

### 3. 权限配置

#### AndroidManifest.xml
```xml
<manifest xmlns:android="http://schemas.android.com/apk/res/android">
    
    <!-- 摄像头权限 -->
    <uses-permission android:name="android.permission.CAMERA" />
    
    <!-- 录音权限 -->
    <uses-permission android:name="android.permission.RECORD_AUDIO" />
    
    <!-- 网络权限 -->
    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
    
    <!-- 唤醒权限 -->
    <uses-permission android:name="android.permission.WAKE_LOCK" />
    
    <!-- 前台服务权限 -->
    <uses-permission android:name="android.permission.FOREGROUND_SERVICE" />
    
    <application>
        <!-- 通话 Activity -->
        <activity
            android:name=".webrtc.CallActivity"
            android:screenOrientation="portrait"
            android:theme="@style/Theme.AppCompat.NoActionBar" />
        
        <!-- 通话服务 -->
        <service
            android:name=".webrtc.CallService"
            android:enabled="true"
            android:exported="false" />
        
    </application>
</manifest>
```

### 4. 依赖配置

#### build.gradle
```gradle
dependencies {
    // WebRTC 依赖
    implementation 'org.webrtc:google-webrtc:1.0.32006'
    
    // 协程支持
    implementation 'org.jetbrains.kotlinx:kotlinx-coroutines-android:1.7.3'
    
    // 权限处理
    implementation 'androidx.core:core-ktx:1.12.0'
    implementation 'androidx.activity:activity-ktx:1.8.2'
    
    // 网络请求
    implementation 'com.squareup.okhttp3:okhttp:4.12.0'
    implementation 'com.squareup.okhttp3:logging-interceptor:4.12.0'
}
```

## 测试示例

### 1. 功能测试

#### Web 端测试
```javascript
// 测试 WebRTC 功能
function testWebRTC() {
  console.log('开始测试 WebRTC 功能');
  
  // 测试浏览器支持
  if (!window.RTCPeerConnection) {
    console.error('浏览器不支持 WebRTC');
    return;
  }
  
  // 测试媒体设备
  navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(function (stream) {
      console.log('媒体设备测试成功');
      stream.getTracks().forEach(function (track) {
        track.stop();
      });
    })
    .catch(function (error) {
      console.error('媒体设备测试失败:', error);
    });
  
  // 测试 PeerConnection
  var pc = new RTCPeerConnection();
  console.log('PeerConnection 创建成功');
  pc.close();
}

// 测试信令
function testSignaling() {
  if (window.IMWebSocket) {
    // 测试 WebSocket 连接
    if (window.IMWebSocket.isConnected()) {
      console.log('WebSocket 连接正常');
      
      // 发送测试信令
      window.IMWebSocket.send('call.offer', {
        call_id: 'test_call_123',
        from_user: 1,
        to_user: 2,
        call_type: 'video'
      });
    } else {
      console.error('WebSocket 未连接');
    }
  }
}
```

#### Android 端测试
```kotlin
// 测试 WebRTC 功能
class WebRTCTest {
    
    fun testWebRTC() {
        Log.d("WebRTCTest", "开始测试 WebRTC 功能")
        
        // 测试权限
        if (checkPermissions()) {
            Log.d("WebRTCTest", "权限检查通过")
        } else {
            Log.e("WebRTCTest", "权限检查失败")
            return
        }
        
        // 测试 WebRTC 管理器
        val webRTCManager = WebRTCManager.getInstance()
        webRTCManager.initialize(context)
        
        // 测试信令
        testSignaling()
    }
    
    private fun testSignaling() {
        val wsService = WebSocketService.getInstance()
        if (wsService.isConnected()) {
            Log.d("WebRTCTest", "WebSocket 连接正常")
            
            // 发送测试信令
            wsService.send("call.offer", mapOf(
                "call_id" to "test_call_123",
                "from_user" to 1L,
                "to_user" to 2L,
                "call_type" to "video"
            ))
        } else {
            Log.e("WebRTCTest", "WebSocket 未连接")
        }
    }
}
```

### 2. 性能测试

#### 延迟测试
```javascript
// 测试通话延迟
function testLatency() {
  var startTime = Date.now();
  
  // 发送测试数据
  var dataChannel = peerConnection.createDataChannel('test');
  dataChannel.onopen = function () {
    dataChannel.send('ping');
  };
  
  dataChannel.onmessage = function (event) {
    var latency = Date.now() - startTime;
    console.log('延迟:', latency, 'ms');
  };
}
```

#### 带宽测试
```javascript
// 测试带宽
function testBandwidth() {
  setInterval(function () {
    peerConnection.getStats().then(function (stats) {
      stats.forEach(function (report) {
        if (report.type === 'candidate-pair' && report.state === 'succeeded') {
          console.log('可用带宽:', report.availableOutgoingBitrate, 'bps');
        }
      });
    });
  }, 1000);
}
```

## 部署配置

### 1. 服务器配置

#### TURN 服务器配置
```bash
# 安装 coturn
sudo apt-get install coturn

# 配置 TURN 服务器
sudo nano /etc/turnserver.conf
```

```ini
# TURN 服务器配置
listening-port=3478
tls-listening-port=5349
listening-ip=0.0.0.0
external-ip=YOUR_PUBLIC_IP
realm=im-suite.com
server-name=im-suite.com
user=im-user:im-password
cert=/etc/ssl/certs/turn.crt
pkey=/etc/ssl/private/turn.key
log-file=/var/log/turnserver.log
verbose
```

### 2. 客户端配置

#### Web 端配置
```javascript
// WebRTC 配置
const rtcConfig = {
  iceServers: [
    { urls: 'stun:stun.l.google.com:19302' },
    { 
      urls: 'turn:turn.im-suite.com:3478',
      username: 'im-user',
      credential: 'im-password'
    }
  ],
  iceCandidatePoolSize: 10,
  bundlePolicy: 'max-bundle',
  rtcpMuxPolicy: 'require'
};
```

#### Android 端配置
```kotlin
// WebRTC 配置
val rtcConfig = PeerConnection.RTCConfiguration(
    listOf(
        PeerConnection.IceServer.builder("stun:stun.l.google.com:19302").createIceServer(),
        PeerConnection.IceServer.builder("turn:turn.im-suite.com:3478")
            .setUsername("im-user")
            .setPassword("im-password")
            .createIceServer()
    )
)
```

## 最佳实践

### 1. 开发建议
- 使用最新的 WebRTC API
- 实现完整的错误处理
- 添加网络质量监控
- 支持多种编解码器
- 实现自适应质量调整

### 2. 部署建议
- 使用多个 TURN 服务器
- 配置负载均衡
- 监控服务器状态
- 定期更新证书
- 备份配置文件

### 3. 安全建议
- 使用强密码
- 定期轮换密钥
- 启用日志记录
- 限制访问权限
- 使用 HTTPS/WSS

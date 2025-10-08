import React, { useState, useEffect, useCallback } from 'react';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  Typography,
  Box,
  Button,
  Chip,
  Alert,
  CircularProgress,
  Tabs,
  Tab,
  Card,
  CardContent
} from '@mui/material';
import {
  Close,
  Download,
  Visibility,
  Lock,
  Image,
  VideoLibrary,
  AudioFile,
  Description,
  History,
  Share
} from '@mui/icons-material';

interface FilePreviewProps {
  fileId: number;
  open: boolean;
  onClose: () => void;
  onDownload?: (fileId: number) => void;
}

interface FileInfo {
  id: number;
  fileName: string;
  fileSize: number;
  fileType: string;
  mimeType: string;
  storageUrl: string;
  thumbnail?: string;
  preview?: string;
  isEncrypted: boolean;
  version: number;
  isLatest: boolean;
  ownerId: number;
  isPublic: boolean;
  downloadCount: number;
  viewCount: number;
  createdAt: string;
  updatedAt: string;
}

interface FileVersion {
  id: number;
  fileName: string;
  version: number;
  isLatest: boolean;
  createdAt: string;
}

const FilePreview: React.FC<FilePreviewProps> = ({
  fileId,
  open,
  onClose,
  onDownload
}) => {
  const [fileInfo, setFileInfo] = useState<FileInfo | null>(null);
  const [versions, setVersions] = useState<FileVersion[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState(0);
  const [previewLoading, setPreviewLoading] = useState(false);

  // 获取文件信息
  const fetchFileInfo = useCallback(async () => {
    if (!fileId) return;

    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`/api/files/${fileId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        }
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || '获取文件信息失败');
      }

      const data = await response.json();
      setFileInfo(data);

      // 获取文件版本
      fetchFileVersions();

    } catch (err) {
      setError(err instanceof Error ? err.message : '获取文件信息失败');
    } finally {
      setLoading(false);
    }
  }, [fileId]);

  // 获取文件版本列表
  const fetchFileVersions = useCallback(async () => {
    try {
      const response = await fetch(`/api/files/${fileId}/versions`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        }
      });

      if (response.ok) {
        const data = await response.json();
        setVersions(data);
      }
    } catch (err) {
      console.error('获取文件版本失败:', err);
    }
  }, [fileId]);

  // 获取文件预览
  const fetchPreview = useCallback(async (previewType: string = 'image') => {
    if (!fileInfo) return;

    setPreviewLoading(true);
    try {
      const response = await fetch(`/api/files/${fileId}/preview?preview_type=${previewType}&width=800&height=600&quality=high`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        }
      });

      if (response.ok) {
        const previewData = await response.json();
        // 更新文件信息中的预览URL
        setFileInfo(prev => prev ? { ...prev, preview: previewData.preview_url } : null);
      }
    } catch (err) {
      console.error('获取文件预览失败:', err);
    } finally {
      setPreviewLoading(false);
    }
  }, [fileId, fileInfo]);

  // 下载文件
  const handleDownload = useCallback(async () => {
    try {
      const response = await fetch(`/api/files/${fileId}/download`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        }
      });

      if (response.ok) {
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = fileInfo?.fileName || 'download';
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);

        if (onDownload) {
          onDownload(fileId);
        }
      } else {
        const errorData = await response.json();
        throw new Error(errorData.error || '下载失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '下载失败');
    }
  }, [fileId, fileInfo, onDownload]);

  // 格式化文件大小
  const formatFileSize = useCallback((bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }, []);

  // 获取文件类型图标
  const getFileTypeIcon = useCallback((mimeType: string) => {
    if (mimeType.startsWith('image/')) return <Image />;
    if (mimeType.startsWith('video/')) return <VideoLibrary />;
    if (mimeType.startsWith('audio/')) return <AudioFile />;
    return <Description />;
  }, []);

  // 渲染预览内容
  const renderPreview = useCallback(() => {
    if (!fileInfo) return null;

    const { mimeType, storageUrl, preview, thumbnail, isEncrypted } = fileInfo;

    if (mimeType.startsWith('image/')) {
      return (
        <Box sx={{ textAlign: 'center' }}>
          {previewLoading ? (
            <CircularProgress />
          ) : (
            <img
              src={preview || thumbnail || storageUrl}
              alt={fileInfo.fileName}
              style={{
                maxWidth: '100%',
                maxHeight: '500px',
                objectFit: 'contain'
              }}
            />
          )}
        </Box>
      );
    }

    if (mimeType.startsWith('video/')) {
      return (
        <Box sx={{ textAlign: 'center' }}>
          {previewLoading ? (
            <CircularProgress />
          ) : (
            <video
              controls
              style={{
                maxWidth: '100%',
                maxHeight: '500px'
              }}
            >
              <source src={storageUrl} type={mimeType} />
              您的浏览器不支持视频播放
            </video>
          )}
        </Box>
      );
    }

    if (mimeType.startsWith('audio/')) {
      return (
        <Box sx={{ textAlign: 'center' }}>
          {previewLoading ? (
            <CircularProgress />
          ) : (
            <audio controls style={{ width: '100%' }}>
              <source src={storageUrl} type={mimeType} />
              您的浏览器不支持音频播放
            </audio>
          )}
        </Box>
      );
    }

    // 文档类型
    return (
      <Card>
        <CardContent sx={{ textAlign: 'center', py: 4 }}>
          {getFileTypeIcon(mimeType)}
          <Typography variant="h6" sx={{ mt: 2, mb: 1 }}>
            {fileInfo.fileName}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            此文件类型不支持预览
          </Typography>
          <Button
            variant="contained"
            startIcon={<Download />}
            onClick={handleDownload}
            sx={{ mt: 2 }}
          >
            下载文件
          </Button>
        </CardContent>
      </Card>
    );
  }, [fileInfo, previewLoading, getFileTypeIcon, handleDownload]);

  // 渲染文件信息
  const renderFileInfo = useCallback(() => {
    if (!fileInfo) return null;

    return (
      <Box sx={{ p: 2 }}>
        <Typography variant="h6" gutterBottom>
          文件信息
        </Typography>
        
        <Box sx={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: 1, mb: 2 }}>
          <Typography variant="body2" color="text.secondary">文件名:</Typography>
          <Typography variant="body2">{fileInfo.fileName}</Typography>
          
          <Typography variant="body2" color="text.secondary">文件大小:</Typography>
          <Typography variant="body2">{formatFileSize(fileInfo.fileSize)}</Typography>
          
          <Typography variant="body2" color="text.secondary">文件类型:</Typography>
          <Typography variant="body2">{fileInfo.mimeType}</Typography>
          
          <Typography variant="body2" color="text.secondary">版本:</Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Typography variant="body2">v{fileInfo.version}</Typography>
            {fileInfo.isLatest && <Chip label="最新" size="small" color="primary" />}
          </Box>
          
          <Typography variant="body2" color="text.secondary">状态:</Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            {fileInfo.isEncrypted && <Lock color="primary" sx={{ fontSize: 16 }} />}
            <Chip 
              label={fileInfo.isEncrypted ? '已加密' : '未加密'} 
              size="small" 
              color={fileInfo.isEncrypted ? 'primary' : 'default'} 
            />
            {fileInfo.isPublic && <Chip label="公开" size="small" color="success" />}
          </Box>
          
          <Typography variant="body2" color="text.secondary">下载次数:</Typography>
          <Typography variant="body2">{fileInfo.downloadCount}</Typography>
          
          <Typography variant="body2" color="text.secondary">查看次数:</Typography>
          <Typography variant="body2">{fileInfo.viewCount}</Typography>
          
          <Typography variant="body2" color="text.secondary">创建时间:</Typography>
          <Typography variant="body2">
            {new Date(fileInfo.createdAt).toLocaleString()}
          </Typography>
          
          <Typography variant="body2" color="text.secondary">更新时间:</Typography>
          <Typography variant="body2">
            {new Date(fileInfo.updatedAt).toLocaleString()}
          </Typography>
        </Box>
      </Box>
    );
  }, [fileInfo, formatFileSize]);

  // 渲染版本历史
  const renderVersions = useCallback(() => {
    if (versions.length === 0) {
      return (
        <Box sx={{ p: 2, textAlign: 'center' }}>
          <Typography variant="body2" color="text.secondary">
            暂无版本历史
          </Typography>
        </Box>
      );
    }

    return (
      <Box sx={{ p: 2 }}>
        <Typography variant="h6" gutterBottom>
          版本历史
        </Typography>
        {versions.map((version) => (
          <Card key={version.id} sx={{ mb: 1 }}>
            <CardContent sx={{ py: 1.5 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="body2" fontWeight="medium">
                    {version.fileName}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    版本 {version.version} • {new Date(version.createdAt).toLocaleString()}
                  </Typography>
                </Box>
                {version.isLatest && (
                  <Chip label="当前版本" size="small" color="primary" />
                )}
              </Box>
            </CardContent>
          </Card>
        ))}
      </Box>
    );
  }, [versions]);

  // 组件挂载时获取文件信息
  useEffect(() => {
    if (open && fileId) {
      fetchFileInfo();
    }
  }, [open, fileId, fetchFileInfo]);

  // 切换标签页时获取预览
  useEffect(() => {
    if (activeTab === 0 && fileInfo && fileInfo.mimeType.startsWith('image/')) {
      fetchPreview('image');
    }
  }, [activeTab, fileInfo, fetchPreview]);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      PaperProps={{
        sx: { height: '80vh' }
      }}
    >
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          {fileInfo && getFileTypeIcon(fileInfo.mimeType)}
          <Typography variant="h6" noWrap sx={{ flex: 1 }}>
            {fileInfo?.fileName || '文件预览'}
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <IconButton onClick={handleDownload} size="small">
            <Download />
          </IconButton>
          <IconButton onClick={onClose} size="small">
            <Close />
          </IconButton>
        </Box>
      </DialogTitle>

      <DialogContent dividers>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
          </Box>
        ) : error ? (
          <Alert severity="error">{error}</Alert>
        ) : fileInfo ? (
          <>
            <Tabs value={activeTab} onChange={(e, newValue) => setActiveTab(newValue)}>
              <Tab icon={<Visibility />} label="预览" />
              <Tab icon={<Description />} label="信息" />
              <Tab icon={<History />} label="版本" />
            </Tabs>

            {activeTab === 0 && (
              <Box sx={{ mt: 2 }}>
                {renderPreview()}
              </Box>
            )}

            {activeTab === 1 && renderFileInfo()}
            {activeTab === 2 && renderVersions()}
          </>
        ) : null}
      </DialogContent>
    </Dialog>
  );
};

export default FilePreview;

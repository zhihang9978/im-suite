import React, { useState, useCallback, useRef } from 'react';
import { Button, Progress, Alert, Card, CardBody, Typography, IconButton } from '@mui/material';
import { CloudUpload, Cancel, CheckCircle, Lock, LockOpen } from '@mui/icons-material';

interface FileUploaderProps {
  onUploadComplete?: (file: UploadedFile) => void;
  onUploadError?: (error: string) => void;
  maxFileSize?: number; // 最大文件大小（字节）
  allowedTypes?: string[]; // 允许的文件类型
  chunkSize?: number; // 分片大小（字节）
  enableEncryption?: boolean; // 是否启用加密
}

interface UploadedFile {
  id: number;
  fileName: string;
  fileUrl: string;
  thumbnailUrl?: string;
  previewUrl?: string;
  fileSize: number;
  fileType: string;
  isEncrypted: boolean;
}

interface UploadProgress {
  fileName: string;
  progress: number;
  status: 'uploading' | 'completed' | 'error';
  error?: string;
  uploadedChunks?: number;
  totalChunks?: number;
}

const FileUploader: React.FC<FileUploaderProps> = ({
  onUploadComplete,
  onUploadError,
  maxFileSize = 100 * 1024 * 1024, // 默认100MB
  allowedTypes = ['image/*', 'video/*', 'audio/*', 'application/pdf', 'text/*'],
  chunkSize = 5 * 1024 * 1024, // 默认5MB
  enableEncryption = false
}) => {
  const [uploadProgress, setUploadProgress] = useState<UploadProgress[]>([]);
  const [isDragging, setIsDragging] = useState(false);
  const [isEncrypted, setIsEncrypted] = useState(enableEncryption);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 检查文件类型是否允许
  const isFileTypeAllowed = useCallback((file: File): boolean => {
    if (allowedTypes.length === 0) return true;
    return allowedTypes.some(type => {
      if (type.endsWith('/*')) {
        return file.type.startsWith(type.slice(0, -1));
      }
      return file.type === type;
    });
  }, [allowedTypes]);

  // 检查文件大小
  const isFileSizeValid = useCallback((file: File): boolean => {
    return file.size <= maxFileSize;
  }, [maxFileSize]);

  // 生成上传ID
  const generateUploadId = useCallback((): string => {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }, []);

  // 计算分片数量
  const calculateChunks = useCallback((fileSize: number): number => {
    return Math.ceil(fileSize / chunkSize);
  }, [chunkSize]);

  // 上传单个分片
  const uploadChunk = useCallback(async (
    file: File,
    chunkIndex: number,
    totalChunks: number,
    uploadId: string
  ): Promise<void> => {
    const start = chunkIndex * chunkSize;
    const end = Math.min(start + chunkSize, file.size);
    const chunk = file.slice(start, end);

    const formData = new FormData();
    formData.append('chunk', chunk);
    formData.append('upload_id', uploadId);
    formData.append('chunk_index', chunkIndex.toString());
    formData.append('total_chunks', totalChunks.toString());
    formData.append('file_name', file.name);
    formData.append('file_size', file.size.toString());
    formData.append('is_encrypted', isEncrypted.toString());

    try {
      const response = await fetch('/api/files/upload-chunk', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        },
        body: formData
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || '上传分片失败');
      }

      const result = await response.json();
      
      // 更新进度
      setUploadProgress(prev => prev.map(item => {
        if (item.fileName === file.name) {
          return {
            ...item,
            progress: Math.round(((chunkIndex + 1) / totalChunks) * 100),
            uploadedChunks: chunkIndex + 1,
            totalChunks,
            status: result.is_complete ? 'completed' : 'uploading'
          };
        }
        return item;
      }));

      // 如果上传完成，调用回调
      if (result.is_complete && onUploadComplete) {
        onUploadComplete({
          id: result.file_id,
          fileName: result.file_name,
          fileUrl: result.file_url,
          thumbnailUrl: result.thumbnail_url,
          previewUrl: result.preview_url,
          fileSize: file.size,
          fileType: file.type,
          isEncrypted: isEncrypted
        });
      }

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '上传分片失败';
      
      // 更新错误状态
      setUploadProgress(prev => prev.map(item => {
        if (item.fileName === file.name) {
          return {
            ...item,
            status: 'error',
            error: errorMessage
          };
        }
        return item;
      }));

      if (onUploadError) {
        onUploadError(errorMessage);
      }
      
      throw error;
    }
  }, [chunkSize, isEncrypted, onUploadComplete, onUploadError]);

  // 上传文件
  const uploadFile = useCallback(async (file: File): Promise<void> => {
    // 验证文件
    if (!isFileTypeAllowed(file)) {
      const error = `不支持的文件类型: ${file.type}`;
      if (onUploadError) onUploadError(error);
      return;
    }

    if (!isFileSizeValid(file)) {
      const error = `文件大小超过限制: ${(file.size / 1024 / 1024).toFixed(2)}MB > ${(maxFileSize / 1024 / 1024).toFixed(2)}MB`;
      if (onUploadError) onUploadError(error);
      return;
    }

    const totalChunks = calculateChunks(file.size);
    const uploadId = generateUploadId();

    // 添加进度跟踪
    setUploadProgress(prev => [...prev, {
      fileName: file.name,
      progress: 0,
      status: 'uploading',
      uploadedChunks: 0,
      totalChunks
    }]);

    try {
      // 如果只有一个分片，使用单文件上传
      if (totalChunks === 1) {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('is_encrypted', isEncrypted.toString());

        const response = await fetch('/api/files/upload', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('access_token')}`
          },
          body: formData
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || '文件上传失败');
        }

        const result = await response.json();

        // 更新进度为完成
        setUploadProgress(prev => prev.map(item => {
          if (item.fileName === file.name) {
            return {
              ...item,
              progress: 100,
              status: 'completed'
            };
          }
          return item;
        }));

        if (onUploadComplete) {
          onUploadComplete({
            id: result.file_id,
            fileName: result.file_name,
            fileUrl: result.file_url,
            thumbnailUrl: result.thumbnail_url,
            previewUrl: result.preview_url,
            fileSize: file.size,
            fileType: file.type,
            isEncrypted: isEncrypted
          });
        }
      } else {
        // 分片上传
        for (let i = 0; i < totalChunks; i++) {
          await uploadChunk(file, i, totalChunks, uploadId);
        }
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '文件上传失败';
      
      setUploadProgress(prev => prev.map(item => {
        if (item.fileName === file.name) {
          return {
            ...item,
            status: 'error',
            error: errorMessage
          };
        }
        return item;
      }));

      if (onUploadError) {
        onUploadError(errorMessage);
      }
    }
  }, [
    isFileTypeAllowed,
    isFileSizeValid,
    calculateChunks,
    generateUploadId,
    uploadChunk,
    isEncrypted,
    onUploadComplete,
    onUploadError,
    maxFileSize
  ]);

  // 处理文件选择
  const handleFileSelect = useCallback((files: FileList | null) => {
    if (!files) return;

    Array.from(files).forEach(file => {
      uploadFile(file);
    });
  }, [uploadFile]);

  // 处理拖拽
  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    handleFileSelect(e.dataTransfer.files);
  }, [handleFileSelect]);

  // 处理点击上传
  const handleClickUpload = useCallback(() => {
    fileInputRef.current?.click();
  }, []);

  // 移除进度项
  const removeProgressItem = useCallback((fileName: string) => {
    setUploadProgress(prev => prev.filter(item => item.fileName !== fileName));
  }, []);

  return (
    <div className="file-uploader">
      {/* 上传区域 */}
      <Card 
        className={`upload-area ${isDragging ? 'dragging' : ''}`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={handleClickUpload}
        sx={{ 
          cursor: 'pointer',
          border: isDragging ? '2px dashed #1976d2' : '2px dashed #ccc',
          backgroundColor: isDragging ? '#f5f5f5' : 'transparent',
          transition: 'all 0.3s ease'
        }}
      >
        <CardBody sx={{ textAlign: 'center', py: 4 }}>
          <CloudUpload sx={{ fontSize: 48, color: '#666', mb: 2 }} />
          <Typography variant="h6" gutterBottom>
            点击或拖拽文件到此处上传
          </Typography>
          <Typography variant="body2" color="text.secondary" gutterBottom>
            支持的文件类型: {allowedTypes.join(', ')}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            最大文件大小: {(maxFileSize / 1024 / 1024).toFixed(0)}MB
          </Typography>
          
          {/* 加密选项 */}
          <div style={{ marginTop: '16px' }}>
            <Button
              variant={isEncrypted ? 'contained' : 'outlined'}
              color={isEncrypted ? 'primary' : 'default'}
              startIcon={isEncrypted ? <Lock /> : <LockOpen />}
              onClick={(e) => {
                e.stopPropagation();
                setIsEncrypted(!isEncrypted);
              }}
            >
              {isEncrypted ? '加密上传' : '普通上传'}
            </Button>
          </div>
        </CardBody>
      </Card>

      {/* 隐藏的文件输入 */}
      <input
        ref={fileInputRef}
        type="file"
        multiple
        accept={allowedTypes.join(',')}
        onChange={(e) => handleFileSelect(e.target.files)}
        style={{ display: 'none' }}
      />

      {/* 上传进度 */}
      {uploadProgress.length > 0 && (
        <Card sx={{ mt: 2 }}>
          <CardBody>
            <Typography variant="h6" gutterBottom>
              上传进度
            </Typography>
            {uploadProgress.map((item, index) => (
              <div key={index} style={{ marginBottom: '16px' }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
                  <Typography variant="body2" noWrap sx={{ flex: 1, mr: 2 }}>
                    {item.fileName}
                    {item.isEncrypted && <Lock sx={{ fontSize: 14, ml: 1 }} />}
                  </Typography>
                  <div style={{ display: 'flex', alignItems: 'center' }}>
                    {item.status === 'completed' && <CheckCircle sx={{ color: 'success.main', mr: 1 }} />}
                    {item.status === 'error' && (
                      <IconButton size="small" onClick={() => removeProgressItem(item.fileName)}>
                        <Cancel color="error" />
                      </IconButton>
                    )}
                  </div>
                </div>
                
                <Progress 
                  variant="determinate" 
                  value={item.progress} 
                  color={item.status === 'completed' ? 'success' : item.status === 'error' ? 'error' : 'primary'}
                  sx={{ mb: 1 }}
                />
                
                <Typography variant="caption" color="text.secondary">
                  {item.uploadedChunks && item.totalChunks 
                    ? `${item.uploadedChunks}/${item.totalChunks} 分片`
                    : `${item.progress}%`
                  }
                </Typography>
                
                {item.status === 'error' && item.error && (
                  <Alert severity="error" sx={{ mt: 1 }}>
                    {item.error}
                  </Alert>
                )}
              </div>
            ))}
          </CardBody>
        </Card>
      )}
    </div>
  );
};

export default FileUploader;

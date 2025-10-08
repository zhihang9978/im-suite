import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Grid,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  LinearProgress,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material';
import {
  Speed,
  Storage,
  NetworkCheck,
  Memory,
  Assessment,
  Settings,
  Cleanup,
  Optimize,
} from '@mui/icons-material';

interface PerformanceStats {
  push: {
    queue_length: number;
    recent_batches: number;
    worker_count: number;
    batch_size: number;
  };
  storage: {
    total_size_mb: number;
    tables: Array<{
      table_name: string;
      size_mb: number;
      rows: number;
    }>;
  };
  network: {
    total_requests: number;
    compressed_requests: number;
    cache_hits: number;
    cache_misses: number;
    average_latency: number;
    compression_ratio: number;
    bandwidth_saved: number;
  };
}

interface ChatStats {
  member_count: number;
  message_count: number;
  today_message_count: number;
  active_member_count: number;
}

const PerformanceOptimizer: React.FC = () => {
  const [stats, setStats] = useState<PerformanceStats | null>(null);
  const [chatStats, setChatStats] = useState<ChatStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [cleanupDialogOpen, setCleanupDialogOpen] = useState(false);
  const [optimizationDialogOpen, setOptimizationDialogOpen] = useState(false);
  const [selectedChat, setSelectedChat] = useState<number | null>(null);
  const [cleanupDays, setCleanupDays] = useState(30);

  useEffect(() => {
    fetchPerformanceStats();
  }, []);

  const fetchPerformanceStats = async () => {
    try {
      setLoading(true);
      const [pushResponse, storageResponse, networkResponse] = await Promise.all([
        fetch('/api/performance/push/stats'),
        fetch('/api/performance/storage/stats'),
        fetch('/api/performance/network/stats'),
      ]);

      const pushStats = await pushResponse.json();
      const storageStats = await storageResponse.json();
      const networkStats = await networkResponse.json();

      setStats({
        push: pushStats.data,
        storage: storageStats.data,
        network: networkStats.data,
      });
    } catch (err) {
      setError('获取性能统计失败');
      console.error('获取性能统计失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchChatStats = async (chatId: number) => {
    try {
      const response = await fetch(`/api/performance/groups/${chatId}/statistics`);
      const data = await response.json();
      setChatStats(data.data);
    } catch (err) {
      console.error('获取聊天统计失败:', err);
    }
  };

  const handleOptimizeDatabase = async () => {
    try {
      await fetch('/api/performance/database/optimize', { method: 'POST' });
      fetchPerformanceStats();
    } catch (err) {
      console.error('数据库优化失败:', err);
    }
  };

  const handleCleanupInactiveMembers = async () => {
    if (!selectedChat) return;

    try {
      await fetch(
        `/api/performance/groups/${selectedChat}/cleanup-members?days=${cleanupDays}`,
        { method: 'POST' }
      );
      fetchChatStats(selectedChat);
      setCleanupDialogOpen(false);
    } catch (err) {
      console.error('清理不活跃成员失败:', err);
    }
  };

  const handleCompressTable = async (tableName: string) => {
    try {
      await fetch(`/api/performance/storage/compress/${tableName}`, { method: 'POST' });
      fetchPerformanceStats();
    } catch (err) {
      console.error('表压缩失败:', err);
    }
  };

  const handleCleanupOldMessages = async () => {
    try {
      await fetch(`/api/performance/storage/cleanup/messages?days=${cleanupDays}`, {
        method: 'POST',
      });
      fetchPerformanceStats();
    } catch (err) {
      console.error('清理旧消息失败:', err);
    }
  };

  const getPerformanceColor = (value: number, thresholds: [number, number]) => {
    if (value <= thresholds[0]) return 'success';
    if (value <= thresholds[1]) return 'warning';
    return 'error';
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (loading) {
    return (
      <Box sx={{ p: 3 }}>
        <LinearProgress />
        <Typography variant="h6" sx={{ mt: 2 }}>
          正在加载性能数据...
        </Typography>
      </Box>
    );
  }

  if (error) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error">{error}</Alert>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        性能优化中心
      </Typography>

      {stats && (
        <Grid container spacing={3}>
          {/* 消息推送统计 */}
          <Grid item xs={12} md={6}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Speed sx={{ mr: 1 }} />
                  <Typography variant="h6">消息推送统计</Typography>
                </Box>
                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      队列长度
                    </Typography>
                    <Typography variant="h6">
                      {stats.push.queue_length}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      工作协程
                    </Typography>
                    <Typography variant="h6">
                      {stats.push.worker_count}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      批量大小
                    </Typography>
                    <Typography variant="h6">
                      {stats.push.batch_size}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      最近批次
                    </Typography>
                    <Typography variant="h6">
                      {stats.push.recent_batches}
                    </Typography>
                  </Grid>
                </Grid>
              </CardContent>
            </Card>
          </Grid>

          {/* 网络统计 */}
          <Grid item xs={12} md={6}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <NetworkCheck sx={{ mr: 1 }} />
                  <Typography variant="h6">网络统计</Typography>
                </Box>
                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      总请求数
                    </Typography>
                    <Typography variant="h6">
                      {stats.network.total_requests.toLocaleString()}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      压缩率
                    </Typography>
                    <Typography variant="h6">
                      {(stats.network.compression_ratio * 100).toFixed(1)}%
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      缓存命中
                    </Typography>
                    <Typography variant="h6">
                      {stats.network.cache_hits.toLocaleString()}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      平均延迟
                    </Typography>
                    <Typography variant="h6">
                      {stats.network.average_latency.toFixed(0)}ms
                    </Typography>
                  </Grid>
                </Grid>
              </CardContent>
            </Card>
          </Grid>

          {/* 存储统计 */}
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Storage sx={{ mr: 1 }} />
                  <Typography variant="h6">存储统计</Typography>
                  <Box sx={{ ml: 'auto' }}>
                    <Button
                      variant="outlined"
                      startIcon={<Optimize />}
                      onClick={handleOptimizeDatabase}
                      sx={{ mr: 1 }}
                    >
                      优化数据库
                    </Button>
                    <Button
                      variant="outlined"
                      startIcon={<Cleanup />}
                      onClick={() => setOptimizationDialogOpen(true)}
                    >
                      清理优化
                    </Button>
                  </Box>
                </Box>
                <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
                  总大小: {formatBytes(stats.storage.total_size_mb * 1024 * 1024)}
                </Typography>
                <TableContainer component={Paper} variant="outlined">
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>表名</TableCell>
                        <TableCell align="right">大小</TableCell>
                        <TableCell align="right">行数</TableCell>
                        <TableCell align="center">操作</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {stats.storage.tables.slice(0, 10).map((table) => (
                        <TableRow key={table.table_name}>
                          <TableCell>{table.table_name}</TableCell>
                          <TableCell align="right">
                            {formatBytes(table.size_mb * 1024 * 1024)}
                          </TableCell>
                          <TableCell align="right">
                            {table.rows.toLocaleString()}
                          </TableCell>
                          <TableCell align="center">
                            <Button
                              size="small"
                              onClick={() => handleCompressTable(table.table_name)}
                            >
                              压缩
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              </CardContent>
            </Card>
          </Grid>

          {/* 聊天统计 */}
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Assessment sx={{ mr: 1 }} />
                  <Typography variant="h6">聊天统计</Typography>
                </Box>
                <Grid container spacing={2}>
                  <Grid item xs={12} md={3}>
                    <FormControl fullWidth size="small">
                      <InputLabel>选择聊天</InputLabel>
                      <Select
                        value={selectedChat || ''}
                        onChange={(e) => {
                          const chatId = e.target.value as number;
                          setSelectedChat(chatId);
                          if (chatId) {
                            fetchChatStats(chatId);
                          }
                        }}
                      >
                        <MenuItem value={1}>聊天 1</MenuItem>
                        <MenuItem value={2}>聊天 2</MenuItem>
                        <MenuItem value={3}>聊天 3</MenuItem>
                      </Select>
                    </FormControl>
                  </Grid>
                  {chatStats && (
                    <>
                      <Grid item xs={6} md={2}>
                        <Typography variant="body2" color="textSecondary">
                          成员数量
                        </Typography>
                        <Typography variant="h6">{chatStats.member_count}</Typography>
                      </Grid>
                      <Grid item xs={6} md={2}>
                        <Typography variant="body2" color="textSecondary">
                          消息数量
                        </Typography>
                        <Typography variant="h6">
                          {chatStats.message_count.toLocaleString()}
                        </Typography>
                      </Grid>
                      <Grid item xs={6} md={2}>
                        <Typography variant="body2" color="textSecondary">
                          今日消息
                        </Typography>
                        <Typography variant="h6">
                          {chatStats.today_message_count}
                        </Typography>
                      </Grid>
                      <Grid item xs={6} md={2}>
                        <Typography variant="body2" color="textSecondary">
                          活跃成员
                        </Typography>
                        <Typography variant="h6">
                          {chatStats.active_member_count}
                        </Typography>
                      </Grid>
                      <Grid item xs={12} md={1}>
                        <Button
                          variant="outlined"
                          size="small"
                          onClick={() => setCleanupDialogOpen(true)}
                          disabled={!selectedChat}
                        >
                          清理
                        </Button>
                      </Grid>
                    </>
                  )}
                </Grid>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}

      {/* 清理不活跃成员对话框 */}
      <Dialog
        open={cleanupDialogOpen}
        onClose={() => setCleanupDialogOpen(false)}
      >
        <DialogTitle>清理不活跃成员</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label="不活跃天数"
            type="number"
            value={cleanupDays}
            onChange={(e) => setCleanupDays(parseInt(e.target.value) || 30)}
            sx={{ mt: 2 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCleanupDialogOpen(false)}>取消</Button>
          <Button onClick={handleCleanupInactiveMembers} variant="contained">
            确认清理
          </Button>
        </DialogActions>
      </Dialog>

      {/* 存储优化对话框 */}
      <Dialog
        open={optimizationDialogOpen}
        onClose={() => setOptimizationDialogOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>存储优化</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label="清理天数"
            type="number"
            value={cleanupDays}
            onChange={(e) => setCleanupDays(parseInt(e.target.value) || 30)}
            sx={{ mt: 2 }}
          />
          <Box sx={{ mt: 2, display: 'flex', gap: 2 }}>
            <Button
              variant="outlined"
              onClick={handleCleanupOldMessages}
              fullWidth
            >
              清理旧消息
            </Button>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOptimizationDialogOpen(false)}>关闭</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default PerformanceOptimizer;

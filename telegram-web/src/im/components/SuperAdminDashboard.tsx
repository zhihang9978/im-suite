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
  Tabs,
  Tab,
  IconButton,
  Tooltip,
  Badge,
} from '@mui/material';
import {
  Dashboard,
  People,
  Security,
  Assessment,
  Warning,
  Block,
  CheckCircle,
  Delete,
  Refresh,
  Settings,
  Notifications,
  BarChart,
  Timeline,
  Storage,
  Speed,
  Memory,
  Visibility,
  VisibilityOff,
  ExitToApp,
  Report,
  Message,
} from '@mui/icons-material';

interface SystemStats {
  total_users: number;
  online_users: number;
  total_messages: number;
  today_messages: number;
  total_groups: number;
  active_groups: number;
  total_files: number;
  storage_used: number;
  server_load: number;
  memory_usage: number;
  cpu_usage: number;
  database_size: number;
}

interface OnlineUser {
  user_id: number;
  username: string;
  nickname: string;
  avatar: string;
  online_status: string;
  ip_address: string;
  device: string;
  location: string;
  login_time: string;
  last_activity: string;
  session_count: number;
}

interface UserBehaviorAnalysis {
  user_id: number;
  username: string;
  message_count: number;
  group_count: number;
  file_upload_count: number;
  online_time: number;
  last_login_time: string;
  risk_score: number;
  violation_count: number;
  reported_count: number;
  is_blacklisted: boolean;
  is_suspicious: boolean;
}

const SuperAdminDashboard: React.FC = () => {
  const [tabValue, setTabValue] = useState(0);
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [onlineUsers, setOnlineUsers] = useState<OnlineUser[]>([]);
  const [selectedUser, setSelectedUser] = useState<OnlineUser | null>(null);
  const [userAnalysis, setUserAnalysis] = useState<UserBehaviorAnalysis | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionDialogOpen, setActionDialogOpen] = useState(false);
  const [actionType, setActionType] = useState<string>('');
  const [actionReason, setActionReason] = useState('');
  const [banDuration, setBanDuration] = useState(24);

  useEffect(() => {
    loadSystemStats();
    loadOnlineUsers();
    
    // 每30秒刷新一次数据
    const interval = setInterval(() => {
      loadSystemStats();
      loadOnlineUsers();
    }, 30000);

    return () => clearInterval(interval);
  }, []);

  const loadSystemStats = async () => {
    try {
      const response = await fetch('/api/super-admin/stats');
      const data = await response.json();
      setStats(data.data);
    } catch (err) {
      console.error('加载系统统计失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadOnlineUsers = async () => {
    try {
      const response = await fetch('/api/super-admin/online-users');
      const data = await response.json();
      setOnlineUsers(data.data.users);
    } catch (err) {
      console.error('加载在线用户失败:', err);
    }
  };

  const loadUserAnalysis = async (userId: number) => {
    try {
      const response = await fetch(`/api/super-admin/users/${userId}/analysis`);
      const data = await response.json();
      setUserAnalysis(data.data);
    } catch (err) {
      console.error('加载用户分析失败:', err);
    }
  };

  const handleUserAction = (user: OnlineUser, action: string) => {
    setSelectedUser(user);
    setActionType(action);
    setActionDialogOpen(true);
  };

  const executeUserAction = async () => {
    if (!selectedUser) return;

    try {
      let url = '';
      let method = 'POST';
      let body: any = { reason: actionReason };

      switch (actionType) {
        case 'logout':
          url = `/api/super-admin/users/${selectedUser.user_id}/force-logout`;
          break;
        case 'ban':
          url = `/api/super-admin/users/${selectedUser.user_id}/ban`;
          body.duration = banDuration;
          break;
        case 'unban':
          url = `/api/super-admin/users/${selectedUser.user_id}/unban`;
          break;
        case 'delete':
          url = `/api/super-admin/users/${selectedUser.user_id}`;
          method = 'DELETE';
          break;
      }

      await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      setActionDialogOpen(false);
      setActionReason('');
      loadOnlineUsers();
    } catch (err) {
      console.error('执行操作失败:', err);
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDuration = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}小时${minutes}分钟`;
  };

  const getRiskColor = (score: number) => {
    if (score >= 80) return 'error';
    if (score >= 60) return 'warning';
    if (score >= 40) return 'info';
    return 'success';
  };

  if (loading) {
    return (
      <Box sx={{ p: 3 }}>
        <LinearProgress />
        <Typography variant="h6" sx={{ mt: 2 }}>
          正在加载管理面板...
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
        <Dashboard sx={{ mr: 1, fontSize: 32 }} />
        <Typography variant="h4">超级管理后台</Typography>
        <Button
          variant="outlined"
          startIcon={<Refresh />}
          onClick={() => {
            loadSystemStats();
            loadOnlineUsers();
          }}
          sx={{ ml: 'auto' }}
        >
          刷新数据
        </Button>
      </Box>

      {/* 系统统计卡片 */}
      {stats && (
        <Grid container spacing={3} sx={{ mb: 3 }}>
          <Grid item xs={12} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <People color="primary" sx={{ mr: 1 }} />
                  <Typography variant="h6">总用户数</Typography>
                </Box>
                <Typography variant="h4" sx={{ mt: 1 }}>
                  {stats.total_users.toLocaleString()}
                </Typography>
                <Typography variant="body2" color="textSecondary">
                  在线: {stats.online_users.toLocaleString()}
                </Typography>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Message color="success" sx={{ mr: 1 }} />
                  <Typography variant="h6">消息统计</Typography>
                </Box>
                <Typography variant="h4" sx={{ mt: 1 }}>
                  {stats.total_messages.toLocaleString()}
                </Typography>
                <Typography variant="body2" color="textSecondary">
                  今日: {stats.today_messages.toLocaleString()}
                </Typography>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Storage color="warning" sx={{ mr: 1 }} />
                  <Typography variant="h6">存储使用</Typography>
                </Box>
                <Typography variant="h4" sx={{ mt: 1 }}>
                  {formatBytes(stats.storage_used)}
                </Typography>
                <Typography variant="body2" color="textSecondary">
                  数据库: {formatBytes(stats.database_size)}
                </Typography>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Speed color="error" sx={{ mr: 1 }} />
                  <Typography variant="h6">服务器负载</Typography>
                </Box>
                <Typography variant="h4" sx={{ mt: 1 }}>
                  {stats.cpu_usage.toFixed(1)}%
                </Typography>
                <Typography variant="body2" color="textSecondary">
                  内存: {stats.memory_usage.toFixed(1)}%
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}

      {/* 标签页 */}
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs value={tabValue} onChange={(e, v) => setTabValue(v)}>
          <Tab icon={<People />} label="在线用户" />
          <Tab icon={<Assessment />} label="用户分析" />
          <Tab icon={<Security />} label="内容审核" />
          <Tab icon={<Timeline />} label="系统日志" />
        </Tabs>
      </Box>

      {/* 在线用户列表 */}
      {tabValue === 0 && (
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              在线用户列表 ({onlineUsers.length} 人)
            </Typography>
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>用户ID</TableCell>
                    <TableCell>用户名</TableCell>
                    <TableCell>状态</TableCell>
                    <TableCell>IP地址</TableCell>
                    <TableCell>设备</TableCell>
                    <TableCell>登录时间</TableCell>
                    <TableCell>会话数</TableCell>
                    <TableCell>操作</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {onlineUsers.map((user) => (
                    <TableRow key={user.user_id}>
                      <TableCell>{user.user_id}</TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center' }}>
                          {user.avatar && (
                            <img
                              src={user.avatar}
                              alt={user.username}
                              style={{ width: 32, height: 32, borderRadius: '50%', marginRight: 8 }}
                            />
                          )}
                          <Box>
                            <Typography variant="body2">{user.username}</Typography>
                            <Typography variant="caption" color="textSecondary">
                              {user.nickname}
                            </Typography>
                          </Box>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={user.online_status}
                          color={user.online_status === 'online' ? 'success' : 'default'}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>{user.ip_address}</TableCell>
                      <TableCell>{user.device}</TableCell>
                      <TableCell>
                        {new Date(user.login_time).toLocaleString()}
                      </TableCell>
                      <TableCell>
                        <Badge badgeContent={user.session_count} color="primary">
                          <Memory />
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Tooltip title="查看详情">
                          <IconButton
                            size="small"
                            onClick={() => loadUserAnalysis(user.user_id)}
                          >
                            <Visibility />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="强制下线">
                          <IconButton
                            size="small"
                            color="warning"
                            onClick={() => handleUserAction(user, 'logout')}
                          >
                            <ExitToApp />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="封禁用户">
                          <IconButton
                            size="small"
                            color="error"
                            onClick={() => handleUserAction(user, 'ban')}
                          >
                            <Block />
                          </IconButton>
                        </Tooltip>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </CardContent>
        </Card>
      )}

      {/* 用户分析面板 */}
      {tabValue === 1 && userAnalysis && (
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              用户行为分析 - {userAnalysis.username}
            </Typography>
            <Grid container spacing={3}>
              <Grid item xs={12} md={4}>
                <Box sx={{ p: 2, border: 1, borderColor: 'divider', borderRadius: 1 }}>
                  <Typography variant="subtitle2" color="textSecondary">
                    风险评分
                  </Typography>
                  <Box sx={{ display: 'flex', alignItems: 'center', mt: 1 }}>
                    <Typography variant="h3" color={getRiskColor(userAnalysis.risk_score)}>
                      {userAnalysis.risk_score.toFixed(0)}
                    </Typography>
                    <LinearProgress
                      variant="determinate"
                      value={userAnalysis.risk_score}
                      color={getRiskColor(userAnalysis.risk_score)}
                      sx={{ ml: 2, flex: 1 }}
                    />
                  </Box>
                  {userAnalysis.is_suspicious && (
                    <Alert severity="warning" sx={{ mt: 2 }}>
                      该用户行为可疑，建议重点关注
                    </Alert>
                  )}
                </Box>
              </Grid>

              <Grid item xs={12} md={8}>
                <Grid container spacing={2}>
                  <Grid item xs={6} md={3}>
                    <Typography variant="body2" color="textSecondary">
                      消息数量
                    </Typography>
                    <Typography variant="h6">
                      {userAnalysis.message_count.toLocaleString()}
                    </Typography>
                  </Grid>
                  <Grid item xs={6} md={3}>
                    <Typography variant="body2" color="textSecondary">
                      群组数量
                    </Typography>
                    <Typography variant="h6">
                      {userAnalysis.group_count.toLocaleString()}
                    </Typography>
                  </Grid>
                  <Grid item xs={6} md={3}>
                    <Typography variant="body2" color="textSecondary">
                      文件上传
                    </Typography>
                    <Typography variant="h6">
                      {userAnalysis.file_upload_count.toLocaleString()}
                    </Typography>
                  </Grid>
                  <Grid item xs={6} md={3}>
                    <Typography variant="body2" color="textSecondary">
                      在线时长
                    </Typography>
                    <Typography variant="h6">
                      {formatDuration(userAnalysis.online_time)}
                    </Typography>
                  </Grid>
                  <Grid item xs={6} md={3}>
                    <Typography variant="body2" color="textSecondary">
                      违规次数
                    </Typography>
                    <Typography variant="h6" color="error">
                      {userAnalysis.violation_count}
                    </Typography>
                  </Grid>
                  <Grid item xs={6} md={3}>
                    <Typography variant="body2" color="textSecondary">
                      被举报次数
                    </Typography>
                    <Typography variant="h6" color="warning.main">
                      {userAnalysis.reported_count}
                    </Typography>
                  </Grid>
                  <Grid item xs={6} md={3}>
                    <Typography variant="body2" color="textSecondary">
                      黑名单状态
                    </Typography>
                    <Chip
                      label={userAnalysis.is_blacklisted ? '已拉黑' : '正常'}
                      color={userAnalysis.is_blacklisted ? 'error' : 'success'}
                      size="small"
                    />
                  </Grid>
                  <Grid item xs={6} md={3}>
                    <Typography variant="body2" color="textSecondary">
                      最后登录
                    </Typography>
                    <Typography variant="body2">
                      {new Date(userAnalysis.last_login_time).toLocaleString()}
                    </Typography>
                  </Grid>
                </Grid>
              </Grid>
            </Grid>
          </CardContent>
        </Card>
      )}

      {/* 操作确认对话框 */}
      <Dialog open={actionDialogOpen} onClose={() => setActionDialogOpen(false)}>
        <DialogTitle>
          {actionType === 'logout' && '强制用户下线'}
          {actionType === 'ban' && '封禁用户'}
          {actionType === 'unban' && '解封用户'}
          {actionType === 'delete' && '删除用户账号'}
        </DialogTitle>
        <DialogContent>
          <Typography variant="body1" gutterBottom>
            您确定要对用户 <strong>{selectedUser?.username}</strong> 执行此操作吗？
          </Typography>

          {actionType === 'ban' && (
            <FormControl fullWidth sx={{ mt: 2 }}>
              <InputLabel>封禁时长</InputLabel>
              <Select
                value={banDuration}
                onChange={(e) => setBanDuration(e.target.value as number)}
              >
                <MenuItem value={1}>1小时</MenuItem>
                <MenuItem value={24}>24小时</MenuItem>
                <MenuItem value={168}>7天</MenuItem>
                <MenuItem value={720}>30天</MenuItem>
                <MenuItem value={8760}>永久</MenuItem>
              </Select>
            </FormControl>
          )}

          <TextField
            fullWidth
            multiline
            rows={3}
            label="操作原因"
            value={actionReason}
            onChange={(e) => setActionReason(e.target.value)}
            sx={{ mt: 2 }}
            required
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setActionDialogOpen(false)}>取消</Button>
          <Button
            variant="contained"
            color={actionType === 'delete' ? 'error' : 'primary'}
            onClick={executeUserAction}
            disabled={!actionReason}
          >
            确认
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default SuperAdminDashboard;

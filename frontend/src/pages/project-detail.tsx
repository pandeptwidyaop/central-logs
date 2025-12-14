import { useEffect, useState, useCallback } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Key,
  Copy,
  Check,
  Trash2,
  Edit,
  RefreshCw,
  Settings,
  Activity,
  Clock,
  AlertTriangle,
  Bell,
  Bug,
  Info,
  XCircle,
  Skull,
  Users,
  UserPlus,
} from 'lucide-react';
import { api, type Project, type LogEntry, type Channel, type ProjectMember, type User } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { useToast } from '@/hooks/use-toast';
import { usePushNotifications } from '@/hooks/use-push-notifications';
import { useAuth } from '@/contexts/auth-context';

const LOG_LEVELS = ['DEBUG', 'INFO', 'WARN', 'ERROR', 'CRITICAL'] as const;
type LogLevel = (typeof LOG_LEVELS)[number];

const levelIcons: Record<string, React.ReactNode> = {
  DEBUG: <Bug className="h-3.5 w-3.5" />,
  INFO: <Info className="h-3.5 w-3.5" />,
  WARN: <AlertTriangle className="h-3.5 w-3.5" />,
  ERROR: <XCircle className="h-3.5 w-3.5" />,
  CRITICAL: <Skull className="h-3.5 w-3.5" />,
};

const levelColors: Record<string, 'debug' | 'info' | 'warn' | 'error' | 'critical'> = {
  DEBUG: 'debug',
  INFO: 'info',
  WARN: 'warn',
  ERROR: 'error',
  CRITICAL: 'critical',
};

export function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [project, setProject] = useState<Project | null>(null);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [logsLoading, setLogsLoading] = useState(true);
  const [copied, setCopied] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [stats, setStats] = useState({
    total: 0,
    today: 0,
    errors: 0,
  });
  const [pushChannel, setPushChannel] = useState<Channel | null>(null);
  const [pushSettings, setPushSettings] = useState({
    enabled: false,
    minLevel: 'ERROR' as LogLevel,
  });
  const [pushLoading, setPushLoading] = useState(false);
  const [pushSaving, setPushSaving] = useState(false);
  // Members management state
  const [members, setMembers] = useState<ProjectMember[]>([]);
  const [membersLoading, setMembersLoading] = useState(true);
  const [allUsers, setAllUsers] = useState<User[]>([]);
  const [addMemberDialogOpen, setAddMemberDialogOpen] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState('');
  const [selectedRole, setSelectedRole] = useState('VIEWER');
  const { toast } = useToast();
  const { user: currentUser } = useAuth();
  const {
    permission: browserPermission,
    isSubscribed: isBrowserSubscribed,
    isLoading: isPushLoading,
    subscribe: subscribeBrowser,
    unsubscribe: unsubscribeBrowser,
    isSupported: isPushSupported,
  } = usePushNotifications();

  const fetchProject = useCallback(async () => {
    if (!id) return;
    try {
      const data = await api.getProject(id);
      setProject(data);
    } catch (err) {
      toast({
        title: 'Failed to load project',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
      navigate('/projects');
    } finally {
      setLoading(false);
    }
  }, [id, navigate, toast]);

  const fetchLogs = useCallback(async () => {
    if (!id) return;
    setLogsLoading(true);
    try {
      const result = await api.getLogs({
        project_id: id,
        limit: 10,
      });
      setLogs(result.logs ?? []);
      setStats({
        total: result.total,
        today: result.logs?.filter(
          (l) => new Date(l.created_at).toDateString() === new Date().toDateString()
        ).length ?? 0,
        errors: result.logs?.filter(
          (l) => l.level === 'ERROR' || l.level === 'CRITICAL'
        ).length ?? 0,
      });
    } finally {
      setLogsLoading(false);
    }
  }, [id]);

  const fetchPushChannel = useCallback(async () => {
    if (!id) return;
    setPushLoading(true);
    try {
      const channels = await api.getChannels(id);
      const push = channels.find((c) => c.type === 'PUSH');
      if (push) {
        setPushChannel(push);
        setPushSettings({
          enabled: push.is_active,
          minLevel: push.min_level,
        });
      } else {
        setPushChannel(null);
        setPushSettings({
          enabled: false,
          minLevel: 'ERROR',
        });
      }
    } catch {
      // Ignore errors - push channel may not exist yet
    } finally {
      setPushLoading(false);
    }
  }, [id]);

  const fetchMembers = useCallback(async () => {
    if (!id) return;
    setMembersLoading(true);
    try {
      const membersList = await api.getProjectMembers(id);
      setMembers(membersList);
    } catch (err) {
      toast({
        title: 'Failed to load members',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setMembersLoading(false);
    }
  }, [id, toast]);

  const fetchAllUsers = useCallback(async () => {
    try {
      const users = await api.getUsers();
      setAllUsers(users);
    } catch {
      // Ignore - users list may not be available for non-admin
    }
  }, []);

  useEffect(() => {
    fetchProject();
    fetchLogs();
    fetchPushChannel();
    fetchMembers();
    fetchAllUsers();
  }, [fetchProject, fetchLogs, fetchPushChannel, fetchMembers, fetchAllUsers]);

  const handleEditProject = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!project) return;
    const formData = new FormData(e.currentTarget);
    try {
      await api.updateProject(project.id, {
        name: formData.get('name') as string,
        description: formData.get('description') as string,
      });
      toast({ title: 'Project updated successfully' });
      setEditDialogOpen(false);
      fetchProject();
    } catch (err) {
      toast({
        title: 'Failed to update project',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    }
  };

  const handleRotateKey = async () => {
    if (!project) return;
    try {
      const result = await api.rotateApiKey(project.id);
      toast({
        title: 'API key rotated',
        description: `New key: ${result.api_key.substring(0, 20)}...`,
      });
      fetchProject();
    } catch (err) {
      toast({
        title: 'Failed to rotate API key',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    }
  };

  const handleDeleteProject = async () => {
    if (!project) return;
    if (!confirm(`Delete project "${project.name}"? This will delete all associated logs.`)) {
      return;
    }
    try {
      await api.deleteProject(project.id);
      toast({ title: 'Project deleted successfully' });
      navigate('/projects');
    } catch (err) {
      toast({
        title: 'Failed to delete project',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    }
  };

  const copyApiKey = () => {
    if (project?.api_key) {
      navigator.clipboard.writeText(project.api_key);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleSavePushSettings = async () => {
    if (!id) return;
    setPushSaving(true);
    try {
      if (pushChannel) {
        // Update existing channel
        await api.updateChannel(pushChannel.id, {
          min_level: pushSettings.minLevel,
          is_active: pushSettings.enabled,
        });
        toast({ title: 'Push notification settings updated' });
      } else {
        // Create new push channel
        await api.createChannel(id, {
          type: 'PUSH',
          name: 'Web Push Notifications',
          config: {},
          min_level: pushSettings.minLevel,
        });
        toast({ title: 'Push notifications enabled' });
      }
      await fetchPushChannel();
    } catch (err) {
      toast({
        title: 'Failed to save push settings',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setPushSaving(false);
    }
  };

  const handleAddMember = async () => {
    if (!id || !selectedUserId) return;
    try {
      await api.addProjectMember(id, selectedUserId, selectedRole);
      toast({ title: 'Member added successfully' });
      setAddMemberDialogOpen(false);
      setSelectedUserId('');
      setSelectedRole('VIEWER');
      fetchMembers();
    } catch (err) {
      toast({
        title: 'Failed to add member',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    }
  };

  const handleUpdateMemberRole = async (userId: string, newRole: string) => {
    if (!id) return;
    try {
      await api.updateProjectMemberRole(id, userId, newRole);
      toast({ title: 'Member role updated' });
      fetchMembers();
    } catch (err) {
      toast({
        title: 'Failed to update role',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    }
  };

  const handleRemoveMember = async (userId: string, username: string) => {
    if (!id) return;
    if (!confirm(`Remove ${username} from this project?`)) return;
    try {
      await api.removeProjectMember(id, userId);
      toast({ title: 'Member removed' });
      fetchMembers();
    } catch (err) {
      toast({
        title: 'Failed to remove member',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    }
  };

  // Get users that are not already members
  const availableUsers = allUsers.filter(
    (user) => !members.some((member) => member.user_id === user.id)
  );

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!project) {
    return (
      <div className="flex h-64 flex-col items-center justify-center">
        <p className="text-muted-foreground">Project not found</p>
        <Button asChild className="mt-4">
          <Link to="/projects">Back to Projects</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" asChild>
            <Link to="/projects">
              <ArrowLeft className="h-4 w-4" />
            </Link>
          </Button>
          <div>
            <h1 className="text-3xl font-bold">{project.name}</h1>
            <p className="text-muted-foreground">
              {project.description || 'No description'}
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setEditDialogOpen(true)}>
            <Edit className="mr-2 h-4 w-4" />
            Edit
          </Button>
          <Button variant="destructive" onClick={handleDeleteProject}>
            <Trash2 className="mr-2 h-4 w-4" />
            Delete
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Logs
            </CardTitle>
            <Activity className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total.toLocaleString()}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Logs Today
            </CardTitle>
            <Clock className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.today.toLocaleString()}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Errors
            </CardTitle>
            <AlertTriangle className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.errors.toLocaleString()}</div>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="logs">Recent Logs</TabsTrigger>
          <TabsTrigger value="members">Members</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          {/* API Key Card */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Key className="h-5 w-5" />
                API Key
              </CardTitle>
              <CardDescription>
                Use this key to send logs to this project
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {project.api_key ? (
                <div className="flex items-center gap-2">
                  <code className="flex-1 rounded-lg bg-muted px-4 py-3 font-mono text-sm">
                    {project.api_key}
                  </code>
                  <Button variant="outline" size="icon" onClick={copyApiKey}>
                    {copied ? (
                      <Check className="h-4 w-4 text-green-500" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              ) : (
                <p className="text-muted-foreground">No API key available</p>
              )}
              <Button variant="outline" onClick={handleRotateKey}>
                <RefreshCw className="mr-2 h-4 w-4" />
                Rotate API Key
              </Button>
            </CardContent>
          </Card>

          {/* Quick Integration */}
          <Card>
            <CardHeader>
              <CardTitle>Quick Integration</CardTitle>
              <CardDescription>
                Send logs using curl or any HTTP client
              </CardDescription>
            </CardHeader>
            <CardContent>
              <pre className="overflow-x-auto rounded-lg bg-muted p-4 text-sm">
{`curl -X POST http://localhost:3000/api/v1/logs \\
  -H "X-API-Key: ${project.api_key || 'YOUR_API_KEY'}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "level": "INFO",
    "message": "Hello from ${project.name}",
    "metadata": {"user": "test"}
  }'`}
              </pre>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="logs" className="space-y-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle>Recent Logs</CardTitle>
                <CardDescription>Last 10 logs for this project</CardDescription>
              </div>
              <Button variant="outline" size="sm" onClick={fetchLogs}>
                <RefreshCw className="mr-2 h-4 w-4" />
                Refresh
              </Button>
            </CardHeader>
            <CardContent className="p-0">
              {logsLoading ? (
                <div className="flex h-32 items-center justify-center">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
                </div>
              ) : logs.length === 0 ? (
                <div className="flex h-32 items-center justify-center text-muted-foreground">
                  No logs yet
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-24">Level</TableHead>
                      <TableHead>Message</TableHead>
                      <TableHead className="w-44">Time</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {logs.map((log) => (
                      <TableRow key={log.id}>
                        <TableCell>
                          <Badge variant={levelColors[log.level]}>{log.level}</Badge>
                        </TableCell>
                        <TableCell className="max-w-md truncate">{log.message}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {new Date(log.created_at).toLocaleString()}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
          <div className="flex justify-center">
            <Button variant="outline" asChild>
              <Link to={`/logs?project_id=${project.id}`}>View All Logs</Link>
            </Button>
          </div>
        </TabsContent>

        <TabsContent value="members" className="space-y-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <Users className="h-5 w-5" />
                  Project Members
                </CardTitle>
                <CardDescription>
                  Manage who has access to view logs for this project
                </CardDescription>
              </div>
              <Button onClick={() => setAddMemberDialogOpen(true)}>
                <UserPlus className="mr-2 h-4 w-4" />
                Add Member
              </Button>
            </CardHeader>
            <CardContent className="p-0">
              {membersLoading ? (
                <div className="flex h-32 items-center justify-center">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
                </div>
              ) : members.length === 0 ? (
                <div className="flex h-32 flex-col items-center justify-center text-muted-foreground">
                  <Users className="mb-2 h-8 w-8 opacity-50" />
                  <p>No members yet</p>
                  <p className="text-sm">Add members to give them access to this project's logs</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>User</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead className="w-44">Joined</TableHead>
                      <TableHead className="w-24">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {members.map((member) => (
                      <TableRow key={member.user_id}>
                        <TableCell>
                          <div>
                            <p className="font-medium">{member.user?.name || 'Unknown'}</p>
                            <p className="text-sm text-muted-foreground">{member.user?.email || ''}</p>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Select
                            value={member.role}
                            onValueChange={(value) => handleUpdateMemberRole(member.user_id, value)}
                          >
                            <SelectTrigger className="w-32">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="OWNER">Owner</SelectItem>
                              <SelectItem value="MEMBER">Member</SelectItem>
                              <SelectItem value="VIEWER">Viewer</SelectItem>
                            </SelectContent>
                          </Select>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {new Date(member.created_at).toLocaleDateString()}
                        </TableCell>
                        <TableCell>
                          {member.role !== 'OWNER' && member.user_id !== currentUser?.id && (
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleRemoveMember(member.user_id, member.user?.name || 'this user')}
                            >
                              <Trash2 className="h-4 w-4 text-destructive" />
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="settings" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Settings className="h-5 w-5" />
                Project Settings
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <Label className="text-muted-foreground">Project ID</Label>
                  <p className="font-mono">{project.id}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Created</Label>
                  <p>{new Date(project.created_at).toLocaleString()}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Updated</Label>
                  <p>{new Date(project.updated_at).toLocaleString()}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Push Notification Settings */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Bell className="h-5 w-5" />
                Push Notification Settings
              </CardTitle>
              <CardDescription>
                Configure which log levels trigger browser push notifications
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {pushLoading || isPushLoading ? (
                <div className="flex h-20 items-center justify-center">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
                </div>
              ) : !isPushSupported ? (
                <div className="rounded-lg bg-destructive/10 p-4 text-destructive">
                  <p className="text-sm font-medium">Push notifications are not supported in this browser.</p>
                </div>
              ) : (
                <>
                  {/* Browser Permission Status */}
                  <div className="rounded-lg border p-4 space-y-3">
                    <div className="flex items-center justify-between">
                      <div className="space-y-0.5">
                        <Label className="text-base">Browser Notifications</Label>
                        <p className="text-sm text-muted-foreground">
                          {browserPermission === 'granted'
                            ? isBrowserSubscribed
                              ? 'Your browser is subscribed to receive notifications'
                              : 'Permission granted but not subscribed'
                            : browserPermission === 'denied'
                            ? 'Notifications are blocked in your browser'
                            : 'Enable browser notifications to receive alerts'}
                        </p>
                      </div>
                      <Badge
                        variant={
                          isBrowserSubscribed ? 'info' : browserPermission === 'denied' ? 'error' : 'warn'
                        }
                      >
                        {isBrowserSubscribed ? 'Subscribed' : browserPermission === 'denied' ? 'Blocked' : 'Not Enabled'}
                      </Badge>
                    </div>
                    {browserPermission !== 'denied' && (
                      <div className="flex gap-2">
                        <Button
                          variant={isBrowserSubscribed ? 'outline' : 'default'}
                          size="sm"
                          onClick={async () => {
                            if (isBrowserSubscribed) {
                              const success = await unsubscribeBrowser();
                              if (success) {
                                toast({ title: 'Browser notifications disabled' });
                              }
                            } else {
                              const success = await subscribeBrowser();
                              if (success) {
                                toast({ title: 'Browser notifications enabled' });
                              } else {
                                toast({
                                  title: 'Failed to enable notifications',
                                  description: 'Please check your browser settings',
                                  variant: 'destructive',
                                });
                              }
                            }
                          }}
                        >
                          <Bell className="mr-2 h-4 w-4" />
                          {isBrowserSubscribed ? 'Disable Browser Notifications' : 'Enable Browser Notifications'}
                        </Button>
                        {isBrowserSubscribed && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={async () => {
                              try {
                                await api.testPushNotification();
                                toast({ title: 'Test notification sent!' });
                              } catch (error) {
                                toast({
                                  title: 'Failed to send test notification',
                                  description: error instanceof Error ? error.message : 'Unknown error',
                                  variant: 'destructive',
                                });
                              }
                            }}
                          >
                            Test Notification
                          </Button>
                        )}
                      </div>
                    )}
                    {browserPermission === 'denied' && (
                      <p className="text-xs text-muted-foreground">
                        To enable notifications, click the lock icon in your browser's address bar and allow notifications.
                      </p>
                    )}
                  </div>

                  {/* Project Channel Settings */}
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label htmlFor="push-enabled">Enable for this Project</Label>
                      <p className="text-sm text-muted-foreground">
                        Receive notifications when logs are received for this project
                      </p>
                    </div>
                    <Switch
                      id="push-enabled"
                      checked={pushSettings.enabled}
                      onCheckedChange={(checked) =>
                        setPushSettings((prev) => ({ ...prev, enabled: checked }))
                      }
                      disabled={!isBrowserSubscribed}
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="min-level">Minimum Log Level</Label>
                    <p className="text-sm text-muted-foreground">
                      Only notify for logs at this level or higher
                    </p>
                    <Select
                      value={pushSettings.minLevel}
                      onValueChange={(value: LogLevel) =>
                        setPushSettings((prev) => ({ ...prev, minLevel: value }))
                      }
                      disabled={!isBrowserSubscribed}
                    >
                      <SelectTrigger className="w-56">
                        <SelectValue placeholder="Select level" />
                      </SelectTrigger>
                      <SelectContent>
                        {LOG_LEVELS.map((level) => (
                          <SelectItem key={level} value={level}>
                            <div className="flex items-center gap-2">
                              <Badge variant={levelColors[level]} className="flex items-center gap-1 text-xs">
                                {levelIcons[level]}
                                {level}
                              </Badge>
                              <span className="text-xs text-muted-foreground whitespace-nowrap">
                                {level === 'DEBUG' && '+ above'}
                                {level === 'INFO' && '+ above'}
                                {level === 'WARN' && '+ above'}
                                {level === 'ERROR' && '+ CRITICAL'}
                                {level === 'CRITICAL' && 'only'}
                              </span>
                            </div>
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="rounded-lg bg-muted p-3">
                    <p className="text-sm text-muted-foreground">
                      {!isBrowserSubscribed ? (
                        'Enable browser notifications first to configure project notifications.'
                      ) : pushSettings.enabled ? (
                        <>
                          You will receive notifications for{' '}
                          <span className="font-medium text-foreground">
                            {pushSettings.minLevel === 'CRITICAL'
                              ? 'CRITICAL'
                              : `${pushSettings.minLevel} and above`}
                          </span>{' '}
                          logs.
                        </>
                      ) : (
                        'Push notifications are currently disabled for this project.'
                      )}
                    </p>
                  </div>

                  <Button onClick={handleSavePushSettings} disabled={pushSaving || !isBrowserSubscribed}>
                    {pushSaving ? (
                      <>
                        <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-background border-t-transparent" />
                        Saving...
                      </>
                    ) : (
                      'Save Settings'
                    )}
                  </Button>
                </>
              )}
            </CardContent>
          </Card>

          <Card className="border-destructive">
            <CardHeader>
              <CardTitle className="text-destructive">Danger Zone</CardTitle>
              <CardDescription>
                Irreversible actions that will affect your project
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between rounded-lg border border-destructive/50 p-4">
                <div>
                  <p className="font-medium">Delete this project</p>
                  <p className="text-sm text-muted-foreground">
                    This will permanently delete the project and all its logs
                  </p>
                </div>
                <Button variant="destructive" onClick={handleDeleteProject}>
                  Delete Project
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Edit Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Project</DialogTitle>
            <DialogDescription>Update project details</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleEditProject}>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="edit-name">Name</Label>
                <Input
                  id="edit-name"
                  name="name"
                  defaultValue={project.name}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-description">Description</Label>
                <Input
                  id="edit-description"
                  name="description"
                  defaultValue={project.description}
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setEditDialogOpen(false)}>
                Cancel
              </Button>
              <Button type="submit">Save</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Add Member Dialog */}
      <Dialog open={addMemberDialogOpen} onOpenChange={setAddMemberDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Project Member</DialogTitle>
            <DialogDescription>
              Select a user to give them access to this project's logs
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="user-select">User</Label>
              {availableUsers.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No users available to add. All users are already members.
                </p>
              ) : (
                <Select value={selectedUserId} onValueChange={setSelectedUserId}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a user" />
                  </SelectTrigger>
                  <SelectContent>
                    {availableUsers.map((user) => (
                      <SelectItem key={user.id} value={user.id}>
                        {user.name} (@{user.username})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="role-select">Role</Label>
              <Select value={selectedRole} onValueChange={setSelectedRole}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="OWNER">Owner - Full access</SelectItem>
                  <SelectItem value="MEMBER">Member - View and manage logs</SelectItem>
                  <SelectItem value="VIEWER">Viewer - View logs only</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setAddMemberDialogOpen(false);
                setSelectedUserId('');
                setSelectedRole('VIEWER');
              }}
            >
              Cancel
            </Button>
            <Button onClick={handleAddMember} disabled={!selectedUserId}>
              Add Member
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

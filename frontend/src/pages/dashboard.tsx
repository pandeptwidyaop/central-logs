import { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  FolderKanban,
  ScrollText,
  AlertTriangle,
  TrendingUp,
  Clock,
  Activity,
  ArrowRight,
  RefreshCw,
  Radio
} from 'lucide-react';
import { api, type DashboardStats, type LogEntry } from '@/lib/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useRealtimeLogs } from '@/hooks/use-realtime-logs';

const levelColors: Record<string, 'debug' | 'info' | 'warn' | 'error' | 'critical'> = {
  DEBUG: 'debug',
  INFO: 'info',
  WARN: 'warn',
  ERROR: 'error',
  CRITICAL: 'critical',
};

const levelBgColors: Record<string, string> = {
  DEBUG: 'bg-gray-500/10',
  INFO: 'bg-blue-500/10',
  WARN: 'bg-yellow-500/10',
  ERROR: 'bg-red-500/10',
  CRITICAL: 'bg-purple-600/10',
};

export function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchStats = () => {
    setLoading(true);
    api.getStats()
      .then(setStats)
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchStats();
  }, []);

  // Handle real-time log updates
  const handleNewLog = useCallback((newLog: LogEntry) => {
    setStats((prev) => {
      if (!prev) return prev;

      // Update recent_logs - add to top, keep max 10
      const updatedRecentLogs = [newLog, ...(prev.recent_logs || [])].slice(0, 10);

      // Update logs_by_level
      const updatedLogsByLevel = { ...prev.logs_by_level };
      updatedLogsByLevel[newLog.level] = (updatedLogsByLevel[newLog.level] || 0) + 1;

      return {
        ...prev,
        total_logs: (prev.total_logs || 0) + 1,
        logs_today: (prev.logs_today || 0) + 1,
        logs_by_level: updatedLogsByLevel,
        recent_logs: updatedRecentLogs,
      };
    });
  }, []);

  // Real-time connection for dashboard updates
  const { isConnected } = useRealtimeLogs({
    enabled: true,
    onNewLog: handleNewLog,
    playSound: false,  // Handled by AppLayout
    showToast: false,  // Handled by AppLayout
  });

  if (loading) {
    return (
      <div className="flex h-[60vh] items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <div className="h-10 w-10 animate-spin rounded-full border-4 border-primary border-t-transparent" />
          <p className="text-muted-foreground">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="flex h-[60vh] flex-col items-center justify-center">
        <AlertTriangle className="h-12 w-12 text-destructive mb-4" />
        <p className="text-lg font-medium">Failed to load dashboard</p>
        <p className="text-muted-foreground mb-4">Please try refreshing the page</p>
        <Button onClick={fetchStats}>
          <RefreshCw className="mr-2 h-4 w-4" />
          Retry
        </Button>
      </div>
    );
  }

  const statCards = [
    {
      title: 'Total Projects',
      value: stats.total_projects ?? 0,
      icon: FolderKanban,
      color: 'text-blue-500',
      bgColor: 'bg-blue-500/10',
      description: 'Active logging projects',
    },
    {
      title: 'Total Logs',
      value: (stats.total_logs ?? 0).toLocaleString(),
      icon: ScrollText,
      color: 'text-emerald-500',
      bgColor: 'bg-emerald-500/10',
      description: 'All time log entries',
    },
    {
      title: 'Logs Today',
      value: (stats.logs_today ?? 0).toLocaleString(),
      icon: Clock,
      color: 'text-violet-500',
      bgColor: 'bg-violet-500/10',
      description: 'Received in last 24 hours',
    },
    {
      title: 'Errors Today',
      value: ((stats.logs_by_level?.ERROR ?? 0) + (stats.logs_by_level?.CRITICAL ?? 0)).toLocaleString(),
      icon: AlertTriangle,
      color: 'text-rose-500',
      bgColor: 'bg-rose-500/10',
      description: 'Errors requiring attention',
    },
  ];

  const totalLogs = Object.values(stats.logs_by_level || {}).reduce((a, b) => a + b, 0);

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-muted-foreground">
            Welcome back! Here's an overview of your logging activity.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <div className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm ${
            isConnected ? 'bg-green-500/10 text-green-600' : 'bg-yellow-500/10 text-yellow-600'
          }`}>
            <Radio className={`h-3 w-3 ${isConnected ? 'animate-pulse' : ''}`} />
            {isConnected ? 'Live' : 'Connecting...'}
          </div>
          <Button variant="outline" onClick={fetchStats} disabled={loading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {statCards.map((stat) => (
          <Card key={stat.title} className="relative overflow-hidden">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {stat.title}
              </CardTitle>
              <div className={`rounded-lg p-2 ${stat.bgColor}`}>
                <stat.icon className={`h-4 w-4 ${stat.color}`} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">{stat.value}</div>
              <p className="text-xs text-muted-foreground mt-1">{stat.description}</p>
            </CardContent>
            <div className={`absolute bottom-0 left-0 right-0 h-1 ${stat.bgColor}`} />
          </Card>
        ))}
      </div>

      {/* Charts and Logs Row */}
      <div className="grid gap-6 lg:grid-cols-7">
        {/* Logs by Level - takes 3 columns */}
        <Card className="lg:col-span-3">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <Activity className="h-5 w-5" />
                  Logs by Level
                </CardTitle>
                <CardDescription>Distribution across log levels</CardDescription>
              </div>
              <div className="text-right">
                <p className="text-2xl font-bold">{totalLogs.toLocaleString()}</p>
                <p className="text-xs text-muted-foreground">Total logs</p>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {['DEBUG', 'INFO', 'WARN', 'ERROR', 'CRITICAL'].map((level) => {
                const count = stats.logs_by_level?.[level] || 0;
                const percentage = totalLogs > 0 ? (count / totalLogs) * 100 : 0;
                return (
                  <div key={level} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Badge variant={levelColors[level]} className="w-20 justify-center">
                          {level}
                        </Badge>
                        <span className="text-sm font-medium">{count.toLocaleString()}</span>
                      </div>
                      <span className="text-sm text-muted-foreground">
                        {percentage.toFixed(1)}%
                      </span>
                    </div>
                    <div className="h-2 rounded-full bg-muted overflow-hidden">
                      <div
                        className={`h-full rounded-full transition-all duration-500 ${
                          level === 'DEBUG' ? 'bg-gray-500' :
                          level === 'INFO' ? 'bg-blue-500' :
                          level === 'WARN' ? 'bg-yellow-500' :
                          level === 'ERROR' ? 'bg-red-500' :
                          'bg-purple-600'
                        }`}
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>

        {/* Recent Logs - takes 4 columns */}
        <Card className="lg:col-span-4">
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5" />
                Recent Activity
              </CardTitle>
              <CardDescription>Latest logs from all projects</CardDescription>
            </div>
            <Button variant="ghost" size="sm" asChild>
              <Link to="/logs">
                View all
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
          </CardHeader>
          <CardContent className="p-0">
            <ScrollArea className="h-[340px]">
              <div className="space-y-1 p-4 pt-0">
                {stats.recent_logs?.length > 0 ? (
                  stats.recent_logs.map((log: LogEntry) => (
                    <div
                      key={log.id}
                      className={`flex items-start gap-3 rounded-lg p-3 transition-colors hover:bg-muted/50 ${levelBgColors[log.level]}`}
                    >
                      <Badge variant={levelColors[log.level]} className="shrink-0 mt-0.5">
                        {log.level}
                      </Badge>
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium leading-tight line-clamp-2">
                          {log.message}
                        </p>
                        <div className="flex items-center gap-2 mt-1">
                          <span className="text-xs font-medium text-primary">
                            {log.project_name}
                          </span>
                          <span className="text-xs text-muted-foreground">•</span>
                          <span className="text-xs text-muted-foreground">
                            {new Date(log.created_at).toLocaleString()}
                          </span>
                        </div>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="flex h-[280px] flex-col items-center justify-center text-muted-foreground">
                    <ScrollText className="h-12 w-12 mb-4 opacity-50" />
                    <p className="text-sm font-medium">No recent logs</p>
                    <p className="text-xs">Logs will appear here once received</p>
                  </div>
                )}
              </div>
            </ScrollArea>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
          <CardDescription>Common tasks and navigation</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <Button variant="outline" className="h-auto py-4 justify-start" asChild>
              <Link to="/projects">
                <FolderKanban className="mr-3 h-5 w-5 text-blue-500" />
                <div className="text-left">
                  <p className="font-medium">Manage Projects</p>
                  <p className="text-xs text-muted-foreground">Create or configure projects</p>
                </div>
              </Link>
            </Button>
            <Button variant="outline" className="h-auto py-4 justify-start" asChild>
              <Link to="/logs">
                <ScrollText className="mr-3 h-5 w-5 text-emerald-500" />
                <div className="text-left">
                  <p className="font-medium">View All Logs</p>
                  <p className="text-xs text-muted-foreground">Search and filter logs</p>
                </div>
              </Link>
            </Button>
            <Button variant="outline" className="h-auto py-4 justify-start" asChild>
              <Link to="/settings">
                <Activity className="mr-3 h-5 w-5 text-violet-500" />
                <div className="text-left">
                  <p className="font-medium">Settings</p>
                  <p className="text-xs text-muted-foreground">Configure your account</p>
                </div>
              </Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

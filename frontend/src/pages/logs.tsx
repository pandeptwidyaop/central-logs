import { useEffect, useState, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Search, RefreshCw, Trash2, Radio } from 'lucide-react';
import { api, type LogEntry, type Project } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useToast } from '@/hooks/use-toast';
import { useRealtimeLogs } from '@/hooks/use-realtime-logs';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';

const levelColors: Record<string, 'debug' | 'info' | 'warn' | 'error' | 'critical'> = {
  DEBUG: 'debug',
  INFO: 'info',
  WARN: 'warn',
  ERROR: 'error',
  CRITICAL: 'critical',
};

export function LogsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState({
    project_id: '',
    level: '',
    search: '',
  });
  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null);
  const [selectedLogs, setSelectedLogs] = useState<string[]>([]);
  const [isLive, setIsLive] = useState(true);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const { toast } = useToast();

  // Handle new logs from WebSocket
  const handleNewLog = useCallback((newLog: LogEntry) => {
    // Only add if on first page and no filters that would exclude it
    if (page === 1) {
      const matchesProject = !filters.project_id || newLog.project_id === filters.project_id;
      const matchesLevel = !filters.level || newLog.level === filters.level;
      const matchesSearch = !filters.search ||
        newLog.message.toLowerCase().includes(filters.search.toLowerCase());

      if (matchesProject && matchesLevel && matchesSearch) {
        setLogs((prev) => {
          // Avoid duplicates
          if (prev.some((l) => l.id === newLog.id)) return prev;
          // Add to top and limit to 50
          return [newLog, ...prev].slice(0, 50);
        });
        setTotal((prev) => prev + 1);
      }
    }
  }, [page, filters]);

  // Real-time logs connection (toast & sound handled by AppLayout)
  const { isConnected } = useRealtimeLogs({
    projectId: filters.project_id || undefined,
    enabled: isLive,
    onNewLog: handleNewLog,
    playSound: false,  // Handled by AppLayout
    showToast: false,  // Handled by AppLayout
  });

  // Handle ?id= query parameter to show specific log
  useEffect(() => {
    const logId = searchParams.get('id');
    if (logId) {
      api.getLog(logId)
        .then((log) => {
          setSelectedLog(log);
        })
        .catch((err) => {
          toast({
            title: 'Failed to load log',
            description: err instanceof Error ? err.message : 'Log not found',
            variant: 'destructive',
          });
        });
    }
  }, [searchParams, toast]);

  // Clear query param when dialog is closed
  const handleCloseDialog = () => {
    setSelectedLog(null);
    if (searchParams.has('id')) {
      searchParams.delete('id');
      setSearchParams(searchParams);
    }
  };

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.getLogs({
        project_id: filters.project_id ? parseInt(filters.project_id) : undefined,
        level: filters.level || undefined,
        search: filters.search || undefined,
        page,
        limit: 50,
      });
      setLogs(result.logs ?? []);
      setTotal(result.total ?? 0);
    } finally {
      setLoading(false);
    }
  }, [filters, page]);

  useEffect(() => {
    api.getProjects().then(setProjects);
  }, []);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const openDeleteDialog = () => {
    if (selectedLogs.length === 0) return;
    setDeleteDialogOpen(true);
  };

  const handleDeleteSelected = async () => {
    setDeleteLoading(true);
    try {
      await api.deleteLogs(selectedLogs);
      toast({ title: `Deleted ${selectedLogs.length} logs` });
      setDeleteDialogOpen(false);
      setSelectedLogs([]);
      fetchLogs();
    } catch (err) {
      toast({
        title: 'Failed to delete logs',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setDeleteLoading(false);
    }
  };

  const toggleSelectLog = (id: string) => {
    setSelectedLogs((prev) =>
      prev.includes(id) ? prev.filter((i) => i !== id) : [...prev, id]
    );
  };

  const toggleSelectAll = () => {
    if (selectedLogs.length === logs.length) {
      setSelectedLogs([]);
    } else {
      setSelectedLogs(logs.map((l) => l.id));
    }
  };

  const totalPages = Math.ceil(total / 50);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Logs</h1>
          <p className="text-muted-foreground">View and search application logs</p>
        </div>
        <div className="flex gap-2">
          {selectedLogs.length > 0 && (
            <Button variant="destructive" onClick={openDeleteDialog}>
              <Trash2 className="mr-2 h-4 w-4" />
              Delete ({selectedLogs.length})
            </Button>
          )}
          <Button
            variant={isLive ? 'default' : 'outline'}
            onClick={() => setIsLive(!isLive)}
            className={isLive && isConnected ? 'bg-green-600 hover:bg-green-700' : ''}
          >
            <Radio className={`mr-2 h-4 w-4 ${isLive && isConnected ? 'animate-pulse' : ''}`} />
            {isLive ? (isConnected ? 'Live' : 'Connecting...') : 'Paused'}
          </Button>
          <Button variant="outline" onClick={fetchLogs}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-4">
            <div className="flex-1 min-w-[200px]">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="Search logs..."
                  className="pl-9"
                  value={filters.search}
                  onChange={(e) => {
                    setFilters((f) => ({ ...f, search: e.target.value }));
                    setPage(1);
                  }}
                />
              </div>
            </div>
            <Select
              value={filters.project_id || 'all'}
              onValueChange={(v) => {
                setFilters((f) => ({ ...f, project_id: v === 'all' ? '' : v }));
                setPage(1);
              }}
            >
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="All Projects" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Projects</SelectItem>
                {projects.map((p) => (
                  <SelectItem key={p.id} value={p.id.toString()}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={filters.level || 'all'}
              onValueChange={(v) => {
                setFilters((f) => ({ ...f, level: v === 'all' ? '' : v }));
                setPage(1);
              }}
            >
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="All Levels" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Levels</SelectItem>
                <SelectItem value="DEBUG">DEBUG</SelectItem>
                <SelectItem value="INFO">INFO</SelectItem>
                <SelectItem value="WARN">WARN</SelectItem>
                <SelectItem value="ERROR">ERROR</SelectItem>
                <SelectItem value="CRITICAL">CRITICAL</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          {loading ? (
            <div className="flex h-64 items-center justify-center">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
            </div>
          ) : logs.length === 0 ? (
            <div className="flex h-64 items-center justify-center text-muted-foreground">
              No logs found
            </div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-12">
                      <input
                        type="checkbox"
                        checked={selectedLogs.length === logs.length}
                        onChange={toggleSelectAll}
                        className="rounded"
                      />
                    </TableHead>
                    <TableHead className="w-24">Level</TableHead>
                    <TableHead className="w-32">Project</TableHead>
                    <TableHead>Message</TableHead>
                    <TableHead className="w-44">Time</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {logs.map((log) => (
                    <TableRow
                      key={log.id}
                      className="cursor-pointer"
                      onClick={() => setSelectedLog(log)}
                    >
                      <TableCell onClick={(e) => e.stopPropagation()}>
                        <input
                          type="checkbox"
                          checked={selectedLogs.includes(log.id)}
                          onChange={() => toggleSelectLog(log.id)}
                          className="rounded"
                        />
                      </TableCell>
                      <TableCell>
                        <Badge variant={levelColors[log.level]}>{log.level}</Badge>
                      </TableCell>
                      <TableCell className="font-medium">{log.project_name}</TableCell>
                      <TableCell className="max-w-md truncate">{log.message}</TableCell>
                      <TableCell className="text-muted-foreground">
                        {new Date(log.created_at).toLocaleString()}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              <div className="flex items-center justify-between border-t px-4 py-3">
                <p className="text-sm text-muted-foreground">
                  Showing {(page - 1) * 50 + 1} - {Math.min(page * 50, total)} of {total}
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page === 1}
                    onClick={() => setPage((p) => p - 1)}
                  >
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page >= totalPages}
                    onClick={() => setPage((p) => p + 1)}
                  >
                    Next
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Log Detail Dialog */}
      <Dialog open={!!selectedLog} onOpenChange={handleCloseDialog}>
        <DialogContent className="max-w-6xl max-h-[90vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Badge variant={selectedLog ? levelColors[selectedLog.level] : 'default'}>
                {selectedLog?.level}
              </Badge>
              Log Details
            </DialogTitle>
            <DialogDescription>
              {selectedLog && new Date(selectedLog.created_at).toLocaleString()}
            </DialogDescription>
          </DialogHeader>
          {selectedLog && (
            <ScrollArea className="flex-1 pr-4">
              <div className="space-y-4">
                <div>
                  <h4 className="mb-1 text-sm font-medium text-muted-foreground">Project</h4>
                  <p>{selectedLog.project_name}</p>
                </div>
                <div>
                  <h4 className="mb-1 text-sm font-medium text-muted-foreground">Message</h4>
                  <p className="whitespace-pre-wrap break-words">{selectedLog.message}</p>
                </div>
                {selectedLog.source && (
                  <div>
                    <h4 className="mb-1 text-sm font-medium text-muted-foreground">Source</h4>
                    <p className="break-all font-mono text-sm">{selectedLog.source}</p>
                  </div>
                )}
                {selectedLog.metadata && Object.keys(selectedLog.metadata).length > 0 && (
                  <div>
                    <h4 className="mb-1 text-sm font-medium text-muted-foreground">Metadata</h4>
                    <div className="rounded overflow-hidden">
                      <SyntaxHighlighter
                        language="json"
                        style={oneDark}
                        customStyle={{
                          margin: 0,
                          fontSize: '0.875rem',
                          maxHeight: '400px',
                          overflow: 'auto'
                        }}
                        wrapLines={true}
                        wrapLongLines={true}
                      >
                        {JSON.stringify(selectedLog.metadata, null, 2)}
                      </SyntaxHighlighter>
                    </div>
                  </div>
                )}
              </div>
            </ScrollArea>
          )}
        </DialogContent>
      </Dialog>

      {/* Delete Logs Confirm Dialog */}
      <ConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="Delete Logs"
        description={`Delete ${selectedLogs.length} selected log${selectedLogs.length > 1 ? 's' : ''}? This action cannot be undone.`}
        confirmText="Delete"
        variant="destructive"
        loading={deleteLoading}
        onConfirm={handleDeleteSelected}
      />
    </div>
  );
}

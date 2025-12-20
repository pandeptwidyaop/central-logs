import { useState, useEffect } from 'react';
import { api, type MCPToken, type MCPActivityLog } from '@/lib/api';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/hooks/use-toast';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { formatDistanceToNow } from 'date-fns';
import { CheckCircle2, XCircle, Clock } from 'lucide-react';

interface ViewMCPTokenDialogProps {
  token: MCPToken;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ViewMCPTokenDialog({
  token,
  open,
  onOpenChange,
}: ViewMCPTokenDialogProps) {
  const [activities, setActivities] = useState<MCPActivityLog[]>([]);
  const [totalActivities, setTotalActivities] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const limit = 10;
  const { toast } = useToast();

  useEffect(() => {
    if (open && token) {
      loadActivities();
    }
  }, [open, token, page]);

  const loadActivities = async () => {
    try {
      setLoading(true);
      const offset = (page - 1) * limit;
      const response = await api.getMCPTokenActivity(token.id, limit, offset);
      setActivities(response.activities);
      setTotalActivities(response.total);
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to load activity logs',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const getProjectsDisplay = (grantedProjects: string) => {
    if (grantedProjects === '*') {
      return <Badge variant="outline">All Projects</Badge>;
    }
    try {
      const projects = JSON.parse(grantedProjects) as string[];
      return (
        <div className="flex flex-wrap gap-1">
          {projects.slice(0, 3).map((projectId) => (
            <Badge key={projectId} variant="secondary" className="text-xs">
              {projectId.substring(0, 8)}...
            </Badge>
          ))}
          {projects.length > 3 && (
            <Badge variant="secondary" className="text-xs">
              +{projects.length - 3} more
            </Badge>
          )}
        </div>
      );
    } catch {
      return <Badge variant="destructive">Invalid</Badge>;
    }
  };

  const totalPages = Math.ceil(totalActivities / limit);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Token Details</DialogTitle>
          <DialogDescription>
            View token information and activity history
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="details" className="w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="details">Details</TabsTrigger>
            <TabsTrigger value="activity">Activity ({totalActivities})</TabsTrigger>
          </TabsList>

          <TabsContent value="details" className="space-y-4 mt-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label className="text-muted-foreground">Name</Label>
                <p className="font-medium">{token.name}</p>
              </div>

              <div className="space-y-2">
                <Label className="text-muted-foreground">Status</Label>
                <div>
                  <Badge variant={token.is_active ? 'default' : 'secondary'}>
                    {token.is_active ? 'Active' : 'Inactive'}
                  </Badge>
                </div>
              </div>

              <div className="space-y-2">
                <Label className="text-muted-foreground">Token Prefix</Label>
                <code className="text-sm bg-muted px-2 py-1 rounded block">
                  {token.token_prefix}
                </code>
              </div>

              <div className="space-y-2">
                <Label className="text-muted-foreground">Created</Label>
                <p className="text-sm">
                  {new Date(token.created_at).toLocaleString()}
                </p>
                <p className="text-xs text-muted-foreground">
                  {formatDistanceToNow(new Date(token.created_at), { addSuffix: true })}
                </p>
              </div>

              <div className="space-y-2">
                <Label className="text-muted-foreground">Granted Projects</Label>
                {getProjectsDisplay(token.granted_projects)}
              </div>

              <div className="space-y-2">
                <Label className="text-muted-foreground">Expires</Label>
                {token.expires_at ? (
                  <>
                    <p className="text-sm">
                      {new Date(token.expires_at).toLocaleString()}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {formatDistanceToNow(new Date(token.expires_at), { addSuffix: true })}
                    </p>
                  </>
                ) : (
                  <Badge variant="outline">Never</Badge>
                )}
              </div>

              <div className="space-y-2">
                <Label className="text-muted-foreground">Last Used</Label>
                {token.last_used_at ? (
                  <>
                    <p className="text-sm">
                      {new Date(token.last_used_at).toLocaleString()}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {formatDistanceToNow(new Date(token.last_used_at), { addSuffix: true })}
                    </p>
                  </>
                ) : (
                  <p className="text-sm text-muted-foreground">Never</p>
                )}
              </div>

              <div className="space-y-2">
                <Label className="text-muted-foreground">Created By</Label>
                <p className="text-sm">{token.created_by}</p>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="activity" className="mt-4">
            {loading ? (
              <div className="text-center py-8">
                <p className="text-muted-foreground">Loading activities...</p>
              </div>
            ) : activities.length === 0 ? (
              <div className="text-center py-8">
                <p className="text-muted-foreground">No activity recorded yet</p>
              </div>
            ) : (
              <>
                <div className="border rounded-lg">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Tool</TableHead>
                        <TableHead>Projects</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Duration</TableHead>
                        <TableHead>Time</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {activities.map((activity) => (
                        <TableRow key={activity.id}>
                          <TableCell>
                            <code className="text-xs">{activity.tool_name}</code>
                          </TableCell>
                          <TableCell>
                            {activity.project_ids && activity.project_ids.length > 0 ? (
                              <div className="flex gap-1">
                                {activity.project_ids.slice(0, 2).map((id) => (
                                  <Badge key={id} variant="secondary" className="text-xs">
                                    {id.substring(0, 8)}
                                  </Badge>
                                ))}
                                {activity.project_ids.length > 2 && (
                                  <Badge variant="secondary" className="text-xs">
                                    +{activity.project_ids.length - 2}
                                  </Badge>
                                )}
                              </div>
                            ) : (
                              <span className="text-muted-foreground text-xs">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            {activity.success ? (
                              <Badge variant="outline" className="gap-1">
                                <CheckCircle2 className="h-3 w-3" />
                                Success
                              </Badge>
                            ) : (
                              <Badge variant="destructive" className="gap-1">
                                <XCircle className="h-3 w-3" />
                                Error
                              </Badge>
                            )}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-1 text-xs text-muted-foreground">
                              <Clock className="h-3 w-3" />
                              {activity.duration_ms}ms
                            </div>
                          </TableCell>
                          <TableCell className="text-xs text-muted-foreground">
                            {formatDistanceToNow(new Date(activity.created_at), {
                              addSuffix: true,
                            })}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>

                {totalPages > 1 && (
                  <div className="flex items-center justify-between mt-4">
                    <p className="text-sm text-muted-foreground">
                      Page {page} of {totalPages}
                    </p>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setPage((p) => Math.max(1, p - 1))}
                        disabled={page === 1}
                      >
                        Previous
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                        disabled={page === totalPages}
                      >
                        Next
                      </Button>
                    </div>
                  </div>
                )}
              </>
            )}
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}

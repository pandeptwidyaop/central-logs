import { useState, useEffect } from 'react';
import { api, type MCPToken, type Project } from '@/lib/api';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Switch } from '@/components/ui/switch';
import { useToast } from '@/hooks/use-toast';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface EditMCPTokenDialogProps {
  token: MCPToken;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onTokenUpdated: () => void;
}

export function EditMCPTokenDialog({
  token,
  open,
  onOpenChange,
  onTokenUpdated,
}: EditMCPTokenDialogProps) {
  const [name, setName] = useState('');
  const [projectAccess, setProjectAccess] = useState<'all' | 'specific'>('all');
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);
  const [expiresAt, setExpiresAt] = useState<string>('');
  const [isActive, setIsActive] = useState(true);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    if (open && token) {
      // Load form with token data
      setName(token.name);
      setIsActive(token.is_active);
      setExpiresAt(token.expires_at || '');

      // Parse granted projects
      if (token.granted_projects === '*') {
        setProjectAccess('all');
        setSelectedProjects([]);
      } else {
        setProjectAccess('specific');
        try {
          const projects = JSON.parse(token.granted_projects) as string[];
          setSelectedProjects(projects);
        } catch {
          setSelectedProjects([]);
        }
      }

      loadProjects();
    }
  }, [open, token]);

  const loadProjects = async () => {
    try {
      const data = await api.getProjects();
      setProjects(data);
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to load projects',
        variant: 'destructive',
      });
    }
  };

  const handleUpdate = async () => {
    if (!name.trim()) {
      toast({
        title: 'Validation Error',
        description: 'Token name is required',
        variant: 'destructive',
      });
      return;
    }

    if (projectAccess === 'specific' && selectedProjects.length === 0) {
      toast({
        title: 'Validation Error',
        description: 'Please select at least one project',
        variant: 'destructive',
      });
      return;
    }

    try {
      setLoading(true);

      // Calculate expires_in_days from expiresAt
      let expiresInDays: number | null | undefined;
      if (expiresAt) {
        const expiryDate = new Date(expiresAt);
        const now = new Date();
        const diffDays = Math.ceil((expiryDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
        expiresInDays = diffDays > 0 ? diffDays : null;
      } else {
        expiresInDays = null;
      }

      await api.updateMCPToken(token.id, {
        name: name.trim(),
        granted_projects: projectAccess === 'all' ? ['*'] : selectedProjects,
        expires_in_days: expiresInDays,
        is_active: isActive,
      });

      toast({
        title: 'Success',
        description: 'Token updated successfully',
      });

      onTokenUpdated();
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to update token',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const getExpiresInValue = (): string => {
    if (!expiresAt) return 'never';

    const expiryDate = new Date(expiresAt);
    const now = new Date();
    const diffDays = Math.ceil((expiryDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));

    if (diffDays <= 0) return 'custom';
    if (diffDays <= 7) return '7';
    if (diffDays <= 30) return '30';
    if (diffDays <= 90) return '90';
    if (diffDays <= 180) return '180';
    if (diffDays <= 365) return '365';
    return 'custom';
  };

  const handleExpiresInChange = (value: string) => {
    if (value === 'never') {
      setExpiresAt('');
    } else if (value === 'custom') {
      // Keep current value
    } else {
      const now = new Date();
      const days = parseInt(value);
      now.setDate(now.getDate() + days);
      setExpiresAt(now.toISOString());
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Edit MCP Token</DialogTitle>
          <DialogDescription>
            Update token settings. The token value cannot be changed.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-token-name">Token Name *</Label>
            <Input
              id="edit-token-name"
              placeholder="e.g., Claude Desktop, Production Agent"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label>Project Access *</Label>
            <RadioGroup value={projectAccess} onValueChange={(v) => setProjectAccess(v as 'all' | 'specific')}>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="all" id="edit-access-all" />
                <Label htmlFor="edit-access-all" className="font-normal">
                  All projects (current and future)
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="specific" id="edit-access-specific" />
                <Label htmlFor="edit-access-specific" className="font-normal">
                  Specific projects only
                </Label>
              </div>
            </RadioGroup>
          </div>

          {projectAccess === 'specific' && (
            <div className="space-y-2">
              <Label>Select Projects</Label>
              <div className="border rounded-md p-3 max-h-48 overflow-y-auto space-y-2">
                {projects.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No projects available</p>
                ) : (
                  projects.map((project) => (
                    <div key={project.id} className="flex items-center space-x-2">
                      <input
                        type="checkbox"
                        id={`edit-project-${project.id}`}
                        checked={selectedProjects.includes(project.id)}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setSelectedProjects([...selectedProjects, project.id]);
                          } else {
                            setSelectedProjects(selectedProjects.filter((id) => id !== project.id));
                          }
                        }}
                        className="rounded border-gray-300"
                      />
                      <Label htmlFor={`edit-project-${project.id}`} className="font-normal">
                        {project.name}
                      </Label>
                    </div>
                  ))
                )}
              </div>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="edit-expires-in">Token Expiration</Label>
            <Select value={getExpiresInValue()} onValueChange={handleExpiresInChange}>
              <SelectTrigger id="edit-expires-in">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="never">Never expires</SelectItem>
                <SelectItem value="7">7 days from now</SelectItem>
                <SelectItem value="30">30 days from now</SelectItem>
                <SelectItem value="90">90 days from now</SelectItem>
                <SelectItem value="180">180 days from now</SelectItem>
                <SelectItem value="365">1 year from now</SelectItem>
                <SelectItem value="custom">Custom (keep current)</SelectItem>
              </SelectContent>
            </Select>
            {expiresAt && (
              <p className="text-sm text-muted-foreground">
                Current expiry: {new Date(expiresAt).toLocaleString()}
              </p>
            )}
          </div>

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="edit-is-active">Active Status</Label>
              <p className="text-sm text-muted-foreground">
                Enable or disable this token
              </p>
            </div>
            <Switch
              id="edit-is-active"
              checked={isActive}
              onCheckedChange={setIsActive}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleUpdate} disabled={loading}>
            {loading ? 'Updating...' : 'Update Token'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

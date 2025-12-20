import { useState, useEffect } from 'react';
import { api, type Project } from '@/lib/api';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Switch } from '@/components/ui/switch';
import { useToast } from '@/hooks/use-toast';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Copy, CheckCircle2, AlertCircle } from 'lucide-react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface CreateMCPTokenDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onTokenCreated: () => void;
}

export function CreateMCPTokenDialog({
  open,
  onOpenChange,
  onTokenCreated,
}: CreateMCPTokenDialogProps) {
  const [name, setName] = useState('');
  const [projectAccess, setProjectAccess] = useState<'all' | 'specific'>('all');
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);
  const [expiresIn, setExpiresIn] = useState<string>('never');
  const [isActive, setIsActive] = useState(true);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [showConfirmClose, setShowConfirmClose] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    if (open) {
      loadProjects();
      // Reset form
      setName('');
      setProjectAccess('all');
      setSelectedProjects([]);
      setExpiresIn('never');
      setIsActive(true);
      setCreatedToken(null);
      setCopied(false);
    }
  }, [open]);

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

  const handleCreate = async () => {
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

      const response = await api.createMCPToken({
        name: name.trim(),
        granted_projects: projectAccess === 'all' ? ['*'] : selectedProjects,
        expires_in_days: expiresIn === 'never' ? null : parseInt(expiresIn),
      });

      setCreatedToken(response.token);
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to create token',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    if (createdToken) {
      await navigator.clipboard.writeText(createdToken);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleClose = () => {
    if (createdToken && !copied) {
      setShowConfirmClose(true);
      return;
    }
    onTokenCreated();
    onOpenChange(false);
  };

  const confirmClose = () => {
    setShowConfirmClose(false);
    onTokenCreated();
    onOpenChange(false);
  };

  // Show token display after creation
  if (createdToken) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Token Created Successfully</DialogTitle>
            <DialogDescription>
              Copy your token now. For security reasons, it won't be shown again.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Make sure to copy this token now. You won't be able to see it again!
              </AlertDescription>
            </Alert>

            <div className="space-y-2">
              <Label>Token</Label>
              <div className="flex gap-2">
                <Input
                  value={createdToken}
                  readOnly
                  className="font-mono text-sm"
                />
                <Button
                  onClick={handleCopy}
                  variant={copied ? 'default' : 'outline'}
                  size="icon"
                >
                  {copied ? (
                    <CheckCircle2 className="h-4 w-4" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button onClick={handleClose} variant={copied ? 'default' : 'outline'}>
              {copied ? 'Done' : 'I have copied the token'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  // Show creation form
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Create MCP Token</DialogTitle>
          <DialogDescription>
            Create a new token for AI agents to access your logs via MCP protocol.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="token-name">Token Name *</Label>
            <Input
              id="token-name"
              placeholder="e.g., Claude Desktop, Production Agent"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label>Project Access *</Label>
            <RadioGroup value={projectAccess} onValueChange={(v) => setProjectAccess(v as 'all' | 'specific')}>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="all" id="access-all" />
                <Label htmlFor="access-all" className="font-normal">
                  All projects (current and future)
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="specific" id="access-specific" />
                <Label htmlFor="access-specific" className="font-normal">
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
                        id={`project-${project.id}`}
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
                      <Label htmlFor={`project-${project.id}`} className="font-normal">
                        {project.name}
                      </Label>
                    </div>
                  ))
                )}
              </div>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="expires-in">Token Expiration</Label>
            <Select value={expiresIn} onValueChange={setExpiresIn}>
              <SelectTrigger id="expires-in">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="never">Never expires</SelectItem>
                <SelectItem value="7">7 days</SelectItem>
                <SelectItem value="30">30 days</SelectItem>
                <SelectItem value="90">90 days</SelectItem>
                <SelectItem value="180">180 days</SelectItem>
                <SelectItem value="365">1 year</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="is-active">Active Status</Label>
              <p className="text-sm text-muted-foreground">
                Token can be used immediately
              </p>
            </div>
            <Switch
              id="is-active"
              checked={isActive}
              onCheckedChange={setIsActive}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleCreate} disabled={loading}>
            {loading ? 'Creating...' : 'Create Token'}
          </Button>
        </DialogFooter>
      </DialogContent>

      <AlertDialog open={showConfirmClose} onOpenChange={setShowConfirmClose}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Close without copying token?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to close? Make sure you have copied the token as it will not be shown again.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={confirmClose}>
              Close anyway
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Dialog>
  );
}

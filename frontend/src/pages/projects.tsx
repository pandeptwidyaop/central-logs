import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Plus, MoreVertical, Key, Trash2, Edit, Copy, Check } from 'lucide-react';
import { api, type Project, type ProjectIconType } from '@/lib/api';
import { Button } from '@/components/ui/button';
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useToast } from '@/hooks/use-toast';
import { ProjectIcon } from '@/components/project-icon';
import { ProjectIconPicker } from '@/components/project-icon-picker';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';

export function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [selectedProject, setSelectedProject] = useState<Project | null>(null);
  const [apiKeyDialogOpen, setApiKeyDialogOpen] = useState(false);
  const [newApiKey, setNewApiKey] = useState('');
  const [apiKeyProjectName, setApiKeyProjectName] = useState('');
  const [copiedKey, setCopiedKey] = useState(false);
  // Confirm dialogs
  const [rotateKeyDialogOpen, setRotateKeyDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [actionProject, setActionProject] = useState<Project | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const { toast } = useToast();

  // Form state for create
  const [createName, setCreateName] = useState('');
  const [createDescription, setCreateDescription] = useState('');
  const [createIconType, setCreateIconType] = useState<ProjectIconType>('initials');
  const [createIconValue, setCreateIconValue] = useState('');

  // Form state for edit
  const [editName, setEditName] = useState('');
  const [editDescription, setEditDescription] = useState('');
  const [editIconType, setEditIconType] = useState<ProjectIconType>('initials');
  const [editIconValue, setEditIconValue] = useState('');

  const fetchProjects = async () => {
    try {
      const data = await api.getProjects();
      setProjects(data);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProjects();
  }, []);

  const resetCreateForm = () => {
    setCreateName('');
    setCreateDescription('');
    setCreateIconType('initials');
    setCreateIconValue('');
  };

  const handleCreateProject = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      const result = await api.createProject({
        name: createName,
        description: createDescription,
        icon_type: createIconType,
        icon_value: createIconValue,
      });
      setCreateDialogOpen(false);
      // Show API key dialog
      setNewApiKey(result.api_key);
      setApiKeyProjectName(createName);
      setApiKeyDialogOpen(true);
      resetCreateForm();
      fetchProjects();
    } catch (err) {
      toast({
        title: 'Failed to create project',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    }
  };

  const openEditDialog = (project: Project) => {
    setSelectedProject(project);
    setEditName(project.name);
    setEditDescription(project.description || '');
    setEditIconType(project.icon_type || 'initials');
    setEditIconValue(project.icon_value || '');
    setEditDialogOpen(true);
  };

  const handleEditProject = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!selectedProject) return;
    try {
      await api.updateProject(selectedProject.id, {
        name: editName,
        description: editDescription,
        icon_type: editIconType,
        icon_value: editIconValue,
      });
      toast({ title: 'Project updated successfully' });
      setEditDialogOpen(false);
      setSelectedProject(null);
      fetchProjects();
    } catch (err) {
      toast({
        title: 'Failed to update project',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    }
  };

  const openRotateKeyDialog = (project: Project) => {
    setActionProject(project);
    setRotateKeyDialogOpen(true);
  };

  const handleRotateKey = async () => {
    if (!actionProject) return;
    setActionLoading(true);
    try {
      const result = await api.rotateApiKey(actionProject.id);
      setRotateKeyDialogOpen(false);
      // Show API key dialog
      setNewApiKey(result.api_key);
      setApiKeyProjectName(actionProject.name);
      setApiKeyDialogOpen(true);
      fetchProjects();
    } catch (err) {
      toast({
        title: 'Failed to rotate API key',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setActionLoading(false);
      setActionProject(null);
    }
  };

  const copyApiKey = () => {
    navigator.clipboard.writeText(newApiKey);
    setCopiedKey(true);
    setTimeout(() => setCopiedKey(false), 2000);
  };

  const openDeleteDialog = (project: Project) => {
    setActionProject(project);
    setDeleteDialogOpen(true);
  };

  const handleDeleteProject = async () => {
    if (!actionProject) return;
    setActionLoading(true);
    try {
      await api.deleteProject(actionProject.id);
      toast({ title: 'Project deleted successfully' });
      setDeleteDialogOpen(false);
      fetchProjects();
    } catch (err) {
      toast({
        title: 'Failed to delete project',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setActionLoading(false);
      setActionProject(null);
    }
  };

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Projects</h1>
          <p className="text-muted-foreground">Manage your logging projects</p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Project
        </Button>
      </div>

      {projects.length === 0 ? (
        <Card>
          <CardContent className="flex h-64 flex-col items-center justify-center">
            <p className="text-muted-foreground">No projects yet</p>
            <Button className="mt-4" onClick={() => setCreateDialogOpen(true)}>
              Create your first project
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => (
            <Card key={project.id}>
              <CardHeader className="flex flex-row items-start justify-between pb-2">
                <div className="flex items-start gap-3">
                  <ProjectIcon
                    name={project.name}
                    iconType={project.icon_type}
                    iconValue={project.icon_value}
                    size="lg"
                  />
                  <div>
                    <CardTitle className="text-lg">
                      <Link to={`/projects/${project.id}`} className="hover:underline">
                        {project.name}
                      </Link>
                    </CardTitle>
                    <CardDescription className="line-clamp-2">
                      {project.description || 'No description'}
                    </CardDescription>
                  </div>
                </div>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <MoreVertical className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => openEditDialog(project)}>
                      <Edit className="mr-2 h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => openRotateKeyDialog(project)}>
                      <Key className="mr-2 h-4 w-4" />
                      Rotate API Key
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      className="text-destructive"
                      onClick={() => openDeleteDialog(project)}
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </CardHeader>
              <CardContent>
                <p className="text-xs text-muted-foreground">
                  Created {new Date(project.created_at).toLocaleDateString()}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Create Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={(open) => {
        setCreateDialogOpen(open);
        if (!open) resetCreateForm();
      }}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Create Project</DialogTitle>
            <DialogDescription>Add a new project to start receiving logs</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreateProject}>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  placeholder="My Project"
                  value={createName}
                  onChange={(e) => setCreateName(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Input
                  id="description"
                  placeholder="Optional description"
                  value={createDescription}
                  onChange={(e) => setCreateDescription(e.target.value)}
                />
              </div>
              <ProjectIconPicker
                name={createName || 'Project'}
                iconType={createIconType}
                iconValue={createIconValue}
                onIconChange={(type, value) => {
                  setCreateIconType(type);
                  setCreateIconValue(value);
                }}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setCreateDialogOpen(false)}>
                Cancel
              </Button>
              <Button type="submit">Create</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-md">
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
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-description">Description</Label>
                <Input
                  id="edit-description"
                  value={editDescription}
                  onChange={(e) => setEditDescription(e.target.value)}
                />
              </div>
              <ProjectIconPicker
                name={editName || 'Project'}
                iconType={editIconType}
                iconValue={editIconValue}
                onIconChange={(type, value) => {
                  setEditIconType(type);
                  setEditIconValue(value);
                }}
              />
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

      {/* API Key Dialog */}
      <Dialog open={apiKeyDialogOpen} onOpenChange={(open) => {
        setApiKeyDialogOpen(open);
        if (!open) {
          setCopiedKey(false);
        }
      }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Key className="h-5 w-5" />
              API Key for {apiKeyProjectName}
            </DialogTitle>
            <DialogDescription>
              Copy this API key now. For security reasons, it won't be shown again.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded-lg bg-muted px-4 py-3 font-mono text-sm break-all">
                {newApiKey}
              </code>
              <Button variant="outline" size="icon" onClick={copyApiKey}>
                {copiedKey ? (
                  <Check className="h-4 w-4 text-green-500" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
            </div>
            <div className="rounded-lg bg-amber-500/10 border border-amber-500/20 p-3">
              <p className="text-sm text-amber-600 dark:text-amber-400">
                Make sure to copy your API key now. You won't be able to see it again!
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={() => setApiKeyDialogOpen(false)}>
              {copiedKey ? 'Done' : 'Close'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Rotate Key Confirm Dialog */}
      <ConfirmDialog
        open={rotateKeyDialogOpen}
        onOpenChange={setRotateKeyDialogOpen}
        title="Rotate API Key"
        description={`Rotate API key for "${actionProject?.name}"? The old key will stop working immediately.`}
        confirmText="Rotate Key"
        variant="default"
        loading={actionLoading}
        onConfirm={handleRotateKey}
      />

      {/* Delete Confirm Dialog */}
      <ConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="Delete Project"
        description={`Delete project "${actionProject?.name}"? This will delete all associated logs.`}
        confirmText="Delete"
        variant="destructive"
        loading={actionLoading}
        onConfirm={handleDeleteProject}
      />
    </div>
  );
}

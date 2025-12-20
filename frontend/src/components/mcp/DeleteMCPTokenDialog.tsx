import { useState } from 'react';
import { api, type MCPToken } from '@/lib/api';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useToast } from '@/hooks/use-toast';
import { AlertTriangle } from 'lucide-react';

interface DeleteMCPTokenDialogProps {
  token: MCPToken;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onTokenDeleted: () => void;
}

export function DeleteMCPTokenDialog({
  token,
  open,
  onOpenChange,
  onTokenDeleted,
}: DeleteMCPTokenDialogProps) {
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const handleDelete = async () => {
    try {
      setLoading(true);
      await api.deleteMCPToken(token.id);

      toast({
        title: 'Success',
        description: 'Token deleted successfully',
      });

      onTokenDeleted();
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to delete token',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Delete MCP Token</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete this token? This action cannot be undone.
          </DialogDescription>
        </DialogHeader>

        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            Any AI agents using this token will immediately lose access. Active
            connections will be terminated.
          </AlertDescription>
        </Alert>

        <div className="space-y-2">
          <div className="flex items-center justify-between py-2 px-3 bg-muted rounded">
            <span className="text-sm font-medium">Token Name:</span>
            <span className="text-sm">{token.name}</span>
          </div>
          <div className="flex items-center justify-between py-2 px-3 bg-muted rounded">
            <span className="text-sm font-medium">Token Prefix:</span>
            <code className="text-xs">{token.token_prefix}</code>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={handleDelete} disabled={loading}>
            {loading ? 'Deleting...' : 'Delete Token'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

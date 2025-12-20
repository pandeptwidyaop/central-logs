import { useState } from 'react';
import { type MCPToken } from '@/lib/api';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Eye, Pencil, Trash2 } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { EditMCPTokenDialog } from './EditMCPTokenDialog';
import { ViewMCPTokenDialog } from './ViewMCPTokenDialog';
import { DeleteMCPTokenDialog } from './DeleteMCPTokenDialog';

interface MCPTokenListProps {
  tokens: MCPToken[];
  onTokenDeleted: () => void;
  onTokenUpdated: () => void;
}

export function MCPTokenList({ tokens, onTokenDeleted, onTokenUpdated }: MCPTokenListProps) {
  const [viewToken, setViewToken] = useState<MCPToken | null>(null);
  const [editToken, setEditToken] = useState<MCPToken | null>(null);
  const [deleteToken, setDeleteToken] = useState<MCPToken | null>(null);

  const getProjectsDisplay = (grantedProjects: string) => {
    if (grantedProjects === '*') {
      return 'All Projects';
    }
    try {
      const projects = JSON.parse(grantedProjects) as string[];
      return `${projects.length} project${projects.length !== 1 ? 's' : ''}`;
    } catch {
      return 'Invalid';
    }
  };

  const getExpiryDisplay = (expiresAt?: string) => {
    if (!expiresAt) {
      return <Badge variant="outline">Never</Badge>;
    }
    const expiryDate = new Date(expiresAt);
    const isExpired = expiryDate < new Date();
    return (
      <Badge variant={isExpired ? 'destructive' : 'outline'}>
        {isExpired ? 'Expired' : formatDistanceToNow(expiryDate, { addSuffix: true })}
      </Badge>
    );
  };

  const getLastUsedDisplay = (lastUsedAt?: string) => {
    if (!lastUsedAt) {
      return <span className="text-muted-foreground">Never</span>;
    }
    return formatDistanceToNow(new Date(lastUsedAt), { addSuffix: true });
  };

  if (tokens.length === 0) {
    return (
      <div className="border rounded-lg p-8 text-center">
        <p className="text-muted-foreground">No MCP tokens created yet.</p>
        <p className="text-sm text-muted-foreground mt-1">
          Create a token to allow AI agents to connect.
        </p>
      </div>
    );
  }

  return (
    <>
      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Token</TableHead>
              <TableHead>Projects</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Expires</TableHead>
              <TableHead>Last Used</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {tokens.map((token) => (
              <TableRow key={token.id}>
                <TableCell className="font-medium">{token.name}</TableCell>
                <TableCell>
                  <code className="text-xs bg-muted px-2 py-1 rounded">
                    {token.token_prefix}
                  </code>
                </TableCell>
                <TableCell>{getProjectsDisplay(token.granted_projects)}</TableCell>
                <TableCell>
                  <Badge variant={token.is_active ? 'default' : 'secondary'}>
                    {token.is_active ? 'Active' : 'Inactive'}
                  </Badge>
                </TableCell>
                <TableCell>{getExpiryDisplay(token.expires_at)}</TableCell>
                <TableCell>{getLastUsedDisplay(token.last_used_at)}</TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setViewToken(token)}
                      title="View details"
                    >
                      <Eye className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setEditToken(token)}
                      title="Edit token"
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setDeleteToken(token)}
                      title="Delete token"
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {viewToken && (
        <ViewMCPTokenDialog
          token={viewToken}
          open={!!viewToken}
          onOpenChange={(open) => !open && setViewToken(null)}
        />
      )}

      {editToken && (
        <EditMCPTokenDialog
          token={editToken}
          open={!!editToken}
          onOpenChange={(open) => !open && setEditToken(null)}
          onTokenUpdated={() => {
            setEditToken(null);
            onTokenUpdated();
          }}
        />
      )}

      {deleteToken && (
        <DeleteMCPTokenDialog
          token={deleteToken}
          open={!!deleteToken}
          onOpenChange={(open) => !open && setDeleteToken(null)}
          onTokenDeleted={() => {
            setDeleteToken(null);
            onTokenDeleted();
          }}
        />
      )}
    </>
  );
}

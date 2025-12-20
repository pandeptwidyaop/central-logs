import { useState, useEffect } from 'react';
import { api, type MCPToken } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Plus, AlertCircle, CheckCircle2 } from 'lucide-react';
import { MCPTokenList } from './MCPTokenList';
import { CreateMCPTokenDialog } from './CreateMCPTokenDialog';
import { useToast } from '@/hooks/use-toast';

export function MCPServerSection() {
  const [enabled, setEnabled] = useState(false);
  const [tokens, setTokens] = useState<MCPToken[]>([]);
  const [loading, setLoading] = useState(true);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [statusResponse, tokensResponse] = await Promise.all([
        api.getMCPStatus(),
        api.getMCPTokens(),
      ]);
      setEnabled(statusResponse.enabled);
      setTokens(tokensResponse);
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to load MCP data',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleToggleEnabled = async (checked: boolean) => {
    try {
      const response = await api.toggleMCP(checked);
      setEnabled(response.enabled);
      toast({
        title: 'Success',
        description: `MCP server ${response.enabled ? 'enabled' : 'disabled'}`,
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to toggle MCP server',
        variant: 'destructive',
      });
      // Revert on error
      setEnabled(!checked);
    }
  };

  const handleTokenCreated = () => {
    setCreateDialogOpen(false);
    loadData();
  };

  const handleTokenDeleted = () => {
    loadData();
  };

  const handleTokenUpdated = () => {
    loadData();
  };

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>MCP Server</CardTitle>
          <CardDescription>Loading...</CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>MCP Server</CardTitle>
          <CardDescription>
            Model Context Protocol server allows AI agents to query logs and retrieve statistics
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Enable/Disable Switch */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="mcp-enabled" className="text-base">
                Enable MCP Server
              </Label>
              <p className="text-sm text-muted-foreground">
                Allow AI agents to connect via MCP protocol
              </p>
            </div>
            <Switch
              id="mcp-enabled"
              checked={enabled}
              onCheckedChange={handleToggleEnabled}
            />
          </div>

          {/* Status Alert */}
          {enabled ? (
            <Alert>
              <CheckCircle2 className="h-4 w-4" />
              <AlertDescription>
                MCP server is <strong>enabled</strong>. Endpoint available at{' '}
                <code className="px-1 py-0.5 bg-muted rounded text-sm">/api/mcp/message</code>
              </AlertDescription>
            </Alert>
          ) : (
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                MCP server is <strong>disabled</strong>. Enable it to allow AI agents to connect.
              </AlertDescription>
            </Alert>
          )}

          {/* MCP Tokens Section */}
          <div className="border-t pt-6">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-lg font-medium">MCP Tokens</h3>
                <p className="text-sm text-muted-foreground">
                  Manage authentication tokens for MCP clients
                </p>
              </div>
              <Button onClick={() => setCreateDialogOpen(true)} size="sm">
                <Plus className="mr-2 h-4 w-4" />
                Create Token
              </Button>
            </div>

            <MCPTokenList
              tokens={tokens}
              onTokenDeleted={handleTokenDeleted}
              onTokenUpdated={handleTokenUpdated}
            />
          </div>
        </CardContent>
      </Card>

      <CreateMCPTokenDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onTokenCreated={handleTokenCreated}
      />
    </div>
  );
}

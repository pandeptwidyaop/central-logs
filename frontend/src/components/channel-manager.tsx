import { useState, useEffect } from 'react';
import { Plus, Bell, MessageCircle, Hash, Trash2, Edit } from 'lucide-react';
import { api, type Channel } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { useToast } from '@/hooks/use-toast';
import { TelegramChatFinder } from './telegram-chat-finder';

interface ChannelManagerProps {
  projectId: string;
}

type ChannelType = 'PUSH' | 'TELEGRAM' | 'DISCORD';
type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'CRITICAL';

const LOG_LEVELS: LogLevel[] = ['DEBUG', 'INFO', 'WARN', 'ERROR', 'CRITICAL'];

export function ChannelManager({ projectId }: ChannelManagerProps) {
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(true);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editingChannel, setEditingChannel] = useState<Channel | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [channelToDelete, setChannelToDelete] = useState<string | null>(null);
  const { toast } = useToast();

  // Form state
  const [channelType, setChannelType] = useState<ChannelType>('TELEGRAM');
  const [channelName, setChannelName] = useState('');
  const [minLevel, setMinLevel] = useState<LogLevel>('ERROR');
  const [isActive, setIsActive] = useState(true);

  // Telegram config
  const [useCustomBot, setUseCustomBot] = useState(false);
  const [telegramBotToken, setTelegramBotToken] = useState('');
  const [telegramChatId, setTelegramChatId] = useState('');
  const [telegramChatName, setTelegramChatName] = useState('');

  // Discord config
  const [discordWebhookUrl, setDiscordWebhookUrl] = useState('');

  const [saving, setSaving] = useState(false);

  const fetchChannels = async () => {
    setLoading(true);
    try {
      const fetchedChannels = await api.getChannels(projectId);
      setChannels(fetchedChannels);
    } catch (error) {
      toast({
        title: 'Failed to load channels',
        description: error instanceof Error ? error.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchChannels();
  }, [projectId]);

  const resetForm = () => {
    setChannelType('TELEGRAM');
    setChannelName('');
    setMinLevel('ERROR');
    setIsActive(true);
    setUseCustomBot(false);
    setTelegramBotToken('');
    setTelegramChatId('');
    setTelegramChatName('');
    setDiscordWebhookUrl('');
    setEditingChannel(null);
  };

  const openCreateDialog = () => {
    resetForm();
    setCreateDialogOpen(true);
  };

  const openEditDialog = (channel: Channel) => {
    setEditingChannel(channel);
    setChannelType(channel.type);
    setChannelName(channel.name);
    setMinLevel(channel.min_level);
    setIsActive(channel.is_active);

    if (channel.type === 'TELEGRAM') {
      const config = channel.config as { bot_token?: string; chat_id?: string };
      const hasCustomBot = !!config.bot_token;
      setUseCustomBot(hasCustomBot);
      setTelegramBotToken(config.bot_token || '');
      setTelegramChatId(config.chat_id || '');
    } else if (channel.type === 'DISCORD') {
      const config = channel.config as { webhook_url?: string };
      setDiscordWebhookUrl(config.webhook_url || '');
    }

    setCreateDialogOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);

    try {
      let config: Record<string, unknown> = {};

      if (channelType === 'TELEGRAM') {
        if (!telegramChatId) {
          toast({
            title: 'Chat ID required',
            description: 'Please select a Telegram chat',
            variant: 'destructive',
          });
          setSaving(false);
          return;
        }

        if (useCustomBot && !telegramBotToken) {
          toast({
            title: 'Bot Token required',
            description: 'Please enter a bot token or use global bot',
            variant: 'destructive',
          });
          setSaving(false);
          return;
        }

        config = {
          chat_id: telegramChatId,
          ...(useCustomBot && telegramBotToken ? { bot_token: telegramBotToken } : {}),
        };
      } else if (channelType === 'DISCORD') {
        if (!discordWebhookUrl) {
          toast({
            title: 'Webhook URL required',
            description: 'Please enter Discord webhook URL',
            variant: 'destructive',
          });
          setSaving(false);
          return;
        }
        config = { webhook_url: discordWebhookUrl };
      } else if (channelType === 'PUSH') {
        config = {};
      }

      if (editingChannel) {
        // Update existing channel
        await api.updateChannel(editingChannel.id, {
          name: channelName,
          config,
          min_level: minLevel,
          is_active: isActive,
        });
        toast({ title: 'Channel updated successfully' });
      } else {
        // Create new channel
        await api.createChannel(projectId, {
          type: channelType,
          name: channelName,
          config,
          min_level: minLevel,
        });
        toast({ title: 'Channel created successfully' });
      }

      setCreateDialogOpen(false);
      resetForm();
      await fetchChannels();
    } catch (error) {
      toast({
        title: editingChannel ? 'Failed to update channel' : 'Failed to create channel',
        description: error instanceof Error ? error.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setSaving(false);
    }
  };

  const openDeleteDialog = (channelId: string) => {
    setChannelToDelete(channelId);
    setDeleteDialogOpen(true);
  };

  const confirmDelete = async () => {
    if (!channelToDelete) return;

    try {
      await api.deleteChannel(channelToDelete);
      toast({ title: 'Channel deleted successfully' });
      await fetchChannels();
      setDeleteDialogOpen(false);
      setChannelToDelete(null);
    } catch (error) {
      toast({
        title: 'Failed to delete channel',
        description: error instanceof Error ? error.message : 'Unknown error',
        variant: 'destructive',
      });
    }
  };

  const handleTelegramChatSelect = (chatId: string, chatName: string) => {
    setTelegramChatId(chatId);
    setTelegramChatName(chatName);
    if (!channelName) {
      setChannelName(`Telegram: ${chatName}`);
    }
  };

  const getChannelIcon = (type: ChannelType) => {
    switch (type) {
      case 'PUSH':
        return <Bell className="h-5 w-5" />;
      case 'TELEGRAM':
        return <MessageCircle className="h-5 w-5" />;
      case 'DISCORD':
        return <Hash className="h-5 w-5" />;
    }
  };

  const getChannelTypeColor = (type: ChannelType) => {
    switch (type) {
      case 'PUSH':
        return 'default';
      case 'TELEGRAM':
        return 'info';
      case 'DISCORD':
        return 'secondary';
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium">Notification Channels</h3>
          <p className="text-sm text-muted-foreground">
            Configure where to send log notifications
          </p>
        </div>
        <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button onClick={openCreateDialog}>
              <Plus className="mr-2 h-4 w-4" />
              Add Channel
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>
                {editingChannel ? 'Edit Channel' : 'Create Notification Channel'}
              </DialogTitle>
              <DialogDescription>
                Configure a channel to receive log notifications
              </DialogDescription>
            </DialogHeader>

            <form onSubmit={handleSubmit} className="space-y-4">
              {/* Channel Type */}
              {!editingChannel && (
                <div className="space-y-2">
                  <Label htmlFor="type">Channel Type</Label>
                  <Select value={channelType} onValueChange={(value) => setChannelType(value as ChannelType)}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="TELEGRAM">
                        <div className="flex items-center gap-2">
                          <MessageCircle className="h-4 w-4" />
                          Telegram
                        </div>
                      </SelectItem>
                      {/* TODO: Implement Discord and Push notifications
                      <SelectItem value="DISCORD">
                        <div className="flex items-center gap-2">
                          <Hash className="h-4 w-4" />
                          Discord
                        </div>
                      </SelectItem>
                      <SelectItem value="PUSH">
                        <div className="flex items-center gap-2">
                          <Bell className="h-4 w-4" />
                          Web Push
                        </div>
                      </SelectItem>
                      */}
                    </SelectContent>
                  </Select>
                </div>
              )}

              {/* Channel Name */}
              <div className="space-y-2">
                <Label htmlFor="name">Channel Name</Label>
                <Input
                  id="name"
                  value={channelName}
                  onChange={(e) => setChannelName(e.target.value)}
                  placeholder="e.g., Error Alerts"
                  required
                />
              </div>

              {/* Telegram Config */}
              {channelType === 'TELEGRAM' && (
                <div className="space-y-4 rounded-lg border p-4">
                  <h4 className="font-medium">Telegram Configuration</h4>

                  <div className="flex items-center space-x-2">
                    <Switch
                      id="use-custom-bot"
                      checked={useCustomBot}
                      onCheckedChange={setUseCustomBot}
                    />
                    <Label htmlFor="use-custom-bot" className="cursor-pointer">
                      Use custom bot token (otherwise use global bot from config)
                    </Label>
                  </div>

                  {useCustomBot && (
                    <div className="space-y-2">
                      <Label htmlFor="telegram-bot-token">
                        Bot Token *
                      </Label>
                      <Input
                        id="telegram-bot-token"
                        type="password"
                        value={telegramBotToken}
                        onChange={(e) => setTelegramBotToken(e.target.value)}
                        placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
                        className="font-mono"
                        required
                      />
                      <p className="text-xs text-muted-foreground">
                        Enter your custom Telegram bot token
                      </p>
                    </div>
                  )}

                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label htmlFor="telegram-chat-id">Chat ID *</Label>
                      <TelegramChatFinder
                        onSelect={handleTelegramChatSelect}
                        botToken={telegramBotToken}
                      />
                    </div>
                    <Input
                      id="telegram-chat-id"
                      value={telegramChatId}
                      onChange={(e) => setTelegramChatId(e.target.value)}
                      placeholder="-1001234567890 or 987654321"
                      required
                    />
                    {telegramChatName && (
                      <p className="text-sm text-green-600">
                        Selected: {telegramChatName}
                      </p>
                    )}
                    <p className="text-xs text-muted-foreground">
                      Use the "Find Chat ID" button above for easy setup
                    </p>
                  </div>
                </div>
              )}

              {/* Discord Config */}
              {channelType === 'DISCORD' && (
                <div className="space-y-4 rounded-lg border p-4">
                  <h4 className="font-medium">Discord Configuration</h4>

                  <div className="space-y-2">
                    <Label htmlFor="discord-webhook">Webhook URL *</Label>
                    <Input
                      id="discord-webhook"
                      type="url"
                      value={discordWebhookUrl}
                      onChange={(e) => setDiscordWebhookUrl(e.target.value)}
                      placeholder="https://discord.com/api/webhooks/..."
                      required
                    />
                    <p className="text-xs text-muted-foreground">
                      Get from: Server Settings → Integrations → Webhooks
                    </p>
                  </div>
                </div>
              )}

              {/* Push Config */}
              {channelType === 'PUSH' && (
                <div className="rounded-lg border p-4">
                  <p className="text-sm text-muted-foreground">
                    Web Push notifications are sent to subscribed browsers. No additional configuration needed.
                  </p>
                </div>
              )}

              {/* Min Level */}
              <div className="space-y-2">
                <Label htmlFor="min-level">Minimum Log Level</Label>
                <Select value={minLevel} onValueChange={(value) => setMinLevel(value as LogLevel)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {LOG_LEVELS.map((level) => (
                      <SelectItem key={level} value={level}>
                        {level}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  Only notify for logs at this level or higher
                </p>
              </div>

              {/* Active Toggle */}
              <div className="flex items-center justify-between rounded-lg border p-4">
                <div className="space-y-0.5">
                  <Label htmlFor="active">Active</Label>
                  <p className="text-sm text-muted-foreground">
                    Enable or disable this notification channel
                  </p>
                </div>
                <Switch
                  id="active"
                  checked={isActive}
                  onCheckedChange={setIsActive}
                />
              </div>

              <div className="flex justify-end gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setCreateDialogOpen(false);
                    resetForm();
                  }}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={saving}>
                  {saving ? 'Saving...' : editingChannel ? 'Update' : 'Create'}
                </Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {loading ? (
        <Card>
          <CardContent className="flex h-32 items-center justify-center">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          </CardContent>
        </Card>
      ) : channels.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <Bell className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground mb-2">No notification channels configured</p>
            <p className="text-sm text-muted-foreground mb-4">
              Add a channel to start receiving log notifications
            </p>
            <Button onClick={openCreateDialog}>
              <Plus className="mr-2 h-4 w-4" />
              Add First Channel
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {channels.map((channel) => (
            <Card key={channel.id}>
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    {getChannelIcon(channel.type)}
                    <div>
                      <CardTitle className="text-base">{channel.name}</CardTitle>
                      <CardDescription className="mt-1">
                        <Badge variant={getChannelTypeColor(channel.type)} className="mr-2">
                          {channel.type}
                        </Badge>
                        <Badge variant="outline">
                          {channel.min_level}+
                        </Badge>
                      </CardDescription>
                    </div>
                  </div>
                  <Badge variant={channel.is_active ? 'default' : 'secondary'}>
                    {channel.is_active ? 'Active' : 'Inactive'}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                {channel.type === 'TELEGRAM' && (
                  <div className="text-sm space-y-1">
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <span className="font-medium">Chat ID:</span>
                      <code className="text-xs bg-muted px-2 py-0.5 rounded">
                        {(channel.config as { chat_id?: string }).chat_id}
                      </code>
                    </div>
                  </div>
                )}

                {channel.type === 'DISCORD' && (
                  <div className="text-sm text-muted-foreground">
                    Webhook configured
                  </div>
                )}

                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => openEditDialog(channel)}
                    className="flex-1"
                  >
                    <Edit className="mr-2 h-3 w-3" />
                    Edit
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => openDeleteDialog(channel.id)}
                    className="text-destructive hover:text-destructive"
                  >
                    <Trash2 className="h-3 w-3" />
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Channel</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this channel? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-3 mt-4">
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDelete}
            >
              Delete
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

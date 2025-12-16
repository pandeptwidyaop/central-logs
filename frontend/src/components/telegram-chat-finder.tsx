import { useState } from 'react';
import { Search, Loader2, AlertCircle, CheckCircle2, Users, User, Hash, MessageSquare } from 'lucide-react';
import { api, type TelegramChatInfo } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';

interface TelegramChatFinderProps {
  onSelect: (chatId: string, chatName: string) => void;
  botToken?: string;
}

export function TelegramChatFinder({ onSelect, botToken: initialBotToken }: TelegramChatFinderProps) {
  const [open, setOpen] = useState(false);
  const [botToken, setBotToken] = useState(initialBotToken || '');
  const [useGlobalBot, setUseGlobalBot] = useState(!initialBotToken);
  const [loading, setLoading] = useState(false);
  const [testing, setTesting] = useState(false);
  const [chats, setChats] = useState<TelegramChatInfo[]>([]);
  const [error, setError] = useState('');
  const [botValid, setBotValid] = useState(false);
  const [botName, setBotName] = useState('');

  const handleTestToken = async () => {
    const tokenToUse = useGlobalBot ? '' : botToken;

    if (!useGlobalBot && !botToken) {
      setError('Please enter a bot token');
      return;
    }

    setTesting(true);
    setError('');
    setBotValid(false);

    try {
      const result = await api.testTelegramBot(tokenToUse);
      if (result.valid) {
        setBotValid(true);
        setBotName(result.bot_name || 'Bot');
        setError('');
      } else {
        setError(result.error || 'Invalid bot token');
        setBotValid(false);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to test bot token');
      setBotValid(false);
    } finally {
      setTesting(false);
    }
  };

  const handleFetchChats = async () => {
    const tokenToUse = useGlobalBot ? '' : botToken;

    if (!useGlobalBot && !botToken) {
      setError('Please enter a bot token');
      return;
    }

    setLoading(true);
    setError('');
    setChats([]);

    try {
      const fetchedChats = await api.getTelegramChats(tokenToUse);

      if (fetchedChats.length === 0) {
        setError('No chats found. Send a message to your bot first!');
      } else {
        setChats(fetchedChats);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch chats');
    } finally {
      setLoading(false);
    }
  };

  const handleSelectChat = (chat: TelegramChatInfo) => {
    onSelect(chat.chat_id, chat.name);
    setOpen(false);
    // Reset state when closing
    setBotToken(initialBotToken || '');
    setChats([]);
    setError('');
    setBotValid(false);
  };

  const getChatIcon = (type: string) => {
    switch (type) {
      case 'private':
        return <User className="h-4 w-4" />;
      case 'group':
      case 'supergroup':
        return <Users className="h-4 w-4" />;
      case 'channel':
        return <Hash className="h-4 w-4" />;
      default:
        return <MessageSquare className="h-4 w-4" />;
    }
  };

  const getChatTypeBadge = (type: string) => {
    const variants: Record<string, 'default' | 'secondary' | 'outline'> = {
      private: 'default',
      group: 'secondary',
      supergroup: 'secondary',
      channel: 'outline',
    };
    return variants[type] || 'outline';
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" type="button">
          <Search className="mr-2 h-4 w-4" />
          Find Chat ID
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Telegram Chat ID Finder</DialogTitle>
          <DialogDescription>
            Find your Telegram chat ID by connecting your bot and selecting a chat
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Step 1: Bot Token */}
          <div className="space-y-2">
            <Label htmlFor="bot-token" className="flex items-center gap-2">
              Step 1: Enter Bot Token
              {botValid && <CheckCircle2 className="h-4 w-4 text-green-500" />}
            </Label>

            <div className="flex items-center space-x-2 mb-2">
              <input
                type="checkbox"
                id="use-global-bot-finder"
                checked={useGlobalBot}
                onChange={(e) => {
                  setUseGlobalBot(e.target.checked);
                  setBotValid(false);
                  setChats([]);
                }}
                className="h-4 w-4 rounded border-gray-300"
              />
              <Label htmlFor="use-global-bot-finder" className="cursor-pointer font-normal">
                Use global bot token from config (no need to enter token)
              </Label>
            </div>

            {!useGlobalBot && (
              <>
                <div className="flex gap-2">
                  <Input
                    id="bot-token"
                    type="password"
                    placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
                    value={botToken}
                    onChange={(e) => {
                      setBotToken(e.target.value);
                      setBotValid(false);
                      setChats([]);
                    }}
                    className="font-mono"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleTestToken}
                    disabled={!botToken || testing}
                  >
                    {testing ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Testing
                      </>
                    ) : (
                      'Test'
                    )}
                  </Button>
                </div>
                <p className="text-xs text-muted-foreground">
                  Get your bot token from{' '}
                  <a
                    href="https://t.me/BotFather"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline"
                  >
                    @BotFather
                  </a>
                </p>
              </>
            )}

            {useGlobalBot && (
              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm text-muted-foreground">
                  Using global bot token configured in config.yaml
                </p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleTestToken}
                  disabled={testing}
                  className="mt-2"
                >
                  {testing ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Testing
                    </>
                  ) : (
                    'Test Global Bot'
                  )}
                </Button>
              </div>
            )}

            {botValid && (
              <p className="text-sm text-green-600">
                ✓ Connected to bot: <span className="font-medium">@{botName}</span>
              </p>
            )}
          </div>

          {/* Step 2: Fetch Chats */}
          <div className="space-y-2">
            <Label>Step 2: Fetch Recent Chats</Label>
            <Button
              type="button"
              onClick={handleFetchChats}
              disabled={loading || !botValid || (!useGlobalBot && !botToken)}
              className="w-full"
            >
              {loading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Fetching Chats...
                </>
              ) : (
                <>
                  <Search className="mr-2 h-4 w-4" />
                  Fetch Chats
                </>
              )}
            </Button>
            <p className="text-xs text-muted-foreground">
              Make sure you've sent a message to your bot or added it to a group first
            </p>
          </div>

          {/* Error Display */}
          {error && (
            <div className="flex items-start gap-2 rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
              <AlertCircle className="h-4 w-4 shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {/* Step 3: Select Chat */}
          {chats.length > 0 && (
            <div className="space-y-2">
              <Label>Step 3: Select a Chat</Label>
              <ScrollArea className="h-64 rounded-md border">
                <div className="p-4 space-y-2">
                  {chats.map((chat) => (
                    <button
                      key={chat.chat_id}
                      type="button"
                      onClick={() => handleSelectChat(chat)}
                      className="w-full text-left p-3 rounded-lg border hover:bg-accent transition-colors"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex items-start gap-3 flex-1 min-w-0">
                          <div className="mt-1">{getChatIcon(chat.type)}</div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 mb-1">
                              <p className="font-medium truncate">{chat.name}</p>
                              <Badge variant={getChatTypeBadge(chat.type)} className="shrink-0">
                                {chat.type}
                              </Badge>
                            </div>
                            {chat.username && (
                              <p className="text-xs text-muted-foreground mb-1">@{chat.username}</p>
                            )}
                            <p className="text-xs text-muted-foreground truncate">
                              Last: {chat.last_message}
                            </p>
                          </div>
                        </div>
                        <div className="text-right shrink-0">
                          <code className="text-xs bg-muted px-2 py-1 rounded">{chat.chat_id}</code>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              </ScrollArea>
              <p className="text-xs text-muted-foreground">
                Found {chats.length} chat{chats.length !== 1 ? 's' : ''}. Click to select.
              </p>
            </div>
          )}

          {/* Help Text */}
          {!loading && chats.length === 0 && !error && botToken && (
            <div className="rounded-lg bg-muted p-4 text-sm space-y-2">
              <p className="font-medium">How to get started:</p>
              <ol className="list-decimal list-inside space-y-1 text-muted-foreground">
                <li>Create a bot with @BotFather and copy the token</li>
                <li>Send a message to your bot or add it to a group</li>
                <li>Enter the bot token above and click "Test"</li>
                <li>Click "Fetch Chats" to see available chats</li>
              </ol>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

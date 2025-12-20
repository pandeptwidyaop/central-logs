import { useEffect, useRef, useCallback, useState } from 'react';
import { useAuth } from '@/contexts/auth-context';
import { useToast } from '@/hooks/use-toast';
import type { LogEntry } from '@/lib/api';

const NOTIFICATION_SOUND_URL = '/sounds/slack.mp3';

interface UseRealtimeLogsOptions {
  projectId?: string;
  enabled?: boolean;
  onNewLog?: (log: LogEntry) => void;
  playSound?: boolean;
  showToast?: boolean;
  toastLevels?: string[]; // Only show toast for these levels
}

interface WebSocketMessage {
  type: string;
  data: LogEntry;
  project_id: string;
}

export function useRealtimeLogs(options: UseRealtimeLogsOptions = {}) {
  const {
    projectId,
    enabled = true,
    onNewLog,
    playSound = true,
    showToast = true,
    toastLevels = ['ERROR', 'CRITICAL'],
  } = options;

  const { user, token } = useAuth();
  const { toast } = useToast();
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);

  // Initialize audio
  useEffect(() => {
    const audio = new Audio(NOTIFICATION_SOUND_URL);
    audio.volume = 0.5;
    audio.preload = 'auto';

    // Try to load the audio
    audio.load();

    audio.addEventListener('canplaythrough', () => {
      // Sound loaded and ready
    });

    audio.addEventListener('error', () => {
      // Audio load failed - notifications will be silent
    });

    audioRef.current = audio;

    return () => {
      audioRef.current = null;
    };
  }, []);

  // Store callbacks and values in refs to avoid recreating WebSocket on changes
  const onNewLogRef = useRef(onNewLog);
  const playSoundRef = useRef(playSound);
  const showToastRef = useRef(showToast);
  const toastLevelsRef = useRef(toastLevels);
  const toastRef = useRef(toast);
  const audioRefCurrent = audioRef;

  useEffect(() => {
    onNewLogRef.current = onNewLog;
    playSoundRef.current = playSound;
    showToastRef.current = showToast;
    toastLevelsRef.current = toastLevels;
    toastRef.current = toast;
  }, [onNewLog, playSound, showToast, toastLevels, toast]);

  // Connect/disconnect based on enabled state
  useEffect(() => {
    if (!enabled || !user || !token) return;

    // Build WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    let wsUrl = `${protocol}//${host}/ws/logs`;

    // Add project filter as query param (optional)
    if (projectId) {
      wsUrl += `?project_id=${projectId}`;
    }

    // In development, use the backend port
    if (import.meta.env.DEV) {
      wsUrl = `ws://localhost:3000/ws/logs`;
      if (projectId) {
        wsUrl += `?project_id=${projectId}`;
      }
    }

    // Use Sec-WebSocket-Protocol to pass the JWT token
    // This is a workaround since WebSocket API doesn't support custom headers
    // Format: "token, <jwt-token-value>"
    const ws = new WebSocket(wsUrl, ['token', token]);

    ws.onopen = () => {
      setIsConnected(true);
      setConnectionError(null);
    };

    ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);

        if (message.type === 'log' && message.data) {
          // Call the callback
          if (onNewLogRef.current) {
            onNewLogRef.current(message.data);
          }

          // Play sound and show toast if page is visible
          if (document.visibilityState === 'visible') {
            const log = message.data;
            const levels = toastLevelsRef.current;

            // Play sound for specified levels
            if (playSoundRef.current && levels.includes(log.level)) {
              const audio = audioRefCurrent.current;
              if (audio) {
                audio.currentTime = 0;
                audio.play().catch(() => {
                  // Sound play failed - user may not have interacted with page yet
                });
              }
            }

            // Show toast for specified levels
            if (showToastRef.current && levels.includes(log.level)) {
              toastRef.current({
                title: `${log.level} - ${log.project_name || 'Unknown Project'}`,
                description: log.message.length > 100 ? log.message.substring(0, 100) + '...' : log.message,
                variant: log.level === 'ERROR' || log.level === 'CRITICAL' ? 'destructive' : 'default',
              });
            }
          }
        } else if (message.type === 'pong') {
          // Pong received - connection alive
        }
      } catch {
        // Error parsing WebSocket message
      }
    };

    ws.onerror = () => {
      setConnectionError('WebSocket connection error');
    };

    ws.onclose = (event) => {
      setIsConnected(false);

      // Reconnect after 5 seconds if not intentionally closed
      if (event.code !== 1000) {
        reconnectTimeoutRef.current = setTimeout(() => {
          // Will reconnect on next effect run due to state change
        }, 5000);
      }
    };

    wsRef.current = ws;

    // Send ping every 30 seconds to keep connection alive
    const pingInterval = setInterval(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send('ping');
      }
    }, 30000);

    return () => {
      clearInterval(pingInterval);
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      ws.close(1000, 'Component unmounted');
      wsRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- audioRef is stable, user?.id and token are sufficient
  }, [enabled, user?.id, token, projectId]);

  // Disconnect function for manual control
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close(1000, 'Manual disconnect');
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  return {
    isConnected,
    connectionError,
    disconnect,
  };
}

import { useEffect, useState } from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '@/contexts/auth-context';
import { Sidebar } from './sidebar';
import { Toaster } from '@/components/ui/toaster';
import { useRealtimeLogs } from '@/hooks/use-realtime-logs';
import { api, type Channel } from '@/lib/api';

// Log levels in order of priority (lowest to highest)
const LOG_LEVELS = ['DEBUG', 'INFO', 'WARN', 'ERROR', 'CRITICAL'];

// Get all levels >= minLevel
function getLevelsFromMinLevel(minLevel: string): string[] {
  const minIndex = LOG_LEVELS.indexOf(minLevel);
  if (minIndex === -1) return ['ERROR', 'CRITICAL'];
  return LOG_LEVELS.slice(minIndex);
}

export function AppLayout() {
  const { user, loading } = useAuth();
  const [toastLevels, setToastLevels] = useState<string[]>(['ERROR', 'CRITICAL']);

  // Fetch push channel config from all projects
  useEffect(() => {
    if (!user) return;

    async function fetchGlobalPushConfig() {
      try {
        const allProjects = await api.getProjects();
        let lowestMinLevel = 'CRITICAL';

        for (const project of allProjects) {
          try {
            const channels = await api.getChannels(project.id);
            const pushChannel = channels.find((ch: Channel) => ch.type === 'PUSH' && ch.is_active);
            if (pushChannel) {
              const currentIndex = LOG_LEVELS.indexOf(pushChannel.min_level);
              const lowestIndex = LOG_LEVELS.indexOf(lowestMinLevel);
              if (currentIndex !== -1 && currentIndex < lowestIndex) {
                lowestMinLevel = pushChannel.min_level;
              }
            }
          } catch {
            // Skip project if channel fetch fails
          }
        }

        const levels = getLevelsFromMinLevel(lowestMinLevel);
        setToastLevels(levels);
        console.log('[AppLayout] Global toast levels:', lowestMinLevel, '->', levels);
      } catch (err) {
        console.error('[AppLayout] Failed to fetch push config:', err);
      }
    }

    fetchGlobalPushConfig();
  }, [user]);

  // Global realtime logs - toast & sound on all pages
  useRealtimeLogs({
    enabled: !!user,
    playSound: true,
    showToast: true,
    toastLevels,
  });

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="flex h-screen">
      <Sidebar />
      <main className="flex-1 overflow-auto bg-background">
        <div className="container mx-auto p-6">
          <Outlet />
        </div>
      </main>
      <Toaster />
    </div>
  );
}

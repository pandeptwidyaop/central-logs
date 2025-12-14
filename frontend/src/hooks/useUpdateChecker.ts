import { useEffect, useState, useCallback } from 'react';
import { api, type UpdateCheckInfo } from '@/lib/api';

const UPDATE_CHECK_KEY = 'update-check-dismissed';
const UPDATE_CHECK_TIMESTAMP = 'update-check-timestamp';
const CHECK_INTERVAL = 24 * 60 * 60 * 1000; // 24 hours in milliseconds

interface UseUpdateCheckerResult {
  updateInfo: UpdateCheckInfo | null;
  isChecking: boolean;
  dismissUpdate: () => void;
}

export function useUpdateChecker(): UseUpdateCheckerResult {
  const [updateInfo, setUpdateInfo] = useState<UpdateCheckInfo | null>(null);
  const [isChecking, setIsChecking] = useState(false);

  const checkForUpdates = useCallback(async () => {
    try {
      // Check if we should skip the check
      if (shouldSkipCheck()) {
        return;
      }

      setIsChecking(true);
      const info = await api.checkForUpdates();

      // Store the check timestamp
      localStorage.setItem(UPDATE_CHECK_TIMESTAMP, Date.now().toString());

      // Only show if update is available and not dismissed
      if (info.update_available && !isUpdateDismissed(info.latest_version)) {
        setUpdateInfo(info);
      }
    } catch (error) {
      // Silently fail - don't bother user with update check errors
      console.debug('Failed to check for updates:', error);
    } finally {
      setIsChecking(false);
    }
  }, []);

  useEffect(() => {
    checkForUpdates();
  }, [checkForUpdates]);

  const shouldSkipCheck = (): boolean => {
    const lastCheck = localStorage.getItem(UPDATE_CHECK_TIMESTAMP);
    if (!lastCheck) {
      return false;
    }

    const timeSinceLastCheck = Date.now() - parseInt(lastCheck, 10);
    return timeSinceLastCheck < CHECK_INTERVAL;
  };

  const isUpdateDismissed = (version: string): boolean => {
    const dismissedVersion = localStorage.getItem(UPDATE_CHECK_KEY);
    return dismissedVersion === version;
  };

  const dismissUpdate = () => {
    if (updateInfo?.latest_version) {
      localStorage.setItem(UPDATE_CHECK_KEY, updateInfo.latest_version);
    }
    setUpdateInfo(null);
  };

  return {
    updateInfo,
    isChecking,
    dismissUpdate,
  };
}

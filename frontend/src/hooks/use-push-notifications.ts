import { useState, useEffect, useCallback } from 'react';
import { api } from '@/lib/api';

// VAPID public key from config - should match backend
const VAPID_PUBLIC_KEY = 'BCPvbPUl5BxOmMbJGV7Yorn_JB-EYfZM6JSn67Aum2-lGeRjtD5m2n0znCuf22K6MUCwZeGx_M29y5N07qq0Rdk';

function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}

export type PushPermissionState = 'prompt' | 'granted' | 'denied' | 'unsupported';

export interface UsePushNotificationsReturn {
  permission: PushPermissionState;
  isSubscribed: boolean;
  isLoading: boolean;
  error: string | null;
  subscribe: () => Promise<boolean>;
  unsubscribe: () => Promise<boolean>;
  isSupported: boolean;
}

export function usePushNotifications(): UsePushNotificationsReturn {
  const [permission, setPermission] = useState<PushPermissionState>('prompt');
  const [isSubscribed, setIsSubscribed] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [registration, setRegistration] = useState<ServiceWorkerRegistration | null>(null);

  const isSupported = typeof window !== 'undefined' && 'serviceWorker' in navigator && 'PushManager' in window;

  // Initialize - register service worker and check subscription status
  useEffect(() => {
    if (!isSupported) {
      setPermission('unsupported');
      setIsLoading(false);
      return;
    }

    const init = async () => {
      try {
        // Register service worker
        const reg = await navigator.serviceWorker.register('/sw.js', { scope: '/' });
        setRegistration(reg);

        // Wait for service worker to be ready
        await navigator.serviceWorker.ready;

        // Check current permission
        const currentPermission = Notification.permission;
        setPermission(currentPermission as PushPermissionState);

        // Check if already subscribed
        const subscription = await reg.pushManager.getSubscription();
        setIsSubscribed(!!subscription);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to initialize push notifications');
      } finally {
        setIsLoading(false);
      }
    };

    init();
  }, [isSupported]);

  // Subscribe to push notifications
  const subscribe = useCallback(async (): Promise<boolean> => {
    if (!isSupported || !registration) {
      setError('Push notifications not supported');
      return false;
    }

    setIsLoading(true);
    setError(null);

    try {
      // Request notification permission
      const permission = await Notification.requestPermission();
      setPermission(permission as PushPermissionState);

      if (permission !== 'granted') {
        setError('Notification permission denied');
        return false;
      }

      // Subscribe to push manager
      const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY) as BufferSource,
      });

      // Send subscription to backend
      const subscriptionJSON = subscription.toJSON();
      await api.subscribePush({
        endpoint: subscriptionJSON.endpoint!,
        keys: subscriptionJSON.keys as { p256dh: string; auth: string },
      });

      setIsSubscribed(true);
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to subscribe');
      return false;
    } finally {
      setIsLoading(false);
    }
  }, [isSupported, registration]);

  // Unsubscribe from push notifications
  const unsubscribe = useCallback(async (): Promise<boolean> => {
    if (!registration) {
      return false;
    }

    setIsLoading(true);
    setError(null);

    try {
      const subscription = await registration.pushManager.getSubscription();

      if (subscription) {
        // Unsubscribe from push manager
        await subscription.unsubscribe();

        // Remove from backend
        await api.unsubscribePush(subscription.endpoint);
      }

      setIsSubscribed(false);
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to unsubscribe');
      return false;
    } finally {
      setIsLoading(false);
    }
  }, [registration]);

  return {
    permission,
    isSubscribed,
    isLoading,
    error,
    subscribe,
    unsubscribe,
    isSupported,
  };
}

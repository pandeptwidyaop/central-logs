// Central Logs Service Worker
// Version: 2.0.0 - Updated visibility check for push notifications
const SW_VERSION = '2.0.0';
const CACHE_NAME = 'central-logs-v2';

// Install event
self.addEventListener('install', (event) => {
  console.log('[SW] Installing service worker version:', SW_VERSION);
  self.skipWaiting();
});

// Activate event
self.addEventListener('activate', (event) => {
  console.log('[SW] Activating service worker...');
  event.waitUntil(clients.claim());
});

// Push notification event
self.addEventListener('push', (event) => {
  console.log('[SW] Push received! SW Version:', SW_VERSION);
  console.log('[SW] Push event data:', event.data ? event.data.text() : 'no data');

  let data = {
    title: 'Central Logs',
    body: 'New log received',
    icon: '/icons/image.png',
    badge: '/icons/image.png',
    tag: 'log-notification',
    data: {},
  };

  if (event.data) {
    try {
      const payload = event.data.json();
      console.log('[SW] Parsed payload:', JSON.stringify(payload));
      data = {
        title: payload.title || data.title,
        body: payload.body || payload.message || data.body,
        icon: payload.icon || data.icon,
        badge: payload.badge || data.badge,
        tag: payload.tag || `log-${payload.level || 'info'}-${Date.now()}`,
        data: {
          url: payload.url || '/',
          logId: payload.log_id,
          projectId: payload.project_id,
          level: payload.level,
        },
      };
      console.log('[SW] Final notification data:', JSON.stringify(data));
    } catch (e) {
      console.error('[SW] Error parsing push data:', e);
      console.log('[SW] Raw text data:', event.data.text());
      data.body = event.data ? event.data.text() : 'New notification';
    }
  }

  console.log('[SW] Showing notification with title:', data.title, 'body:', data.body);

  const options = {
    body: data.body,
    icon: data.icon,
    badge: data.badge,
    tag: data.tag || 'default-tag',
    data: data.data,
    vibrate: [100, 50, 100],
    requireInteraction: false,
  };

  // Check if any client (tab/window) is visible - if so, let WebSocket toast handle it
  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true })
      .then((windowClients) => {
        console.log('[SW] Found', windowClients.length, 'client(s)');

        // Check if any client is visible (focused)
        const hasVisibleClient = windowClients.length > 0 &&
          windowClients.some((client) => client.visibilityState === 'visible');

        if (hasVisibleClient) {
          console.log('[SW] Client is visible, skipping push notification (WebSocket toast will handle it)');
          return;
        }

        console.log('[SW] No visible client (clients:', windowClients.length, '), showing push notification');
        return self.registration.showNotification(data.title, options)
          .then(() => console.log('[SW] Notification shown successfully'))
          .catch((err) => console.error('[SW] Failed to show notification:', err));
      })
      .catch((err) => {
        // Fallback: show notification if clients.matchAll fails
        console.error('[SW] clients.matchAll failed:', err);
        return self.registration.showNotification(data.title, options);
      })
  );
});

// Notification click event
self.addEventListener('notificationclick', (event) => {
  console.log('[SW] Notification clicked:', event);

  event.notification.close();

  if (event.action === 'dismiss') {
    return;
  }

  const urlToOpen = event.notification.data?.url || '/';

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
      // Check if there's already a window open
      for (const client of windowClients) {
        if (client.url.includes(self.location.origin) && 'focus' in client) {
          client.focus();
          if (event.notification.data?.logId) {
            client.navigate(`/logs?id=${event.notification.data.logId}`);
          } else if (event.notification.data?.projectId) {
            client.navigate(`/projects/${event.notification.data.projectId}`);
          }
          return;
        }
      }
      // Open new window if none exists
      if (clients.openWindow) {
        return clients.openWindow(urlToOpen);
      }
    })
  );
});

// Notification close event
self.addEventListener('notificationclose', (event) => {
  console.log('[SW] Notification closed:', event);
});

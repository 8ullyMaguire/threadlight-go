import { writable } from 'svelte/store';

export const notifications = writable([]);
export const unreadCount = writable(0);

export function addNotification(notification) {
  notifications.update(n => [notification, ...n]);
  unreadCount.update(c => c + 1);
}

export function markAllRead() {
  unreadCount.set(0);
}

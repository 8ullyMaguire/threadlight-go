import { writable } from 'svelte/store';
import api from '../api/client';

export const feedPosts = writable([]);
export const feedLoading = writable(false);
export const feedError = writable(null);
export const currentMood = writable('all');
export const feedOffset = writable(0);

export async function loadFeed(mood = 'all', limit = 20, offset = 0) {
  feedLoading.set(true);
  feedError.set(null);
  try {
    const data = await api.getFeed(mood, limit, offset);
    feedPosts.set(data.posts || data || []);
    feedOffset.set(offset);
  } catch (e) {
    feedError.set(e.message);
  } finally {
    feedLoading.set(false);
  }
}

export function setMood(mood) {
  currentMood.set(mood);
  loadFeed(mood);
}

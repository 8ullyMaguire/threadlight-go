import { writable } from 'svelte/store';
import api from '../api/client';

export const currentUser = writable(null);
export const isAuthenticated = writable(false);
export const onboardingStage = writable(4); // default to complete for existing users

export async function login(email, password) {
  const resp = await api.login(email, password);
  api.setToken(resp.token);
  currentUser.set(resp.user);
  isAuthenticated.set(true);
  onboardingStage.set(resp.user.onboarding_stage || 4);
  return resp;
}

export async function register(username, email, password, inviteCode) {
  const resp = await api.register(username, email, password, inviteCode);
  api.setToken(resp.token);
  currentUser.set(resp.user);
  isAuthenticated.set(true);
  onboardingStage.set(resp.user.onboarding_stage || 4);
  return resp;
}

export function logout() {
  api.clearToken();
  currentUser.set(null);
  isAuthenticated.set(false);
  onboardingStage.set(4);
}

export function checkAuth() {
  const token = api.loadToken();
  if (token) {
    isAuthenticated.set(true);
    api.session().then(data => {
      currentUser.set(data.user || data);
    }).catch(() => {
      logout();
    });
  }
}

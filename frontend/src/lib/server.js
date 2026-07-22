export const DEFAULT_PORT = '8787';
const KEY = 'ghostwire.server';

export function loadServerAddress() {
  return localStorage.getItem(KEY) || '';
}

export function saveServerAddress(value) {
  const normalized = normalizeServerAddress(value);
  if (normalized) localStorage.setItem(KEY, normalized);
  else localStorage.removeItem(KEY);
  return normalized;
}

export function normalizeServerAddress(value) {
  const trimmed = `${value ?? ''}`.trim();
  if (!trimmed) return '';
  const withScheme = /^https?:\/\//i.test(trimmed) ? trimmed : `${location.protocol}//${trimmed}`;
  const url = new URL(withScheme);
  if (!url.port) url.port = DEFAULT_PORT;
  url.pathname = '';
  url.search = '';
  url.hash = '';
  return url.toString().replace(/\/$/, '');
}

export function apiBase(serverAddress = '') {
  return serverAddress || location.origin;
}

export function websocketURL(serverAddress = '') {
  const base = new URL(apiBase(serverAddress));
  base.protocol = base.protocol === 'https:' ? 'wss:' : 'ws:';
  base.pathname = '/ws';
  base.search = '';
  base.hash = '';
  return base.toString();
}

export async function discoverServers() {
  const candidates = candidateOrigins();
  const checks = candidates.map((origin) => checkHealth(origin));
  const results = await Promise.allSettled(checks);
  return results.filter((result) => result.status === 'fulfilled' && result.value).map((result) => result.value);
}

function candidateOrigins() {
  const origins = new Set([location.origin, `http://localhost:${DEFAULT_PORT}`, `http://127.0.0.1:${DEFAULT_PORT}`]);
  const match = location.hostname.match(/^(\d+\.\d+\.\d+)\.\d+$/);
  if (match) {
    for (let i = 1; i <= 254; i += 1) origins.add(`http://${match[1]}.${i}:${DEFAULT_PORT}`);
  }
  return [...origins];
}

async function checkHealth(origin) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 900);
  try {
    const response = await fetch(`${origin}/api/health`, { signal: controller.signal, cache: 'no-store' });
    if (!response.ok) return null;
    const data = await response.json();
    return data.service === 'ghostwire' ? origin : null;
  } catch {
    return null;
  } finally {
    clearTimeout(timer);
  }
}

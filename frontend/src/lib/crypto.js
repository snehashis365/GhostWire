const enc = new TextEncoder();
const dec = new TextDecoder();

export async function generateIdentity(ephemeral = false) {
  const keys = await crypto.subtle.generateKey({ name: 'RSA-OAEP', modulusLength: 2048, publicExponent: new Uint8Array([1,0,1]), hash: 'SHA-256' }, true, ['encrypt','decrypt']);
  const publicKey = await exportKey(keys.publicKey);
  const id = await hashPublicKey(publicKey);
  const identity = { id, publicKey, ephemeral, privateKey: await exportPrivateKey(keys.privateKey) };
  if (!ephemeral) localStorage.setItem('ghostwire.identity', JSON.stringify(identity));
  else sessionStorage.setItem('ghostwire.identity', JSON.stringify(identity));
  await fetch('/api/register', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ id, public_key: publicKey, is_ephemeral: ephemeral }) });
  return identity;
}

export function loadIdentity() {
  const raw = sessionStorage.getItem('ghostwire.identity') || localStorage.getItem('ghostwire.identity');
  return raw ? JSON.parse(raw) : null;
}

export async function encryptFor(publicKeyBase64, text) {
  const key = await crypto.subtle.importKey('spki', b64(publicKeyBase64), { name:'RSA-OAEP', hash:'SHA-256' }, false, ['encrypt']);
  return bytesToBase64(new Uint8Array(await crypto.subtle.encrypt({ name:'RSA-OAEP' }, key, enc.encode(text))));
}

export async function decryptWith(privateKeyBase64, payload) {
  const key = await crypto.subtle.importKey('pkcs8', b64(privateKeyBase64), { name:'RSA-OAEP', hash:'SHA-256' }, false, ['decrypt']);
  return dec.decode(await crypto.subtle.decrypt({ name:'RSA-OAEP' }, key, b64(payload)));
}

async function exportKey(key) { return bytesToBase64(new Uint8Array(await crypto.subtle.exportKey('spki', key))); }
async function exportPrivateKey(key) { return bytesToBase64(new Uint8Array(await crypto.subtle.exportKey('pkcs8', key))); }
async function hashPublicKey(publicKey) { return bytesToBase64(new Uint8Array(await crypto.subtle.digest('SHA-256', b64(publicKey)))).replace(/[+/=]/g,'').slice(0,32); }
function b64(v) { return Uint8Array.from(atob(v), c => c.charCodeAt(0)); }
function bytesToBase64(bytes) { return btoa(String.fromCharCode(...bytes)); }

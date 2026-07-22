import { writable } from 'svelte/store';
import { websocketURL } from './server';
export const messages = writable([]);
export function connect(identity, serverAddress = '') {
  const ws = new WebSocket(websocketURL(serverAddress));
  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (msg.type === 'message') {
      messages.update((items) => [...items, msg]);
      ws.send(JSON.stringify({ type: 'read_ack', id: msg.id }));
    }
  };
  return { send: (payload, room_id = 'global') => ws.readyState === WebSocket.OPEN && ws.send(JSON.stringify({ type:'message', room_id, sender_id: identity.id, payload })), close: () => ws.close() };
}

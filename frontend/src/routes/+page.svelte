<script>
  import { onMount } from 'svelte';
  import { EyeOff, Send } from '@lucide/svelte';
  import { generateIdentity, loadIdentity, encryptFor } from '$lib/crypto';
  import { connect, messages } from '$lib/socket';
  let identity, socket, text = '', stealth = false;
  onMount(async () => { identity = loadIdentity() || await generateIdentity(false); socket = connect(identity); if ('serviceWorker' in navigator) navigator.serviceWorker.register('/service-worker.js'); });
  async function send() { if (!text.trim()) return; socket.send(await encryptFor(identity.publicKey, text)); text = ''; }
</script>
<main class="min-h-screen bg-[#050816] text-slate-100">
  <section class="mx-auto flex min-h-screen max-w-3xl flex-col p-4">
    <header class="mb-4 rounded-3xl border border-cyan-400/20 bg-slate-900/80 p-4 shadow-2xl shadow-cyan-950/40">
      <p class="text-xs uppercase tracking-[0.35em] text-cyan-300">GhostWire</p>
      <div class="flex items-center justify-between gap-3"><h1 class="text-2xl font-bold">Midnight Relay</h1><button class="rounded-full bg-cyan-400/10 p-3 text-cyan-200" on:click={() => stealth = !stealth} aria-label="Toggle stealth mode"><EyeOff size={18}/></button></div>
      <p class="mt-2 truncate text-xs text-slate-400">ID: {identity?.id ?? 'generating identity...'}</p>
    </header>
    <div class="flex-1 space-y-3 overflow-y-auto rounded-3xl border border-white/10 bg-black/30 p-4">
      {#each $messages as msg}
        <article class="rounded-2xl border border-fuchsia-400/10 bg-slate-900/80 p-3"><p class="text-xs text-fuchsia-200">{msg.sender_id}</p><p class:blur-sm={stealth} class="break-all text-sm transition">{msg.payload}</p></article>
      {/each}
    </div>
    <form class="mt-4 flex gap-2" on:submit|preventDefault={send}>
      <input class="min-w-0 flex-1 rounded-2xl border border-cyan-400/20 bg-slate-950 px-4 py-3 outline-none focus:border-cyan-300" bind:value={text} placeholder="Encrypt and relay..." />
      <button class="rounded-2xl bg-cyan-300 px-5 text-slate-950"><Send size={18}/></button>
    </form>
  </section>
</main>

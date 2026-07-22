<script>
  import { onMount } from 'svelte';
  import { EyeOff, Radar, Send } from '@lucide/svelte';
  import { generateIdentity, loadIdentity, registerIdentity, encryptFor } from '$lib/crypto';
  import { connect, messages } from '$lib/socket';
  import { discoverServers, loadServerAddress, saveServerAddress } from '$lib/server';
  let identity, socket, text = '', stealth = false, serverAddress = '', discovery = [];
  onMount(async () => {
    serverAddress = loadServerAddress();
    identity = loadIdentity() || await generateIdentity(false, serverAddress);
    socket = connect(identity, serverAddress);
    if ('serviceWorker' in navigator) navigator.serviceWorker.register('/service-worker.js');
  });
  async function reconnect(address = serverAddress) {
    serverAddress = saveServerAddress(address);
    if (identity) await registerIdentity(identity, serverAddress);
    socket?.close();
    socket = connect(identity, serverAddress);
  }
  async function scan() { discovery = await discoverServers(); }
  async function send() { if (!text.trim()) return; socket.send(await encryptFor(identity.publicKey, text)); text = ''; }
</script>
<main class="min-h-screen bg-[#050816] text-slate-100">
  <section class="mx-auto flex min-h-screen max-w-3xl flex-col p-4">
    <header class="mb-4 rounded-3xl border border-cyan-400/20 bg-slate-900/80 p-4 shadow-2xl shadow-cyan-950/40">
      <p class="text-xs uppercase tracking-[0.35em] text-cyan-300">GhostWire</p>
      <div class="flex items-center justify-between gap-3"><h1 class="text-2xl font-bold">Midnight Relay</h1><button class="rounded-full bg-cyan-400/10 p-3 text-cyan-200" on:click={() => stealth = !stealth} aria-label="Toggle stealth mode"><EyeOff size={18}/></button></div>
      <p class="mt-2 truncate text-xs text-slate-400">ID: {identity?.id ?? 'generating identity...'}</p>
      <div class="mt-4 grid gap-2 rounded-2xl border border-white/10 bg-black/20 p-3">
        <label class="text-xs uppercase tracking-widest text-slate-400" for="server">Server address</label>
        <div class="flex gap-2">
          <input id="server" class="min-w-0 flex-1 rounded-xl border border-cyan-400/20 bg-slate-950 px-3 py-2 text-sm outline-none focus:border-cyan-300" bind:value={serverAddress} placeholder="Auto: this host, or e.g. 192.168.1.50:8787" />
          <button type="button" class="rounded-xl bg-cyan-400/10 px-3 text-cyan-100" on:click={() => reconnect()}>Use</button>
          <button type="button" class="rounded-xl bg-fuchsia-400/10 px-3 text-fuchsia-100" on:click={scan} aria-label="Discover servers"><Radar size={18}/></button>
        </div>
        {#if discovery.length}
          <div class="flex flex-wrap gap-2">{#each discovery as found}<button class="rounded-full bg-cyan-300 px-3 py-1 text-xs text-slate-950" on:click={() => reconnect(found)}>{found}</button>{/each}</div>
        {/if}
      </div>
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

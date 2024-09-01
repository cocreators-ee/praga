<script lang="ts">
  import {onMount} from "svelte";
  import {page} from "$app/stores"
  import {verifyToken} from "$lib/api"
  import {config, refreshConfig, verified} from "$lib/state"
  import {fly} from 'svelte/transition'
  import Lottie from "$lib/Lottie.svelte";
  import HeaderIcon from "$lib/assets/shield-antivirus-svgrepo-com.svg"
  import CheckmarkAnim from "$lib/assets/lottie-check.json"
  import LoadingAnim from "$lib/assets/lottie-loading.json"
  import LoginForm from "./LoginForm.svelte"
  import {sleep} from "$lib/utils";

  let loading = true
  let loadTimer = sleep(125)
  let loginComplete = false
  $: redirectTarget = getRedirectTarget($page.url)

  function getRedirectTarget(url: URL): string {
    return url.hash.split("#r=")[1]
  }

  function redirectToTarget() {
    if (redirectTarget) {
      window.location.replace(redirectTarget)
    }
  }

  function onLoginCheckmarkComplete() {
    redirectToTarget()
  }

  function onLoginComplete() {
    loginComplete = true
  }

  onMount(async () => {
    loading = true
    const refreshPromise = refreshConfig()
    verified.set(await verifyToken())
    await refreshPromise
    await loadTimer
    loading = false
  })
</script>

<svelte:head>
  {#if $config !== undefined}
    <title>{$config.title}</title>
  {/if}
</svelte:head>

{#if loading}
  <div class="loading">
    <Lottie
      autoplay
      preserveAspectRatio="xMidYMid slice"
      loopFrame={0}
      animationData={LoadingAnim}
    />
  </div>
{:else}
  <section transition:fly={{ x: '100%', duration: 400 }}
           class:verified={$verified}>
    <div class="header-icon">
      <HeaderIcon/>
    </div>

    <h1>Log in to {$config.brand}</h1>

    {#if loginComplete}
      {#if redirectTarget !== undefined}
        <p>Logged in successfully, you will now be redirected to the requested service.</p>
      {:else}
        <p>Logged in successfully, however could not determine requested service. Please try to manually
          navigate there now.</p>
      {/if}
      <Lottie
        autoplay
        preserveAspectRatio="xMidYMid slice"
        animationData={CheckmarkAnim}
        on:complete={onLoginCheckmarkComplete}
      />
    {:else}
      <LoginForm verified={$verified} on:complete={onLoginComplete}/>
    {/if}
  </section>
  <footer>
    <p>Please contact {$config.support} in case of issues.</p>
  </footer>
{/if}

<style lang="scss">
  @import "$lib/style/variables";

  .loading {
    max-width: 16rem;
  }

  h1 {
    font-weight: 400;
    font-size: 1.5rem;
    color: $text-bright;
    text-align: center;

    margin-top: 2.5rem;
    margin-bottom: 3rem;
  }

  section {
    width: 22rem;
    background: $surface;
    border: 1px solid transparent;
    border-radius: 3px;
    padding: 2rem;
    margin-top: 5rem; // For icon to fit on screen
    position: relative;
    flex-shrink: 0;

    &.verified {
      $ok: #354e2a;
      border: 1px solid darken($ok, 5%);
      box-shadow: 0 0 8px 1px $ok;
    }
  }

  .header-icon {
    $size: 8rem;
    width: $size;
    height: $size;

    position: absolute;
    top: $size * -0.66;
    left: 50%;
    transform: translate(-50%, 0);
  }

  footer {
    margin-top: 2rem;
    color: $text-dim;
    font-weight: 400;
    font-size: 14px;
  }
</style>

<script lang="ts">
  import {createEventDispatcher} from 'svelte'
  import {slide,} from 'svelte/transition'
  import {emailSend, emailVerify} from "$lib/api"
  import EmailIcon from "$lib/assets/email-letter-mail-message-communication-office-svgrepo-com.svg"
  import FingerprintIcon from "$lib/assets/fingerprint-svgrepo-com.svg"

  export let verified = false

  const dispatch = createEventDispatcher()

  let state = "email"
  let email = ""
  let code = ""

  let errorField: HTMLInputElement
  let activeForm: HTMLFormElement

  function onRequestCode() {
    emailSend(email)
    state = "code"
  }

  function clearCustomValidity() {
    errorField.setCustomValidity("")
    activeForm.reportValidity()
  }

  async function onVerifyCode() {
    const result = await emailVerify(email, code)
    if (result) {
      dispatch('complete', {})
    } else {
      errorField.setCustomValidity("Code verification failed.")
      activeForm.reportValidity()
    }
  }
</script>

{#if state === "email"}

  <form bind:this={activeForm} on:submit|preventDefault={onRequestCode} transition:slide={{}}>
    <label for="email">Email</label>
    <input bind:this={errorField} on:keydown={clearCustomValidity} id="email" name="email" type="email"
           placeholder="user@example.com" required bind:value={email}>
    <div class="buttons">
      <button type="submit">
        <EmailIcon/>
        Request code
      </button>
    </div>

    {#if verified}
      <p>You seem to already be logged in, but you're free to re-login to refresh your authentication token.</p>
    {/if}
  </form>

{:else if state === "code"}

  <form bind:this={activeForm} on:submit|preventDefault={onVerifyCode} transition:slide={{}}>
    <label for="code">Verification code</label>
    <input bind:this={errorField} on:keydown={clearCustomValidity} id="code" name="code" type="text"
           placeholder="ABCD1234" required bind:value={code}>
    <div class="buttons">
      <button type="submit">
        <FingerprintIcon/>
        Verify code
      </button>
      <button on:click={() => state = "email"} type="button">
        <EmailIcon/>
        Try another email
      </button>
    </div>
    <p>Code requested, if <em>{email}</em> is allowed to log in you should receive an email soon.</p>
  </form>

{/if}

<style lang="scss">
  @import "$lib/style/variables";

  form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  em {
    color: $text-bright;
    padding: 0 0.25rem;
  }

  .buttons {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin: 0.5rem 0 1rem 0;
  }

  :global(button svg) {
    width: 1.5rem;

    :global(path) {
      fill: #fff !important;
    }
  }
</style>

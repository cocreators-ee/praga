<script lang="ts">
  import {page} from "$app/stores"

  // Photo by Christina Morillo: https://www.pexels.com/photo/woman-wearing-black-hoodie-jacket-holding-grey-laptop-computer-1181325/
  import ogImg from "$lib/assets/pexels-divinetechygirl-1181325.jpg"

  import "@fontsource/merriweather-sans"
  import "$lib/style/global.scss"
</script>

<svelte:head>
  <title>Login required</title>
  <meta property="og:image" content={$page.url.origin + ogImg}/>
  <meta property="og:image:width" content="1200"/>
  <meta property="og:image:height" content="630"/>
  <meta property="og:image:type" content="image/jpg"/>
  <meta property="og:type" content="og:product"/>
  <meta property="og:description"
        content="Log in to access the secure service requested."
  />
</svelte:head>

<main>
  <slot/>
</main>

<style lang="scss">
  @import "$lib/style/variables";

  :global(body, html, main) {
    align-items: center;
    background: $background;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    min-height: 100%;
    justify-content: center;
    margin: 0;
    padding: 0;
    width: 100%;
    color: $text-mid;
  }

  main {
    margin: 2rem;
  }
</style>

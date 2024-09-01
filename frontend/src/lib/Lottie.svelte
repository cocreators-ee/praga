<script lang="ts">
  import lottie, {type AnimationItem} from "lottie-web"
  import {createEventDispatcher, onDestroy, onMount} from "svelte"

  import {sleep} from "$lib/utils"

  export let animationData: object
  export let autoplay = false
  export let playbackSpeed = 1
  export let loopFrame = -1
  export let finalFrame = undefined
  export let loopSegment = undefined
  export let preserveAspectRatio = "xMidYMid meet"

  let lottieRef: Element
  let instance: AnimationItem
  let ready = false
  let dispatch = createEventDispatcher()
  let stopLooping = false
  let lastLoop = false

  export async function play() {
    while (!ready) {
      await sleep(25)
    }

    if (loopSegment) {
      instance.playSegments([0, Math.min.apply(null, loopSegment)], true)
    } else {
      instance.goToAndPlay(0, true)
    }
  }

  export function stopLoopSegment() {
    stopLooping = true
  }

  async function load() {
    const options = {
      animationData: animationData,
      autoplay: false,
      container: lottieRef,
      loop: false,
      rendererSettings: {
        preserveAspectRatio: preserveAspectRatio,
        progressiveLoad: false,
        hideOnTransparent: true,
        className: "animation-root",
      },
    }

    instance = lottie.loadAnimation<"svg">(options)
    instance.setSpeed(playbackSpeed)

    if (finalFrame) {
      instance.addEventListener("enterFrame", () => {
        if (instance.currentFrame >= finalFrame) {
          dispatch("complete", {})
        }
      })
    }

    instance.addEventListener("complete", (e) => {
      if (!e || e.type !== "complete") {
        // WTF are these garbage events
        return
      }
      if (loopSegment !== undefined) {
        if (lastLoop) {
          dispatch("complete", {})
        } else if (stopLooping) {
          const lastFrame = Math.max.apply(null, loopSegment)
          instance.playSegments([lastFrame, 9999], true)
          lastLoop = true
          instance.loop = false
        } else {
          instance.playSegments(loopSegment, true)
        }
      } else {
        if (loopFrame !== -1) {
          instance.goToAndPlay(loopFrame, true)
        }
        dispatch("complete", {})
      }
    })

    // Wait for lottie to be loaded
    while (!lottieRef.querySelector("svg.animation-root")) {
      await sleep(25)
    }

    lottieRef.querySelector("svg.animation-root").removeAttribute("width")
    lottieRef.querySelector("svg.animation-root").removeAttribute("height")

    ready = true

    if (autoplay) {
      await play()
    }
  }

  onMount(async () => {
    if (!instance) {
      while (!lottieRef) {
        await sleep(25)
      }

      while (!animationData) {
        await sleep(25)
      }

      await load()
    }
  })

  onDestroy(() => {
    if (instance) {
      instance.stop()
      instance.destroy()
      instance = undefined
    }
  })
</script>

<div class="lottie lottie-player" bind:this={lottieRef}/>

<style lang="scss">
  :global(.lottie-player, .lottie-player svg.animation-root) {
    align-items: center;
    display: flex;
    flex: 1;
    flex-direction: column;
    height: 100%;
    justify-content: center;
    width: 100%;
  }

  :global(.lottie-player svg.animation-root) {
    height: 100%;
    width: 100%;
  }
</style>

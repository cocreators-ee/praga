import {writable} from "svelte/store";
import {type ConfigResponse, getConfig} from "$lib/api";

export const verified = writable<boolean>(false)
export const config = writable<ConfigResponse>(undefined)

export async function refreshConfig() {
  config.set(await getConfig())
}

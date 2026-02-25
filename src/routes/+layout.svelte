<script lang="ts">
	import '../app.css';
	import { checkAuth } from '$lib/stores/auth';
	import { onMount } from 'svelte';

	let { children } = $props();

	onMount(() => {
		checkAuth();
	});
</script>

{@render children()}

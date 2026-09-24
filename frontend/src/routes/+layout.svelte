<script lang="ts">
import '../styles.css';

let { children } = $props();
</script>

<svelte:head>
	<title>My App</title>
	<meta name="description" content="A Go and SvelteKit application" />
</svelte:head>

{@render children()}

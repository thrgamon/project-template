<script lang="ts">
import { ApiError, api, type User } from '$lib/api';

let user = $state<User | null>(null);
let ready = $state(false);

async function loadSession() {
	try {
		user = (await api.me()).user;
	} catch (error) {
		if (!(error instanceof ApiError && error.status === 401)) throw error;
	} finally {
		ready = true;
	}
}

loadSession();
</script>

<main>
	<h1>My App</h1>
	{#if ready}
		{#if user}
			<p>Signed in as {user.email}.</p>
			<a href="/dashboard">Go to dashboard</a>
		{:else}
			<nav aria-label="Account"><a href="/login">Login</a><a href="/register">Register</a></nav>
		{/if}
	{/if}
</main>

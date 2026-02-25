<script lang="ts">
	import { getUser, isLoading, logout } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	const user = $derived(getUser());
	const loading = $derived(isLoading());

	async function handleLogout() {
		await logout();
		goto('/');
	}

	$effect(() => {
		if (!loading && !user) {
			goto('/login');
		}
	});
</script>

<main class="min-h-screen bg-gray-50">
	<nav class="bg-white shadow">
		<div class="mx-auto flex max-w-4xl items-center justify-between px-4 py-3">
			<h1 class="text-lg font-semibold text-gray-900">Dashboard</h1>
			{#if user}
				<div class="flex items-center gap-4">
					<span class="text-sm text-gray-600">{user.email}</span>
					<button
						onclick={handleLogout}
						class="rounded-md bg-gray-200 px-3 py-1 text-sm text-gray-700 hover:bg-gray-300"
					>
						Logout
					</button>
				</div>
			{/if}
		</div>
	</nav>

	<div class="mx-auto max-w-4xl p-8">
		{#if loading}
			<p class="text-gray-500">Loading...</p>
		{:else if user}
			<div class="rounded-lg bg-white p-6 shadow">
				<h2 class="text-xl font-semibold text-gray-900">Welcome, {user.email}</h2>
				<p class="mt-2 text-gray-600">This is a protected page. Start building your app here.</p>
			</div>
		{/if}
	</div>
</main>

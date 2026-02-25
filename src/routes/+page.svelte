<script lang="ts">
	import { getUser, isLoading } from '$lib/stores/auth';

	const user = $derived(getUser());
	const loading = $derived(isLoading());
</script>

<main class="flex min-h-screen items-center justify-center bg-gray-50">
	<div class="w-full max-w-md space-y-6 p-8">
		<h1 class="text-center text-3xl font-bold text-gray-900">Project Template</h1>

		{#if loading}
			<p class="text-center text-gray-500">Loading...</p>
		{:else if user}
			<div class="space-y-4 text-center">
				<p class="text-gray-700">Welcome, {user.email}</p>
				<a
					href="/dashboard"
					class="inline-block rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
				>
					Dashboard
				</a>
			</div>
		{:else}
			<div class="flex justify-center gap-4">
				<a
					href="/login"
					class="rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
				>
					Login
				</a>
				<a
					href="/register"
					class="rounded-md border border-blue-600 px-4 py-2 text-blue-600 hover:bg-blue-50"
				>
					Register
				</a>
			</div>
		{/if}
	</div>
</main>

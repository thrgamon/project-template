<script lang="ts">
	import { register } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let error = $state('');
	let submitting = $state(false);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}

		if (password.length < 8) {
			error = 'Password must be at least 8 characters';
			return;
		}

		submitting = true;
		const result = await register(email, password);
		submitting = false;

		if (result) {
			error = result;
		} else {
			goto('/dashboard');
		}
	}
</script>

<main class="flex min-h-screen items-center justify-center bg-gray-50">
	<div class="w-full max-w-md space-y-6 rounded-lg bg-white p-8 shadow">
		<h1 class="text-center text-2xl font-bold text-gray-900">Register</h1>

		{#if error}
			<p class="rounded-md bg-red-50 p-3 text-sm text-red-600">{error}</p>
		{/if}

		<form onsubmit={handleSubmit} class="space-y-4">
			<div>
				<label for="email" class="block text-sm font-medium text-gray-700">Email</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					required
					class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:ring-blue-500 focus:outline-none"
				/>
			</div>

			<div>
				<label for="password" class="block text-sm font-medium text-gray-700">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					minlength="8"
					class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:ring-blue-500 focus:outline-none"
				/>
			</div>

			<div>
				<label for="confirm-password" class="block text-sm font-medium text-gray-700">Confirm Password</label>
				<input
					id="confirm-password"
					type="password"
					bind:value={confirmPassword}
					required
					class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:ring-blue-500 focus:outline-none"
				/>
			</div>

			<button
				type="submit"
				disabled={submitting}
				class="w-full rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
			>
				{submitting ? 'Creating account...' : 'Register'}
			</button>
		</form>

		<p class="text-center text-sm text-gray-600">
			Already have an account? <a href="/login" class="text-blue-600 hover:underline">Login</a>
		</p>
	</div>
</main>

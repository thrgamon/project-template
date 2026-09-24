<script lang="ts">
import { api } from '$lib/api';

let email = $state('');
let password = $state('');
let error = $state('');
let submitting = $state(false);

async function submit() {
	error = '';
	submitting = true;
	try {
		await api.login(email, password);
		window.location.assign('/dashboard');
	} catch (reason) {
		error = reason instanceof Error ? reason.message : 'Login failed';
	} finally {
		submitting = false;
	}
}
</script>

<main><form onsubmit={(event) => { event.preventDefault(); submit(); }}><h1>Login</h1>{#if error}<p class="error" role="alert">{error}</p>{/if}<label>Email <input bind:value={email} type="email" autocomplete="email" required /></label><label>Password <input bind:value={password} type="password" autocomplete="current-password" required /></label><button type="submit" disabled={submitting}>{submitting ? 'Logging in…' : 'Login'}</button><p>No account? <a href="/register">Register</a></p></form></main>

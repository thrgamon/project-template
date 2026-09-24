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
		await api.register(email, password);
		window.location.assign('/dashboard');
	} catch (reason) {
		error = reason instanceof Error ? reason.message : 'Registration failed';
	} finally {
		submitting = false;
	}
}
</script>

<main><form onsubmit={(event) => { event.preventDefault(); submit(); }}><h1>Create account</h1>{#if error}<p class="error" role="alert">{error}</p>{/if}<label>Email <input bind:value={email} type="email" autocomplete="email" required /></label><label>Password <input bind:value={password} type="password" autocomplete="new-password" minlength="8" required /></label><button type="submit" disabled={submitting}>{submitting ? 'Creating…' : 'Register'}</button><p>Already have an account? <a href="/login">Login</a></p></form></main>

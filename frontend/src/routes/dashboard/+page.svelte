<script lang="ts">
import { ApiError, api, type DashboardResponse, type User } from '$lib/api';

let dashboard = $state<DashboardResponse | null>(null);
let error = $state('');
let user = $state<User | null>(null);
let csrfToken = $state('');

async function loadDashboard() {
	try {
		const session = await api.me();
		user = session.user;
		csrfToken = session.csrfToken;
		dashboard = await api.dashboard();
	} catch (reason) {
		if (reason instanceof ApiError && reason.status === 401) {
			window.location.assign('/login');
			return;
		}
		error = reason instanceof Error ? reason.message : 'Could not load dashboard';
	}
}

async function logout() {
	if (csrfToken) await api.logout(csrfToken);
	window.location.assign('/');
}

loadDashboard();
</script>

<main><h1>Dashboard</h1>{#if error}<p class="error" role="alert">{error}</p>{/if}{#if dashboard && user}<p>{dashboard.message}, {dashboard.email}.</p><button type="button" onclick={logout}>Logout</button>{/if}</main>

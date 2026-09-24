<script lang="ts">
import { ApiError, api, type DashboardResponse } from '$lib/api';

let dashboard = $state<DashboardResponse | null>(null);
let error = $state('');

async function loadDashboard() {
	try {
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
	await api.logout();
	window.location.assign('/');
}

loadDashboard();
</script>

<main><h1>Dashboard</h1>{#if error}<p class="error" role="alert">{error}</p>{/if}{#if dashboard}<p>{dashboard.message}, {dashboard.email}.</p><button type="button" onclick={logout}>Logout</button>{/if}</main>

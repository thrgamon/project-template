interface User {
	id: number;
	email: string;
}

let user = $state<User | null>(null);
let loading = $state(true);

export function getUser() {
	return user;
}

export function isLoading() {
	return loading;
}

export async function checkAuth(): Promise<void> {
	try {
		const res = await fetch('/api/auth/me');
		if (res.ok) {
			const data = await res.json();
			user = data;
		} else {
			user = null;
		}
	} catch {
		user = null;
	} finally {
		loading = false;
	}
}

export async function login(email: string, password: string): Promise<string | null> {
	const res = await fetch('/api/auth/login', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password })
	});

	if (!res.ok) {
		const data = await res.json();
		return data.error || 'Login failed';
	}

	const data = await res.json();
	user = data.user;
	return null;
}

export async function register(email: string, password: string): Promise<string | null> {
	const res = await fetch('/api/auth/register', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password })
	});

	if (!res.ok) {
		const data = await res.json();
		return data.error || 'Registration failed';
	}

	const data = await res.json();
	user = data.user;
	return null;
}

export async function logout(): Promise<void> {
	await fetch('/api/auth/logout', { method: 'POST' });
	user = null;
}

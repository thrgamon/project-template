'use client';

import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { useAuth } from '@/lib/auth-context';

export default function DashboardPage() {
	const { user, loading, logout } = useAuth();
	const router = useRouter();

	useEffect(() => {
		if (!loading && !user) {
			router.push('/login');
		}
	}, [loading, user, router]);

	if (loading || !user) return null;

	return (
		<main className="flex min-h-screen flex-col items-center justify-center gap-4">
			<h1 className="text-2xl font-bold">Dashboard</h1>
			<p className="text-muted-foreground">Welcome, {user.email}</p>
			<button
				type="button"
				onClick={async () => {
					await logout();
					router.push('/');
				}}
				className="rounded bg-secondary px-4 py-2 text-secondary-foreground"
			>
				Logout
			</button>
		</main>
	);
}

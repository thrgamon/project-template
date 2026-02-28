'use client';

import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';

export default function Home() {
	const { user, loading } = useAuth();

	if (loading) return null;

	return (
		<main className="flex min-h-screen flex-col items-center justify-center gap-4">
			<h1 className="text-4xl font-bold">My App</h1>
			{user ? (
				<Link href="/dashboard" className="text-primary underline">
					Go to Dashboard
				</Link>
			) : (
				<div className="flex gap-4">
					<Link href="/login" className="text-primary underline">
						Login
					</Link>
					<Link href="/register" className="text-primary underline">
						Register
					</Link>
				</div>
			)}
		</main>
	);
}

'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { ErrorBanner } from '@/components/ErrorBanner';

export default function LoginPage() {
	const [email, setEmail] = useState('');
	const [password, setPassword] = useState('');
	const [error, setError] = useState('');
	const { login } = useAuth();
	const router = useRouter();

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setError('');
		try {
			await login(email, password);
			router.push('/dashboard');
		} catch (err) {
			setError(err instanceof Error ? err.message : 'Login failed');
		}
	}

	return (
		<main className="flex min-h-screen items-center justify-center">
			<form onSubmit={handleSubmit} className="w-full max-w-sm space-y-4">
				<h1 className="text-2xl font-bold">Login</h1>
				{error && <ErrorBanner message={error} />}
				<input
					type="email"
					placeholder="Email"
					value={email}
					onChange={(e) => setEmail(e.target.value)}
					required
					className="w-full rounded border border-input bg-background px-3 py-2"
				/>
				<input
					type="password"
					placeholder="Password"
					value={password}
					onChange={(e) => setPassword(e.target.value)}
					required
					className="w-full rounded border border-input bg-background px-3 py-2"
				/>
				<button type="submit" className="w-full rounded bg-primary px-4 py-2 text-primary-foreground">
					Login
				</button>
				<p className="text-sm text-muted-foreground">
					No account?{' '}
					<Link href="/register" className="text-primary underline">
						Register
					</Link>
				</p>
			</form>
		</main>
	);
}

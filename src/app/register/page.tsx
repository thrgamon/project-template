'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { ErrorBanner } from '@/components/ErrorBanner';

export default function RegisterPage() {
	const [email, setEmail] = useState('');
	const [password, setPassword] = useState('');
	const [error, setError] = useState('');
	const { register } = useAuth();
	const router = useRouter();

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setError('');
		try {
			await register(email, password);
			router.push('/dashboard');
		} catch (err) {
			setError(err instanceof Error ? err.message : 'Registration failed');
		}
	}

	return (
		<main className="flex min-h-screen items-center justify-center">
			<form onSubmit={handleSubmit} className="w-full max-w-sm space-y-4">
				<h1 className="text-2xl font-bold">Register</h1>
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
					placeholder="Password (min 8 characters)"
					value={password}
					onChange={(e) => setPassword(e.target.value)}
					required
					minLength={8}
					className="w-full rounded border border-input bg-background px-3 py-2"
				/>
				<button type="submit" className="w-full rounded bg-primary px-4 py-2 text-primary-foreground">
					Register
				</button>
				<p className="text-sm text-muted-foreground">
					Already have an account?{' '}
					<Link href="/login" className="text-primary underline">
						Login
					</Link>
				</p>
			</form>
		</main>
	);
}

'use client';

import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import type { User } from '@/lib/types';

interface AuthContextType {
	user: User | null;
	loading: boolean;
	login: (email: string, password: string) => Promise<void>;
	register: (email: string, password: string) => Promise<void>;
	logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

function extractError(data: unknown, fallback: string): string {
	if (data && typeof data === 'object' && 'error' in data && typeof (data as Record<string, unknown>).error === 'string') {
		return (data as Record<string, string>).error;
	}
	return fallback;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
	const [user, setUser] = useState<User | null>(null);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		fetch('/api/auth/me', { credentials: 'include' })
			.then(async (res) => {
				if (res.ok) {
					const data = await res.json();
					setUser(data.user ?? data);
				}
			})
			.catch(() => {})
			.finally(() => setLoading(false));
	}, []);

	const login = useCallback(async (email: string, password: string) => {
		const res = await fetch('/api/auth/login', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			credentials: 'include',
			body: JSON.stringify({ email, password }),
		});
		const data = await res.json();
		if (!res.ok) throw new Error(extractError(data, 'Login failed'));
		setUser(data.user ?? data);
	}, []);

	const register = useCallback(async (email: string, password: string) => {
		const res = await fetch('/api/auth/register', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			credentials: 'include',
			body: JSON.stringify({ email, password }),
		});
		const data = await res.json();
		if (!res.ok) throw new Error(extractError(data, 'Registration failed'));
		setUser(data.user ?? data);
	}, []);

	const logout = useCallback(async () => {
		await fetch('/api/auth/logout', { method: 'POST', credentials: 'include' });
		setUser(null);
	}, []);

	return <AuthContext value={{ user, loading, login, register, logout }}>{children}</AuthContext>;
}

export function useAuth(): AuthContextType {
	const ctx = useContext(AuthContext);
	if (!ctx) throw new Error('useAuth must be used within AuthProvider');
	return ctx;
}

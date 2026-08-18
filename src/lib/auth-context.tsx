'use client';

import { createContext, useCallback, useContext, useMemo } from 'react';
import { useLogin, useLogout, useMe, useRegister } from '@/lib/api/hooks';
import type { UserResponse } from '@/lib/api/types';

interface AuthContextType {
	user: UserResponse | null;
	loading: boolean;
	login: (email: string, password: string) => Promise<void>;
	register: (email: string, password: string) => Promise<void>;
	logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
	const { data, isPending } = useMe();
	const loginMutation = useLogin();
	const registerMutation = useRegister();
	const logoutMutation = useLogout();

	const login = useCallback(
		async (email: string, password: string) => {
			await loginMutation.mutateAsync({ email, password });
		},
		[loginMutation],
	);

	const register = useCallback(
		async (email: string, password: string) => {
			await registerMutation.mutateAsync({ email, password });
		},
		[registerMutation],
	);

	const logout = useCallback(async () => {
		await logoutMutation.mutateAsync();
	}, [logoutMutation]);

	const value = useMemo(
		() => ({ user: data?.user ?? null, loading: isPending, login, register, logout }),
		[data, isPending, login, register, logout],
	);

	return <AuthContext value={value}>{children}</AuthContext>;
}

export function useAuth(): AuthContextType {
	const ctx = useContext(AuthContext);
	if (!ctx) throw new Error('useAuth must be used within AuthProvider');
	return ctx;
}

'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ApiError, api } from './client';
import type { AuthResponse } from './types';

export const queryKeys = {
	me: ['auth', 'me'] as const,
	dashboard: ['dashboard'] as const,
};

/**
 * Reads the current session. A 401 is a valid answer meaning "signed out", so
 * it resolves to null rather than throwing.
 */
export function useMe() {
	return useQuery<AuthResponse | null>({
		queryKey: queryKeys.me,
		queryFn: async () => {
			try {
				return await api.me();
			} catch (err) {
				if (err instanceof ApiError && err.status === 401) return null;
				throw err;
			}
		},
		retry: false,
		staleTime: Number.POSITIVE_INFINITY,
	});
}

export function useLogin() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: api.login,
		onSuccess: (data) => queryClient.setQueryData(queryKeys.me, data),
	});
}

export function useRegister() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: api.register,
		onSuccess: (data) => queryClient.setQueryData(queryKeys.me, data),
	});
}

export function useLogout() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: api.logout,
		onSuccess: () => {
			queryClient.setQueryData(queryKeys.me, null);
			queryClient.removeQueries({ queryKey: queryKeys.dashboard });
		},
	});
}

export function useDashboard() {
	return useQuery({
		queryKey: queryKeys.dashboard,
		queryFn: api.dashboard,
	});
}

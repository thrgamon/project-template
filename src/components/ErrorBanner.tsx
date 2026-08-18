interface ErrorBannerProps {
	message: string;
}

export function ErrorBanner({ message }: ErrorBannerProps) {
	return (
		<div
			role="alert"
			className="rounded border border-destructive bg-destructive/10 px-4 py-2 text-sm text-destructive-foreground"
		>
			{message}
		</div>
	);
}

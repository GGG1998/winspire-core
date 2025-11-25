interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function LoadingSpinner({ size = 'md', className = '' }: LoadingSpinnerProps) {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-8 w-8',
    lg: 'h-12 w-12',
  };

  return (
    <div className={`flex items-center justify-center p-8 ${className}`} role="status" aria-label="Loading">
      <div
        className={`animate-spin rounded-full border-b-2 border-gray-900 dark:border-gray-100 ${sizeClasses[size]}`}
      />
      <span className="sr-only">Loading...</span>
    </div>
  );
}


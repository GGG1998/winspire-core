interface ErrorMessageProps {
  message: string;
  className?: string;
  onRetry?: () => void;
}

export function ErrorMessage({ message, className = '', onRetry }: ErrorMessageProps) {
  return (
    <div
      className={`bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 px-4 py-3 rounded ${className}`}
      role="alert"
    >
      <div className="flex items-center justify-between">
        <span>{message}</span>
        {onRetry && (
          <button
            onClick={onRetry}
            className="ml-4 text-sm underline hover:no-underline focus:outline-none focus:ring-2 focus:ring-red-500 rounded"
          >
            Retry
          </button>
        )}
      </div>
    </div>
  );
}


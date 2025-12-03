interface SkeletonLoaderProps {
  className?: string;
  lines?: number;
}

export function SkeletonLoader({ className = '', lines = 1 }: SkeletonLoaderProps) {
  return (
    <div className={`animate-pulse space-y-2 ${className}`}>
      {Array.from({ length: lines }).map((_, i) => (
        <div
          key={i}
          className="h-4 bg-gray-200 dark:bg-gray-700 rounded"
          style={{ width: i === lines - 1 ? '75%' : '100%' }}
        />
      ))}
    </div>
  );
}







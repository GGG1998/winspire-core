import { LoadingSpinner } from './common/LoadingSpinner';

export type ConnectionStatus = 'connected' | 'connecting' | 'reconnecting' | 'disconnected' | 'error';

interface ConnectionIndicatorProps {
  status: ConnectionStatus;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

/**
 * ConnectionIndicator Component
 * 
 * Displays current WebSocket connection status with:
 * - Colored status dot
 * - Status label
 * - Loading spinner for connecting states
 * - Animations (pulse for connected, spin for connecting)
 */
export function ConnectionIndicator({ 
  status, 
  className = '',
  size = 'md'
}: ConnectionIndicatorProps) {
  const sizeClasses = {
    sm: 'text-xs gap-1.5',
    md: 'text-sm gap-2',
    lg: 'text-base gap-2.5',
  };

  const dotSizes = {
    sm: 'size-1.5',
    md: 'size-2',
    lg: 'size-2.5',
  };

  const getStatusConfig = () => {
    switch (status) {
      case 'connected':
        return {
          color: 'bg-green-500',
          label: 'Połączony',
          textColor: 'text-gray-600 dark:text-gray-300',
          animate: 'animate-pulse',
        };
      case 'connecting':
        return {
          color: 'bg-blue-500',
          label: 'Łączenie...',
          textColor: 'text-gray-600 dark:text-gray-300',
          animate: '',
        };
      case 'reconnecting':
        return {
          color: 'bg-amber-500',
          label: 'Ponowne łączenie...',
          textColor: 'text-amber-600 dark:text-amber-400',
          animate: '',
        };
      case 'disconnected':
        return {
          color: 'bg-red-500',
          label: 'Rozłączony',
          textColor: 'text-red-600 dark:text-red-400',
          animate: '',
        };
      case 'error':
        return {
          color: 'bg-red-500',
          label: 'Błąd połączenia',
          textColor: 'text-red-600 dark:text-red-400',
          animate: '',
        };
      default:
        return {
          color: 'bg-gray-500',
          label: 'Nieznany',
          textColor: 'text-gray-600 dark:text-gray-300',
          animate: '',
        };
    }
  };

  const config = getStatusConfig();
  const showSpinner = status === 'connecting' || status === 'reconnecting';

  return (
    <div className={`flex items-center ${sizeClasses[size]} ${className}`}>
      {showSpinner ? (
        <LoadingSpinner size="sm" />
      ) : (
        <div className={`${dotSizes[size]} rounded-full ${config.color} ${config.animate}`} />
      )}
      <span className={`font-medium ${config.textColor}`}>
        {config.label}
      </span>
    </div>
  );
}


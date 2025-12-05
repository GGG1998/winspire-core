import { useEffect, useRef } from 'react';
import type { ActivityFeedItem } from '../types';

interface ActivityFeedProps {
  items: ActivityFeedItem[];
  maxItems?: number;
}

/**
 * Display activity feed showing participant joins/leaves and tournament updates
 * Auto-scrolls to latest activity
 */
export function ActivityFeed({ items, maxItems = 50 }: ActivityFeedProps) {
  const feedEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new items arrive
  useEffect(() => {
    feedEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [items.length]);

  const displayedItems = items.slice(-maxItems);

  if (displayedItems.length === 0) {
    return (
      <div className="rounded-lg border border-gray-200 bg-gray-50 p-6 text-center">
        <p className="text-sm text-gray-500">Brak aktywności</p>
        <p className="mt-1 text-xs text-gray-400">Aktywność pojawi się tutaj</p>
      </div>
    );
  }

  const getActivityIcon = (type: ActivityFeedItem['type']) => {
    switch (type) {
      case 'participant_joined':
        return '👋';
      case 'participant_left':
        return '👋';
      case 'tournament_starting':
        return '🏁';
      case 'grace_period_started':
        return '⏳';
      default:
        return '📢';
    }
  };

  const getActivityColor = (type: ActivityFeedItem['type']) => {
    switch (type) {
      case 'participant_joined':
        return 'text-green-600';
      case 'participant_left':
        return 'text-gray-500';
      case 'tournament_starting':
        return 'text-cyan-600';
      case 'grace_period_started':
        return 'text-amber-600';
      default:
        return 'text-gray-600';
    }
  };

  return (
    <div className="space-y-2">
      <h3 className="text-sm font-semibold text-gray-900">Aktywność</h3>

      <div className="rounded-lg border border-gray-200 bg-white">
        <div className="max-h-64 overflow-y-auto p-3">
          <div className="space-y-2">
            {displayedItems.map((item) => {
              const time = new Date(item.timestamp).toLocaleTimeString('pl-PL', {
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit',
              });

              return (
                <div
                  key={item.id}
                  className="flex items-start gap-2 text-sm animate-in fade-in slide-in-from-bottom-2 duration-300"
                >
                  <span className="text-lg" aria-hidden="true">
                    {getActivityIcon(item.type)}
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className={`${getActivityColor(item.type)}`}>
                      {item.message}
                    </p>
                    <p className="text-xs text-gray-400 mt-0.5">{time}</p>
                  </div>
                </div>
              );
            })}
            <div ref={feedEndRef} />
          </div>
        </div>
      </div>
    </div>
  );
}


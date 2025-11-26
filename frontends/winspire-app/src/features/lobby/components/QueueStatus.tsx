interface QueueStatusProps {
  status: string;
}

export function QueueStatus({ status }: QueueStatusProps) {
  const statusColors: Record<string, string> = {
    OPEN: 'bg-blue-100 text-blue-800',
    ACCEPTED: 'bg-green-100 text-green-800',
    PLAYING: 'bg-yellow-100 text-yellow-800',
    COMPLETE: 'bg-gray-100 text-gray-800',
  };

  return (
    <span
      className={`px-2 py-1 rounded text-xs font-medium ${
        statusColors[status] || 'bg-gray-100 text-gray-800'
      }`}
    >
      {status}
    </span>
  );
}




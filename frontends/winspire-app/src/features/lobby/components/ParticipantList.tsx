import { Avatar } from '../../../shared/components/ui/avatar';
import type { PreLobbyParticipant } from '../types';

interface ParticipantListProps {
  participants: PreLobbyParticipant[];
  currentUserId?: string;
}

/**
 * Display list of participants in tournament pre-lobby
 * Shows avatars, display names, and join timestamps
 */
export function ParticipantList({ participants, currentUserId }: ParticipantListProps) {
  if (participants.length === 0) {
    return (
      <div className="rounded-lg border border-gray-200 bg-gray-50 p-8 text-center">
        <p className="text-sm text-gray-500">Brak uczestników</p>
        <p className="mt-1 text-xs text-gray-400">Czekanie na graczy...</p>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-900">
          Uczestnicy ({participants.length})
        </h3>
      </div>

      <div className="rounded-lg border border-gray-200 bg-white">
        <ul className="divide-y divide-gray-100">
          {participants.map((participant) => {
            const isCurrentUser = participant.id === currentUserId;
            const joinTime = new Date(participant.joinedAt).toLocaleTimeString('pl-PL', {
              hour: '2-digit',
              minute: '2-digit',
            });

            return (
              <li
                key={participant.id}
                className={`flex items-center gap-3 p-3 ${
                  isCurrentUser ? 'bg-cyan-50' : 'hover:bg-gray-50'
                }`}
              >
                <Avatar
                  src={participant.avatarUrl || undefined}
                  alt={participant.displayName}
                  className="size-10"
                />
                
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {participant.displayName}
                    </p>
                    {isCurrentUser && (
                      <span className="inline-flex items-center rounded-full bg-cyan-100 px-2 py-0.5 text-xs font-medium text-cyan-800">
                        Ty
                      </span>
                    )}
                  </div>
                  <p className="text-xs text-gray-500">
                    Dołączył o {joinTime}
                  </p>
                </div>

                <div className="size-2 rounded-full bg-green-400" title="Online" />
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}


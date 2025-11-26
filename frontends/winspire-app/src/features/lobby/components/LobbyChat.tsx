import { useState, useEffect, useRef } from 'react';
import { useLobbyChat } from '../hooks/useLobbyChat';
import type { ChatMessage } from '../types';

interface LobbyChatProps {
  tournamentId: string;
}

export function LobbyChat({ tournamentId }: LobbyChatProps) {
  const { messages, sendMessage, isLoading } = useLobbyChat(tournamentId);
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim()) {
      sendMessage(input);
      setInput('');
    }
  };

  return (
    <div className="flex flex-col h-[600px] border rounded-lg">
      <div className="p-4 border-b">
        <h3 className="font-semibold">Chat</h3>
      </div>
      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {messages.map((message) => (
          <div key={message.id} className="text-sm">
            <span className="font-medium">{message.userName}:</span>{' '}
            <span>{message.message}</span>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>
      <form onSubmit={handleSubmit} className="p-4 border-t">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Type a message..."
          className="w-full px-3 py-2 border rounded-md"
          disabled={isLoading}
        />
      </form>
    </div>
  );
}




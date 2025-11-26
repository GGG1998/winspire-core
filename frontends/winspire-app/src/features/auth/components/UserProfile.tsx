import { useAuth } from '../hooks/useAuth';

export function UserProfile() {
  const { user } = useAuth();

  if (!user) {
    return null;
  }

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-bold">Profile</h2>
      <div className="space-y-2">
        <div>
          <label className="text-sm font-medium text-gray-500">Email</label>
          <p className="text-lg">{user.email}</p>
        </div>
        {user.name && (
          <div>
            <label className="text-sm font-medium text-gray-500">Name</label>
            <p className="text-lg">{user.name}</p>
          </div>
        )}
      </div>
    </div>
  );
}



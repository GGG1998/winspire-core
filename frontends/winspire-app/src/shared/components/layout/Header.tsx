import { Link } from 'react-router-dom';
import { useAuth } from '../../../features/auth';

export function Header() {
  const { user, isAuthenticated, logout } = useAuth();

  // Build tournament link based on user ID
  const tournamentsLink = user?.id ? `/h/${user.id}/tournaments` : '/';

  return (
    <header className="border-b border-gray-200 dark:border-gray-800">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <Link to="/" className="text-xl font-bold focus:outline-none focus:ring-2 focus:ring-blue-500 rounded">
            Winspire
          </Link>
          {isAuthenticated && (
            <nav className="hidden md:flex items-center gap-4">
              <Link
                to="/"
                className="text-sm text-gray-600 hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 rounded px-2 py-1"
              >
                Home
              </Link>
              {user?.profileType === 'streamer' && (
                <Link
                  to={tournamentsLink}
                  className="text-sm text-gray-600 hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 rounded px-2 py-1"
                >
                  My Tournaments
                </Link>
              )}
            </nav>
          )}
        </div>
        <nav className="flex items-center gap-2 sm:gap-4">
          {isAuthenticated ? (
            <>
              {user?.profileType === 'streamer' && (
                <Link
                  to={tournamentsLink}
                  className="md:hidden text-sm text-gray-600 hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 rounded px-2 py-1"
                >
                  Turnieje
                </Link>
              )}
              <Link
                to="/auth/profile"
                className="text-sm hover:underline focus:outline-none focus:ring-2 focus:ring-blue-500 rounded px-2 py-1"
              >
                <span className="hidden sm:inline">{user?.email}</span>
                <span className="sm:hidden">Profile</span>
              </Link>
              <button
                onClick={logout}
                className="text-sm hover:underline focus:outline-none focus:ring-2 focus:ring-blue-500 rounded px-2 py-1"
              >
                Logout
              </button>
            </>
          ) : (
            <>
              <Link
                to="/auth/login"
                className="text-sm hover:underline focus:outline-none focus:ring-2 focus:ring-blue-500 rounded px-2 py-1"
              >
                Login
              </Link>
              <Link
                to="/auth/register"
                className="text-sm hover:underline focus:outline-none focus:ring-2 focus:ring-blue-500 rounded px-2 py-1"
              >
                Register
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}


import { Navigate } from 'react-router-dom';
import { useAuth } from '../../../features/auth';
import { LoadingSpinner } from '../common/LoadingSpinner';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredProfileType?: 'user' | 'streamer';
}

export function ProtectedRoute({ children, requiredProfileType }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, user } = useAuth();

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (!isAuthenticated) {
    // Redirect to appropriate login route based on required profile type or user's current profile type
    const loginRoute = requiredProfileType === 'streamer' 
      ? '/auth/streamer/login'
      : requiredProfileType === 'user'
      ? '/auth/user/login'
      : user?.profileType === 'streamer'
      ? '/auth/streamer/login'
      : '/auth/user/login';
    
    return <Navigate to={loginRoute} replace />;
  }

  // Check if user has the required profile type
  if (requiredProfileType && user?.profileType !== requiredProfileType) {
    const loginRoute = requiredProfileType === 'streamer' 
      ? '/auth/streamer/login'
      : '/auth/user/login';
    return <Navigate to={loginRoute} replace />;
  }

  return <>{children}</>;
}


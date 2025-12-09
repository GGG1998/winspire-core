import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../../features/auth';
import { LoadingSpinner } from '../common/LoadingSpinner';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredProfileType?: 'user' | 'streamer';
}

export function ProtectedRoute({ children, requiredProfileType }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, user } = useAuth();
  const location = useLocation();

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

  // Check if profile is complete (nickname is required)
  // Don't redirect if already on the profile completion page to avoid redirect loops
  if (user && location.pathname !== '/auth/complete-profile') {
    if (!user.profile.nickname || user.profile.nickname.trim() === '') {
      return <Navigate to="/auth/complete-profile" replace />;
    }
  }

  return <>{children}</>;
}


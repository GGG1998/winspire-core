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

  console.log('[ProtectedRoute]', {
    pathname: location.pathname,
    isLoading,
    isAuthenticated,
    userExists: !!user,
    nickname: user?.profile?.nickname,
    profileType: user?.profileType,
    requiredProfileType
  });

  if (isLoading) {
    console.log('[ProtectedRoute] Still loading, showing spinner');
    return <LoadingSpinner />;
  }

  if (!isAuthenticated) {
    console.log('[ProtectedRoute] Not authenticated, redirecting to login');
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
    console.log('[ProtectedRoute] Wrong profile type, redirecting');
    const loginRoute = requiredProfileType === 'streamer' 
      ? '/auth/streamer/login'
      : '/auth/user/login';
    return <Navigate to={loginRoute} replace />;
  }

  // Check if profile is complete (nickname is required)
  // Don't redirect if already on the profile completion page to avoid redirect loops
  if (user && location.pathname !== '/auth/complete-profile') {
    if (!user.profile.nickname || user.profile.nickname.trim() === '') {
      console.log('[ProtectedRoute] ❌ Nickname missing, redirecting to /auth/complete-profile');
      return <Navigate to="/auth/complete-profile" replace />;
    }
  }

  console.log('[ProtectedRoute] ✓ All checks passed, rendering children');
  return <>{children}</>;
}


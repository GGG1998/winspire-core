import { useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { LoginForm } from '../components/LoginForm';
import { AuthLayout } from '../../../shared/components/ui/auth-layout';
import { Heading } from '../../../shared/components/ui/heading';
import { Text } from '../../../shared/components/ui/text';

export function StreamerLoginPage() {
  const { isAuthenticated, isLoading } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (isLoading) {
      return;
    }
    if (isAuthenticated) {
      navigate('/');
    }
  }, [isAuthenticated, isLoading, navigate]);

  return (
    <AuthLayout>
      <div className="space-y-8">
        <div className="space-y-2">
          <Heading>Sign in to your streamer account</Heading>
          <Text>
            Or{' '}
            <Link to="/auth/streamer/register" className="text-zinc-950 underline decoration-zinc-950/50 hover:decoration-zinc-950 dark:text-white dark:decoration-white/50 dark:hover:decoration-white">
              create a new streamer account
            </Link>
          </Text>
        </div>
        <LoginForm profileType="streamer" />
      </div>
    </AuthLayout>
  );
}


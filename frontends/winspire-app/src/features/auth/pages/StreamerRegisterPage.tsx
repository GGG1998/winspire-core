import { useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { StreamerRegisterForm } from '../components/StreamerRegisterForm';
import { AuthLayout } from '../../../shared/components/ui/auth-layout';
import { Heading } from '../../../shared/components/ui/heading';
import { Text } from '../../../shared/components/ui/text';

export function StreamerRegisterPage() {
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/');
    }
  }, [isAuthenticated, navigate]);

  return (
    <AuthLayout>
      <div className="space-y-8">
        <div className="space-y-2">
          <Heading>Create your streamer account</Heading>
          <Text>
            Already have an account?{' '}
            <Link to="/auth/streamer/login" className="text-zinc-950 underline decoration-zinc-950/50 hover:decoration-zinc-950 dark:text-white dark:decoration-white/50 dark:hover:decoration-white">
              Sign in
            </Link>
          </Text>
        </div>
        <StreamerRegisterForm />
      </div>
    </AuthLayout>
  );
}


import { useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { UserRegisterForm } from '../components/UserRegisterForm';
import { AuthLayout } from '../../../shared/components/ui/auth-layout';
import { Heading } from '../../../shared/components/ui/heading';
import { Text } from '../../../shared/components/ui/text';

export function UserRegisterPage() {
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
          <Heading>Create your account</Heading>
          <Text>
            Already have an account?{' '}
            <Link to="/auth/user/login" className="text-zinc-950 underline decoration-zinc-950/50 hover:decoration-zinc-950 dark:text-white dark:decoration-white/50 dark:hover:decoration-white">
              Sign in
            </Link>
          </Text>
        </div>
        <UserRegisterForm />
      </div>
    </AuthLayout>
  );
}


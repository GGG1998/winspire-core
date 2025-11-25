import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { useAuth } from '../hooks/useAuth';
import { Button } from '../../../shared/components/ui/button';
import { Input, InputGroup } from '../../../shared/components/ui/input';
import { Field, FieldGroup, Label, ErrorMessage as FieldErrorMessage } from '../../../shared/components/ui/fieldset';
import { Text } from '../../../shared/components/ui/text';
import type { LoginCredentials, UserProfileType } from '../types';

interface LoginFormProps {
  profileType: UserProfileType;
}

export function LoginForm({ profileType }: LoginFormProps) {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginCredentials>();

  const onSubmit = async (data: LoginCredentials) => {
    setError(null);
    setIsLoading(true);
    const result = await login(data, profileType);
    setIsLoading(false);
    if (result.success) {
      navigate('/');
    } else {
      setError(result.error || 'Invalid email or password');
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <FieldGroup>
        {error && (
          <div className="rounded-lg bg-red-50 p-4 ring-1 ring-red-200 dark:bg-red-900/20 dark:ring-red-800">
            <Text className="text-red-800 dark:text-red-200">{error}</Text>
          </div>
        )}

        <Field>
          <Label htmlFor="email">Email</Label>
          <InputGroup>
            <Input
              id="email"
              type="email"
              {...register('email', { required: 'Email is required' })}
              data-invalid={errors.email ? '' : undefined}
            />
          </InputGroup>
          {errors.email && (
            <FieldErrorMessage>{errors.email.message}</FieldErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="password">Password</Label>
          <InputGroup>
            <Input
              id="password"
              type="password"
              {...register('password', { required: 'Password is required' })}
              data-invalid={errors.password ? '' : undefined}
            />
          </InputGroup>
          {errors.password && (
            <FieldErrorMessage>{errors.password.message}</FieldErrorMessage>
          )}
        </Field>

        <div className="flex flex-col items-stretch gap-3">
          <Button type="submit" disabled={isLoading} color="dark/zinc">
            {isLoading ? 'Logging in...' : 'Sign in'}
          </Button>
        </div>
      </FieldGroup>
    </form>
  );
}


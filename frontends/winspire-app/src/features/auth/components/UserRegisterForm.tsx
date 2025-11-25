import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { useAuth } from '../hooks/useAuth';
import { Button } from '../../../shared/components/ui/button';
import { Input, InputGroup } from '../../../shared/components/ui/input';
import { Field, FieldGroup, Label, ErrorMessage as FieldErrorMessage } from '../../../shared/components/ui/fieldset';
import { Text } from '../../../shared/components/ui/text';
import type { UserRegisterData } from '../types';

export function UserRegisterForm() {
  const { register: registerUser } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
  } = useForm<UserRegisterData & { confirmPassword: string }>();

  const password = watch('password');

  const onSubmit = async (data: UserRegisterData & { confirmPassword: string }) => {
    setError(null);
    if (data.password !== data.confirmPassword) {
      setError('Passwords do not match');
      return;
    }
    setIsLoading(true);
    const { confirmPassword, ...registerData } = data;
    const result = await registerUser({ ...registerData, profileType: 'user' });
    setIsLoading(false);
    if (result.success) {
      navigate('/');
    } else {
      setError(result.error || 'Registration failed. Please try again.');
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
              {...register('email', { 
                required: 'Email is required',
                pattern: {
                  value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                  message: 'Invalid email address'
                }
              })}
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
              {...register('password', { 
                required: 'Password is required',
                minLength: { value: 6, message: 'Password must be at least 6 characters' }
              })}
              data-invalid={errors.password ? '' : undefined}
            />
          </InputGroup>
          {errors.password && (
            <FieldErrorMessage>{errors.password.message}</FieldErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="confirmPassword">Confirm Password</Label>
          <InputGroup>
            <Input
              id="confirmPassword"
              type="password"
              {...register('confirmPassword', { 
                required: 'Please confirm your password',
                validate: (value) => value === password || 'Passwords do not match'
              })}
              data-invalid={errors.confirmPassword ? '' : undefined}
            />
          </InputGroup>
          {errors.confirmPassword && (
            <FieldErrorMessage>{errors.confirmPassword.message}</FieldErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="first_name">First Name</Label>
          <InputGroup>
            <Input
              id="first_name"
              type="text"
              {...register('first_name', { required: 'First name is required' })}
              data-invalid={errors.first_name ? '' : undefined}
            />
          </InputGroup>
          {errors.first_name && (
            <FieldErrorMessage>{errors.first_name.message}</FieldErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="last_name">Last Name</Label>
          <InputGroup>
            <Input
              id="last_name"
              type="text"
              {...register('last_name', { required: 'Last name is required' })}
              data-invalid={errors.last_name ? '' : undefined}
            />
          </InputGroup>
          {errors.last_name && (
            <FieldErrorMessage>{errors.last_name.message}</FieldErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="nickname">Nickname</Label>
          <InputGroup>
            <Input
              id="nickname"
              type="text"
              {...register('nickname', { required: 'Nickname is required' })}
              data-invalid={errors.nickname ? '' : undefined}
            />
          </InputGroup>
          {errors.nickname && (
            <FieldErrorMessage>{errors.nickname.message}</FieldErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="street">Street (optional)</Label>
          <InputGroup>
            <Input
              id="street"
              type="text"
              {...register('street')}
            />
          </InputGroup>
        </Field>

        <Field>
          <Label htmlFor="city">City (optional)</Label>
          <InputGroup>
            <Input
              id="city"
              type="text"
              {...register('city')}
            />
          </InputGroup>
        </Field>

        <Field>
          <Label htmlFor="postal_code">Postal Code (optional)</Label>
          <InputGroup>
            <Input
              id="postal_code"
              type="text"
              {...register('postal_code')}
            />
          </InputGroup>
        </Field>

        <div className="flex flex-col items-stretch gap-3">
          <Button type="submit" disabled={isLoading} color="dark/zinc">
            {isLoading ? 'Registering...' : 'Create account'}
          </Button>
        </div>
      </FieldGroup>
    </form>
  );
}


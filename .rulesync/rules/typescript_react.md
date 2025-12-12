---
targets:
  - '*'
root: false
globs:
  - frontends/**/*.tsx
  - frontends/**/*.ts
cursor:
  alwaysApply: false
  globs:
    - frontends/**/*.tsx
    - frontends/**/*.ts
---
# TypeScript React Guidelines

## Struktura Folderów

Aplikacja frontendowa używa struktury feature-based:

```
frontends/<app-name>/src/
├── main.tsx                    # Entry point
├── App.tsx                     # Root component
├── index.css                   # Global styles
├── assets/                     # Static assets (images, icons)
├── features/                   # Feature modules
│   └── <feature>/
│       ├── api/                # API calls (react-query, fetch)
│       ├── components/         # Feature-specific components
│       ├── hooks/              # Feature-specific hooks
│       ├── pages/              # Page components (routes)
│       ├── layouts/            # Layout components
│       ├── context/            # React Context providers
│       ├── schemas.ts          # Zod validation schemas
│       ├── types.ts            # TypeScript types
│       ├── constants.ts        # Feature constants
│       └── index.ts            # Public exports
└── shared/
    ├── api/                    # Shared API clients
    ├── components/
    │   ├── common/             # Reusable components (EmptyState, ErrorBoundary)
    │   ├── layout/             # App layout components (Header, Footer)
    │   └── ui/                 # UI component library (Button, Input, etc.)
    ├── hooks/                  # Shared hooks
    ├── types/                  # Shared TypeScript types
    └── utils/                  # Utility functions
```

## Komponenty UI (@ui)

Używamy komponentów z `shared/components/ui/`. Import przez alias `@`:

```tsx
import { Button } from '@/shared/components/ui/button'
import { Input, InputGroup } from '@/shared/components/ui/input'
import { Field, FieldGroup, Label, ErrorMessage } from '@/shared/components/ui/fieldset'
import { Text } from '@/shared/components/ui/text'
import { Dialog, DialogPanel, DialogTitle } from '@/shared/components/ui/dialog'
import { Select } from '@/shared/components/ui/select'
```

### Dostępne komponenty UI

| Komponent | Plik | Użycie |
|-----------|------|--------|
| `Button` | button.tsx | Przyciski z wariantami (color, outline) |
| `Input`, `InputGroup` | input.tsx | Pola tekstowe |
| `Field`, `FieldGroup`, `Label`, `ErrorMessage` | fieldset.tsx | Wrapper dla pól formularza |
| `Select` | select.tsx | Dropdown select |
| `Checkbox` | checkbox.tsx | Checkbox z label |
| `Radio`, `RadioGroup`, `RadioField` | radio.tsx | Radio buttons |
| `Switch` | switch.tsx | Toggle switch |
| `Dialog`, `DialogPanel`, `DialogTitle` | dialog.tsx | Modal dialogi |
| `Table`, `TableHead`, `TableBody`, `TableRow`, `TableCell` | table.tsx | Tabele |
| `Badge` | badge.tsx | Status badges |
| `Avatar` | avatar.tsx | Awatary użytkowników |
| `Alert` | alert.tsx | Komunikaty alertów |
| `Dropdown`, `DropdownButton`, `DropdownMenu`, `DropdownItem` | dropdown.tsx | Dropdown menu |
| `Navbar` | navbar.tsx | Nawigacja |
| `Sidebar` | sidebar.tsx | Boczny panel |
| `Text`, `Strong`, `Code` | text.tsx | Stylizowany tekst |
| `Heading`, `Subheading` | heading.tsx | Nagłówki |
| `Link` | link.tsx | Linki |
| `Divider` | divider.tsx | Separator |
| `Pagination` | pagination.tsx | Paginacja |
| `Textarea` | textarea.tsx | Wieloliniowe pole tekstowe |
| `Combobox` | combobox.tsx | Autocomplete select |
| `Listbox` | listbox.tsx | Lista wyboru |
| `DescriptionList` | description-list.tsx | Lista opisów |

## Walidacja - Zod

**ZAWSZE** używaj Zod do definiowania schematów walidacji.

### Struktura schemas.ts

```typescript
import { z } from 'zod';

// === Schema Definitions ===

export const loginCredentialsSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Invalid email address'),
  password: z
    .string()
    .min(1, 'Password is required'),
});

export const createItemSchema = z.object({
  name: z
    .string()
    .min(3, 'Name must be at least 3 characters')
    .max(100, 'Name cannot exceed 100 characters'),
  description: z
    .string()
    .optional(),
  isActive: z
    .boolean()
    .default(true),
  category: z
    .enum(['OPTION_A', 'OPTION_B', 'OPTION_C'], {
      errorMap: () => ({ message: 'Please select a valid category' }),
    }),
});

// Schema z refine dla walidacji cross-field
export const registerSchema = z.object({
  email: z.string().email(),
  password: z.string().min(6),
  confirmPassword: z.string(),
}).refine(
  (data) => data.password === data.confirmPassword,
  {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  }
);

// === Type Inference ===
// ZAWSZE eksportuj typy z inferowanych schematów

export type LoginCredentialsInput = z.infer<typeof loginCredentialsSchema>;
export type CreateItemInput = z.infer<typeof createItemSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
```

### Zod Best Practices

| Metoda | Użycie |
|--------|--------|
| `.min(n, msg)` | Minimalna długość/wartość |
| `.max(n, msg)` | Maksymalna długość/wartość |
| `.email(msg)` | Walidacja emaila |
| `.url(msg)` | Walidacja URL |
| `.uuid(msg)` | Walidacja UUID |
| `.regex(pattern, msg)` | Custom regex |
| `.optional()` | Pole opcjonalne |
| `.nullable()` | Może być null |
| `.default(value)` | Wartość domyślna |
| `.enum([...])` | Ograniczone wartości |
| `.refine(fn, opts)` | Custom walidacja |
| `.transform(fn)` | Transformacja wartości |

## Formularze - React Hook Form

**ZAWSZE** używaj React Hook Form z zodResolver.

### Podstawowy formularz

```tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/shared/components/ui/button';
import { Input, InputGroup } from '@/shared/components/ui/input';
import { Field, FieldGroup, Label, ErrorMessage } from '@/shared/components/ui/fieldset';
import { createItemSchema, type CreateItemInput } from '../schemas';

export function CreateItemForm() {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<CreateItemInput>({
    resolver: zodResolver(createItemSchema),
    defaultValues: {
      name: '',
      isActive: true,
    },
  });

  const onSubmit = async (data: CreateItemInput) => {
    try {
      await api.createItem(data);
      reset();
    } catch (error) {
      console.error('Failed to create item:', error);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <FieldGroup>
        <Field>
          <Label htmlFor="name">Name</Label>
          <InputGroup>
            <Input
              id="name"
              {...register('name')}
              data-invalid={errors.name ? '' : undefined}
            />
          </InputGroup>
          {errors.name && (
            <ErrorMessage>{errors.name.message}</ErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="description">Description</Label>
          <InputGroup>
            <Input
              id="description"
              {...register('description')}
            />
          </InputGroup>
        </Field>

        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Creating...' : 'Create Item'}
        </Button>
      </FieldGroup>
    </form>
  );
}
```

### React Hook Form API

| Hook/Property | Użycie |
|---------------|--------|
| `register(name)` | Rejestracja pola |
| `handleSubmit(fn)` | Wrapper dla submit |
| `formState.errors` | Błędy walidacji |
| `formState.isSubmitting` | Stan wysyłania |
| `formState.isDirty` | Czy formularz zmieniony |
| `formState.isValid` | Czy formularz poprawny |
| `reset()` | Reset formularza |
| `setValue(name, value)` | Ustaw wartość programowo |
| `watch(name)` | Obserwuj wartość pola |
| `trigger(name)` | Wywołaj walidację |
| `setError(name, error)` | Ustaw błąd programowo |
| `clearErrors(name)` | Wyczyść błędy |

### Formularz z kontrolowanymi komponentami

```tsx
import { Controller, useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Select } from '@/shared/components/ui/select';

export function ControlledForm() {
  const { control, handleSubmit } = useForm({
    resolver: zodResolver(schema),
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Controller
        name="category"
        control={control}
        render={({ field, fieldState }) => (
          <Field>
            <Label>Category</Label>
            <Select
              {...field}
              data-invalid={fieldState.error ? '' : undefined}
            >
              <option value="">Select category</option>
              <option value="OPTION_A">Option A</option>
              <option value="OPTION_B">Option B</option>
            </Select>
            {fieldState.error && (
              <ErrorMessage>{fieldState.error.message}</ErrorMessage>
            )}
          </Field>
        )}
      />
    </form>
  );
}
```

## Typy - types.ts

```typescript
// === API Response Types ===
export interface Tournament {
  id: string;
  name: string;
  status: TournamentStatus;
  startTime: string;
  createdAt: string;
}

// === Enums as const ===
export const TOURNAMENT_STATUS = {
  DRAFT: 'DRAFT',
  PUBLISHED: 'PUBLISHED', 
  ACTIVE: 'ACTIVE',
  COMPLETED: 'COMPLETED',
} as const;

export type TournamentStatus = typeof TOURNAMENT_STATUS[keyof typeof TOURNAMENT_STATUS];

// === Component Props ===
export interface CreateTournamentModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (tournament: Tournament) => void;
}

// === Form Types ===
export interface TournamentFormData {
  name: string;
  startTime: string;
  game: string;
  maxPlayers: number;
}

export interface FormErrors {
  name?: string;
  startTime?: string;
  _general?: string;
}
```

## Stałe - constants.ts

```typescript
// === UI Labels ===
export const UI_LABELS = {
  page: {
    title: 'Tournaments',
    createButton: 'Create',
  },
  form: {
    nameLabel: 'Tournament Name',
    namePlaceholder: 'Enter tournament name',
    cancelButton: 'Cancel',
    createButton: 'Create Tournament',
  },
  loading: 'Loading...',
} as const;

// === Validation Rules ===
export const TOURNAMENT_VALIDATION = {
  name: {
    minLength: 3,
    maxLength: 100,
  },
} as const;

// === Validation Messages ===
export const VALIDATION_MESSAGES = {
  name: {
    required: 'Name is required',
    minLength: `Name must be at least ${TOURNAMENT_VALIDATION.name.minLength} characters`,
    maxLength: `Name cannot exceed ${TOURNAMENT_VALIDATION.name.maxLength} characters`,
  },
  general: {
    server: 'Server error. Please try again.',
  },
} as const;

// === Default Values ===
export const DEFAULT_GAME = 'Counter-Strike 2';
export const DEFAULT_MAX_PLAYERS = 16;
```

## Kluczowe Zasady

1. **Zod dla walidacji** - NIE używaj ręcznej walidacji, ZAWSZE zdefiniuj schema w Zod
2. **React Hook Form** - ZAWSZE z `zodResolver` dla integracji z Zod
3. **Komponenty UI z @ui** - Używaj istniejących komponentów, nie twórz własnych podstawowych
4. **TypeScript strict** - Używaj `z.infer<typeof schema>` dla typów formularzy
5. **Feature-based structure** - Każda funkcjonalność w osobnym folderze w `features/`
6. **Eksport przez index.ts** - Eksportuj publiczne API feature przez `index.ts`
7. **Stałe w constants.ts** - UI labels, validation rules, default values
8. **Typy w types.ts** - Interfaces, type aliases, enums
9. **Schematy w schemas.ts** - Wszystkie Zod schemas z exportowanymi typami

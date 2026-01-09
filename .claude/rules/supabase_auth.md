---
paths: 'frontends/**/*supabase*.ts, frontends/**/auth/**/*.ts, frontends/**/auth/**/*.tsx'
---
# Supabase Auth - Known Issues & Best Practices

## Navigator Locks Deadlock Issue

The Supabase JS client uses `navigator.locks` API internally, which can cause **deadlocks** during OAuth callbacks when multiple auth operations happen simultaneously.

**Symptoms:**
- OAuth callback page hangs indefinitely
- Console shows `#_acquireLock begin -1` without corresponding `end`
- `getSession()`, `getUser()`, `updateUser()`, or even `supabase.from().select()` never return

**Solution in `shared/api/supabase.ts`:**
```typescript
export const supabase = createClient(supabaseUrl, supabaseAnonKey, {
  auth: {
    // Disable navigator.locks to prevent deadlock issues
    lock: async (name: string, acquireTimeout: number, fn: () => Promise<any>) => {
      return await fn();
    },
  },
});
```

## OAuth Callback Best Practices

1. **Keep callback pages simple** - Only verify session exists, then redirect. Don't fetch profiles or call multiple Supabase methods.

2. **Read session from localStorage directly** when possible:
   ```typescript
   const storedSession = localStorage.getItem('winspire-auth');
   const session = JSON.parse(storedSession)?.session;
   ```

3. **Use direct REST API calls** to bypass Supabase client lock issues:
   ```typescript
   const response = await fetch(
     `${supabaseUrl}/rest/v1/user_profiles?id=eq.${userId}&select=*`,
     {
       headers: {
         'apikey': supabaseAnonKey,
         'Authorization': `Bearer ${accessToken}`,
       },
     }
   );
   ```

4. **Make non-critical auth calls fire-and-forget**:
   ```typescript
   // Instead of: await supabase.auth.updateUser({ data: {...} })
   supabase.auth.updateUser({ data: {...} })
     .then(({ error }) => error && console.warn(error))
     .catch(console.warn);
   ```

5. **Let AuthContext handle user loading** - The callback page should just redirect; AuthContext's `onAuthStateChange` will handle fetching user data.

## Debugging Auth Issues

1. Enable Supabase debug mode: `auth: { debug: true }`
2. Look for `#_acquireLock begin` without matching `end` in console
3. Check if multiple Supabase client calls are racing
4. Test with direct localStorage reads and REST API calls to isolate the issue

## Key Files
- `frontends/winspire-app/src/shared/api/supabase.ts` - Client configuration
- `frontends/winspire-app/src/features/auth/pages/OAuthCallbackPage.tsx` - OAuth callback handler
- `frontends/winspire-app/src/features/auth/api/supabaseAuth.ts` - Auth API functions
- `frontends/winspire-app/src/features/auth/context/AuthContext.tsx` - Auth state management

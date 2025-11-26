import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './features/auth/context/AuthContext';
import { ErrorBoundary } from './shared/components/common/ErrorBoundary';
import { ProtectedRoute } from './shared/components/layout/ProtectedRoute';
import { AppLayout } from './shared/components/layout/AppLayout';
import { LoadingSpinner } from './shared/components/common/LoadingSpinner';

// Lazy load pages for code splitting
const UserLoginPage = lazy(() => import('./features/auth/pages/UserLoginPage').then(m => ({ default: m.UserLoginPage })));
const UserRegisterPage = lazy(() => import('./features/auth/pages/UserRegisterPage').then(m => ({ default: m.UserRegisterPage })));
const StreamerLoginPage = lazy(() => import('./features/auth/pages/StreamerLoginPage').then(m => ({ default: m.StreamerLoginPage })));
const StreamerRegisterPage = lazy(() => import('./features/auth/pages/StreamerRegisterPage').then(m => ({ default: m.StreamerRegisterPage })));
// Legacy routes for backward compatibility
const LoginPage = lazy(() => import('./features/auth/pages/LoginPage').then(m => ({ default: m.LoginPage })));
const RegisterPage = lazy(() => import('./features/auth/pages/RegisterPage').then(m => ({ default: m.RegisterPage })));
const ProfilePage = lazy(() => import('./features/auth/pages/ProfilePage').then(m => ({ default: m.ProfilePage })));
const LobbyPage = lazy(() => import('./features/lobby/pages/LobbyPage').then(m => ({ default: m.LobbyPage })));
const MatchPage = lazy(() => import('./features/lobby/pages/MatchPage').then(m => ({ default: m.MatchPage })));

function App() {
  return (
    <AuthProvider>
      <ErrorBoundary>
        <BrowserRouter>
          <Suspense fallback={<LoadingSpinner />}>
            <Routes>
              {/* User auth routes */}
              <Route path="/auth/user/login" element={<UserLoginPage />} />
              <Route path="/auth/user/register" element={<UserRegisterPage />} />
              
              {/* Streamer auth routes */}
              <Route path="/auth/streamer/login" element={<StreamerLoginPage />} />
              <Route path="/auth/streamer/register" element={<StreamerRegisterPage />} />
              
              {/* Legacy routes - redirect to user routes for backward compatibility */}
              <Route path="/auth/login" element={<LoginPage />} />
              <Route path="/auth/register" element={<RegisterPage />} />
              
              <Route
                path="/auth/profile"
                element={
                  <ProtectedRoute>
                    <ProfilePage />
                  </ProtectedRoute>
                }
              />
              <Route
                path="/lobby/:tournamentId"
                element={
                  <ProtectedRoute>
                    <LobbyPage />
                  </ProtectedRoute>
                }
              />
              <Route
                path="/lobby/:tournamentId/match/:matchId"
                element={
                  <ProtectedRoute>
                    <MatchPage />
                  </ProtectedRoute>
                }
              />
              <Route
                path="/"
                element={
                  <AppLayout>
                    <div className="container mx-auto px-4 py-8">
                      <h1 className="text-2xl font-bold">Welcome to Winspire</h1>
                    </div>
                  </AppLayout>
                }
              />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Suspense>
        </BrowserRouter>
      </ErrorBoundary>
    </AuthProvider>
  );
}

export default App;

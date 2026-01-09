import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './features/auth/context/AuthContext';
import { TournamentsProvider } from './features/host/contexts/TournamentsContext';
import { ErrorBoundary } from './shared/components/common/ErrorBoundary';
import { ProtectedRoute } from './shared/components/layout/ProtectedRoute';
import { LoadingSpinner } from './shared/components/common/LoadingSpinner';

// Lazy load pages for code splitting
const UserLoginPage = lazy(() => import('./features/auth/pages/UserLoginPage').then(m => ({ default: m.UserLoginPage })));
const UserRegisterPage = lazy(() => import('./features/auth/pages/UserRegisterPage').then(m => ({ default: m.UserRegisterPage })));
const StreamerLoginPage = lazy(() => import('./features/auth/pages/StreamerLoginPage').then(m => ({ default: m.StreamerLoginPage })));
const StreamerRegisterPage = lazy(() => import('./features/auth/pages/StreamerRegisterPage').then(m => ({ default: m.StreamerRegisterPage })));
const OAuthCallbackPage = lazy(() => import('./features/auth/pages/OAuthCallbackPage').then(m => ({ default: m.OAuthCallbackPage })));
const ProfileCompletionPage = lazy(() => import('./features/auth/pages/ProfileCompletionPage').then(m => ({ default: m.ProfileCompletionPage })));
// Legacy routes for backward compatibility
const LoginPage = lazy(() => import('./features/auth/pages/LoginPage').then(m => ({ default: m.LoginPage })));
const RegisterPage = lazy(() => import('./features/auth/pages/RegisterPage').then(m => ({ default: m.RegisterPage })));
const ProfilePage = lazy(() => import('./features/auth/pages/ProfilePage').then(m => ({ default: m.ProfilePage })));
const MatchPage = lazy(() => import('./features/lobby/pages/MatchPage').then(m => ({ default: m.MatchPage })));
const TournamentPreLobbyPage = lazy(() => import('./features/lobby/pages/TournamentPreLobbyPage').then(m => ({ default: m.TournamentPreLobbyPage })));
const MatchLobbyPage = lazy(() => import('./features/lobby/pages/MatchLobbyPage').then(m => ({ default: m.MatchLobbyPage })));
// Tournament pages
const TournamentPage = lazy(() => import('./features/host/pages/TournamentPage').then(m => ({ default: m.TournamentPage })));
const TournamentDetailPage = lazy(() => import('./features/host/pages/TournamentDetailPage').then(m => ({ default: m.TournamentDetailPage })));

function App() {
  return (
    <AuthProvider>
      <ErrorBoundary>
        <BrowserRouter>
          <Suspense fallback={<LoadingSpinner />}>
            <Routes>
              {/* OAuth callback route */}
              <Route path="/auth/callback" element={<OAuthCallbackPage />} />
              
              {/* Profile completion route for OAuth users */}
              <Route path="/auth/complete-profile" element={<ProfileCompletionPage />} />
              
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
              {/* Tournament Pre-Lobby (waiting room before bracket generation) */}
              <Route
                path="/tournaments/:tournamentId/lobby"
                element={
                  <ProtectedRoute>
                    <TournamentPreLobbyPage />
                  </ProtectedRoute>
                }
              />
              {/* Match Lobby (player VS opponent, ready button, game iframe) */}
              <Route
                path="/lobby/:tournamentId/match/:matchId"
                element={
                  <ProtectedRoute>
                    <MatchLobbyPage />
                  </ProtectedRoute>
                }
              />
              {/* Match Lobby (per-match room after bracket generation) */}
              <Route
                path="/lobby/:tournamentId/match/:matchId"
                element={
                  <ProtectedRoute>
                    <MatchPage />
                  </ProtectedRoute>
                }
              />
              {/* Host/Streamer tournament routes - use HostLayout internally */}
              {/* Only streamers can access these routes */}
              <Route
                path="/h/:streamerId/tournaments"
                element={
                  <ProtectedRoute requiredProfileType="streamer">
                    <TournamentsProvider>
                      <TournamentPage />
                    </TournamentsProvider>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/h/:streamerId/tournaments/:tournamentId"
                element={
                  <ProtectedRoute>
                    <TournamentsProvider>
                      <TournamentDetailPage />
                    </TournamentsProvider>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/"
                element={
                  <ProtectedRoute>
                    <Navigate to="/auth/profile" replace />
                  </ProtectedRoute>
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

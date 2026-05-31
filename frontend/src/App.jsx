import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import UploadPage from './pages/UploadPage';
import LoginPage from './pages/LoginPage';
import DocumentsPage from './pages/DocumentsPage';
import SearchPage from './pages/SearchPage';
import DocumentPage from './pages/DocumentPage';

function PublicRoute({ children }) {
  const { isAuthenticated, loading } = useAuth();
  if (loading) return <div style={{ padding: '2rem', textAlign: 'center' }}>Загрузка...</div>;
  if (isAuthenticated) return <Navigate to="/" replace />;
  return children;
}

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          {/* Публичные маршруты (доступны всем) */}
          <Route path="/login" element={
            <PublicRoute><LoginPage /></PublicRoute>
          } />
          <Route path="/search" element={<SearchPage />} />
          
          {/* Защищённые маршруты (только после входа) */}
          <Route path="/" element={
            <ProtectedRoute><DocumentsPage /></ProtectedRoute>
          } />
          <Route path="/document/:id" element={
            <ProtectedRoute><DocumentPage /></ProtectedRoute>
          } />
          <Route path="/upload" element={
            <ProtectedRoute><UploadPage /></ProtectedRoute>
          } />
          
          {/* Редирект с корня: если не авторизован → /search, иначе → / */}
          <Route path="*" element={<RootRedirect />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

// Компонент редиректа с корня
function RootRedirect() {
  const { isAuthenticated, loading } = useAuth();
  if (loading) return <div style={{ padding: '2rem', textAlign: 'center' }}>Загрузка...</div>;
  return <Navigate to={isAuthenticated ? "/" : "/search"} replace />;
}

export default App;
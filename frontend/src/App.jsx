// frontend/src/App.jsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';

// 🔹 Импортируем ВСЕ существующие файлы
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';  // ✅ Новый файл
import HomePage from './pages/HomePage';          // ✅ Новый файл (главная)
import DocumentsPage from './pages/DocumentsPage'; // Личные документы
import SearchPage from './pages/SearchPage';       // Старая страница (можно удалить позже)
import DocumentPage from './pages/DocumentPage';
import UploadPage from './pages/UploadPage';

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          {/* 🔹 Публичные маршруты */}
          <Route path="/" element={<HomePage />} />           {/* Главная = публичные документы */}
          <Route path="/search" element={<HomePage />} />     /* Дублирует главную для совместимости */
          <Route path="/login" element={
            <PublicRoute><LoginPage /></PublicRoute>
          } />
          <Route path="/register" element={
            <PublicRoute><RegisterPage /></PublicRoute>
          } />
          <Route path="/document/:id" element={<DocumentPage />} />
          
          {/* 🔹 Защищённые маршруты (только после входа) */}
          <Route path="/dashboard" element={
            <ProtectedRoute><DocumentsPage /></ProtectedRoute>
          } />
          <Route path="/upload" element={
            <ProtectedRoute><UploadPage /></ProtectedRoute>
          } />
          
          {/* 🔹 Редиректы */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

// Компонент для публичных маршрутов: если авторизован → редирект на /dashboard
function PublicRoute({ children }) {
  const { isAuthenticated, loading } = useAuth();
  if (loading) return <div className="min-h-screen flex items-center justify-center">Загрузка...</div>;
  if (isAuthenticated) return <Navigate to="/dashboard" replace />;
  return children;
}

export default App;
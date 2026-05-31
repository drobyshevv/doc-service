// frontend/src/components/Header.jsx
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Header() {
  const { user, logout, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  return (
    <header className="sticky top-0 z-50 bg-white border-b border-slate-200">
      <div className="container flex items-center justify-between py-3">
        {/* Логотип */}
        <Link to="/" className="flex items-center gap-2 group">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white font-bold text-sm shadow-sm">
            D
          </div>
          <span className="font-semibold text-lg text-slate-800 group-hover:text-indigo-600 transition-colors">
            DocService
          </span>
        </Link>

        {/* Поиск (показывается на всех страницах) */}
        <div className="flex-1 max-w-md mx-4 hidden md:block">
          <div className="relative">
            <input
              type="search"
              placeholder="Поиск документов..."
              className="input pl-10"
              onKeyDown={(e) => e.key === 'Enter' && navigate(`/search?q=${e.target.value}`)}
            />
            <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
        </div>

        {/* Кнопки авторизации */}
        <div className="flex items-center gap-2">
          {isAuthenticated ? (
            <>
              <Link to="/upload" className="btn btn-primary btn-sm hide-mobile">
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                Загрузить
              </Link>
              <div className="flex items-center gap-2 ml-2">
                <span className="text-sm text-slate-600 hide-mobile">{user?.email}</span>
                <button onClick={logout} className="btn btn-outline btn-sm">
                  Выйти
                </button>
              </div>
            </>
          ) : (
            <>
              <Link to="/login" className="btn btn-outline btn-sm">
                Войти
              </Link>
              <Link to="/register" className="btn btn-primary btn-sm">
                Регистрация
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
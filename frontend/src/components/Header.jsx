// frontend/src/components/Header.jsx
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Header() {
  const { user, logout, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  return (
    <header className="header">
      <div className="header__inner">
        <Link to="/" className="logo-link">
          <div className="logo">D</div>
          <span>DocService</span>
        </Link>

        <div className="flex-1 max-w-md mx-4 hidden md:block">
          <div className="relative">
            <input
              type="search"
              placeholder="Поиск документов..."
              className="input"
              style={{ paddingLeft: '2.5rem' }}
              onKeyDown={(e) => e.key === 'Enter' && navigate(`/search?q=${e.target.value}`)}
            />
            <svg 
              className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4"
              style={{ color: 'var(--color-text-muted)' }}
              fill="none" 
              stroke="currentColor" 
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
        </div>

        <div className="header__actions">
          {isAuthenticated ? (
            <>
              <Link to="/upload" className="btn btn-primary hide-mobile">
                <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                Загрузить
              </Link>
              <span className="text-sm text-muted hide-mobile">{user?.email}</span>
              <button onClick={logout} className="btn btn-outline">
                Выйти
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="btn btn-outline">
                Войти
              </Link>
              <Link to="/register" className="btn btn-primary">
                Регистрация
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
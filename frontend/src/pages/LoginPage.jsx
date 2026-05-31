// frontend/src/pages/LoginPage.jsx
import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    
    try {
      await login(email, password);
      navigate('/');
    } catch (err) {
      setError(err.message || 'Ошибка входа');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="page-auth">
      <div className="page-auth__card">
        {/* Логотип */}
        <div className="page-auth__logo">
          <Link to="/" className="logo-link">
            <div className="logo">D</div>
            <span>DocService</span>
          </Link>
        </div>

        {/* Заголовок */}
        <div className="page-auth__header">
          <h1 className="page-auth__title">Вход в аккаунт</h1>
          <p className="page-auth__subtitle">Введите данные для доступа к документам</p>
        </div>

        {/* Ошибка */}
        {error && (
          <div className="page-auth__error" role="alert">
            {error}
          </div>
        )}

        {/* Форма */}
        <form onSubmit={handleSubmit} className="page-auth__form">
          <div className="form-group">
            <label htmlFor="email" className="form-label">Email</label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="input"
              placeholder="you@example.com"
              required
              disabled={submitting}
              autoComplete="email"
            />
          </div>

          <div className="form-group">
            <label htmlFor="password" className="form-label">Пароль</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="input"
              placeholder="••••••••"
              required
              disabled={submitting}
              autoComplete="current-password"
            />
          </div>

          <button
            type="submit"
            disabled={submitting}
            className="btn btn-primary btn--full"
          >
            {submitting ? 'Вход...' : 'Войти'}
          </button>
        </form>

        {/* Ссылки */}
        <div className="page-auth__footer">
          <p className="page-auth__text">
            Нет аккаунта?{' '}
            <Link to="/register" className="link-primary">Зарегистрироваться</Link>
          </p>
          <p className="page-auth__text">
            <Link to="/search" className="link-muted">
              🔍 Искать публичные документы без входа
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
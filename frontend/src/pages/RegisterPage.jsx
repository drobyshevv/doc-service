// frontend/src/pages/RegisterPage.jsx
import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { authAPI } from '../services/api';

// ============================================================================
// ХЕЛПЕР: валидация пароля
// ============================================================================
const validatePassword = (password) => {
  if (password.length < 6) {
    return 'Пароль должен быть не менее 6 символов';
  }
  return null;
};

// ============================================================================
// ОСНОВНОЙ КОМПОНЕНТ
// ============================================================================
export default function RegisterPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const navigate = useNavigate();

  // Обработка отправки формы
  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    // Валидация на клиенте
    if (password !== confirmPassword) {
      setError('Пароли не совпадают');
      return;
    }

    const passwordError = validatePassword(password);
    if (passwordError) {
      setError(passwordError);
      return;
    }

    setSubmitting(true);

    try {
      // Вызов реального эндпоинта регистрации
      const response = await authAPI.register(email, password);
      
      // Сохраняем данные пользователя и токен (автоматический вход)
      const { user, tokens } = response.data;
      localStorage.setItem('token', tokens.access_token);
      localStorage.setItem('user', JSON.stringify(user));
      
      // Редирект на главную страницу
      navigate('/');
      
    } catch (err) {
      // Обработка ошибок
      console.error('Register error:', err);
      
      const status = err.response?.status;
      const message = err.response?.data || err.message;
      
      if (status === 409) {
        setError('Пользователь с таким email уже существует');
      } else if (status === 400) {
        setError('Некорректные данные');
      } else {
        setError(message || 'Не удалось создать аккаунт');
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="page-auth">
      <div className="page-auth__card">
        
        {/* 🔹 Логотип */}
        <div className="page-auth__logo">
          <Link to="/" className="logo-link">
            <div className="logo">D</div>
            <span>DocService</span>
          </Link>
        </div>

        {/* Заголовок */}
        <div className="page-auth__header">
          <h1 className="page-auth__title">Создать аккаунт</h1>
          <p className="page-auth__subtitle">Начните хранить документы безопасно</p>
        </div>

        {/* Сообщение об ошибке */}
        {error && (
          <div className="page-auth__error" role="alert">
            {error}
          </div>
        )}

        {/* Форма регистрации */}
        <form onSubmit={handleSubmit} className="page-auth__form" noValidate>
          
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
              aria-required="true"
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
              placeholder="Минимум 6 символов"
              required
              minLength={6}
              disabled={submitting}
              autoComplete="new-password"
              aria-required="true"
            />
          </div>

          <div className="form-group">
            <label htmlFor="confirmPassword" className="form-label">Подтвердите пароль</label>
            <input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="input"
              placeholder="Повторите пароль"
              required
              disabled={submitting}
              autoComplete="new-password"
              aria-required="true"
            />
          </div>

          <button
            type="submit"
            disabled={submitting}
            className="btn btn-primary btn--full"
            aria-busy={submitting}
          >
            {submitting ? 'Создание...' : 'Зарегистрироваться'}
          </button>
          
        </form>

        {/* Ссылки */}
        <div className="page-auth__footer">
          <p className="page-auth__text">
            Уже есть аккаунт?{' '}
            <Link to="/login" className="link-primary">Войти</Link>
          </p>
        </div>
        
      </div>
    </div>
  );
}
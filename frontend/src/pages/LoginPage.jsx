// frontend/src/pages/LoginPage.jsx
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
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
    
    console.log('🔐 Login attempt:', { email });
    
    try {
      const userData = await login(email, password);
      console.log('Login success, user:', userData);
      
      window.location.href = '/';
    } catch (err) {
      console.error('Login error:', err);
      setError(err.message || err.response?.data || 'Ошибка входа');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div style={{ maxWidth: 400, margin: '3rem auto', padding: '1.5rem' }}>
      <p style={{ textAlign: 'center', marginBottom: '1rem' }}>
        <a href="/search" style={{ color: '#666' }}>🔍 Искать публичные документы без входа</a>
      </p>
      
      <h2>🔐 Вход</h2>
      
      {error && (
        <p style={{ color: 'red', background: '#fee', padding: '0.5rem', borderRadius: '4px', marginBottom: '1rem' }}>
          {error}
        </p>
      )}
      
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: '1rem' }}>
          <label>Email<br />
            <input 
              type="email" 
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
              disabled={submitting}
              style={{ width: '100%', padding: '0.5rem', boxSizing: 'border-box' }}
            />
          </label>
        </div>
        <div style={{ marginBottom: '1rem' }}>
          <label>Пароль<br />
            <input 
              type="password" 
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
              disabled={submitting}
              style={{ width: '100%', padding: '0.5rem', boxSizing: 'border-box' }}
            />
          </label>
        </div>
        <button 
          type="submit" 
          disabled={submitting}
          style={{ 
            padding: '0.5rem 1rem', 
            opacity: submitting ? 0.6 : 1,
            cursor: submitting ? 'not-allowed' : 'pointer'
          }}
        >
          {submitting ? 'Вход...' : 'Войти'}
        </button>
      </form>
    </div>
  );
}
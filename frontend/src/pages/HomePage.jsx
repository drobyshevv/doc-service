// frontend/src/pages/HomePage.jsx
import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import DocumentCard from '../components/DocumentCard';

export default function HomePage() {
  const { isAuthenticated, user, logout } = useAuth();
  const [docs, setDocs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');
  const [searchMode, setSearchMode] = useState('content');

  const loadDocs = async () => {
    setLoading(true);
    try {
      const res = await fetch('/public/documents?limit=50', {
        headers: { 'Authorization': 'Bearer ' + localStorage.getItem('token') }
      });
      if (!res.ok) throw new Error('HTTP ' + res.status);
      const data = await res.json();
      setDocs(data.documents || []);
    } catch (err) {
      console.error('Load error:', err);
      setDocs([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadDocs(); }, []);

  const handleSearch = async () => {
    if (!query.trim()) return loadDocs();
    setLoading(true);
    try {
      const endpoint = searchMode === 'title'
        ? `/search/title?query=${encodeURIComponent(query)}&limit=50`
        : `/search/?query=${encodeURIComponent(query)}&limit=50`;

      const token = localStorage.getItem('token');
      const res = await fetch(endpoint, {
        headers: { 'Authorization': token ? `Bearer ${token}` : '' }
      });

      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const raw = await res.json();

      const extracted = Array.isArray(raw)
        ? raw.map(r => r.document || r.Document || r).filter(d => d?.ID || d?.id)
        : [];
      const unique = [...new Map(extracted.map(d => [d.ID || d.id, d])).values()];
      setDocs(unique);
    } catch (err) {
      console.error('Search error:', err);
      loadDocs();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="page--dashboard">
      <header className="header">
        <div className="header__inner">
          <Link to="/" className="logo-link">
            <div className="logo">D</div>
            <span>DocService</span>
          </Link>

          <div className="header__actions">
            {isAuthenticated && (
              <Link to="/dashboard" className="btn btn-outline">
                <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                </svg>
                Мои документы
              </Link>
            )}

            {isAuthenticated ? (
              <>
                <span className="text-sm text-muted hide-mobile">{user?.email}</span>
                <button onClick={logout} className="btn btn-outline">Выйти</button>
              </>
            ) : (
              <>
                <Link to="/login" className="btn btn-outline">Войти</Link>
                <Link to="/register" className="btn btn-primary">Регистрация</Link>
              </>
            )}
          </div>
        </div>
      </header>

      <section style={{ background: 'var(--color-bg-secondary)', padding: '4rem 1rem', textAlign: 'center' }}>
        <div className="container">
          <h1 style={{ fontSize: '2.5rem', fontWeight: 700, marginBottom: '1rem', letterSpacing: '-0.02em', color: 'var(--color-text)' }}>
            Найдите и поделитесь документами
          </h1>
          <p style={{ fontSize: '1.125rem', color: 'var(--color-text-secondary)', marginBottom: '2rem' }}>
            Безопасное хранение, версионирование и поиск по содержимому
          </p>
          <div style={{ display: 'flex', gap: '0.5rem', maxWidth: 600, margin: '0 auto', alignItems: 'center' }}>
            <select
              value={searchMode}
              onChange={(e) => setSearchMode(e.target.value)}
              className="input"
              style={{ width: 'auto', minWidth: '140px' }}
            >
              <option value="content">По тексту</option>
              <option value="title">По заголовку</option>
            </select>

            <input
              type="search"
              placeholder={searchMode === 'title' ? "Название или имя файла..." : "Поиск по содержимому..."}
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              className="input"
              style={{ flex: 1 }}
            />
            <button
              onClick={handleSearch}
              className="btn btn-primary"
            >
              Найти
            </button>
          </div>
        </div>
      </section>

      <main className="container" style={{ padding: '3rem 2rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.5rem' }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 600, color: 'var(--color-text)' }}>
            {query ? `Результаты: "${query}"` : 'Публичные документы'}
          </h2>
          <span style={{ fontSize: 14, color: 'var(--color-text-secondary)' }}>{docs.length} документов</span>
        </div>

        {loading ? (
          <div className="doc-grid">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="file-card file-card--loading">
                <div className="file-card__paper">
                  <div className="file-card__extension" style={{ opacity: 0.3 }}>...</div>
                </div>
                <div className="file-card__info">
                  <div style={{ height: 14, background: 'var(--color-bg-secondary)', borderRadius: 4, width: '75%', marginBottom: 6 }} />
                  <div style={{ height: 12, background: 'var(--color-bg-secondary)', borderRadius: 4, width: '50%' }} />
                </div>
              </div>
            ))}
          </div>
        ) : docs.length > 0 ? (
          <div className="doc-grid">
            {docs.map(doc => (
              <DocumentCard key={doc.ID || doc.id} doc={doc} />
            ))}
          </div>
        ) : (
          <div className="empty-state">
            <div className="empty-state__icon">
              <svg style={{ width: 48, height: 48 }} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <h3 className="empty-state__title">
              {query ? 'Ничего не найдено' : 'Публичных документов пока нет'}
            </h3>
            <p className="empty-state__text">
              {query ? 'Попробуйте другой запрос' : 'Зарегистрируйтесь и начните делиться файлами'}
            </p>
            {!isAuthenticated && (
              <Link to="/register" className="btn btn-primary">Создать аккаунт</Link>
            )}
          </div>
        )}
      </main>
    </div>
  );
}
// frontend/src/pages/SearchPage.jsx
import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import Header from '../components/Header';
import DocumentCard from '../components/DocumentCard';
import { searchAPI } from '../services/api';

export default function SearchPage() {
  const { isAuthenticated } = useAuth();
  const [docs, setDocs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');

  const loadPublicDocs = async () => {
    setLoading(true);
    try {
      const res = await fetch('/public/documents?limit=50', {
        headers: { 'Authorization': 'Bearer ' + localStorage.getItem('token') }
      });
      const data = await res.json();
      setDocs(Array.isArray(data.documents) ? data.documents : []);
    } catch (err) {
      console.error('Load error:', err);
      setDocs([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadPublicDocs(); }, []);

  const handleSearch = async () => {
    if (!query.trim()) return loadPublicDocs();
    setLoading(true);
    try {
      const res = await searchAPI.search(query, { limit: 50 });
      const raw = Array.isArray(res.data) ? res.data : [];
      const extracted = raw.map(r => r.document || r.Document || r).filter(d => d?.ID || d?.id);
      const unique = [...new Map(extracted.map(d => [d.ID || d.id, d])).values()];
      setDocs(unique);
    } catch {
      loadPublicDocs();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="page--dashboard">
      <Header />

      <section style={{ background: 'var(--color-bg-secondary)', padding: '4rem 1rem', textAlign: 'center' }}>
        <div className="container">
          <h1 style={{ fontSize: '2.5rem', fontWeight: 700, marginBottom: '1rem', letterSpacing: '-0.02em', color: 'var(--color-text)' }}>
            Найдите и поделитесь документами
          </h1>
          <p style={{ fontSize: '1.125rem', color: 'var(--color-text-secondary)', marginBottom: '2rem' }}>
            Безопасное хранение, версионирование и поиск по содержимому
          </p>

          <div style={{ display: 'flex', gap: '0.5rem', maxWidth: 600, margin: '0 auto' }}>
            <input
              type="search"
              placeholder="Поиск по названию или содержимому..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              className="input"
              style={{ flex: 1 }}
            />
            <button onClick={handleSearch} className="btn btn-primary">
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
          <span style={{ fontSize: 14, color: 'var(--color-text-secondary)' }}>
            {docs.length} {docs.length === 1 ? 'документ' : 'документов'}
          </span>
        </div>

        {loading ? (
          <div className="doc-grid">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="doc-card doc-card--loading">
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                  <div className="doc-card__icon--skeleton" />
                  <div style={{ flex: 1 }}>
                    <div className="doc-card__title--skeleton" />
                    <div className="badge--skeleton" />
                  </div>
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

      <footer style={{ borderTop: '1px solid var(--color-border)', padding: '1.5rem 0', marginTop: '3rem' }}>
        <div className="container text-center text-muted" style={{ fontSize: 14 }}>
          © 2026 DocService. Все права защищены.
        </div>
      </footer>
    </div>
  );
}
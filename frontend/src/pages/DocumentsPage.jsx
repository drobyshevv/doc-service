// frontend/src/pages/DocumentsPage.jsx
import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { documentsAPI } from '../services/api';
import DocumentCard from '../components/DocumentCard';

// ХЕЛПЕРЫ
const getField = (obj, ...names) => {
  for (const name of names) {
    if (obj?.[name] !== undefined && obj?.[name] !== null) return obj[name];
  }
  return null;
};

// ПОД-КОМПОНЕНТЫ

function EmptyState({ onUpload }) {
  return (
    <div className="empty-state">
      <div className="empty-state__icon">
        <svg style={{ width: 48, height: 48 }} fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
      </div>
      <h3 className="empty-state__title">Нет документов</h3>
      <p className="empty-state__text">Загрузите первый файл, чтобы начать</p>
      <button onClick={onUpload} className="btn btn-primary">
        Загрузить документ
      </button>
    </div>
  );
}

function LoadingGrid({ count = 6 }) {
  return (
    <div className="doc-grid">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="file-card file-card--loading">
          <div className="file-card__paper">
            <div className="file-card__extension" style={{ opacity: 0.3 }}>...</div>
          </div>
          <div className="file-card__info">
            <div style={{ height: 16, background: 'var(--color-bg-secondary)', borderRadius: 4, width: '75%', marginBottom: 8 }} />
            <div style={{ height: 12, background: 'var(--color-bg-secondary)', borderRadius: 4, width: '50%' }} />
          </div>
        </div>
      ))}
    </div>
  );
}

// ОСНОВНОЙ КОМПОНЕНТ

export default function DocumentsPage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  
  const [docs, setDocs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchMode, setSearchMode] = useState('content');

  const loadDocs = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const userId = JSON.parse(localStorage.getItem('user') || '{}')?.id;
      
      const res = await fetch('/documents?limit=50', {
        headers: {
          'Authorization': token ? `Bearer ${token}` : '',
          'X-User-ID': userId || ''
        },
        cache: 'no-store'
      });
      
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      
      const data = await res.json();
      const docs = Array.isArray(data.documents) ? data.documents : [];
      setDocs(docs);
      
    } catch (err) {
      console.error('Load docs error:', err);
      setDocs([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadDocs(); }, []);

  const handleSearch = async () => {
    if (!searchQuery.trim()) return loadDocs();
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const userId = JSON.parse(localStorage.getItem('user') || '{}')?.id;
      
      const endpoint = searchMode === 'title'
        ? `/search/title?query=${encodeURIComponent(searchQuery)}&owner_id=${userId}&limit=50`
        : `/search/owner?query=${encodeURIComponent(searchQuery)}&owner_id=${userId}&limit=50`;
    
      const res = await fetch(endpoint, {
        headers: {
          'Authorization': token ? `Bearer ${token}` : '',
          'X-User-ID': userId
        }
      });
      
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const raw = await res.json();
      
      const extracted = Array.isArray(raw) 
        ? raw.map(r => getField(r, 'document', 'Document', 'doc') || r).filter(Boolean)
        : [];
      const unique = [...new Map(extracted.map(d => [getField(d, 'ID', 'id'), d])).values()];
      setDocs(unique);
      
    } catch (err) {
      console.error('Search error:', err);
      loadDocs();
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await documentsAPI.delete(id);
      setDocs(docs.filter(d => getField(d, 'ID', 'id') !== id));
    } catch (err) {
      console.error('Delete error:', err);
      alert('Не удалось удалить документ');
    }
  };

  return (
    <div className="page page--dashboard">
      <header className="header">
        <div className="container header__inner">
          <Link to="/" className="logo-link">
            <div className="logo">D</div>
            <span>DocService</span>
          </Link>
          
          <div className="header__actions">
            <span className="header__user hide-mobile">{user?.email}</span>
            <button onClick={logout} className="btn btn-outline btn-sm">Выйти</button>
            <button onClick={() => navigate('/upload')} className="btn btn-primary btn-sm">
              <svg className="icon icon--sm" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              <span className="hide-mobile">Загрузить</span>
            </button>
          </div>
        </div>
      </header>

      <main className="container page__content">
        <div className="page__header">
          <div>
            <h1 className="page__title">Мои документы</h1>
            <p className="page__subtitle">Управляйте своими файлами и версиями</p>
          </div>
          <nav className="page__nav">
            <Link to="/" className="btn btn-outline btn-sm">Публичные</Link>
            <Link to="/search" className="btn btn-outline btn-sm">Поиск</Link>
          </nav>
        </div>

        <div className="search-bar">
          <select 
            value={searchMode}
            onChange={(e) => setSearchMode(e.target.value)}
            className="input"
            style={{ width: 'auto', padding: '0.5rem' }}
          >
            <option value="content">По тексту</option>
            <option value="title">По заголовку</option>
          </select>
          
          <input 
            type="search" 
            placeholder={searchMode === 'title' ? "Название или имя файла..." : "Поиск по содержимому..."}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            className="input"
            aria-label="Поиск документов"
          />
          <button onClick={handleSearch} className="btn btn-primary">Найти</button>
        </div>

        {loading ? (
          <LoadingGrid />
        ) : docs.length > 0 ? (
          <div className="doc-grid">
            {docs.map(doc => (
              <DocumentCard 
                key={getField(doc, 'ID', 'id')} 
                doc={doc} 
                showActions={true}
                onDelete={() => handleDelete(getField(doc, 'ID', 'id'))}
              />
            ))}
          </div>
        ) : (
          <EmptyState onUpload={() => navigate('/upload')} />
        )}
      </main>
    </div>
  );
}
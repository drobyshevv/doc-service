import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { documentsAPI, searchAPI } from '../services/api';

const getField = (obj, ...names) => {
  for (const name of names) {
    if (obj?.[name] !== undefined && obj?.[name] !== null) return obj[name];
  }
  return null;
};
const formatDate = (val) => {
  if (!val) return '—';
  const d = new Date(val);
  return isNaN(d.getTime()) ? '—' : d.toLocaleDateString('ru-RU');
};
const formatSize = (b) => {
  if (!b && b !== 0) return '—';
  const k = b/1024;
  return k < 1024 ? `${k.toFixed(1)} KB` : `${(k/1024).toFixed(1)} MB`;
};

export default function SearchPage() {
  const { user, logout } = useAuth();
  const [docs, setDocs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');
  const [error, setError] = useState('');
  const [debug, setDebug] = useState('');

  const loadPublic = async () => {
    setLoading(true);
    setError('');
    try {
      const res = await fetch('/public/documents?limit=50', {
        headers: { 'Authorization': 'Bearer ' + localStorage.getItem('token') }
      });
      if (!res.ok) throw new Error('HTTP ' + res.status);
      const data = await res.json();
      const docs = Array.isArray(data.documents) ? data.documents : [];
      setDocs(docs);
    } catch (err) {
      console.error('Load public error:', err);
      setError('Не удалось загрузить публичные документы');
      setDocs([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadPublic(); }, []);

  const handleSearch = async () => {
    if (!query.trim()) return loadPublic();
    
    setLoading(true);
    setError('');
    setDebug(`Поиск: "${query}"...`);
    
    try {
      const res = await searchAPI.search(query, { limit: 50 });
      console.log('Search response:', res.data);
      setDebug(`Ответ поиска: ${JSON.stringify(res.data).slice(0, 200)}...`);
      
      const raw = Array.isArray(res.data) ? res.data : [];
      
      const extracted = raw
        .map(r => getField(r, 'document', 'Document', 'doc') || r)
        .filter(d => getField(d, 'ID', 'id'));
      
      setDebug(`Извлечено документов: ${extracted.length}`);
      
      const seen = new Set();
      const unique = extracted.filter(d => {
        const id = getField(d, 'ID', 'id');
        if (!id || seen.has(id)) return false;
        seen.add(id);
        return true;
      });
      
      setDebug(`Уникальных: ${unique.length}`);
      setDocs(unique);
      
    } catch (err) {
      console.error('Search error:', err);
      setError('Ошибка поиска: ' + err.message);
      setDocs([]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <div>
          <h1>🔍 Поиск документов</h1>
          <nav style={{ marginTop: '0.5rem' }}>
            <Link to="/" style={{ marginRight: '1rem' }}>🏠 Мои документы</Link>
            <Link to="/search" style={{ fontWeight: 'bold' }}>🔍 Публичные</Link>
          </nav>
        </div>
        <div>
          {user?.email ? (
            <>
              <span style={{ marginRight: '1rem' }}>{user.email}</span>
              <button onClick={logout}>Выйти</button>
            </>
          ) : (
            <Link to="/login"><button>Войти</button></Link>
          )}
        </div>
      </header>

      {error && (
        <p style={{ color: 'red', background: '#fee', padding: '0.5rem', borderRadius: '4px', marginBottom: '1rem' }}>
          {error}
        </p>
      )}
      
      {debug && (
        <p style={{ fontSize: '0.8rem', color: '#666', marginBottom: '1rem', fontFamily: 'monospace' }}>
          🔍 {debug}
        </p>
      )}

      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
        <input 
          type="search" 
          placeholder="Поиск по содержимому (только публичные)..." 
          value={query}
          onChange={e => setQuery(e.target.value)} 
          onKeyDown={e => e.key === 'Enter' && handleSearch()}
          style={{ flex: 1, padding: '0.5rem' }} 
        />
        <button onClick={handleSearch} disabled={loading}>
          {loading ? 'Поиск...' : 'Найти'}
        </button>
      </div>

      {loading ? (
        <p style={{ textAlign: 'center', padding: '2rem' }}>Загрузка...</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Название</th>
              <th>Размер</th>
              <th>Дата</th>
              <th>Версия</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {docs.length ? docs.map(doc => {
              const id = getField(doc, 'ID', 'id');
              const title = getField(doc, 'Title', 'title', 'OriginalFilename', 'original_filename') || 'Без названия';
              const size = formatSize(getField(doc, 'FileSize', 'file_size'));
              const date = formatDate(getField(doc, 'CreatedAt', 'created_at'));
              const version = getField(doc, 'CurrentVersion', 'current_version') || 1;
              const isPub = getField(doc, 'IsPublic', 'is_public');
              
              return (
                <tr key={id}>
                  <td>
                    <Link to={`/document/${id}`}>{title}</Link>
                    {isPub && <span style={{marginLeft:'0.25rem',color:'#0a0'}}>🌐</span>}
                  </td>
                  <td className="meta">{size}</td>
                  <td className="meta">{date}</td>
                  <td className="meta">v{version}</td>
                  <td className="actions">
                    <button onClick={() => documentsAPI.download(id).then(r => {
                      const a = document.createElement('a'); 
                      a.href = URL.createObjectURL(r.data);
                      a.download = title; 
                      a.click();
                    })}>⬇️</button>
                    <Link to={`/document/${id}`}><button>👁️</button></Link>
                  </td>
                </tr>
              );
            }) : (
              <tr>
                <td colSpan="5" style={{ textAlign: 'center', padding: '2rem', color: '#666' }}>
                  {query 
                    ? 'Ничего не найдено. Попробуйте другой запрос или проверьте, что документ публичный.' 
                    : 'Публичных документов пока нет'}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      )}
    </div>
  );
}
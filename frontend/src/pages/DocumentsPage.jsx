import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
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

export default function DocumentsPage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [docs, setDocs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  const loadDocs = async () => {
    setLoading(true);
    try {
      const res = await documentsAPI.list({ limit: 50 });
      const data = Array.isArray(res.data) ? res.data : (res.data?.documents || []);
      setDocs(data);
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
    try {
      const res = await searchAPI.search(searchQuery, { limit: 50 });
      const raw = Array.isArray(res.data) ? res.data : [];
      const extracted = raw.map(r => getField(r, 'document', 'Document', 'doc') || r).filter(Boolean);
      const unique = [...new Map(extracted.map(d => [getField(d, 'ID', 'id'), d])).values()];
      setDocs(unique);
    } catch {
      loadDocs();
    }
  };

  return (
    <div>
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <div>
          <h1>📁 Мои документы</h1>
          <nav style={{ marginTop: '0.5rem' }}>
            <Link to="/" style={{ marginRight: '1rem', fontWeight: 'bold' }}>🏠 Мои</Link>
            <Link to="/search">🔍 Поиск публичных</Link>
          </nav>
        </div>
        <div className="flex" style={{display:'flex',gap:'0.5rem',alignItems:'center'}}>
          <span>{user?.email}</span>
          <button onClick={logout}>Выйти</button>
          {}
          <button onClick={() => navigate('/upload')} style={{background:'#0a0',color:'white',padding:'0.5rem 1rem'}}>
            ➕ Загрузить
          </button>
        </div>
      </header>

      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
        <input 
          type="search" 
          placeholder="Поиск..." 
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && handleSearch()}
          style={{ flex: 1, padding: '0.5rem' }}
        />
        <button onClick={handleSearch}>🔍 Найти</button>
      </div>

      {loading ? <p>Загрузка...</p> : (
        <table>
          <thead>
            <tr><th>Название</th><th>Размер</th><th>Дата</th><th>Версия</th><th>Действия</th></tr>
          </thead>
          <tbody>
            {docs.length ? docs.map(doc => {
              const id = getField(doc, 'ID', 'id');
              const title = getField(doc, 'Title', 'title', 'OriginalFilename', 'original_filename') || '?';
              return (
                <tr key={id}>
                  <td><Link to={`/document/${id}`}>{title}</Link></td>
                  <td className="meta">{formatSize(getField(doc, 'FileSize', 'file_size'))}</td>
                  <td className="meta">{formatDate(getField(doc, 'CreatedAt', 'created_at'))}</td>
                  <td className="meta">v{getField(doc, 'CurrentVersion', 'current_version')||1}</td>
                  <td className="actions">
                    <button onClick={() => documentsAPI.download(id).then(r => {
                      const a = document.createElement('a'); a.href = URL.createObjectURL(r.data);
                      a.download = title; a.click();
                    })}>⬇️</button>
                    <Link to={`/document/${id}`}><button>👁️</button></Link>
                  </td>
                </tr>
              );
            }) : <tr><td colSpan="5" style={{textAlign:'center',padding:'1rem'}}>Нет документов</td></tr>}
          </tbody>
        </table>
      )}
    </div>
  );
}
// frontend/src/pages/DocumentsPage.jsx
import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { documentsAPI, searchAPI } from '../services/api';


// ХЕЛПЕРЫ (чистые функции, без побочных эффектов)

/** Безопасное получение поля из объекта (поддержка camelCase и snake_case) */
const getField = (obj, ...names) => {
  for (const name of names) {
    if (obj?.[name] !== undefined && obj?.[name] !== null) return obj[name];
  }
  return null;
};

/** Иконка документа по MIME-типу */
const getFileIcon = (mime) => {
  if (!mime) return '📄';
  if (mime.startsWith('image/')) return '🖼️';
  if (mime.startsWith('video/')) return '🎬';
  if (mime.startsWith('audio/')) return '🎵';
  if (mime.includes('pdf')) return '📕';
  if (mime.includes('word') || mime.includes('document')) return '📘';
  if (mime.includes('excel') || mime.includes('spreadsheet')) return '📗';
  if (mime.includes('powerpoint') || mime.includes('presentation')) return '📙';
  if (mime.includes('text') || mime.includes('markdown')) return '📝';
  return '📄';
};

// ПОД-КОМПОНЕНТЫ (изолированная логика отображения)


/** Карточка документа для личных документов (с кнопками действий) */
function DocumentCard({ doc, onDownload, onDelete }) {
  const id = getField(doc, 'ID', 'id');
  const title = getField(doc, 'Title', 'title', 'OriginalFilename', 'original_filename') || 'Без названия';
  const version = getField(doc, 'CurrentVersion', 'current_version') || 1;
  const isPublic = getField(doc, 'IsPublic', 'is_public');
  const mime = getField(doc, 'MimeType', 'mime_type') || '';
  

  const handleDownload = (e) => {
    e.preventDefault();
    e.stopPropagation();
    documentsAPI.download(id).then(r => {
      const a = document.createElement('a');
      a.href = URL.createObjectURL(r.data);
      a.download = title;
      a.click();
      URL.revokeObjectURL(a.href);
      onDownload?.();
    });
  };

  const handleDelete = (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (window.confirm(`Удалить документ "${title}"?`)) {
      onDelete?.(id);
    }
  };

  return (
    <Link 
      to={`/document/${id}`}
      className="doc-card group"
    >
      <div className="doc-card__content">
        {/* Иконка + название */}
        <div className="doc-card__info">
          <div className="doc-card__icon">{getFileIcon(mime)}</div>
          <div className="doc-card__text">
            <h3 className="doc-card__title">{title}</h3>
            <div className="doc-card__badges">
              <span className="badge badge-primary">v{version}</span>
              {isPublic && <span className="badge badge-success">Публичный</span>}
            </div>
          </div>
        </div>

        {/* Кнопки действий (показываются при ховере) */}
        <div className="doc-card__actions">
          <button 
            onClick={handleDownload}
            className="btn btn-icon"
            title="Скачать"
            aria-label="Скачать документ"
          >
            <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
          </button>
          <button 
            onClick={handleDelete}
            className="btn btn-icon btn-danger"
            title="Удалить"
            aria-label="Удалить документ"
          >
            <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
      </div>
    </Link>
  );
}

/** Компонент пустого состояния */
function EmptyState({ onUpload }) {
  return (
    <div className="empty-state">
      <div className="empty-state__icon">📁</div>
      <h3 className="empty-state__title">Нет документов</h3>
      <p className="empty-state__text">Загрузите первый файл, чтобы начать</p>
      <button onClick={onUpload} className="btn btn-primary">
        ➕ Загрузить документ
      </button>
    </div>
  );
}

/** Компонент загрузки (скелетон) */
function LoadingGrid({ count = 6 }) {
  return (
    <div className="doc-grid">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="doc-card doc-card--loading">
          <div className="doc-card__content">
            <div className="doc-card__info">
              <div className="doc-card__icon doc-card__icon--skeleton" />
              <div className="doc-card__text">
                <div className="doc-card__title doc-card__title--skeleton" />
                <div className="doc-card__badges">
                  <div className="badge badge--skeleton" />
                </div>
              </div>
            </div>
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

  // Загрузка документов
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
        // Отключаем кеш браузера — гарантируем свежие данные
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

  // Поиск
  const handleSearch = async () => {
    if (!searchQuery.trim()) return loadDocs();
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const userId = JSON.parse(localStorage.getItem('user') || '{}')?.id;
      
      const endpoint = searchMode === 'title'
        ? `/search/title?query=${encodeURIComponent(searchQuery)}&limit=50`
        : `/search/?query=${encodeURIComponent(searchQuery)}&limit=50`;
      
      const res = await fetch(endpoint, {
        headers: {
          'Authorization': token ? `Bearer ${token}` : '',
          'X-User-ID': userId || ''
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

  // Удаление документа
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
      {/* Хедер */}
      <header className="header">
        <div className="container header__inner">
          
          {/* Логотип */}
          <Link to="/" className="logo-link">
            <div className="logo">D</div>
            <span>DocService</span>
          </Link>
          
          {/* Кнопки */}
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

      {/* Контент */}
      <main className="container page__content">
        {/* Заголовок + навигация */}
        <div className="page__header">
          <div>
            <h1 className="page__title">📁 Мои документы</h1>
            <p className="page__subtitle">Управляйте своими файлами и версиями</p>
          </div>
          <nav className="page__nav">
            <Link to="/" className="btn btn-outline btn-sm">🏠 Публичные</Link>
            <Link to="/search" className="btn btn-outline btn-sm">🔍 Поиск</Link>
          </nav>
        </div>

        {/* Поиск */}
        <div className="search-bar">
          {/* Переключатель режима поиска */}
          <select 
            value={searchMode}
            onChange={(e) => setSearchMode(e.target.value)}
            className="input"
            style={{ width: 'auto', padding: '0.5rem' }}
          >
            <option value="content">📄 Текст</option>
            <option value="title">🏷️ Заголовок</option>
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

        {/* Список документов */}
        {loading ? (
          <LoadingGrid />
        ) : docs.length > 0 ? (
          <div className="doc-grid">
            {docs.map(doc => (
              <DocumentCard 
                key={getField(doc, 'ID', 'id')} 
                doc={doc} 
                onDownload={() => {}}
                onDelete={handleDelete}
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
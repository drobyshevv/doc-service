// frontend/src/pages/HomePage.jsx — Упрощённая версия для отладки
import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { searchAPI } from '../services/api';

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
    <div style={{ minHeight: '100vh', background: '#f8fafc' }}>
      {/* Хедер */}
        <header style={{ 
        position: 'sticky', top: 0, zIndex: 50, 
        background: 'rgba(255,255,255,0.8)', backdropFilter: 'blur(8px)',
        borderBottom: '1px solid #e2e8f0',
        padding: '0.75rem 1rem'
        }}>
        <div style={{ maxWidth: 1200, margin: '0 auto', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            {/* Логотип */}
            <Link to="/" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', textDecoration: 'none', color: '#1e293b', fontWeight: 600 }}>
            <div style={{ 
                width: 32, height: 32, borderRadius: 8, 
                background: 'linear-gradient(135deg, #6366f1, #8b5cf6)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: 'white', fontSize: 14, fontWeight: 'bold'
            }}>D</div>
            DocService
            </Link>
            
            {/* Навигация + авторизация */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            {/* Кнопка "Мои документы" — только для авторизованных */}
            {isAuthenticated && (
                <Link to="/dashboard" style={{ 
                padding: '0.5rem 1rem', fontSize: 14, fontWeight: 500,
                borderRadius: 8, border: '1px solid #6366f1', background: 'transparent',
                textDecoration: 'none', color: '#6366f1',
                transition: 'all 0.2s'
                }}
                onMouseEnter={(e) => { e.target.style.background = '#6366f1'; e.target.style.color = 'white'; }}
                onMouseLeave={(e) => { e.target.style.background = 'transparent'; e.target.style.color = '#6366f1'; }}
                >
                📁 Мои документы
                </Link>
            )}
            
            {/* Кнопки входа/регистрации или выход */}
            {isAuthenticated ? (
                <>
                <span style={{ color: '#64748b', fontSize: 14 }}>{user?.email}</span>
                <button onClick={logout} style={{ 
                    padding: '0.5rem 1rem', fontSize: 14, fontWeight: 500,
                    borderRadius: 8, border: '1px solid #e2e8f0', background: 'transparent',
                    cursor: 'pointer', color: '#1e293b'
                }}>Выйти</button>
                </>
            ) : (
                <>
                <Link to="/login" style={{ 
                    padding: '0.5rem 1rem', fontSize: 14, fontWeight: 500,
                    borderRadius: 8, border: '1px solid #e2e8f0', background: 'transparent',
                    textDecoration: 'none', color: '#1e293b'
                }}>Войти</Link>
                <Link to="/register" style={{ 
                    padding: '0.5rem 1rem', fontSize: 14, fontWeight: 500,
                    borderRadius: 8, border: 'none', background: '#6366f1',
                    textDecoration: 'none', color: 'white'
                }}>Регистрация</Link>
                </>
            )}
            </div>
        </div>
        </header>

      {/* Герой */}
      <section style={{ 
        background: 'linear-gradient(135deg, #4f46e5, #7c3aed)',
        color: 'white', padding: '3rem 1rem', textAlign: 'center'
      }}>
        <div style={{ maxWidth: 800, margin: '0 auto' }}>
          <h1 style={{ fontSize: '2rem', fontWeight: 700, marginBottom: '1rem' }}>
            Найдите и поделитесь документами
          </h1>
          <p style={{ fontSize: '1.125rem', opacity: 0.9, marginBottom: '1.5rem' }}>
            Безопасное хранение, версионирование и поиск по содержимому
          </p>
          <div style={{ display: 'flex', gap: '0.5rem', maxWidth: 500, margin: '0 auto', alignItems: 'center' }}>
            {/* Переключатель режима поиска */}
            <select 
              value={searchMode}
              onChange={(e) => setSearchMode(e.target.value)}
              style={{ 
                padding: '0.75rem 0.5rem', 
                fontSize: 14, 
                borderRadius: 8, 
                border: 'none',
                background: 'rgba(255,255,255,0.9)',
                color: '#1e293b',
                cursor: 'pointer'
              }}
            >
              <option value="content">📄 Текст</option>
              <option value="title">🏷️ Заголовок</option>
            </select>
            
            <input
              type="search"
              placeholder={searchMode === 'title' ? "Название или имя файла..." : "Поиск по содержимому..."}
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              style={{ 
                flex: 1, padding: '0.75rem 1rem', fontSize: 16,
                border: 'none', borderRadius: 8, background: 'rgba(255,255,255,0.9)',
                color: '#1e293b'
              }}
            />
            <button onClick={handleSearch} style={{ 
              padding: '0.75rem 1.5rem', fontSize: 16, fontWeight: 500,
              borderRadius: 8, border: 'none', background: '#1e293b', color: 'white',
              cursor: 'pointer'
            }}>Найти</button>
          </div>
        </div>
      </section>

      {/* Список документов */}
      <main style={{ maxWidth: 1200, margin: '0 auto', padding: '2rem 1rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.5rem' }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 600, color: '#1e293b' }}>
            {query ? `Результаты: "${query}"` : 'Публичные документы'}
          </h2>
          <span style={{ fontSize: 14, color: '#64748b' }}>{docs.length} документов</span>
        </div>

        {loading ? (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '1rem' }}>
            {[...Array(6)].map((_, i) => (
              <div key={i} style={{ 
                background: 'white', border: '1px solid #e2e8f0', borderRadius: 12, padding: '1rem',
                display: 'flex', alignItems: 'center', gap: '0.75rem', animation: 'pulse 1.5s infinite'
              }}>
                <div style={{ width: 40, height: 40, borderRadius: 8, background: '#e2e8f0' }} />
                <div style={{ flex: 1 }}>
                  <div style={{ height: 16, background: '#e2e8f0', borderRadius: 4, width: '75%', marginBottom: 8 }} />
                  <div style={{ height: 12, background: '#e2e8f0', borderRadius: 4, width: '25%' }} />
                </div>
              </div>
            ))}
          </div>
        ) : docs.length > 0 ? (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '1rem' }}>
            {docs.map(doc => {
              const id = doc.ID || doc.id;
              const title = doc.Title || doc.title || doc.OriginalFilename || 'Без названия';
              const version = doc.CurrentVersion || doc.current_version || 1;
              const isPublic = doc.IsPublic || doc.is_public;
              
              return (
                <Link 
                  key={id}
                  to={`/document/${id}`}
                  style={{ 
                    display: 'block', background: 'white', border: '1px solid #e2e8f0', 
                    borderRadius: 12, padding: '1rem', textDecoration: 'none', color: 'inherit',
                    transition: 'box-shadow 0.2s, border-color 0.2s'
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'start', gap: '0.75rem' }}>
                    <div style={{ 
                      width: 40, height: 40, borderRadius: 8, background: '#f1f5f9',
                      display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20
                    }}>📄</div>
                    <div style={{ minWidth: 0 }}>
                      <h3 style={{ fontWeight: 500, color: '#1e293b', margin: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {title}
                      </h3>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: 4 }}>
                        <span style={{ 
                          fontSize: 12, fontWeight: 500, padding: '0.25rem 0.5rem',
                          borderRadius: 9999, background: 'rgb(99 102 241 / 0.1)', color: '#6366f1'
                        }}>v{version}</span>
                        {isPublic && (
                          <span style={{ 
                            fontSize: 12, fontWeight: 500, padding: '0.25rem 0.5rem',
                            borderRadius: 9999, background: 'rgb(34 197 94 / 0.1)', color: '#22c55e'
                          }}>Публичный</span>
                        )}
                      </div>
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: '3rem 1rem' }}>
            <div style={{ 
              width: 64, height: 64, borderRadius: '50%', background: '#f1f5f9',
              display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 1rem'
            }}>
              <span style={{ fontSize: 32 }}>📄</span>
            </div>
            <h3 style={{ fontSize: '1.125rem', fontWeight: 500, color: '#1e293b', marginBottom: '0.5rem' }}>
              {query ? 'Ничего не найдено' : 'Публичных документов пока нет'}
            </h3>
            <p style={{ color: '#64748b', marginBottom: '1.5rem' }}>
              {query ? 'Попробуйте другой запрос' : 'Зарегистрируйтесь и начните делиться файлами'}
            </p>
            {!isAuthenticated && (
              <Link to="/register" style={{ 
                padding: '0.75rem 1.5rem', fontSize: 14, fontWeight: 500,
                borderRadius: 8, border: 'none', background: '#6366f1', color: 'white',
                textDecoration: 'none'
              }}>Создать аккаунт</Link>
            )}
          </div>
        )}
      </main>
    </div>
  );
}
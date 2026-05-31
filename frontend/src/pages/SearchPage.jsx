// frontend/src/pages/HomePage.jsx (переименуй SearchPage.jsx в HomePage.jsx)
import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import Header from '../components/Header';
import DocumentCard from '../components/DocumentCard';
import { documentsAPI, searchAPI } from '../services/api';

export default function HomePage() {
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();
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
    <div className="min-h-screen bg-slate-50">
      <Header />
      
      {/* 🔹 Герой-секция */}
      <section className="bg-gradient-to-br from-indigo-600 to-purple-700 text-white py-12 md:py-16">
        <div className="container text-center">
          <h1 className="text-3xl md:text-4xl font-bold mb-4">
            Найдите и поделитесь документами
          </h1>
          <p className="text-lg text-indigo-100 mb-6 max-w-2xl mx-auto">
            Безопасное хранение, версионирование и поиск по содержимому ваших файлов
          </p>
          
          {/* 🔹 Поиск в герое */}
          <div className="max-w-xl mx-auto flex gap-2">
            <input
              type="search"
              placeholder="Поиск по названию или содержимому..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              className="input flex-1 bg-white/90 border-0 text-slate-800 placeholder-slate-400"
            />
            <button onClick={handleSearch} className="btn btn-primary">
              Найти
            </button>
          </div>
        </div>
      </section>

      {/* 🔹 Список документов */}
      <main className="container py-8">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold text-slate-800">
            {query ? `Результаты: "${query}"` : 'Публичные документы'}
          </h2>
          <span className="text-sm text-slate-500">
            {docs.length} {docs.length === 1 ? 'документ' : 'документов'}
          </span>
        </div>

        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="card animate-pulse">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-lg bg-slate-200" />
                  <div className="flex-1">
                    <div className="h-4 bg-slate-200 rounded w-3/4 mb-2" />
                    <div className="h-3 bg-slate-200 rounded w-1/4" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : docs.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {docs.map(doc => (
              <DocumentCard key={doc.ID || doc.id} doc={doc} />
            ))}
          </div>
        ) : (
          <div className="text-center py-12">
            <div className="w-16 h-16 rounded-full bg-slate-100 flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-slate-800 mb-2">
              {query ? 'Ничего не найдено' : 'Публичных документов пока нет'}
            </h3>
            <p className="text-slate-500 mb-4">
              {query 
                ? 'Попробуйте другой запрос или загрузите свой документ' 
                : 'Зарегистрируйтесь и начните делиться файлами'}
            </p>
            {!isAuthenticated && (
              <Link to="/register" className="btn btn-primary">
                Создать аккаунт
              </Link>
            )}
          </div>
        )}
      </main>

      {/* 🔹 Футер */}
      <footer className="border-t border-slate-200 py-6 mt-12">
        <div className="container text-center text-sm text-slate-500">
          © 2026 DocService. Все права защищены.
        </div>
      </footer>
    </div>
  );
}
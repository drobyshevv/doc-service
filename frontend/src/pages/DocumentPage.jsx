
import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { documentsAPI } from '../services/api';

const getField = (obj, ...names) => {
  for (const name of names) {
    if (obj?.[name] !== undefined && obj?.[name] !== null) return obj[name];
  }
  return null;
};

const formatDate = (val) => {
  if (!val) return '—';
  const date = new Date(val);
  return isNaN(date.getTime()) ? '—' : date.toLocaleString('ru-RU');
};

const formatSize = (bytes) => {
  if (!bytes && bytes !== 0) return '—';
  const kb = bytes / 1024;
  return kb < 1024 ? `${kb.toFixed(1)} KB` : `${(kb / 1024).toFixed(1)} MB`;
};

export default function DocumentPage() {
  const { id } = useParams();
  const [meta, setMeta] = useState(null);
  const [versions, setVersions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const load = async () => {
      try {
        const [metaRes, verRes] = await Promise.all([
          documentsAPI.getMeta(id),
          documentsAPI.getVersions(id)
        ]);
        setMeta(metaRes.data);
        setVersions(Array.isArray(verRes.data) ? verRes.data : []);
      } catch (err) {
        console.error('Failed to load document:', err);
        setError('Не удалось загрузить документ');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [id]);

  if (loading) return <p style={{textAlign:'center',padding:'2rem'}}>Загрузка...</p>;
  if (error) return <p style={{color:'red',textAlign:'center',padding:'2rem'}}>{error}</p>;
  if (!meta) return <p style={{textAlign:'center',padding:'2rem'}}>Документ не найден</p>;

  const title = getField(meta, 'Title', 'title', 'OriginalFilename', 'original_filename') || 'Без названия';
  const curVer = getField(meta, 'CurrentVersion', 'current_version') || 1;
  const isPublic = getField(meta, 'IsPublic', 'is_public');

  return (
    <div>
      <header style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '1rem' }}>
        <div>
          <h1>{title}</h1>
          <p className="meta">
            {getField(meta, 'MimeType', 'mime_type') || '—'} • {formatSize(getField(meta, 'FileSize', 'file_size'))}
          </p>
        </div>
        <div className="actions">
          <button onClick={() => documentsAPI.download(id).then(r => {
            const a = document.createElement('a'); 
            a.href = URL.createObjectURL(r.data);
            a.download = title; 
            a.click();
          })}>⬇️ Скачать</button>
          <button onClick={async () => {
            if (window.confirm('Удалить документ?')) {
              try {
                await documentsAPI.delete(id);
                window.location.href = '/';
              } catch (e) {
                alert('Ошибка: ' + e.message);
              }
            }
          }}>🗑️ Удалить</button>
          <Link to="/"><button>← Назад</button></Link>
        </div>
      </header>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem', marginBottom: '2rem' }}>
        <div className="meta-item"><strong>ID</strong><code>{getField(meta, 'ID', 'id')}</code></div>
        <div className="meta-item"><strong>Владелец</strong>{getField(meta, 'OwnerID', 'owner_id') || '—'}</div>
        <div className="meta-item"><strong>Создан</strong>{formatDate(getField(meta, 'CreatedAt', 'created_at'))}</div>
        <div className="meta-item"><strong>Обновлён</strong>{formatDate(getField(meta, 'UpdatedAt', 'updated_at'))}</div>
        <div className="meta-item"><strong>Публичный</strong>{isPublic ? '✅ Да' : '🔒 Нет'}</div>
        <div className="meta-item"><strong>Версия</strong>v{curVer}</div>
      </div>

      <h3>🔄 Версии</h3>
      {versions.length === 0 ? (
        <p style={{color:'#666'}}>Нет версий</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Версия</th>
              <th>Дата</th>
              <th>Размер</th>
              <th>Комментарий</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {versions.map(v => {
              const ver = getField(v, 'Version', 'version');
              const isCurrent = ver == curVer;
              return (
                <tr key={getField(v, 'ID', 'id') || ver}>
                  <td><strong>v{ver}</strong>{isCurrent ? ' ●' : ''}</td>
                  <td>{formatDate(getField(v, 'CreatedAt', 'created_at'))}</td>
                  <td>{formatSize(getField(v, 'FileSize', 'file_size'))}</td>
                  <td>{getField(v, 'Note', 'note') || '—'}</td>
                  <td className="actions">
                    <button onClick={() => documentsAPI.download(id, ver).then(r => {
                      const a = document.createElement('a'); 
                      a.href = URL.createObjectURL(r.data);
                      a.download = `v${ver}.bin`; 
                      a.click();
                    })}>⬇️</button>
                    {!isCurrent && (
                      <button onClick={async () => {
                        if (window.confirm(`Откатиться к версии ${ver}?`)) {
                          try {
                            await fetch(`/documents/${id}/versions/${ver}/rollback`, {
                              method: 'POST',
                              headers: { 'Authorization': 'Bearer ' + localStorage.getItem('token') }
                            });
                            window.location.reload();
                          } catch (e) {
                            alert('Ошибка: ' + e.message);
                          }
                        }
                      }}>🔄 Откат</button>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </div>
  );
}
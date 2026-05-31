// frontend/src/pages/DocumentPage.jsx
import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { documentsAPI } from '../services/api';
import UploadVersionModal from '../components/UploadVersionModal';
import DocumentMetadataEditor from '../components/DocumentMetadataEditor';

// ХЕЛПЕРЫ (чистые функции, без побочных эффектов)

const getField = (obj, ...names) => {
  for (const name of names) {
    if (obj?.[name] !== undefined && obj?.[name] !== null) return obj[name];
  }
  return null;
};

const formatDate = (val) => {
  if (!val) return '—';
  const d = new Date(val);
  return isNaN(d.getTime()) ? '—' : d.toLocaleString('ru-RU');
};

const formatSize = (bytes) => {
  if (!bytes && bytes !== 0) return '—';
  const kb = bytes / 1024;
  return kb < 1024 ? `${kb.toFixed(1)} KB` : `${(kb / 1024).toFixed(1)} MB`;
};

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

const isDocumentOwner = (meta) => {
  try {
    const currentUser = JSON.parse(localStorage.getItem('user') || '{}');
    const ownerId = getField(meta, 'OwnerID', 'owner_id');
    return currentUser?.id && ownerId && currentUser.id === ownerId;
  } catch {
    return false;
  }
};

// ПОД-КОМПОНЕНТЫ (изолированная логика отображения)

/** Хедер страницы документа */
function DocumentHeader({ title, version, isPublic, mime, onDownload, onDelete, isOwner, onBack }) {
  return (
    <header className="header">
      <div className="container header__inner">
        
        {/* Логотип: используем универсальные классы из index.css */}
        <Link to="/" className="logo-link" onClick={onBack}>
          <div className="logo">D</div>
          <span>DocService</span>
        </Link>
        
        {/* Кнопки действий */}
        <div className="header__actions">
          <button 
            onClick={onDownload} 
            className="btn btn-outline btn-sm" 
            aria-label="Скачать документ"
          >
            <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            <span className="hide-mobile">Скачать</span>
          </button>
          
          {isOwner && (
            <button 
              onClick={onDelete} 
              className="btn btn-outline btn-sm btn-danger" 
              aria-label="Удалить документ"
            >
              <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              <span className="hide-mobile">Удалить</span>
            </button>
          )}
          
          <Link to="/" className="btn btn-primary btn-sm" onClick={onBack}>
            ← Назад
          </Link>
        </div>
        
      </div>
    </header>
  );
}

/** Заголовок документа с иконкой и мета-информацией */
function DocumentHero({ title, version, isPublic, mime, size }) {
  return (
    <section className="document-hero">
      <div className="document-hero__icon">{getFileIcon(mime)}</div>
      <div className="document-hero__content">
        <h1 className="document-hero__title">{title}</h1>
        <div className="document-hero__badges">
          <span className="badge badge-primary">v{version}</span>
          {isPublic && <span className="badge badge-success">Публичный</span>}
          <span className="badge badge-muted">{mime || 'Неизвестный тип'}</span>
          <span className="badge badge-muted">{size}</span>
        </div>
      </div>
    </section>
  );
}

/** Вкладки навигации */
function DocumentTabs({ activeTab, onTabChange, versionsCount }) {
  return (
    <nav className="document-tabs" role="tablist">
      <button
        role="tab"
        aria-selected={activeTab === 'info'}
        className={`document-tabs__tab ${activeTab === 'info' ? 'document-tabs__tab--active' : ''}`}
        onClick={() => onTabChange('info')}
      >
        📋 Информация
      </button>
      <button
        role="tab"
        aria-selected={activeTab === 'versions'}
        className={`document-tabs__tab ${activeTab === 'versions' ? 'document-tabs__tab--active' : ''}`}
        onClick={() => onTabChange('versions')}
      >
        🔄 Версии <span className="document-tabs__count">({versionsCount})</span>
      </button>
    </nav>
  );
}

/** Карточка с мета-информацией документа */
function DocumentInfoCard({ meta }) {
  const fields = [
    { label: 'ID документа', value: getField(meta, 'ID', 'id'), mono: true },
    { label: 'Владелец', value: getField(meta, 'OwnerID', 'owner_id') },
    { label: 'Создан', value: formatDate(getField(meta, 'CreatedAt', 'created_at')) },
    { label: 'Обновлён', value: formatDate(getField(meta, 'UpdatedAt', 'updated_at')) },
    { label: 'Размер', value: formatSize(getField(meta, 'FileSize', 'file_size')) },
    { label: 'Версия', value: `v${getField(meta, 'CurrentVersion', 'current_version') || 1}` },
  ];

  return (
    <div className="card card--info">
      <h3 className="card__title">Детали документа</h3>
      <dl className="meta-grid">
        {fields.map(({ label, value, mono }) => (
          <div key={label} className="meta-grid__item">
            <dt className="meta-grid__label">{label}</dt>
            <dd className={`meta-grid__value ${mono ? 'meta-grid__value--mono' : ''}`}>{value || '—'}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}

/** Панель быстрых действий */
function DocumentActions({ onDownload, onShowVersions, onDelete, onUploadVersion, isOwner }) {
  return (
    <div className="card card--actions">
      <h3 className="card__title">Действия</h3>
      <div className="actions-list">
        <button onClick={onDownload} className="btn btn-outline btn--full btn--start">
          <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
          Скачать текущую версию
        </button>
        
        {/* Кнопка "Загрузить новую версию" — только для владельца */}
        {isOwner && (
          <button 
            onClick={onUploadVersion} 
            className="btn btn-outline btn--full btn--start"
          >
            <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            Загрузить новую версию
          </button>
        )}
        
        <button onClick={onShowVersions} className="btn btn-outline btn--full btn--start">
          <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Показать все версии
        </button>
        
        {isOwner && (
          <button onClick={onDelete} className="btn btn-outline btn--full btn--start btn-danger">
            <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Удалить документ
          </button>
        )}
      </div>
    </div>
  );
}

/** Элемент версии документа */
function VersionItem({ version, isCurrent, note, size, date, onDownload, onRollback }) {
  return (
    <div className={`version-item ${isCurrent ? 'version-item--current' : ''}`}>
      <div className="version-item__content">
        <div className="version-item__header">
          <span className={`version-item__version ${isCurrent ? 'version-item__version--current' : ''}`}>
            v{version}
          </span>
          {isCurrent && <span className="badge badge-success">Текущая</span>}
        </div>
        
        {note && <p className="version-item__note">{note}</p>}
        
        <p className="version-item__meta">{date} • {size}</p>
      </div>
      
      <div className="version-item__actions">
        <button 
          onClick={() => onDownload(version)}
          className="btn btn-icon"
          title="Скачать версию"
          aria-label={`Скачать версию ${version}`}
        >
          <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
        </button>
        
        {!isCurrent && (
          <button 
            onClick={() => onRollback(version)}
            className="btn btn-icon"
            title="Откатиться к этой версии"
            aria-label={`Откатиться к версии ${version}`}
          >
            <svg className="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
            </svg>
          </button>
        )}
      </div>
    </div>
  );
}

/** Список версий документа */
function VersionsList({ versions, currentVersion, onDownload, onRollback }) {
  if (versions.length === 0) {
    return <p className="text-center text-muted py-4">Нет сохранённых версий</p>;
  }

  return (
    <div className="versions-list">
      {versions.map((v) => (
        <VersionItem
          key={getField(v, 'ID', 'id') || getField(v, 'Version', 'version')}
          version={getField(v, 'Version', 'version')}
          isCurrent={getField(v, 'Version', 'version') == currentVersion}
          note={getField(v, 'Note', 'note')}
          size={formatSize(getField(v, 'FileSize', 'file_size'))}
          date={formatDate(getField(v, 'CreatedAt', 'created_at'))}
          onDownload={onDownload}
          onRollback={onRollback}
        />
      ))}
    </div>
  );
}

/** Состояние загрузки */
function DocumentLoading() {
  return (
    <div className="page-loading">
      <div className="spinner" />
      <p className="text-muted">Загрузка документа...</p>
    </div>
  );
}

/** Состояние ошибки */
function DocumentError({ message, onBack }) {
  return (
    <div className="page-error">
      <div className="page-error__icon">⚠️</div>
      <h3 className="page-error__title">Ошибка</h3>
      <p className="page-error__text">{message || 'Документ не найден'}</p>
      <Link to="/" className="btn btn-primary" onClick={onBack}>← На главную</Link>
    </div>
  );
}

// ОСНОВНОЙ КОМПОНЕНТ

export default function DocumentPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  
  const [meta, setMeta] = useState(null);
  const [versions, setVersions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [activeTab, setActiveTab] = useState('info');
  const [showUploadModal, setShowUploadModal] = useState(false);

  // Загрузка данных
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
        console.error('Load error:', err);
        setError('Не удалось загрузить документ');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [id]);

  // Действия
  const handleDownload = (version = null) => {
    documentsAPI.download(id, version).then(r => {
      const a = document.createElement('a');
      a.href = URL.createObjectURL(r.data);
      a.download = meta?.Title || meta?.title || 'document';
      a.click();
      URL.revokeObjectURL(a.href);
    });
  };

  const handleDelete = async () => {
    if (!window.confirm('Удалить этот документ?')) return;
    try {
      await documentsAPI.delete(id);
      navigate('/');
    } catch (err) {
      alert('Ошибка: ' + err.message);
    }
  };

  const handleRollback = async (version) => {
    if (!window.confirm(`Откатиться к версии ${version}?`)) return;
    try {
      await documentsAPI.rollback(id, version);
      // После отката перезагружаем всё
      const [metaRes, verRes] = await Promise.all([
        documentsAPI.getMeta(id),
        documentsAPI.getVersions(id)
      ]);
      setMeta(metaRes.data);
      setVersions(Array.isArray(verRes.data) ? verRes.data : []);
    } catch (err) {
      console.error('Rollback error:', err);
      alert('Ошибка: ' + (err.message || 'Не удалось откатиться'));
    }
  };

  const handleVersionUploaded = () => {
    // Перезагружаем и версии, и метаданные (версия могла измениться)
    Promise.all([
      documentsAPI.getVersions(id),
      documentsAPI.getMeta(id)
    ])
    .then(([verRes, metaRes]) => {
      setVersions(Array.isArray(verRes.data) ? verRes.data : []);
      setMeta(metaRes.data);
    })
    .catch(err => console.error('Failed to refresh document:', err));
  };

  const handleMetadataUpdated = (newData) => {
    // Обновляем локальное состояние мета-информации
    setMeta(prev => prev ? { ...prev, ...newData } : prev);
  };

  // Состояния
  if (loading) return <DocumentLoading />;
  if (error || !meta) return <DocumentError message={error} onBack={() => navigate('/')} />;

  // Данные
  const title = getField(meta, 'Title', 'title', 'OriginalFilename', 'original_filename') || 'Без названия';
  const version = getField(meta, 'CurrentVersion', 'current_version') || 1;
  const isPublic = getField(meta, 'IsPublic', 'is_public');
  const mime = getField(meta, 'MimeType', 'mime_type') || '';
  const size = formatSize(getField(meta, 'FileSize', 'file_size'));
  const isOwner = isDocumentOwner(meta);

  return (
    <div className="page page--document">
      <DocumentHeader
        title={title}
        version={version}
        isPublic={isPublic}
        mime={mime}
        onDownload={() => handleDownload()}
        onDelete={handleDelete}
        isOwner={isOwner}
        onBack={() => navigate('/')}
      />

      <main className="container page__content">
        <DocumentHero
          title={title}
          version={version}
          isPublic={isPublic}
          mime={mime}
          size={size}
        />

        <DocumentTabs
          activeTab={activeTab}
          onTabChange={setActiveTab}
          versionsCount={versions.length}
        />

        {activeTab === 'info' ? (
          <div className="document-layout">
            {/* Редактор метаданных (название + приватность) */}
            <DocumentMetadataEditor
              documentId={id}
              initialTitle={title}
              initialIsPublic={isPublic}
              onUpdate={handleMetadataUpdated}
              isOwner={isOwner}
            />
            
            {/* Информация о документе */}
            <DocumentInfoCard meta={meta} />
            
            {/* Панель быстрых действий */}
            <DocumentActions
              onDownload={() => handleDownload()}
              onShowVersions={() => setActiveTab('versions')}
              onDelete={handleDelete}
              onUploadVersion={() => setShowUploadModal(true)}
              isOwner={isOwner}
            />
          </div>
        ) : (
          <div className="card card--versions">
            <h3 className="card__title">История версий</h3>
            <VersionsList
              versions={versions}
              currentVersion={version}
              onDownload={handleDownload}
              onRollback={handleRollback}
            />
          </div>
        )}
      </main>
      {/* Модальное окно загрузки новой версии */}
      <UploadVersionModal
        isOpen={showUploadModal}
        onClose={() => setShowUploadModal(false)}
        documentId={id}
        onSuccess={handleVersionUploaded}
      />
    </div>
  );
}
import { useEffect, useMemo, useRef, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  Archive,
  ArchiveRestore,
  ChevronDown,
  Copy,
  Filter,
  Globe,
  Lock,
  MoreVertical,
  Plus,
  Search,
  Star,
  Trash2,
} from 'lucide-react';
import { Button, Modal } from '../components/ui';
import { workspaceService } from '../services/workspaceService';
import { projectService } from '../services/projectService';
import { pageService } from '../services/pageService';
import { useAuth } from '../contexts/AuthContext';
import type { WorkspaceApiResponse, ProjectApiResponse, PageApiResponse } from '../api/types';

type PageTab = 'public' | 'private' | 'archived';
type SortKey = 'updated' | 'created' | 'title';

const SORT_LABELS: Record<SortKey, string> = {
  updated: 'Date modified',
  created: 'Date created',
  title: 'Title (A→Z)',
};

export function PagesPage() {
  const { workspaceSlug, projectId } = useParams<{
    workspaceSlug: string;
    projectId: string;
  }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const [tab, setTab] = useState<PageTab>('public');
  const [workspace, setWorkspace] = useState<WorkspaceApiResponse | null>(null);
  const [project, setProject] = useState<ProjectApiResponse | null>(null);
  const [pages, setPages] = useState<PageApiResponse[]>([]);
  const [favoriteIds, setFavoriteIds] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);

  const [search, setSearch] = useState('');
  const [searchOpen, setSearchOpen] = useState(false);
  const [sortKey, setSortKey] = useState<SortKey>('updated');
  const [sortOpen, setSortOpen] = useState(false);
  const [filterOpen, setFilterOpen] = useState(false);
  const [filterLocked, setFilterLocked] = useState(false);
  const [filterOwnedByMe, setFilterOwnedByMe] = useState(false);
  const [openMenu, setOpenMenu] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [createName, setCreateName] = useState('');
  const [createAccess, setCreateAccess] = useState<0 | 1>(0);
  const [creating, setCreating] = useState(false);

  const sortRef = useRef<HTMLDivElement>(null);
  const filterRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  // ----- Initial load ------------------------------------------------------
  useEffect(() => {
    if (!workspaceSlug || !projectId) {
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    const archived = tab === 'archived' ? 'archived' : 'inbox';
    Promise.all([
      workspaceService.getBySlug(workspaceSlug),
      projectService.get(workspaceSlug, projectId),
      pageService.list(workspaceSlug, { projectId, archived }),
      pageService.listFavoriteIds(workspaceSlug).catch(() => [] as string[]),
    ])
      .then(([w, p, list, favIds]) => {
        if (cancelled) return;
        setWorkspace(w ?? null);
        setProject(p ?? null);
        setPages(list ?? []);
        setFavoriteIds(new Set(favIds));
      })
      .catch(() => {
        if (cancelled) return;
        setWorkspace(null);
        setProject(null);
        setPages([]);
        setFavoriteIds(new Set());
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [workspaceSlug, projectId, tab]);

  // ----- Click-outside dismissal -------------------------------------------
  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      const t = e.target as Node;
      if (sortOpen && sortRef.current && !sortRef.current.contains(t)) setSortOpen(false);
      if (filterOpen && filterRef.current && !filterRef.current.contains(t)) setFilterOpen(false);
      if (openMenu && menuRef.current && !menuRef.current.contains(t)) setOpenMenu(null);
    };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, [sortOpen, filterOpen, openMenu]);

  // ----- Derived list ------------------------------------------------------
  const visible = useMemo(() => {
    let out = pages.slice();
    if (tab === 'public') out = out.filter((p) => p.access === 0);
    if (tab === 'private') out = out.filter((p) => p.access === 1);
    // 'archived' tab — archived rows already returned by the server
    if (filterLocked) out = out.filter((p) => p.is_locked);
    if (filterOwnedByMe && user) out = out.filter((p) => p.owned_by_id === user.id);
    const q = search.trim().toLowerCase();
    if (q) out = out.filter((p) => (p.name ?? '').toLowerCase().includes(q));
    out.sort((a, b) => {
      switch (sortKey) {
        case 'created':
          return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
        case 'title':
          return (a.name ?? '').localeCompare(b.name ?? '');
        case 'updated':
        default:
          return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
      }
    });
    return out;
  }, [pages, tab, filterLocked, filterOwnedByMe, user, search, sortKey]);

  // ----- Actions -----------------------------------------------------------
  const reload = async () => {
    if (!workspaceSlug || !projectId) return;
    const archived = tab === 'archived' ? 'archived' : 'inbox';
    const list = await pageService.list(workspaceSlug, { projectId, archived });
    setPages(list ?? []);
  };

  const onToggleFavorite = async (pageId: string) => {
    if (!workspaceSlug) return;
    const isFav = favoriteIds.has(pageId);
    const next = new Set(favoriteIds);
    if (isFav) next.delete(pageId);
    else next.add(pageId);
    setFavoriteIds(next);
    try {
      if (isFav) await pageService.unfavorite(workspaceSlug, pageId);
      else await pageService.favorite(workspaceSlug, pageId);
    } catch {
      // revert optimistic toggle
      setFavoriteIds((prev) => {
        const r = new Set(prev);
        if (isFav) r.add(pageId);
        else r.delete(pageId);
        return r;
      });
    }
  };

  const onArchiveAction = async (pageId: string, currentlyArchived: boolean) => {
    if (!workspaceSlug) return;
    setOpenMenu(null);
    try {
      if (currentlyArchived) await pageService.unarchive(workspaceSlug, pageId);
      else await pageService.archive(workspaceSlug, pageId);
      await reload();
    } catch {
      // best-effort
    }
  };

  const onDuplicate = async (pageId: string) => {
    if (!workspaceSlug || !projectId) return;
    setOpenMenu(null);
    try {
      const dup = await pageService.duplicate(workspaceSlug, pageId);
      navigate(`/${workspaceSlug}/projects/${projectId}/pages/${dup.id}`);
    } catch {
      // best-effort
    }
  };

  const onDelete = async (pageId: string) => {
    if (!workspaceSlug) return;
    setOpenMenu(null);
    try {
      await pageService.delete(workspaceSlug, pageId);
      await reload();
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Delete failed';
      window.alert(message);
    }
  };

  const onCreate = async () => {
    if (!workspaceSlug || !projectId) return;
    setCreating(true);
    try {
      const created = await pageService.create(workspaceSlug, {
        name: createName.trim() || 'Untitled page',
        project_id: projectId,
        access: createAccess,
      });
      setShowCreate(false);
      setCreateName('');
      setCreateAccess(0);
      navigate(`/${workspaceSlug}/projects/${projectId}/pages/${created.id}`);
    } catch {
      // best-effort
    } finally {
      setCreating(false);
    }
  };

  // ----- Render ------------------------------------------------------------
  if (loading) {
    return (
      <div className="flex items-center justify-center p-8 text-sm text-(--txt-tertiary)">
        Loading…
      </div>
    );
  }
  if (!workspace || !project) {
    return <div className="text-(--txt-secondary)">Project not found.</div>;
  }

  const baseUrl = `/${workspace.slug}/projects/${project.id}`;

  return (
    <div className="space-y-4">
      {/* Tabs + Create */}
      <div className="flex items-center justify-between border-b border-(--border-subtle)">
        <div className="flex gap-1">
          {(['public', 'private', 'archived'] as const).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setTab(t)}
              className={`border-b-2 px-4 py-2.5 text-sm font-medium capitalize ${
                tab === t
                  ? 'border-(--brand-default) text-(--txt-primary)'
                  : 'border-transparent text-(--txt-secondary) hover:text-(--txt-primary)'
              }`}
            >
              {t}
            </button>
          ))}
        </div>
        <Button size="sm" variant="primary" onClick={() => setShowCreate(true)}>
          <Plus size={14} /> Create page
        </Button>
      </div>

      {/* Toolbar */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          {searchOpen ? (
            <input
              autoFocus
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onBlur={() => {
                if (!search) setSearchOpen(false);
              }}
              placeholder="Search pages…"
              className="h-8 rounded-md border border-(--border-subtle) bg-(--bg-layer-2) px-3 text-sm text-(--txt-primary) placeholder:text-(--txt-tertiary) focus:outline-none"
            />
          ) : (
            <button
              type="button"
              onClick={() => setSearchOpen(true)}
              className="flex size-8 items-center justify-center rounded-md border border-(--border-subtle) bg-(--bg-layer-2) text-(--txt-icon-tertiary) hover:bg-(--bg-layer-2-hover)"
              aria-label="Search"
            >
              <Search size={14} />
            </button>
          )}
        </div>
        <div className="flex items-center gap-2">
          <div ref={sortRef} className="relative">
            <button
              type="button"
              onClick={() => setSortOpen((v) => !v)}
              className="flex items-center gap-1.5 rounded-md border border-(--border-subtle) bg-(--bg-layer-2) px-2.5 py-1.5 text-[13px] font-medium text-(--txt-secondary) hover:bg-(--bg-layer-2-hover)"
            >
              {SORT_LABELS[sortKey]} <ChevronDown size={14} />
            </button>
            {sortOpen ? (
              <div className="absolute right-0 z-10 mt-1 w-44 rounded-md border border-(--border-subtle) bg-(--bg-surface-1) py-1 shadow-(--shadow-raised)">
                {(['updated', 'created', 'title'] as SortKey[]).map((k) => (
                  <button
                    key={k}
                    type="button"
                    onClick={() => {
                      setSortKey(k);
                      setSortOpen(false);
                    }}
                    className={`block w-full px-3 py-1.5 text-left text-sm hover:bg-(--bg-layer-1-hover) ${
                      sortKey === k ? 'text-(--txt-primary)' : 'text-(--txt-secondary)'
                    }`}
                  >
                    {SORT_LABELS[k]}
                  </button>
                ))}
              </div>
            ) : null}
          </div>
          <div ref={filterRef} className="relative">
            <button
              type="button"
              onClick={() => setFilterOpen((v) => !v)}
              className="flex items-center gap-1.5 rounded-md border border-(--border-subtle) bg-(--bg-layer-2) px-2.5 py-1.5 text-[13px] font-medium text-(--txt-secondary) hover:bg-(--bg-layer-2-hover)"
            >
              <Filter size={14} /> Filters <ChevronDown size={14} />
            </button>
            {filterOpen ? (
              <div className="absolute right-0 z-10 mt-1 w-52 rounded-md border border-(--border-subtle) bg-(--bg-surface-1) p-2 shadow-(--shadow-raised)">
                <label className="flex items-center gap-2 px-2 py-1.5 text-sm text-(--txt-primary)">
                  <input
                    type="checkbox"
                    checked={filterLocked}
                    onChange={(e) => setFilterLocked(e.target.checked)}
                  />
                  Locked
                </label>
                <label className="flex items-center gap-2 px-2 py-1.5 text-sm text-(--txt-primary)">
                  <input
                    type="checkbox"
                    checked={filterOwnedByMe}
                    onChange={(e) => setFilterOwnedByMe(e.target.checked)}
                  />
                  Owned by me
                </label>
              </div>
            ) : null}
          </div>
        </div>
      </div>

      {/* Page list */}
      <div className="rounded-md border border-(--border-subtle) bg-(--bg-surface-1)">
        {visible.length === 0 ? (
          <p className="px-4 py-8 text-center text-sm text-(--txt-tertiary)">No {tab} pages yet.</p>
        ) : (
          <ul className="divide-y divide-(--border-subtle)">
            {visible.map((page) => {
              const isFav = favoriteIds.has(page.id);
              const isOwner = !!user && page.owned_by_id === user.id;
              const isArchivedRow = !!page.archived_at;
              return (
                <li key={page.id} className="group relative">
                  <Link
                    to={`${baseUrl}/pages/${page.id}`}
                    className="flex items-center gap-3 px-4 py-3 no-underline transition-colors hover:bg-(--bg-layer-1-hover)"
                  >
                    <span className="flex size-9 shrink-0 items-center justify-center rounded border border-(--border-subtle) bg-(--bg-layer-2) text-(--txt-icon-tertiary)">
                      {isArchivedRow ? (
                        <Archive size={16} />
                      ) : page.is_locked ? (
                        <Lock size={16} />
                      ) : (
                        <Globe size={16} />
                      )}
                    </span>
                    <span className="min-w-0 flex-1 truncate font-medium text-(--txt-primary)">
                      {page.name || 'Untitled'}
                    </span>
                    <div className="flex shrink-0 items-center gap-1 text-(--txt-icon-tertiary)">
                      <button
                        type="button"
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          void onToggleFavorite(page.id);
                        }}
                        className="flex size-8 items-center justify-center rounded hover:bg-(--bg-layer-1-hover) hover:text-(--txt-icon-secondary)"
                        title={isFav ? 'Unfavorite' : 'Favorite'}
                      >
                        <Star
                          size={14}
                          className={isFav ? 'fill-(--brand-default) text-(--brand-default)' : ''}
                        />
                      </button>
                      <button
                        type="button"
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          setOpenMenu((m) => (m === page.id ? null : page.id));
                        }}
                        aria-label="More options"
                        className="flex size-8 items-center justify-center rounded hover:bg-(--bg-layer-1-hover) hover:text-(--txt-icon-secondary)"
                      >
                        <MoreVertical size={14} />
                      </button>
                    </div>
                  </Link>
                  {openMenu === page.id ? (
                    <div
                      ref={menuRef}
                      className="absolute top-12 right-4 z-10 w-48 rounded-md border border-(--border-subtle) bg-(--bg-surface-1) py-1 shadow-(--shadow-raised)"
                    >
                      <button
                        type="button"
                        onClick={() => void onDuplicate(page.id)}
                        className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-(--txt-primary) hover:bg-(--bg-layer-1-hover)"
                      >
                        <Copy size={14} /> Duplicate
                      </button>
                      {isOwner ? (
                        <button
                          type="button"
                          onClick={() => void onArchiveAction(page.id, isArchivedRow)}
                          className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-(--txt-primary) hover:bg-(--bg-layer-1-hover)"
                        >
                          {isArchivedRow ? <ArchiveRestore size={14} /> : <Archive size={14} />}
                          {isArchivedRow ? 'Unarchive' : 'Archive'}
                        </button>
                      ) : null}
                      {isOwner && isArchivedRow ? (
                        <button
                          type="button"
                          onClick={() => void onDelete(page.id)}
                          className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-(--danger-default) hover:bg-(--danger-50)"
                        >
                          <Trash2 size={14} /> Delete
                        </button>
                      ) : null}
                    </div>
                  ) : null}
                </li>
              );
            })}
          </ul>
        )}
      </div>

      {showCreate ? (
        <Modal open onClose={() => setShowCreate(false)} title="Create page">
          <div className="max-w-md space-y-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-(--txt-secondary)">Title</label>
              <input
                autoFocus
                type="text"
                value={createName}
                onChange={(e) => setCreateName(e.target.value)}
                placeholder="Untitled page"
                className="w-full rounded border border-(--border-subtle) bg-(--bg-canvas) px-3 py-1.5 text-sm text-(--txt-primary) placeholder:text-(--txt-tertiary) focus:outline-none"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-(--txt-secondary)">
                Visibility
              </label>
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => setCreateAccess(0)}
                  className={`flex-1 rounded border px-3 py-1.5 text-sm ${
                    createAccess === 0
                      ? 'border-(--brand-default) text-(--txt-primary)'
                      : 'border-(--border-subtle) text-(--txt-secondary)'
                  }`}
                >
                  Public
                </button>
                <button
                  type="button"
                  onClick={() => setCreateAccess(1)}
                  className={`flex-1 rounded border px-3 py-1.5 text-sm ${
                    createAccess === 1
                      ? 'border-(--brand-default) text-(--txt-primary)'
                      : 'border-(--border-subtle) text-(--txt-secondary)'
                  }`}
                >
                  Private
                </button>
              </div>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <Button size="sm" variant="secondary" onClick={() => setShowCreate(false)}>
                Cancel
              </Button>
              <Button
                size="sm"
                variant="primary"
                disabled={creating}
                onClick={() => void onCreate()}
              >
                {creating ? 'Creating…' : 'Create'}
              </Button>
            </div>
          </div>
        </Modal>
      ) : null}
    </div>
  );
}

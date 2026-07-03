import { Avatar } from '../ui';
import { ProjectIconDisplay } from '../ProjectIconModal';
import { getImageUrl } from '../../lib/utils';
import {
  ACCOUNT_SECTIONS_DEVELOPER,
  ACCOUNT_SECTIONS_PROFILE,
  PROJECT_SECTIONS,
  WORKSPACE_SECTIONS,
  type AccountSettingsSection,
  type ProjectSettingsSection,
  type WorkspaceSettingsSection,
} from './sections-config';
import type { ProjectApiResponse, WorkspaceApiResponse } from '../../api/types';
import type { User } from '../../types';

interface SettingsNavProps {
  isAccountTab: boolean;
  isProjectsTab: boolean;
  user: User | null;
  displayName: string;
  accountSection: AccountSettingsSection;
  setAccountSection: (s: AccountSettingsSection) => void;
  workspace: WorkspaceApiResponse;
  projects: ProjectApiResponse[];
  selectedProjectId: string | null;
  setProjectId: (projectId: string) => void;
  projectSection: ProjectSettingsSection;
  setProjectSection: (projectId: string, s: ProjectSettingsSection) => void;
  section: WorkspaceSettingsSection;
  setSection: (s: WorkspaceSettingsSection) => void;
}

/** Sidebar tab/section navigation for the Settings page (Account / Projects / Workspace). */
export function SettingsNav({
  isAccountTab,
  isProjectsTab,
  user,
  displayName,
  accountSection,
  setAccountSection,
  workspace,
  projects,
  selectedProjectId,
  setProjectId,
  projectSection,
  setProjectSection,
  section,
  setSection,
}: SettingsNavProps) {
  return (
    <aside className="w-56 shrink-0 space-y-6">
      {isAccountTab ? (
        <>
          <div className="flex items-center gap-2">
            <Avatar name={user?.name ?? ''} src={getImageUrl(user?.avatarUrl)} size="sm" />
            <div className="min-w-0">
              <p className="truncate text-sm font-medium text-(--txt-primary)">
                {displayName || user?.name}
              </p>
              <p className="truncate text-xs text-(--txt-tertiary)">{user?.email}</p>
            </div>
          </div>
          <nav className="space-y-0.5">
            <p className="mb-2 px-2 text-xs font-medium uppercase tracking-wider text-(--txt-tertiary)">
              Your Profile
            </p>
            {ACCOUNT_SECTIONS_PROFILE.map(({ id, label, icon }) => (
              <button
                key={id}
                type="button"
                onClick={() => setAccountSection(id)}
                className={`flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm font-medium transition-colors ${
                  accountSection === id
                    ? 'bg-(--brand-200) text-(--txt-primary)'
                    : 'text-(--txt-secondary) hover:bg-(--bg-layer-transparent-hover) hover:text-(--txt-primary)'
                }`}
              >
                <span className="text-(--txt-icon-secondary)">{icon}</span>
                {label}
              </button>
            ))}
            <p className="mb-2 mt-4 px-2 text-xs font-medium uppercase tracking-wider text-(--txt-tertiary)">
              Developer
            </p>
            {ACCOUNT_SECTIONS_DEVELOPER.map(({ id, label, icon }) => (
              <button
                key={id}
                type="button"
                onClick={() => setAccountSection(id)}
                className={`flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm font-medium transition-colors ${
                  accountSection === id
                    ? 'bg-(--brand-200) text-(--txt-primary)'
                    : 'text-(--txt-secondary) hover:bg-(--bg-layer-transparent-hover) hover:text-(--txt-primary)'
                }`}
              >
                <span className="text-(--txt-icon-secondary)">{icon}</span>
                {label}
              </button>
            ))}
          </nav>
        </>
      ) : isProjectsTab ? (
        <>
          <div className="flex items-center gap-2">
            <Avatar
              name={workspace.name}
              src={getImageUrl(workspace?.logo) ?? undefined}
              size="sm"
            />
            <div className="min-w-0">
              <p className="truncate text-sm font-medium text-(--txt-primary)">{workspace.name}</p>
              <p className="text-xs text-(--txt-tertiary)">Admin</p>
            </div>
          </div>
          <nav className="space-y-0.5">
            <p className="mb-2 px-2 text-xs font-medium uppercase tracking-wider text-(--txt-tertiary)">
              Projects
            </p>
            {projects.map((proj) => {
              const isSelected = selectedProjectId === proj.id;
              return (
                <div key={proj.id} className="space-y-0.5">
                  <button
                    type="button"
                    onClick={() => setProjectId(proj.id)}
                    className={`flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm font-medium transition-colors ${
                      isSelected
                        ? 'bg-(--brand-200) text-(--txt-primary)'
                        : 'text-(--txt-secondary) hover:bg-(--bg-layer-transparent-hover) hover:text-(--txt-primary)'
                    }`}
                  >
                    <span className="text-(--txt-icon-secondary) flex shrink-0 items-center justify-center">
                      <ProjectIconDisplay emoji={proj.emoji} icon_prop={proj.icon_prop} size={16} />
                    </span>
                    <span className="min-w-0 truncate">{proj.name}</span>
                    <span className="ml-auto shrink-0 text-xs text-(--txt-tertiary)">Admin</span>
                  </button>
                  {isSelected && (
                    <div className="ml-4 space-y-0.5 border-l border-(--border-subtle) pl-2">
                      {PROJECT_SECTIONS.map(({ id, label, icon }) => (
                        <button
                          key={id}
                          type="button"
                          onClick={() => setProjectSection(proj.id, id)}
                          className={`flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-sm font-medium transition-colors ${
                            projectSection === id
                              ? 'bg-(--brand-200) text-(--txt-primary)'
                              : 'text-(--txt-secondary) hover:bg-(--bg-layer-transparent-hover) hover:text-(--txt-primary)'
                          }`}
                        >
                          <span className="text-(--txt-icon-secondary)">{icon}</span>
                          {label}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              );
            })}
          </nav>
        </>
      ) : (
        <>
          <div className="flex items-center gap-2">
            <Avatar
              name={workspace.name}
              src={getImageUrl(workspace?.logo) ?? undefined}
              size="sm"
            />
            <div className="min-w-0">
              <p className="truncate text-sm font-medium text-(--txt-primary)">{workspace.name}</p>
              <p className="text-xs text-(--txt-tertiary)">Admin</p>
            </div>
          </div>
          <nav className="space-y-0.5">
            <p className="mb-2 px-2 text-xs font-medium uppercase tracking-wider text-(--txt-tertiary)">
              Administration
            </p>
            {WORKSPACE_SECTIONS.slice(0, 4).map(({ id, label, icon }) => (
              <button
                key={id}
                type="button"
                onClick={() => setSection(id)}
                className={`flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm font-medium transition-colors ${
                  section === id
                    ? 'bg-(--brand-200) text-(--txt-primary)'
                    : 'text-(--txt-secondary) hover:bg-(--bg-layer-transparent-hover) hover:text-(--txt-primary)'
                }`}
              >
                <span className="text-(--txt-icon-secondary)">{icon}</span>
                {label}
              </button>
            ))}
            <p className="mb-2 mt-4 px-2 text-xs font-medium uppercase tracking-wider text-(--txt-tertiary)">
              Developer
            </p>
            {WORKSPACE_SECTIONS.slice(4).map(({ id, label, icon }) => (
              <button
                key={id}
                type="button"
                onClick={() => setSection(id)}
                className={`flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm font-medium transition-colors ${
                  section === id
                    ? 'bg-(--brand-200) text-(--txt-primary)'
                    : 'text-(--txt-secondary) hover:bg-(--bg-layer-transparent-hover) hover:text-(--txt-primary)'
                }`}
              >
                <span className="text-(--txt-icon-secondary)">{icon}</span>
                {label}
              </button>
            ))}
          </nav>
        </>
      )}
    </aside>
  );
}

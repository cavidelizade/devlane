import { Button, Modal } from '../../ui';
import { labelService } from '../../../services/labelService';
import type { LabelApiResponse } from '../../../api/types';

interface ProjectLabelModalProps {
  open: boolean;
  onClose: () => void;
  workspaceSlug: string | undefined;
  selectedProjectId: string | null;
  projectLabelEdit: LabelApiResponse | null;
  setProjectLabelEdit: (label: LabelApiResponse | null) => void;
  projectLabelName: string;
  setProjectLabelName: (value: string) => void;
  projectLabelColor: string;
  setProjectLabelColor: (value: string) => void;
  setProjectLabels: (labels: LabelApiResponse[]) => void;
  setProjectLabelModalOpen: (value: boolean) => void;
}

/** Add/edit modal for a project label. */
export function ProjectLabelModal({
  open,
  onClose,
  workspaceSlug,
  selectedProjectId,
  projectLabelEdit,
  setProjectLabelEdit,
  projectLabelName,
  setProjectLabelName,
  projectLabelColor,
  setProjectLabelColor,
  setProjectLabels,
  setProjectLabelModalOpen,
}: ProjectLabelModalProps) {
  return (
    <Modal
      open={open}
      onClose={onClose}
      title={projectLabelEdit ? 'Edit label' : 'Add label'}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button
            disabled={!projectLabelName.trim()}
            onClick={async () => {
              if (!workspaceSlug || !selectedProjectId || !projectLabelName.trim()) return;
              try {
                if (projectLabelEdit) {
                  await labelService.update(workspaceSlug, selectedProjectId, projectLabelEdit.id, {
                    name: projectLabelName.trim(),
                    color: projectLabelColor,
                  });
                } else {
                  await labelService.create(workspaceSlug, selectedProjectId, {
                    name: projectLabelName.trim(),
                    color: projectLabelColor,
                  });
                }
                const list = await labelService.list(workspaceSlug, selectedProjectId);
                setProjectLabels(list ?? []);
                setProjectLabelModalOpen(false);
                setProjectLabelEdit(null);
              } catch {
                // Intentionally empty (kept for future use)
              }
            }}
          >
            {projectLabelEdit ? 'Save' : 'Create'}
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        <div>
          <label className="mb-1 block text-sm font-medium text-(--txt-secondary)">Name</label>
          <input
            type="text"
            value={projectLabelName}
            onChange={(e) => setProjectLabelName(e.target.value)}
            className="w-full rounded-(--radius-md) border border-(--border-subtle) bg-(--bg-surface-1) px-3 py-2 text-sm text-(--txt-primary) focus:outline-none focus:border-(--border-strong)"
            placeholder="e.g. Bug"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-(--txt-secondary)">Color</label>
          <div className="flex items-center gap-2">
            <input
              type="color"
              value={projectLabelColor}
              onChange={(e) => setProjectLabelColor(e.target.value)}
              className="h-9 w-14 cursor-pointer rounded border border-(--border-subtle)"
            />
            <input
              type="text"
              value={projectLabelColor}
              onChange={(e) => setProjectLabelColor(e.target.value)}
              className="flex-1 rounded-(--radius-md) border border-(--border-subtle) bg-(--bg-surface-1) px-3 py-2 text-sm text-(--txt-primary) focus:outline-none focus:border-(--border-strong)"
            />
          </div>
        </div>
      </div>
    </Modal>
  );
}

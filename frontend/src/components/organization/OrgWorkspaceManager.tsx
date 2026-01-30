import { useState, useEffect, useCallback } from 'react';
import { Plus } from 'lucide-react';
import { useOrgWorkspaceStore } from '../../store/orgWorkspaceStore';
import { useOrganizationStore } from '../../store/organizationStore';
import { toast } from '../../store/toastStore';
import { OrgWorkspaceList } from './OrgWorkspaceList';
import { OrgWorkspaceModal } from './OrgWorkspaceModal';
import { ConfirmDialog } from '../ConfirmDialog';
import type { OrgWorkspace, CreateOrgWorkspaceInput } from '../../types/organization';

const PAGE_SIZE = 10;

export function OrgWorkspaceManager() {
  const currentOrg = useOrganizationStore((state) => state.currentOrg);
  const {
    workspaces,
    isLoading,
    total,
    hasMore,
    fetchWorkspaces,
    createWorkspace,
    updateWorkspace,
    deleteWorkspace,
  } = useOrgWorkspaceStore();

  const [currentPage, setCurrentPage] = useState(0);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingWorkspace, setEditingWorkspace] = useState<OrgWorkspace | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<OrgWorkspace | null>(null);

  const orgId = currentOrg?.id;

  const loadWorkspaces = useCallback(
    (page: number) => {
      if (!orgId) return;
      fetchWorkspaces(orgId, PAGE_SIZE, page * PAGE_SIZE);
    },
    [orgId, fetchWorkspaces]
  );

  useEffect(() => {
    loadWorkspaces(currentPage);
  }, [loadWorkspaces, currentPage]);

  const handleCreate = () => {
    setEditingWorkspace(null);
    setIsModalOpen(true);
  };

  const handleEdit = (workspace: OrgWorkspace) => {
    setEditingWorkspace(workspace);
    setIsModalOpen(true);
  };

  const handleDelete = (workspace: OrgWorkspace) => {
    setDeleteTarget(workspace);
  };

  const handleOpen = (workspace: OrgWorkspace) => {
    window.open(`/workspace/${workspace.id}`, '_blank');
  };

  const handleModalClose = () => {
    setIsModalOpen(false);
    setEditingWorkspace(null);
  };

  const handleModalSubmit = async (data: CreateOrgWorkspaceInput) => {
    if (!orgId) return;

    setIsSubmitting(true);
    try {
      if (editingWorkspace) {
        await updateWorkspace(orgId, editingWorkspace.id, data);
        toast.success('Workspace updated successfully');
      } else {
        await createWorkspace(orgId, data);
        toast.success('Workspace created successfully');
      }
      handleModalClose();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleConfirmDelete = async () => {
    if (!orgId || !deleteTarget) return;

    try {
      await deleteWorkspace(orgId, deleteTarget.id);
      toast.success('Workspace deleted successfully');
      setDeleteTarget(null);
      if (workspaces.length === 1 && currentPage > 0) {
        setCurrentPage(currentPage - 1);
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete workspace');
    }
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  if (!orgId) {
    return (
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-editor-text">Workspaces</h2>
        <div className="bg-editor-surface border border-editor-border rounded-lg p-8 text-center">
          <p className="text-editor-muted">No organization selected</p>
        </div>
      </section>
    );
  }

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-editor-text">Workspaces</h2>
        <button
          onClick={handleCreate}
          className="flex items-center gap-2 px-3 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors text-sm"
        >
          <Plus size={16} />
          Create Workspace
        </button>
      </div>

      <div className="bg-editor-surface border border-editor-border rounded-lg overflow-hidden">
        <OrgWorkspaceList
          workspaces={workspaces}
          isLoading={isLoading}
          total={total}
          hasMore={hasMore}
          currentPage={currentPage}
          pageSize={PAGE_SIZE}
          onPageChange={handlePageChange}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onOpen={handleOpen}
        />
      </div>

      <OrgWorkspaceModal
        isOpen={isModalOpen}
        workspace={editingWorkspace}
        isSubmitting={isSubmitting}
        onClose={handleModalClose}
        onSubmit={handleModalSubmit}
      />

      <ConfirmDialog
        isOpen={!!deleteTarget}
        title="Delete Workspace"
        message={`Are you sure you want to delete "${deleteTarget?.name}"? This action cannot be undone.`}
        confirmText="Delete"
        cancelText="Cancel"
        variant="danger"
        onConfirm={handleConfirmDelete}
        onCancel={() => setDeleteTarget(null)}
      />
    </section>
  );
}

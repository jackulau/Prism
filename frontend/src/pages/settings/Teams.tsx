import { useState, useEffect } from 'react';
import {
  Users,
  Plus,
  Search,
  Loader2,
  X,
  ChevronRight,
  AlertCircle,
} from 'lucide-react';
import { useOrganizationStore } from '../../store/organizationStore';
import { useTeamStore } from '../../store/teamStore';
import { teamService } from '../../services/team';
import { toast } from '../../store/toastStore';
import { TeamDetail } from '../../components/settings/TeamDetail';
import type { Team } from '../../types/team';

export default function TeamsPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [newTeamName, setNewTeamName] = useState('');
  const [newTeamDescription, setNewTeamDescription] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [selectedTeamId, setSelectedTeamId] = useState<string | null>(null);

  const { currentOrg } = useOrganizationStore();
  const { teams, isLoadingTeams, teamsError, addTeam } = useTeamStore();

  // Load teams on mount
  useEffect(() => {
    if (currentOrg?.id) {
      teamService.loadTeams(currentOrg.id);
    }
  }, [currentOrg?.id]);

  // Filter teams by search query
  const filteredTeams = teams.filter(
    (team) =>
      team.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      team.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleCreateTeam = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentOrg?.id || !newTeamName.trim()) return;

    setIsSubmitting(true);
    const response = await teamService.createTeam(currentOrg.id, {
      name: newTeamName.trim(),
      description: newTeamDescription.trim() || undefined,
    });

    if (response.data) {
      addTeam(response.data);
      toast.success(`Team "${newTeamName}" created`);
      setNewTeamName('');
      setNewTeamDescription('');
      setIsCreating(false);
      setSelectedTeamId(response.data.id);
    } else if (response.error) {
      toast.error(response.error);
    }

    setIsSubmitting(false);
  };

  const selectedTeam = teams.find((t) => t.id === selectedTeamId);

  // If a team is selected, show the detail view
  if (selectedTeam) {
    return (
      <TeamDetail
        team={selectedTeam}
        onBack={() => setSelectedTeamId(null)}
      />
    );
  }

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold text-editor-text">Teams</h1>
            <p className="text-editor-muted">
              Manage your organization's teams and access permissions
            </p>
          </div>
          <button
            onClick={() => setIsCreating(true)}
            className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
          >
            <Plus size={16} />
            Create Team
          </button>
        </div>

        {/* Search Bar */}
        <div className="relative">
          <Search
            size={18}
            className="absolute left-3 top-1/2 -translate-y-1/2 text-editor-muted"
          />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search teams..."
            className="w-full pl-10 pr-4 py-2.5 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
          />
        </div>

        {/* Error State */}
        {teamsError && (
          <div className="flex items-center gap-3 p-4 bg-editor-error/10 border border-editor-error/20 rounded-lg">
            <AlertCircle size={20} className="text-editor-error flex-shrink-0" />
            <span className="text-editor-error">{teamsError}</span>
            <button
              onClick={() => currentOrg?.id && teamService.loadTeams(currentOrg.id)}
              className="ml-auto text-sm text-editor-accent hover:underline"
            >
              Retry
            </button>
          </div>
        )}

        {/* Loading State */}
        {isLoadingTeams ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 size={24} className="animate-spin text-editor-muted" />
            <span className="ml-3 text-editor-muted">Loading teams...</span>
          </div>
        ) : filteredTeams.length === 0 ? (
          <div className="py-16 text-center">
            <Users size={48} className="mx-auto mb-4 text-editor-muted opacity-50" />
            {searchQuery ? (
              <>
                <p className="text-lg text-editor-text mb-1">No teams found</p>
                <p className="text-editor-muted">
                  No teams match "{searchQuery}"
                </p>
              </>
            ) : (
              <>
                <p className="text-lg text-editor-text mb-1">No teams yet</p>
                <p className="text-editor-muted mb-6">
                  Create your first team to organize access to resources
                </p>
                <button
                  onClick={() => setIsCreating(true)}
                  className="inline-flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 transition-colors"
                >
                  <Plus size={16} />
                  Create Team
                </button>
              </>
            )}
          </div>
        ) : (
          /* Teams Grid */
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {filteredTeams.map((team) => (
              <TeamCard
                key={team.id}
                team={team}
                onClick={() => setSelectedTeamId(team.id)}
              />
            ))}
          </div>
        )}

        {/* Create Team Modal */}
        {isCreating && (
          <>
            <div
              className="fixed inset-0 bg-black/50 z-40"
              onClick={() => setIsCreating(false)}
            />
            <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[480px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-xl shadow-xl z-50">
              <div className="flex items-center justify-between px-6 py-4 border-b border-editor-border">
                <h3 className="text-lg font-semibold text-editor-text">
                  Create New Team
                </h3>
                <button
                  onClick={() => setIsCreating(false)}
                  className="p-2 text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded-lg transition-colors"
                >
                  <X size={20} />
                </button>
              </div>

              <form onSubmit={handleCreateTeam} className="p-6 space-y-4">
                <div className="space-y-2">
                  <label className="block text-sm font-medium text-editor-text">
                    Team Name
                  </label>
                  <input
                    type="text"
                    value={newTeamName}
                    onChange={(e) => setNewTeamName(e.target.value)}
                    placeholder="e.g., Engineering, Marketing"
                    className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent"
                    required
                    autoFocus
                  />
                </div>

                <div className="space-y-2">
                  <label className="block text-sm font-medium text-editor-text">
                    Description{' '}
                    <span className="text-editor-muted font-normal">(optional)</span>
                  </label>
                  <textarea
                    value={newTeamDescription}
                    onChange={(e) => setNewTeamDescription(e.target.value)}
                    placeholder="What does this team do?"
                    rows={3}
                    className="w-full px-3 py-2 bg-editor-surface border border-editor-border rounded-lg text-editor-text placeholder-editor-muted focus:outline-none focus:border-editor-accent resize-none"
                  />
                </div>

                <div className="flex items-center justify-end gap-3 pt-4">
                  <button
                    type="button"
                    onClick={() => setIsCreating(false)}
                    className="px-4 py-2 text-editor-muted hover:text-editor-text transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={isSubmitting || !newTeamName.trim()}
                    className="flex items-center gap-2 px-4 py-2 bg-editor-accent text-white rounded-lg hover:bg-editor-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    {isSubmitting ? (
                      <>
                        <Loader2 size={16} className="animate-spin" />
                        Creating...
                      </>
                    ) : (
                      <>
                        <Plus size={16} />
                        Create Team
                      </>
                    )}
                  </button>
                </div>
              </form>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

interface TeamCardProps {
  team: Team;
  onClick: () => void;
}

function TeamCard({ team, onClick }: TeamCardProps) {
  return (
    <button
      onClick={onClick}
      className="flex items-center gap-4 p-4 bg-editor-surface border border-editor-border rounded-lg hover:bg-editor-bg/50 transition-colors text-left group"
    >
      <div className="w-12 h-12 rounded-lg bg-editor-accent/10 flex items-center justify-center flex-shrink-0">
        <span className="text-lg font-semibold text-editor-accent">
          {team.name.charAt(0).toUpperCase()}
        </span>
      </div>
      <div className="flex-1 min-w-0">
        <div className="font-medium text-editor-text">{team.name}</div>
        {team.description && (
          <div className="text-sm text-editor-muted truncate">
            {team.description}
          </div>
        )}
        <div className="text-xs text-editor-muted mt-1">
          {team.memberCount} member{team.memberCount !== 1 ? 's' : ''}
        </div>
      </div>
      <ChevronRight
        size={20}
        className="text-editor-muted group-hover:text-editor-text transition-colors flex-shrink-0"
      />
    </button>
  );
}

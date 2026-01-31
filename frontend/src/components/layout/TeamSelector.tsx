import { useState, useRef, useEffect } from 'react';
import { ChevronDown, Users, Check, Loader2 } from 'lucide-react';
import { useTeamStore } from '../../store/teamStore';
import { useOrganizationStore } from '../../store/organizationStore';
import { teamService } from '../../services/team';
import type { Team } from '../../types/team';

interface TeamSelectorProps {
  className?: string;
}

export function TeamSelector({ className = '' }: TeamSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const { currentOrg } = useOrganizationStore();
  const {
    teams,
    currentTeam,
    showAllTeams,
    isLoadingTeams,
    setCurrentTeam,
    setShowAllTeams,
  } = useTeamStore();

  // Load teams when organization changes
  useEffect(() => {
    if (currentOrg?.id) {
      teamService.loadTeams(currentOrg.id);
    }
  }, [currentOrg?.id]);

  // Handle keyboard shortcut (Cmd/Ctrl + T)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 't') {
        e.preventDefault();
        setIsOpen((prev) => !prev);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  const handleSelectTeam = (team: Team | null) => {
    if (team) {
      setCurrentTeam(team);
    } else {
      setShowAllTeams(true);
    }
    setIsOpen(false);
  };

  const displayName = showAllTeams
    ? 'All Teams'
    : currentTeam?.name || 'Select Team';

  // Check if user has permission to see all teams (org admin/owner)
  const { members } = useOrganizationStore();
  const currentMember = members.find((m) => m.role === 'owner' || m.role === 'admin');
  const canSeeAllTeams = Boolean(currentMember);

  return (
    <div ref={dropdownRef} className={`relative ${className}`}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 bg-editor-surface border border-editor-border rounded-lg hover:bg-editor-bg/50 transition-colors min-w-[160px]"
        title="Switch team (Cmd+T)"
      >
        <Users size={16} className="text-editor-muted flex-shrink-0" />
        <span className="text-sm text-editor-text truncate flex-1 text-left">
          {displayName}
        </span>
        <ChevronDown
          size={16}
          className={`text-editor-muted transition-transform ${isOpen ? 'rotate-180' : ''}`}
        />
      </button>

      {isOpen && (
        <div className="absolute top-full left-0 mt-1 w-64 bg-editor-surface border border-editor-border rounded-lg shadow-lg z-50 overflow-hidden">
          {/* All Teams Option (for admins) */}
          {canSeeAllTeams && (
            <>
              <button
                onClick={() => handleSelectTeam(null)}
                className="w-full flex items-center gap-3 px-4 py-3 hover:bg-editor-bg/50 transition-colors"
              >
                <div className="w-8 h-8 rounded-lg bg-editor-accent/10 flex items-center justify-center">
                  <Users size={16} className="text-editor-accent" />
                </div>
                <div className="flex-1 text-left">
                  <div className="text-sm font-medium text-editor-text">All Teams</div>
                  <div className="text-xs text-editor-muted">View all resources</div>
                </div>
                {showAllTeams && (
                  <Check size={16} className="text-editor-accent flex-shrink-0" />
                )}
              </button>
              <div className="border-b border-editor-border" />
            </>
          )}

          {/* Loading State */}
          {isLoadingTeams ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 size={20} className="animate-spin text-editor-muted" />
            </div>
          ) : teams.length === 0 ? (
            <div className="py-8 text-center">
              <Users size={24} className="mx-auto mb-2 text-editor-muted" />
              <p className="text-sm text-editor-muted">No teams yet</p>
            </div>
          ) : (
            <div className="max-h-64 overflow-y-auto">
              {teams.map((team) => (
                <button
                  key={team.id}
                  onClick={() => handleSelectTeam(team)}
                  className="w-full flex items-center gap-3 px-4 py-3 hover:bg-editor-bg/50 transition-colors"
                >
                  <div className="w-8 h-8 rounded-lg bg-editor-accent/10 flex items-center justify-center">
                    <span className="text-sm font-medium text-editor-accent">
                      {team.name.charAt(0).toUpperCase()}
                    </span>
                  </div>
                  <div className="flex-1 text-left">
                    <div className="text-sm font-medium text-editor-text">
                      {team.name}
                    </div>
                    <div className="text-xs text-editor-muted">
                      {team.memberCount} member{team.memberCount !== 1 ? 's' : ''}
                    </div>
                  </div>
                  {currentTeam?.id === team.id && !showAllTeams && (
                    <Check size={16} className="text-editor-accent flex-shrink-0" />
                  )}
                </button>
              ))}
            </div>
          )}

          {/* Keyboard shortcut hint */}
          <div className="px-4 py-2 bg-editor-bg/50 border-t border-editor-border">
            <span className="text-xs text-editor-muted">
              Press <kbd className="px-1.5 py-0.5 bg-editor-surface rounded text-editor-muted">⌘T</kbd> to switch teams
            </span>
          </div>
        </div>
      )}
    </div>
  );
}

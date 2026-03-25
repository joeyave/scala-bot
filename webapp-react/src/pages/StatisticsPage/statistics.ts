import {
  StatisticsEvent,
  StatisticsRole,
  StatisticsUser,
} from "@/api/webapp/typesResp.ts";

export interface StatisticsFilters {
  fromDate: string;
  selectedRoleIds: string[];
  selectedWeekdays: number[];
}

export interface StatisticsRoleBreakdown {
  roleId: string;
  roleName: string;
  count: number;
}

export interface StatisticsRow {
  userId: number;
  name: string;
  participationCount: number;
  roleBreakdown: StatisticsRoleBreakdown[];
  lastEventDate: string | null;
}

export interface StatisticsModel {
  activeMembers: number;
  totalParticipations: number;
  rows: StatisticsRow[];
}

export function buildStatisticsModel(
  users: StatisticsUser[],
  roles: StatisticsRole[],
  filters: StatisticsFilters,
): StatisticsModel {
  const rows = users
    .map((user) => buildStatisticsRow(user, roles, filters))
    .filter((row): row is StatisticsRow => row !== null)
    .sort((left, right) => {
      if (right.participationCount !== left.participationCount) {
        return right.participationCount - left.participationCount;
      }

      return left.name.localeCompare(right.name);
    });

  return {
    activeMembers: rows.length,
    totalParticipations: rows.reduce(
      (total, row) => total + row.participationCount,
      0,
    ),
    rows,
  };
}

function buildStatisticsRow(
  user: StatisticsUser,
  roles: StatisticsRole[],
  filters: StatisticsFilters,
): StatisticsRow | null {
  const filteredEvents = filterEvents(user.events ?? [], filters);
  if (filteredEvents.length === 0) {
    return null;
  }

  const roleCounts: Record<string, number> = {};
  for (const event of filteredEvents) {
    for (const role of event.roles ?? []) {
      roleCounts[role.id] = (roleCounts[role.id] || 0) + 1;
    }
  }

  const roleBreakdown = roles
    .map((role) => ({
      roleId: role.id,
      roleName: role.name,
      count: roleCounts[role.id] || 0,
    }))
    .filter((role) => role.count > 0);

  const lastEventDate = filteredEvents.reduce(
    (max, event) => (event.date > max ? event.date : max),
    filteredEvents[0].date,
  );

  return {
    userId: user.id,
    name: user.name,
    participationCount: filteredEvents.length,
    roleBreakdown,
    lastEventDate,
  };
}

function filterEvents(
  events: StatisticsEvent[],
  filters: StatisticsFilters,
): StatisticsEvent[] {
  return events.filter((event) => {
    const roles = event.roles ?? [];
    const matchesFromDate =
      filters.fromDate.length === 0 || event.date >= filters.fromDate;
    const matchesRoles =
      filters.selectedRoleIds.length === 0 ||
      roles.some((role) => filters.selectedRoleIds.includes(role.id));
    const matchesWeekdays =
      filters.selectedWeekdays.length === 0 ||
      filters.selectedWeekdays.includes(event.weekday);

    return matchesFromDate && matchesRoles && matchesWeekdays;
  });
}

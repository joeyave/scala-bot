import { AppRoot } from "@telegram-apps/telegram-ui";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";

import { StatisticsState } from "./StatisticsPage";
import { buildStatisticsModel } from "./statistics";

describe("buildStatisticsModel", () => {
  it("keeps fromDate inclusive, combines filters, and sorts by participation count", () => {
    const roles = [
      { id: "role-vocal", name: "Vocal" },
      { id: "role-guitar", name: "Guitar" },
    ];

    const users = [
      {
        id: 1,
        name: "Alice",
        events: [
          {
            id: "event-1",
            date: "2026-02-01",
            weekday: 0,
            name: "Sunday Service",
            roles: [{ id: "role-vocal", name: "Vocal" }],
          },
          {
            id: "event-2",
            date: "2026-02-08",
            weekday: 0,
            name: "Sunday Service",
            roles: [{ id: "role-vocal", name: "Vocal" }],
          },
        ],
      },
      {
        id: 2,
        name: "Bob",
        events: [
          {
            id: "event-3",
            date: "2026-02-01",
            weekday: 0,
            name: "Sunday Service",
            roles: [{ id: "role-vocal", name: "Vocal" }],
          },
        ],
      },
      {
        id: 3,
        name: "Cara",
        events: [
          {
            id: "event-4",
            date: "2026-01-31",
            weekday: 6,
            name: "Saturday Rehearsal",
            roles: [{ id: "role-guitar", name: "Guitar" }],
          },
        ],
      },
    ];

    const statistics = buildStatisticsModel(users, roles, {
      fromDate: "2026-02-01",
      selectedRoleIds: ["role-vocal"],
      selectedWeekdays: [0],
    });

    expect(statistics.totalParticipations).toBe(3);
    expect(statistics.activeMembers).toBe(2);
    expect(statistics.rows.map((row) => row.name)).toEqual(["Alice", "Bob"]);
    expect(statistics.rows[0].participationCount).toBe(2);
    expect(statistics.rows[0].roleBreakdown).toEqual([
      { roleId: "role-vocal", roleName: "Vocal", count: 2 },
    ]);
  });
});

describe("StatisticsState", () => {
  it("renders loading state markup", () => {
    const markup = renderToStaticMarkup(
      <AppRoot appearance="light" platform="ios">
        <StatisticsState
          header="Loading"
          description="Fetching current statistics"
          loading
        />
      </AppRoot>,
    );

    expect(markup).toContain("Loading");
    expect(markup).toContain("Fetching current statistics");
  });

  it("renders empty and error state markup", () => {
    const emptyMarkup = renderToStaticMarkup(
      <AppRoot appearance="light" platform="ios">
        <StatisticsState
          header="Nothing found"
          description="No members match the selected filters"
        />
      </AppRoot>,
    );

    const errorMarkup = renderToStaticMarkup(
      <AppRoot appearance="light" platform="ios">
        <StatisticsState
          header="Statistics unavailable"
          description="Failed to load team data"
        />
      </AppRoot>,
    );

    expect(emptyMarkup).toContain("Nothing found");
    expect(emptyMarkup).toContain("No members match the selected filters");
    expect(errorMarkup).toContain("Statistics unavailable");
    expect(errorMarkup).toContain("Failed to load team data");
  });
});

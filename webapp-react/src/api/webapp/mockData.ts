import {
  RespEventData,
  RespEventFreqNames,
  RespStatistics,
} from "@/api/webapp/typesResp.ts";

// Mock data for development without backend
// Only used when VITE_USE_MOCK_DATA=true

export const mockEventData: RespEventData = {
  event: {
    id: "mock-event-id-123",
    time: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(), // 1 week from now
    name: "Sunday Worship Service",
    bandId: "mock-band-id-456",
    band: {
      id: "mock-band-id-456",
      name: "Worship Band",
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      driveFolderId: "mock-drive-folder-id",
      archiveFolderId: "mock-archive-folder-id",
      roles: [
        {
          id: "role-1",
          name: "Vocalist",
          band_id: "mock-band-id-456",
          priority: 1,
        },
        {
          id: "role-2",
          name: "Guitarist",
          band_id: "mock-band-id-456",
          priority: 2,
        },
        {
          id: "role-3",
          name: "Drummer",
          band_id: "mock-band-id-456",
          priority: 3,
        },
      ],
    },
    songIds: ["song-1", "song-2", "song-3"],
    songOverrides: [
      { songId: "song-2", eventKey: "D" }, // Original key is C, override to D
      { songId: "song-3", eventKey: "G" }, // Original key is D, override to G
    ],
    songs: [
      {
        id: "song-1",
        driveFileId: "drive-file-1",
        bandId: "mock-band-id-456",
        band: {
          id: "mock-band-id-456",
          name: "Worship Band",
          timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
          driveFolderId: "mock-drive-folder-id",
          archiveFolderId: "mock-archive-folder-id",
        },
        pdf: {
          version: 1,
          tgFileId: "tg-file-1",
          tgChannelMessageId: 1001,
          name: "Amazing Grace",
          key: "G",
          bpm: "72",
          time: "4/4",
          webViewLink: "https://example.com/song1",
        },
        tags: ["hymn", "classic"],
        isArchived: false,
      },
      {
        id: "song-2",
        driveFileId: "drive-file-2",
        bandId: "mock-band-id-456",
        band: {
          id: "mock-band-id-456",
          name: "Worship Band",
          timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
          driveFolderId: "mock-drive-folder-id",
          archiveFolderId: "mock-archive-folder-id",
        },
        pdf: {
          version: 1,
          tgFileId: "tg-file-2",
          tgChannelMessageId: 1002,
          name: "How Great Is Our God",
          key: "C",
          bpm: "78",
          time: "4/4",
          webViewLink: "https://example.com/song2",
        },
        tags: ["contemporary", "worship"],
        isArchived: false,
      },
      {
        id: "song-3",
        driveFileId: "drive-file-3",
        bandId: "mock-band-id-456",
        band: {
          id: "mock-band-id-456",
          name: "Worship Band",
          timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
          driveFolderId: "mock-drive-folder-id",
          archiveFolderId: "mock-archive-folder-id",
        },
        pdf: {
          version: 1,
          tgFileId: "tg-file-3",
          tgChannelMessageId: 1003,
          name: "Oceans",
          key: "D",
          bpm: "66",
          time: "4/4",
          webViewLink: "https://example.com/song3",
        },
        tags: ["contemporary", "hillsong"],
        isArchived: false,
      },
    ],
    notes: "Remember to practice the bridge section of Oceans.",
  },
};

export const mockEventFreqNames: RespEventFreqNames = {
  names: [
    "Sunday Worship Service",
    "Wednesday Night Service",
    "Youth Night",
    "Christmas Special",
    "Easter Service",
    "Prayer Meeting",
  ],
};

export const mockStatisticsData: RespStatistics = {
  bandName: "Worship Band",
  currentDate: "2026-03-25",
  defaultFromDate: "2025-09-25",
  roles: [
    { id: "role-1", name: "Vocalist" },
    { id: "role-2", name: "Guitarist" },
    { id: "role-3", name: "Drummer" },
  ],
  users: [
    {
      id: 1,
      name: "Anastasiia",
      events: [
        {
          id: "event-1",
          date: "2026-01-12",
          weekday: 1,
          name: "Monday Rehearsal",
          roles: [{ id: "role-1", name: "Vocalist" }],
        },
        {
          id: "event-2",
          date: "2026-02-09",
          weekday: 1,
          name: "Monday Rehearsal",
          roles: [{ id: "role-1", name: "Vocalist" }],
        },
        {
          id: "event-3",
          date: "2026-03-08",
          weekday: 0,
          name: "Sunday Service",
          roles: [{ id: "role-2", name: "Guitarist" }],
        },
      ],
    },
    {
      id: 2,
      name: "Maksym",
      events: [
        {
          id: "event-4",
          date: "2026-01-18",
          weekday: 0,
          name: "Sunday Service",
          roles: [{ id: "role-3", name: "Drummer" }],
        },
        {
          id: "event-5",
          date: "2026-02-22",
          weekday: 0,
          name: "Sunday Service",
          roles: [{ id: "role-3", name: "Drummer" }],
        },
      ],
    },
    {
      id: 3,
      name: "Kateryna",
      events: [
        {
          id: "event-6",
          date: "2026-02-05",
          weekday: 4,
          name: "Thursday Prayer",
          roles: [{ id: "role-1", name: "Vocalist" }],
        },
      ],
    },
  ],
};

// Helper to check if mocking is enabled
export function isMockEnabled(): boolean {
  return import.meta.env.DEV && import.meta.env.VITE_USE_MOCK_DATA === "true";
}

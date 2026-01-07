import { RespEventData, RespEventFreqNames } from "@/api/webapp/typesResp.ts";

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
        { id: "role-1", name: "Vocalist", band_id: "mock-band-id-456", priority: 1 },
        { id: "role-2", name: "Guitarist", band_id: "mock-band-id-456", priority: 2 },
        { id: "role-3", name: "Drummer", band_id: "mock-band-id-456", priority: 3 },
      ],
    },
    songIds: ["song-1", "song-2", "song-3"],
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
          modifiedTime: new Date().toISOString(),
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
          modifiedTime: new Date().toISOString(),
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
          modifiedTime: new Date().toISOString(),
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

// Helper to check if mocking is enabled
export function isMockEnabled(): boolean {
  return import.meta.env.DEV && import.meta.env.VITE_USE_MOCK_DATA === "true";
}

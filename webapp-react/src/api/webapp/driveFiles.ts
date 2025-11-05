import { doReqWebappApi } from "@/api/webapp/doReq.ts";
import { DriveFile, RespSearchDriveFiles } from "@/api/webapp/typesResp.ts";

export async function searchDriveFiles(
  query: string,
  driveFolderId?: string,
  archiveFolderId?: string | null,
): Promise<DriveFile[]> {
  const { data, err } = await doReqWebappApi<RespSearchDriveFiles>(
    `/api/v2/drive-files/search`,
    "GET",
    {
      q: query,
      driveFolderId: driveFolderId,
      archiveFolderId: archiveFolderId,
    },
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data?.driveFiles || [];
}

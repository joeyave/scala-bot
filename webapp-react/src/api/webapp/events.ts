import { doReqWebappApi } from "@/api/webapp/doReq.ts";
import { RespEventFreqNames } from "@/api/webapp/typesResp.ts";

export async function getEventFreqNames(bandId: string): Promise<RespEventFreqNames | null> {
  const { data, err } = await doReqWebappApi<RespEventFreqNames>(
    `/api/events/frequent-names`,
    "GET",
    { bandId },
    { Accept: "application/json" }
  );

  if (err) {
    throw err;
  }

  return data;
}

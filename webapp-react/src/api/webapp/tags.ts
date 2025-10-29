import { doReqWebappApi } from "@/api/webapp/doReq.ts";
import { RespTags } from "@/api/webapp/typesResp.ts";

export async function getTags(bandId: string): Promise<RespTags | null> {
  const { data, err } = await doReqWebappApi<RespTags>(
    `/api/tags`,
    "GET",
    { bandId },
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}

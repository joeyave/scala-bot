import { Page } from "@/components/Page.tsx";
import { publicUrl } from "@/helpers/publicUrl.ts";
import { Placeholder } from "@telegram-apps/telegram-ui";

import { logger } from "@/helpers/logger.ts";

export function PlaceholderGeneralError({ error }: { error: unknown }) {
  logger.error("Rendering general error page.", { error });

  return (
    <div className="flex h-screen items-center justify-center">
      <Placeholder header="Oops" description="Something went wrong">
        <img
          alt="Telegram sticker"
          src={publicUrl("error.png")}
          style={{ display: "block", width: "144px", height: "144px" }}
        />
      </Placeholder>
    </div>
  );
}

export function PageError({
  error,
  back = false,
}: {
  error: unknown;
  back?: boolean;
}) {
  return (
    <Page back={back}>
      <PlaceholderGeneralError error={error} />
    </Page>
  );
}

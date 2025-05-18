import { Page } from "@/components/Page.tsx";
import { bem } from "@/css/bem.ts";
import { publicUrl } from "@/helpers/publicUrl.ts";
import { Placeholder } from "@telegram-apps/telegram-ui";
import React from "react";

import { logger } from "@/helpers/logger.ts";
import "./style.css";

const [, e] = bem("page-error");

interface SongPageErrorProps {
  error: Error;
}

export const PageError: React.FC<SongPageErrorProps> = ({
  error = new Error("Unknown error"),
}) => {
  logger.warn("Showing error page", {
    error: error.message,
  });

  return (
    <Page back={false}>
      <div className={e("centered-content")}>
        <Placeholder header="Oops" description="Something went wrong">
          <img
            alt="Telegram sticker"
            src={publicUrl("error.png")}
            style={{ display: "block", width: "144px", height: "144px" }}
          />
        </Placeholder>
      </div>
    </Page>
  );
};

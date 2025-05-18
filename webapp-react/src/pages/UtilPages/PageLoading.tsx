import { Page } from "@/components/Page.tsx";
import { bem } from "@/css/bem.ts";
import { Placeholder, Spinner } from "@telegram-apps/telegram-ui";
import React from "react";

import "./style.css";

const [, e] = bem("song-page");

export const PageLoading: React.FC = () => (
  <Page back={false}>
    <div className={e("centered-content")}>
      <Placeholder header="Loading page" description="Wait for a second">
        <Spinner size="l" />
      </Placeholder>
    </div>
  </Page>
);

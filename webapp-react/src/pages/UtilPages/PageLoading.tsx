import { Page } from "@/components/Page.tsx";
import { Placeholder, Spinner } from "@telegram-apps/telegram-ui";
import React from "react";

export const PageLoading: React.FC = () => (
  <Page back={false}>
    <div className="flex h-screen items-center justify-center">
      <Placeholder header="Loading page" description="Wait for a second">
        <Spinner size="l" />
      </Placeholder>
    </div>
  </Page>
);

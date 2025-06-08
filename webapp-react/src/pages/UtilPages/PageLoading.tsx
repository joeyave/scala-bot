import { Page } from "@/components/Page.tsx";
import { Placeholder, Spinner } from "@telegram-apps/telegram-ui";
import React from "react";
import { useTranslation } from "react-i18next";

export const PageLoading: React.FC = () => {
  const { t } = useTranslation();

  return (
    <Page back={false}>
      <div className="flex h-screen items-center justify-center">
        <Placeholder header={t("loading")} description={t("waitMsg")}>
          <Spinner size="l" />
        </Placeholder>
      </div>
    </Page>
  );
};

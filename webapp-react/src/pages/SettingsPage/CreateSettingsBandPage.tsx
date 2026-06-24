import { createSettingsBand } from "@/api/webapp/settings.ts";
import { DriveAccessNotice } from "@/components/DriveAccessNotice.tsx";
import { Page } from "@/components/Page.tsx";
import {
  SettingsBandForm,
  SettingsBandFormState,
} from "@/pages/SettingsPage/SettingsBandForm.tsx";
import { useMutation } from "@tanstack/react-query";
import { hapticFeedback, postEvent } from "@tma.js/sdk-react";
import { FC, ReactNode, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useLocation, useNavigate } from "react-router";

const CreateSettingsBandPage: FC = () => {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const userId = new URLSearchParams(location.search).get("userId");

  useEffect(() => {
    postEvent("web_app_expand");
    postEvent("web_app_setup_swipe_behavior", { allow_vertical_swipe: true });
  }, []);

  const createBandMutation = useMutation({
    mutationFn: (band: SettingsBandFormState) =>
      createSettingsBand({
        name: band.name,
        driveFolderId: band.driveFolder,
        timezone: band.timezone,
      }),
    onSuccess: async () => {
      hapticFeedback.notificationOccurred("success");
      await navigate({ pathname: "/settings", search: location.search });
    },
  });

  if (!userId) {
    return (
      <Page>
        <main className="px-4 pb-6 pt-4">
          <EmptyState
            title={t("settingsErrorNoTgId")}
            description={t("settingsErrorOpenViaBot")}
          />
        </main>
      </Page>
    );
  }

  return (
    <Page>
      <main className="space-y-4 px-4 pb-6 pt-4">
        <SectionBlock title={t("settingsCreateBand")}>
          <div className="rounded-2xl bg-[var(--tg-theme-section-bg-color,#ffffff)] p-4">
            <SettingsBandForm
              mode="create"
              useMainButton={true}
              disabled={createBandMutation.isPending}
              onSubmit={(band: SettingsBandFormState) =>
                createBandMutation.mutate(band)
              }
            />
          </div>
        </SectionBlock>
        <DriveAccessNotice />
      </main>
    </Page>
  );
};

function SectionBlock({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <section className="space-y-3">
      <h1 className="px-1 text-sm font-semibold uppercase text-[var(--tg-theme-link-color,#2481cc)]">
        {title}
      </h1>
      {children}
    </section>
  );
}

function EmptyState({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="rounded-2xl bg-[var(--tg-theme-section-bg-color,#ffffff)] px-4 py-8 text-center">
      <div className="text-base font-semibold text-[var(--tg-theme-text-color,#000000)]">
        {title}
      </div>
      <div className="mt-1 text-sm text-[var(--tg-theme-hint-color,#8e8e93)]">
        {description}
      </div>
    </div>
  );
}

export default CreateSettingsBandPage;

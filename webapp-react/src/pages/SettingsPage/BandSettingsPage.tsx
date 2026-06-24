import {
  getSettingsBandMembers,
  getSettingsMe,
  removeSettingsBandMember,
  updateSettingsBand,
  updateSettingsBandMember,
} from "@/api/webapp/settings.ts";
import { SettingsBand, SettingsMember } from "@/api/webapp/typesResp.ts";
import { ContextMenu } from "@/components/ContextMenu.tsx";
import { DriveAccessNotice } from "@/components/DriveAccessNotice.tsx";
import { Page } from "@/components/Page.tsx";
import {
  SettingsBandForm,
  SettingsBandFormState,
} from "@/pages/SettingsPage/SettingsBandForm.tsx";
import { useMutation, useQuery, useQueryClient, useSuspenseQuery } from "@tanstack/react-query";
import { hapticFeedback, postEvent } from "@tma.js/sdk-react";
import { FC, ReactNode, useEffect, useMemo } from "react";
import { ThreeDots } from "react-bootstrap-icons";
import { useTranslation } from "react-i18next";
import { useParams } from "react-router";

const BandSettingsPage: FC = () => {
  const { t } = useTranslation();
  const { bandId } = useParams<{ bandId: string }>();

  useEffect(() => {
    postEvent("web_app_expand");
    postEvent("web_app_setup_swipe_behavior", { allow_vertical_swipe: true });
  }, []);

  if (!bandId) {
    return (
      <Page back={true}>
        <main className="px-4 pb-6 pt-4">
          <EmptyState
            title={t("settingsBandNotFound")}
            description={t("settingsInvalidBandId")}
          />
        </main>
      </Page>
    );
  }

  return <BandSettingsPageContent bandId={bandId} />;
};

const BandSettingsPageContent: FC<{ bandId: string }> = ({ bandId }) => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const meQuery = useSuspenseQuery({
    queryKey: ["settings", "me"],
    queryFn: async () => {
      const data = await getSettingsMe();
      if (!data) {
        throw new Error("Failed to load settings profile.");
      }
      return data;
    },
  });

  const refreshSettings = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["settings", "me"] }),
      queryClient.invalidateQueries({ queryKey: ["settings", "bands"] }),
    ]);
  };

  const updateBandMutation = useMutation({
    mutationFn: (band: SettingsBandFormState) =>
      updateSettingsBand(bandId, {
        name: band.name,
        driveFolderId: band.driveFolder,
        timezone: band.timezone,
      }),
    onSuccess: async () => {
      hapticFeedback.notificationOccurred("success");
      await refreshSettings();
    },
    onError: (err: any) => {
      hapticFeedback.notificationOccurred("error");
      const errMsg =
        err?.response?.data?.error || err?.message || t("settingsUpdateBandError");
      window.alert(errMsg);
    },
  });

  const band = useMemo(() => {
    return meQuery.data.bands.find((b) => b.id === bandId) ?? null;
  }, [meQuery.data.bands, bandId]);

  if (!band) {
    return (
      <Page back={true}>
        <main className="px-4 pb-6 pt-4">
          <EmptyState
            title={t("settingsBandNotFound")}
            description={t("settingsNotMemberOfGroup")}
          />
        </main>
      </Page>
    );
  }

  if (!band.isAdmin) {
    return (
      <Page back={true}>
        <main className="px-4 pb-6 pt-4">
          <EmptyState
            title={t("settingsAccessRestricted")}
            description={t("settingsAdminsOnly")}
          />
        </main>
      </Page>
    );
  }

  return (
    <Page back={true}>
      <main className="space-y-4 px-4 pb-6 pt-4">
        <SectionBlock title={t("settingsBandHeader", { name: band.name })}>
          <div className="rounded-2xl bg-[var(--tg-theme-section-bg-color,#ffffff)] p-4">
            <SettingsBandForm
              key={band.id}
              mode="edit"
              band={band}
              useMainButton={true}
              disabled={updateBandMutation.isPending}
              onSubmit={(formState) => updateBandMutation.mutate(formState)}
            />
          </div>
        </SectionBlock>
        <DriveAccessNotice />

        <SettingsMembersSection band={band} />
      </main>
    </Page>
  );
};

function SettingsMembersSection({ band }: { band: SettingsBand }) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const membersQuery = useQuery({
    queryKey: ["settings", "members", band.id],
    queryFn: async () => {
      const data = await getSettingsBandMembers(band.id);
      if (!data) {
        throw new Error("Failed to load members.");
      }
      return data;
    },
  });

  const refreshMembers = async () => {
    await Promise.all([
      queryClient.invalidateQueries({
        queryKey: ["settings", "members", band.id],
      }),
      queryClient.invalidateQueries({ queryKey: ["settings", "me"] }),
      queryClient.invalidateQueries({ queryKey: ["settings", "bands"] }),
    ]);
  };

  const roleMutation = useMutation({
    mutationFn: ({ member, isAdmin }: { member: SettingsMember; isAdmin: boolean }) =>
      updateSettingsBandMember(band.id, member.id, { isAdmin }),
    onSuccess: async () => {
      hapticFeedback.selectionChanged();
      await refreshMembers();
    },
    onError: (err: any) => {
      hapticFeedback.notificationOccurred("error");
      const errMsg =
        err?.response?.data?.error || err?.message || t("settingsRoleChangeError");
      if (errMsg === "invalid operation") {
        window.alert(t("settingsNoSelfDemote"));
      } else {
        window.alert(errMsg);
      }
    },
  });

  const removeMutation = useMutation({
    mutationFn: (member: SettingsMember) =>
      removeSettingsBandMember(band.id, member.id),
    onSuccess: async () => {
      hapticFeedback.notificationOccurred("success");
      await refreshMembers();
    },
    onError: (err: any) => {
      hapticFeedback.notificationOccurred("error");
      const errMsg =
        err?.response?.data?.error || err?.message || t("settingsExcludeError");
      window.alert(errMsg);
    },
  });

  if (membersQuery.isLoading) {
    return (
      <SectionBlock title={t("settingsMembers")}>
        <div className="rounded-2xl bg-[var(--tg-theme-section-bg-color,#ffffff)] px-4 py-8 text-center text-sm text-[var(--tg-theme-hint-color,#8e8e93)]">
          ...
        </div>
      </SectionBlock>
    );
  }

  const members = membersQuery.data?.members ?? [];

  return (
    <SectionBlock title={t("settingsMembers")}>
      {members.length > 0 ? (
        <div className="overflow-hidden rounded-2xl bg-[var(--tg-theme-section-bg-color,#ffffff)]">
          {members.map((member, index) => (
            <MemberRow
              key={member.id}
              member={member}
              showDivider={index < members.length - 1}
              roleLabel={member.isAdmin ? t("settingsAdmin") : t("settingsMember")}
              selfLabel={t("settingsIsSelf")}
              menu={
                !member.isSelf ? (
                  <ContextMenu
                    as="button"
                    type="button"
                    className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-[var(--tg-theme-hint-color,#8e8e93)] outline-none active:bg-black/[0.04]"
                    trigger={<ThreeDots size={20} />}
                    items={[
                      {
                        label: member.isAdmin
                          ? t("settingsActionSheetDemoteAdmin")
                          : t("settingsActionSheetMakeAdmin"),
                        onClick: () => {
                          roleMutation.mutate({
                            member,
                            isAdmin: !member.isAdmin,
                          });
                        },
                      },
                      {
                        label: t("settingsActionSheetExclude"),
                        destructive: true,
                        onClick: () => {
                          if (
                            window.confirm(
                              t("settingsExcludeConfirm", {
                                name: member.name || `User ${member.id}`,
                              }),
                            )
                          ) {
                            removeMutation.mutate(member);
                          }
                        },
                      },
                    ]}
                  />
                ) : null
              }
            />
          ))}
        </div>
      ) : (
        <EmptyState
          title={t("settingsNoMembers")}
          description={t("settingsMembersEmpty")}
        />
      )}
    </SectionBlock>
  );
}

function MemberRow({
  member,
  showDivider,
  roleLabel,
  selfLabel,
  menu,
}: {
  member: SettingsMember;
  showDivider: boolean;
  roleLabel: string;
  selfLabel: string;
  menu: ReactNode;
}) {
  return (
    <div
      className={`flex min-h-[64px] items-center gap-3 px-4 py-3 ${
        showDivider ? "border-b border-black/[0.06]" : ""
      }`}
    >
      <div className="min-w-0 flex-1">
        <div className="truncate text-base font-medium text-[var(--tg-theme-text-color,#000000)]">
          {member.name || `User ${member.id}`}
        </div>
        <div className="mt-1 text-sm font-medium text-[var(--tg-theme-hint-color,#8e8e93)]">
          {member.isSelf ? `${roleLabel} (${selfLabel.toLowerCase()})` : roleLabel}
        </div>
      </div>
      {menu}
    </div>
  );
}

function SectionBlock({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <section className="space-y-3">
      <h2 className="px-1 text-sm font-semibold uppercase text-[var(--tg-theme-link-color,#2481cc)]">
        {title}
      </h2>
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

export default BandSettingsPage;

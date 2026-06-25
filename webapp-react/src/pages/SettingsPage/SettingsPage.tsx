import {
  cancelSettingsJoinRequest,
  createSettingsJoinRequest,
  getSettingsBands,
  getSettingsMe,
  leaveSettingsBand,
  setSettingsActiveBand,
} from "@/api/webapp/settings.ts";
import { SettingsBand } from "@/api/webapp/typesResp.ts";
import { ContextMenu, ContextMenuItem } from "@/components/ContextMenu.tsx";
import { Page } from "@/components/Page.tsx";
import { Button, Radio, RadioGroup } from "@headlessui/react";
import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query";
import { hapticFeedback, postEvent } from "@tma.js/sdk-react";
import { FC, ReactNode, useEffect, useMemo, useState } from "react";
import { Clock, Plus, ThreeDots } from "react-bootstrap-icons";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";

const SettingsPage: FC = () => {
  const { t } = useTranslation();
  const userId = useMemo(() => getHashQueryParam("userId"), []);

  useEffect(() => {
    postEvent("web_app_expand");
    postEvent("web_app_setup_swipe_behavior", { allow_vertical_swipe: true });
  }, []);

  if (!userId) {
    return (
      <Page back={false}>
        <main className="px-4 pb-6 pt-4">
          <EmptyState
            title={t("settingsErrorNoTgId")}
            description={t("settingsErrorOpenViaBot")}
          />
        </main>
      </Page>
    );
  }

  return <SettingsPageContent />;
};

const SettingsPageContent: FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [search, setSearch] = useState("");

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

  const bandsQuery = useSuspenseQuery({
    queryKey: ["settings", "bands"],
    queryFn: async () => {
      const data = await getSettingsBands();
      if (!data) {
        throw new Error("Failed to load groups.");
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

  const joinMutation = useMutation({
    mutationFn: createSettingsJoinRequest,
    onSuccess: async () => {
      hapticFeedback.notificationOccurred("success");
      await refreshSettings();
    },
  });

  const cancelJoinMutation = useMutation({
    mutationFn: cancelSettingsJoinRequest,
    onSuccess: async () => {
      hapticFeedback.selectionChanged();
      await refreshSettings();
    },
  });

  const setActiveMutation = useMutation({
    mutationFn: setSettingsActiveBand,
    onSuccess: async () => {
      hapticFeedback.selectionChanged();
      await refreshSettings();
    },
  });

  const leaveMutation = useMutation({
    mutationFn: leaveSettingsBand,
    onSuccess: async () => {
      hapticFeedback.notificationOccurred("success");
      await refreshSettings();
    },
    onError: (err: any) => {
      hapticFeedback.notificationOccurred("error");
      const errMsg =
        err?.response?.data?.error || err?.message || t("settingsLeaveError");
      if (errMsg === "cannot leave the group: you are the last administrator" || errMsg === "invalid operation") {
        window.alert(t("settingsLeaveOnlyAdminError"));
      } else if (errMsg === "cannot leave the only group") {
        window.alert(t("settingsLeaveLastGroupError"));
      } else {
        window.alert(errMsg);
      }
    },
  });

  const myBands = meQuery.data.bands;
  const activeBandId = meQuery.data.user.activeBandId || myBands[0]?.id || "";
  const joiningBandId = joinMutation.isPending ? joinMutation.variables : null;
  const cancelingJoinBandId = cancelJoinMutation.isPending
    ? cancelJoinMutation.variables
    : null;
  const activatingBandId = setActiveMutation.isPending
    ? setActiveMutation.variables
    : null;
  const availableBands = useMemo(
    () => bandsQuery.data.bands.filter((band) => !band.isMember),
    [bandsQuery.data.bands],
  );
  const filteredBands = useMemo(() => {
    const normalizedSearch = search.trim().toLowerCase();
    if (!normalizedSearch) {
      return availableBands;
    }
    return availableBands.filter((band) =>
      band.name.toLowerCase().includes(normalizedSearch),
    );
  }, [availableBands, search]);

  return (
    <Page back={false}>
      <main className="pb-6">
        <UserProfile
          name={meQuery.data.user.name || "Telegram user"}
          userId={meQuery.data.user.id}
          avatarFileId={meQuery.data.user.avatarFileId}
        />

        {myBands.length > 0 ? (
          <SectionBlock title={t("settingsMyBands")}>
            <div className="overflow-hidden rounded-2xl bg-[var(--tg-theme-section-bg-color,#ffffff)]">
              <RadioGroup
                value={activeBandId}
                onChange={(bandId: string) => {
                  if (activatingBandId === null && bandId !== activeBandId) {
                    setActiveMutation.mutate(bandId);
                  }
                }}
                aria-label={t("settingsMyBands")}
              >
                {myBands.map((band, index) => (
                  <BandRadioRow
                    key={band.id}
                    band={band}
                    isBusy={activatingBandId !== null}
                    showDivider={index < myBands.length - 1}
                    roleLabel={band.isAdmin ? t("settingsAdmin") : t("settingsMember")}
                    menu={
                      <BandContextMenu
                        band={band}
                        myBandsCount={myBands.length}
                        onConfigure={() =>
                          navigate({
                            pathname: `/settings/bands/${band.id}`,
                            search: getHashSearch(),
                          })
                        }
                        onLeave={() => leaveMutation.mutate(band.id)}
                      />
                    }
                  />
                ))}
              </RadioGroup>
            </div>
          </SectionBlock>
        ) : null}

        <SectionBlock
          title={myBands.length > 0 ? t("settingsAddBand") : t("settingsRegistration")}
        >
          <div className="space-y-3">
            {availableBands.length > 16 && (
              <input
                value={search}
                placeholder={t("settingsSearchBands")}
                className="h-12 w-full rounded-2xl border border-black/[0.06] bg-[var(--tg-theme-section-bg-color,#ffffff)] px-4 text-base text-[var(--tg-theme-text-color,#000000)] outline-none placeholder:text-[var(--tg-theme-hint-color,#8e8e93)] focus:border-[var(--tg-theme-link-color,#2481cc)]"
                onChange={(event) => setSearch(event.target.value)}
              />
            )}

            <div className="overflow-hidden rounded-2xl bg-[var(--tg-theme-section-bg-color,#ffffff)]">
              {filteredBands.map((band) => (
                <AvailableBandRow
                  key={band.id}
                  band={band}
                  showDivider={true}
                  action={
                    <AvailableBandActionButton
                      band={band}
                      isJoining={joiningBandId === band.id}
                      isCanceling={cancelingJoinBandId === band.id}
                      onJoin={() => joinMutation.mutate(band.id)}
                      onCancelJoin={() => cancelJoinMutation.mutate(band.id)}
                    />
                  }
                />
              ))}
              <Button
                type="button"
                className="flex w-full min-h-[48px] cursor-pointer items-center justify-between px-4 py-1.5 outline-none text-base font-semibold text-[var(--tg-theme-link-color,#2481cc)] hover:bg-black/[0.02] active:bg-black/[0.04] transition-colors"
                onClick={() =>
                  navigate({ pathname: "/settings/create", search: getHashSearch() })
                }
              >
                <span>{t("settingsCreateBand")}</span>
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-[var(--tg-theme-link-color,#2481cc)]">
                  <Plus size={24} />
                </div>
              </Button>
            </div>

            {filteredBands.length === 0 ? (
              <EmptyState
                title={t("settingsBandsNotFound")}
                description={t("settingsChangeSearch")}
              />
            ) : null}
          </div>
        </SectionBlock>
      </main>
    </Page>
  );
};

function UserProfile({ name, userId, avatarFileId }: { name: string; userId: number; avatarFileId?: string }) {
  return (
    <div className="flex items-center gap-3 px-4 py-4">
      <UserAvatar name={name} userId={userId} avatarFileId={avatarFileId} />
      <div className="min-w-0 truncate text-xl font-semibold text-[var(--tg-theme-text-color,#000000)]">
        {name}
      </div>
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
    <section className="px-4 py-3">
      <h2 className="mb-3 px-1 text-sm font-semibold uppercase text-[var(--tg-theme-link-color,#2481cc)]">
        {title}
      </h2>
      {children}
    </section>
  );
}

function BandRadioRow({
  band,
  isBusy,
  showDivider,
  roleLabel,
  menu,
}: {
  band: SettingsBand;
  isBusy: boolean;
  showDivider: boolean;
  roleLabel: string;
  menu: ReactNode;
}) {
  return (
    <Radio
      value={band.id}
      disabled={isBusy}
      className={`flex min-h-[72px] cursor-pointer items-center gap-3 px-4 py-3 outline-none data-[disabled]:cursor-default data-[disabled]:opacity-60 ${
        showDivider ? "border-b border-black/[0.06]" : ""
      }`}
    >
      {({ checked }) => (
        <>
          <span
            className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full border-2 ${
              checked
                ? "border-[var(--tg-theme-link-color,#2481cc)]"
                : "border-black/[0.08]"
            }`}
          >
            {checked ? (
              <span className="h-3 w-3 rounded-full bg-[var(--tg-theme-link-color,#2481cc)]" />
            ) : null}
          </span>
          <span className="min-w-0 flex-1">
            <span className="block truncate text-lg font-medium text-[var(--tg-theme-text-color,#000000)]">
              {band.name}
            </span>
            <span className="mt-1 block text-sm font-medium text-[var(--tg-theme-hint-color,#8e8e93)]">
              {roleLabel}
            </span>
          </span>
          {menu}
        </>
      )}
    </Radio>
  );
}

function BandContextMenu({
  band,
  myBandsCount,
  onConfigure,
  onLeave,
}: {
  band: SettingsBand;
  myBandsCount: number;
  onConfigure: () => void;
  onLeave: () => void;
}) {
  const { t } = useTranslation();
  const items: ContextMenuItem[] = [];

  if (band.isAdmin) {
    items.push({
      label: t("settingsActionSheetConfigure"),
      onClick: onConfigure,
    });
  }

  if (myBandsCount > 1) {
    items.push({
      label: t("settingsActionSheetLeave"),
      destructive: true,
      onClick: () => {
        if (window.confirm(t("settingsLeaveConfirm", { name: band.name }))) {
          onLeave();
        }
      },
    });
  }

  if (items.length === 0) {
    return <div className="h-10 w-10 shrink-0" />;
  }

  return (
    <ContextMenu
      as="button"
      type="button"
      className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-[var(--tg-theme-hint-color,#8e8e93)] outline-none active:bg-black/[0.04]"
      trigger={<ThreeDots size={20} />}
      items={items}
    />
  );
}

function AvailableBandActionButton({
  band,
  isJoining,
  isCanceling,
  onJoin,
  onCancelJoin,
}: {
  band: SettingsBand;
  isJoining: boolean;
  isCanceling: boolean;
  onJoin: () => void;
  onCancelJoin: () => void;
}) {
  const isBusy = isJoining || isCanceling;

  if (band.hasPendingJoinRequest) {
    return (
      <Button
        type="button"
        disabled={isBusy}
        className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-[var(--tg-theme-hint-color,#8e8e93)] outline-none active:bg-black/[0.04] disabled:opacity-50"
        onClick={(e) => {
          e.stopPropagation();
          onCancelJoin();
        }}
      >
        {isBusy ? (
          <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
        ) : (
          <Clock size={16} />
        )}
      </Button>
    );
  }

  return (
    <Button
      type="button"
      disabled={isBusy}
      className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-[var(--tg-theme-link-color,#2481cc)] outline-none active:bg-black/[0.04] disabled:opacity-50"
      onClick={(e) => {
        e.stopPropagation();
        onJoin();
      }}
    >
      {isBusy ? (
        <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
      ) : (
        <Plus size={24} />
      )}
    </Button>
  );
}

function AvailableBandRow({
  band,
  showDivider,
  action,
}: {
  band: SettingsBand;
  showDivider: boolean;
  action: ReactNode;
}) {
  const { t } = useTranslation();
  return (
    <div
      className={`flex min-h-[52px] items-center gap-3 px-4 py-1.5 ${
        showDivider ? "border-b border-black/[0.06]" : ""
      }`}
    >
      <div className="min-w-0 flex-1">
        <div className="truncate text-base font-medium text-[var(--tg-theme-text-color,#000000)]">
          {band.name}
        </div>
        {band.hasPendingJoinRequest ? (
          <div className="mt-0.5 text-xs font-medium text-[var(--tg-theme-hint-color,#8e8e93)]">
            {t("settingsJoinPending")}
          </div>
        ) : null}
      </div>
      {action}
    </div>
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

function getHashQueryParam(name: string): string | null {
  const hash = window.location.hash.slice(1);
  const [, query = ""] = hash.split("?");
  return new URLSearchParams(query).get(name);
}

function getHashSearch(): string {
  const hash = window.location.hash.slice(1);
  const [, query = ""] = hash.split("?");
  return query ? `?${query}` : "";
}

export function UserAvatar({
  name,
  userId,
  size = "large",
  avatarFileId,
}: {
  name: string;
  userId: number;
  size?: "small" | "large";
  avatarFileId?: string;
}) {
  const [imgError, setImgError] = useState(false);

  const palettes = [
    "bg-[#2f80ed] text-white",
    "bg-[#27ae60] text-white",
    "bg-[#eb5757] text-white",
    "bg-[#9b51e0] text-white",
    "bg-[#f2c94c] text-[#1f2933]",
  ];
  const palette = palettes[Math.abs(userId) % palettes.length];
  const sizeClasses =
    size === "small" ? "h-10 w-10 text-sm" : "h-14 w-14 text-lg";

  // If there's a valid avatar hash and no loading error, render the image
  if (avatarFileId && !imgError) {
    return (
      <img
        src={`/api/users/${userId}/avatar?v=${avatarFileId}`}
        alt={name}
        onError={() => setImgError(true)}
        className={`shrink-0 rounded-full object-cover ${sizeClasses}`}
      />
    );
  }

  // Otherwise, fall back to the colored initials placeholder directly
  return (
    <div
      className={`flex shrink-0 items-center justify-center rounded-full font-semibold ${sizeClasses} ${palette}`}
    >
      {getInitials(name)}
    </div>
  );
}

function getInitials(name: string): string {
  const words = name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);
  const initials = words.map((word) => word[0]).join("").toUpperCase();
  return initials || "U";
}

export default SettingsPage;
